package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vignesh/livepoll/internal/auth"
	"github.com/vignesh/livepoll/internal/idgen"
	"github.com/vignesh/livepoll/internal/live"
	"github.com/vignesh/livepoll/internal/models"
	"github.com/vignesh/livepoll/internal/store"
	"github.com/vignesh/livepoll/internal/validate"
)

type PollHandler struct {
	store     *store.Store
	live      *live.Live
	isProd    bool
	voterSalt string
}

func NewPollHandler(s *store.Store, l *live.Live, isProd bool, voterSalt string) *PollHandler {
	return &PollHandler{
		store:     s,
		live:      l,
		isProd:    isProd,
		voterSalt: voterSalt,
	}
}

func (h *PollHandler) Create(c *gin.Context) {
	var req struct {
		Question string       `json:"question"`
		Options  []string     `json:"options"`
		Multi    bool         `json:"multi"`
		ClosesAt *time.Time   `json:"closesAt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid json")
		return
	}

	r := validate.New()
	q := validate.Text(r, "question", req.Question, 3, 200)

	r.Check(len(req.Options) >= 2 && len(req.Options) <= 10, "options", "must have between 2 and 10 options")

	var options []models.Option
	for i, opt := range req.Options {
		cleanOpt := validate.Text(r, "options", opt, 1, 100)
		id, _ := idgen.Code(4) // Short ID for option
		options = append(options, models.Option{
			ID:   id,
			Text: cleanOpt,
		})
		// If an option failed length checks, validate will record it, but we can just use the short loop here since validate.Text accumulates errors.
		_ = i
	}

	if req.ClosesAt != nil {
		r.Check(req.ClosesAt.After(time.Now()), "closesAt", "must be in the future")
	}

	if !r.OK() {
		respondValidation(c, r.Errors)
		return
	}

	userID, _ := getUserID(c)
	claims := c.MustGet("claims").(*auth.Claims)

	code, err := idgen.Code(8)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to generate code")
		return
	}

	now := time.Now()
	p := &models.Poll{
		Code:       code,
		Question:   q,
		Options:    options,
		OwnerID:    userID,
		OwnerName:  claims.Name,
		Status:     models.StatusOpen,
		Multi:      req.Multi,
		ClosesAt:   req.ClosesAt,
		Totals:     make(map[string]int64),
		TotalVotes: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.store.Polls.Create(c.Request.Context(), p); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create poll")
		return
	}

	respondJSON(c, http.StatusCreated, mapPollToResponse(p))
}

func (h *PollHandler) ListMine(c *gin.Context) {
	userID, _ := getUserID(c)
	polls, err := h.store.Polls.ListByOwner(c.Request.Context(), userID, 50)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list polls")
		return
	}

	res := make([]PollResponse, 0, len(polls))
	for _, p := range polls {
		res = append(res, mapPollToResponse(p))
	}
	respondJSON(c, http.StatusOK, res)
}

func (h *PollHandler) GetPoll(c *gin.Context) {
	code := c.Param("code")
	p, err := h.store.Polls.FindByCode(c.Request.Context(), code)
	if err != nil {
		if err == store.ErrNotFound {
			respondError(c, http.StatusNotFound, "poll not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}

	tally, err := h.live.GetTally(c.Request.Context(), p.ID.Hex())
	if err != nil {
		// Fallback to rebuilding from mongo
		counts, total, err := h.store.Votes.CountByPoll(c.Request.Context(), p.ID)
		if err == nil {
			tally = &models.Tally{Counts: counts, Total: total, Seq: 0}
			_ = h.live.SeedTally(c.Request.Context(), p.ID.Hex(), counts, total)
		} else {
			tally = &models.Tally{Counts: p.Totals, Total: p.TotalVotes, Seq: 0}
		}
	}

	viewers, _ := h.live.GetViewers(c.Request.Context(), p.ID.Hex())

	hasVoted := false
	if uid, ok := getUserID(c); ok {
		hasVoted, _ = h.store.Votes.Exists(c.Request.Context(), p.ID, "u:"+uid.Hex())
	} else {
		// Anonymous check
		voterKey := "a:" + h.hashVoter(c.GetString("realIP"), c.Request.UserAgent(), p.ID.Hex())
		hasVoted, _ = h.store.Votes.Exists(c.Request.Context(), p.ID, voterKey)
	}

	respondJSON(c, http.StatusOK, PollDetailResponse{
		Poll:     mapPollToResponse(p),
		Tally:    tally,
		HasVoted: hasVoted,
		Viewers:  viewers,
	})
}

func (h *PollHandler) GetResults(c *gin.Context) {
	code := c.Param("code")
	p, err := h.store.Polls.FindByCode(c.Request.Context(), code)
	if err != nil {
		respondError(c, http.StatusNotFound, "poll not found")
		return
	}

	tally, err := h.live.GetTally(c.Request.Context(), p.ID.Hex())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to get tally")
		return
	}

	respondJSON(c, http.StatusOK, tally)
}

func (h *PollHandler) Vote(c *gin.Context) {
	code := c.Param("code")
	var req struct {
		OptionIDs []string `json:"optionIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid json")
		return
	}

	p, err := h.store.Polls.FindByCode(c.Request.Context(), code)
	if err != nil {
		respondError(c, http.StatusNotFound, "poll not found")
		return
	}

	if !p.IsOpen(time.Now()) {
		respondError(c, http.StatusForbidden, "poll is closed")
		return
	}

	r := validate.New()
	r.Check(len(req.OptionIDs) > 0, "optionIds", "must select at least one option")
	if !p.Multi {
		r.Check(len(req.OptionIDs) == 1, "optionIds", "cannot select multiple options")
	}

	// Check deduplication of optionIds and ownership
	seen := make(map[string]bool)
	for _, opt := range req.OptionIDs {
		if !p.HasOption(opt) {
			r.Add("optionIds", "invalid option")
		}
		if seen[opt] {
			r.Add("optionIds", "duplicate option")
		}
		seen[opt] = true
	}

	if !r.OK() {
		respondValidation(c, r.Errors)
		return
	}

	var voterKey string
	if uid, ok := getUserID(c); ok {
		voterKey = "u:" + uid.Hex()
	} else {
		voterKey = "a:" + h.hashVoter(c.GetString("realIP"), c.Request.UserAgent(), p.ID.Hex())
	}

	// Fast path dedup
	isNew, err := h.live.CheckVote(c.Request.Context(), p.ID.Hex(), voterKey)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "redis error")
		return
	}
	if !isNew {
		respondError(c, http.StatusConflict, "already voted")
		return
	}

	v := &models.Vote{
		PollID:    p.ID,
		OptionIDs: req.OptionIDs,
		VoterKey:  voterKey,
		CreatedAt: time.Now(),
	}

	// Authoritative insert
	if err := h.store.Votes.Create(c.Request.Context(), v); err != nil {
		h.live.ClearVote(context.Background(), p.ID.Hex(), voterKey) // Compensating action
		if err == store.ErrDuplicate {
			respondError(c, http.StatusConflict, "already voted")
			return
		}
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}

	// Durable stats update
	_ = h.store.Polls.IncrementTotals(c.Request.Context(), p.ID, req.OptionIDs)

	// Live tally publish
	tally, err := h.live.IncrementAndPublish(c.Request.Context(), p.ID.Hex(), p.Code, req.OptionIDs)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to publish vote")
		return
	}

	respondJSON(c, http.StatusOK, tally)
}

func (h *PollHandler) Close(c *gin.Context) {
	code := c.Param("code")
	userID, _ := getUserID(c)

	p, err := h.store.Polls.FindByCode(c.Request.Context(), code)
	if err != nil {
		respondError(c, http.StatusNotFound, "poll not found")
		return
	}

	if p.OwnerID != userID {
		respondError(c, http.StatusForbidden, "not owner")
		return
	}

	err = h.store.Polls.SetStatus(c.Request.Context(), p.ID, userID, models.StatusClosed)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to close poll")
		return
	}

	h.live.PublishClose(c.Request.Context(), p.ID.Hex(), p.Code)
	
	p.Status = models.StatusClosed
	respondJSON(c, http.StatusOK, mapPollToResponse(p))
}

func (h *PollHandler) Delete(c *gin.Context) {
	code := c.Param("code")
	userID, _ := getUserID(c)

	p, err := h.store.Polls.FindByCode(c.Request.Context(), code)
	if err != nil {
		c.Status(http.StatusNoContent)
		return
	}

	err = h.store.Polls.Delete(c.Request.Context(), p.ID, userID)
	if err != nil {
		if err == store.ErrNotFound {
			respondError(c, http.StatusForbidden, "not owner")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to delete poll")
		return
	}

	h.live.InvalidatePoll(c.Request.Context(), p.Code)
	c.Status(http.StatusNoContent)
}

func (h *PollHandler) hashVoter(ip, ua, pollID string) string {
	sum := sha256.Sum256([]byte(ip + "|" + ua + "|" + pollID + "|" + h.voterSalt))
	return hex.EncodeToString(sum[:])
}

func mapPollToResponse(p *models.Poll) PollResponse {
	opts := make([]OptionResponse, len(p.Options))
	for i, o := range p.Options {
		opts[i] = OptionResponse{ID: o.ID, Text: o.Text}
	}
	return PollResponse{
		Code:      p.Code,
		Question:  p.Question,
		Options:   opts,
		OwnerID:   p.OwnerID.Hex(),
		OwnerName: p.OwnerName,
		Status:    p.EffectiveStatus(time.Now()),
		Multi:     p.Multi,
		ClosesAt:  p.ClosesAt,
		CreatedAt: p.CreatedAt,
	}
}
