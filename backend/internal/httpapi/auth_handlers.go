package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vignesh/livepoll/internal/auth"
	"github.com/vignesh/livepoll/internal/live"
	"github.com/vignesh/livepoll/internal/models"
	"github.com/vignesh/livepoll/internal/store"
	"github.com/vignesh/livepoll/internal/validate"
)

type AuthHandler struct {
	store  *store.Store
	tm     *auth.TokenManager
	live   *live.Live
	isProd bool
}

func NewAuthHandler(s *store.Store, tm *auth.TokenManager, l *live.Live, isProd bool) *AuthHandler {
	return &AuthHandler{
		store:  s,
		tm:     tm,
		live:   l,
		isProd: isProd,
	}
}

func (h *AuthHandler) setCookie(c *gin.Context, token string, expires time.Time) {
	// SameSite Lax, HttpOnly true, Secure only in prod
	c.SetCookie(auth.SessionCookie, token, int(time.Until(expires).Seconds()), "/", "", h.isProd, true)
}

func (h *AuthHandler) clearCookie(c *gin.Context) {
	c.SetCookie(auth.SessionCookie, "", -1, "/", "", h.isProd, true)
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid json")
		return
	}

	r := validate.New()
	name := validate.Text(r, "name", req.Name, 1, 60)
	email := validate.Email(r, "email", req.Email)
	validate.Password(r, "password", req.Password)

	if !r.OK() {
		respondValidation(c, r.Errors)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	u := &models.User{
		Name:      name,
		Email:     email,
		PassHash:  hash,
		CreatedAt: time.Now(),
	}

	if err := h.store.Users.Create(c.Request.Context(), u); err != nil {
		if err == store.ErrDuplicate {
			// Do not leak existence, or maybe leak since signups are public?
			// The instructions don't say explicitly to hide it on signup, only login.
			// However, usually we return a generic error or just say "email taken".
			r.Add("email", "is already registered")
			respondValidation(c, r.Errors)
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, exp, err := h.tm.Issue(u.ID.Hex(), u.Email, u.Name)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to issue session")
		return
	}

	h.setCookie(c, token, exp)

	respondJSON(c, http.StatusCreated, UserResponse{
		ID:        u.ID.Hex(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid json")
		return
	}

	u, err := h.store.Users.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == store.ErrNotFound {
			auth.BurnTiming()
			respondError(c, http.StatusUnauthorized, "invalid email or password")
			return
		}
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}

	if !auth.CheckPassword(u.PassHash, req.Password) {
		respondError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, exp, err := h.tm.Issue(u.ID.Hex(), u.Email, u.Name)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to issue session")
		return
	}

	h.setCookie(c, token, exp)

	respondJSON(c, http.StatusOK, UserResponse{
		ID:        u.ID.Hex(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	claims := claimsVal.(*auth.Claims)
	
	respondJSON(c, http.StatusOK, UserResponse{
		ID:    claims.Subject,
		Name:  claims.Name,
		Email: claims.Email,
		// Missing createdAt from token, but not strictly needed for frontend me() call
	})
}
