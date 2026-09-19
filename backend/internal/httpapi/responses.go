package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vignesh/livepoll/internal/models"
	"github.com/vignesh/livepoll/internal/validate"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type OptionResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type PollResponse struct {
	Code      string           `json:"code"`
	Question  string           `json:"question"`
	Options   []OptionResponse `json:"options"`
	OwnerID   string           `json:"ownerId"`
	OwnerName string           `json:"ownerName"`
	Status    string           `json:"status"`
	Multi     bool             `json:"multi"`
	ClosesAt  *time.Time       `json:"closesAt,omitempty"`
	CreatedAt time.Time        `json:"createdAt"`
}

type PollDetailResponse struct {
	Poll     PollResponse  `json:"poll"`
	Tally    *models.Tally `json:"tally"`
	HasVoted bool          `json:"hasVoted"`
	Viewers  int64         `json:"viewers"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationResponse struct {
	Error  string                `json:"error"`
	Fields []validate.FieldError `json:"fields"`
}

func respondJSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Error: message})
}

func respondValidation(c *gin.Context, fields []validate.FieldError) {
	c.JSON(http.StatusUnprocessableEntity, ValidationResponse{
		Error:  "validation failed",
		Fields: fields,
	})
}
