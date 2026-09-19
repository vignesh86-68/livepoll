package models_test

import (
	"testing"
	"time"

	"github.com/vignesh/livepoll/internal/models"
)

func TestPollStatusAndOptions(t *testing.T) {
	now := time.Now()
	past := now.Add(-10 * time.Minute)
	future := now.Add(10 * time.Minute)

	p := &models.Poll{
		Status: models.StatusOpen,
		Options: []models.Option{
			{ID: "opt1", Text: "Go"},
			{ID: "opt2", Text: "TypeScript"},
		},
	}

	// Case 1: Open with no deadline
	if !p.IsOpen(now) {
		t.Fatal("poll should be open when status is open and closesAt is nil")
	}
	if p.EffectiveStatus(now) != models.StatusOpen {
		t.Fatalf("expected effective status open, got %s", p.EffectiveStatus(now))
	}

	// Case 2: Open with future deadline
	p.ClosesAt = &future
	if !p.IsOpen(now) {
		t.Fatal("poll should be open when closesAt is in future")
	}

	// Case 3: Open status but past deadline
	p.ClosesAt = &past
	if p.IsOpen(now) {
		t.Fatal("poll should not be open when closesAt is in past")
	}
	if p.EffectiveStatus(now) != models.StatusClosed {
		t.Fatalf("expected effective status closed for expired poll, got %s", p.EffectiveStatus(now))
	}

	// Case 4: Explicitly closed status
	p.Status = models.StatusClosed
	p.ClosesAt = &future
	if p.IsOpen(now) {
		t.Fatal("poll should not be open when status is closed")
	}

	// Option checks
	if !p.HasOption("opt1") {
		t.Fatal("expected poll to have opt1")
	}
	if !p.HasOption("opt2") {
		t.Fatal("expected poll to have opt2")
	}
	if p.HasOption("opt3") {
		t.Fatal("poll should not have opt3")
	}

	ids := p.OptionIDs()
	if len(ids) != 2 || ids[0] != "opt1" || ids[1] != "opt2" {
		t.Fatalf("unexpected option IDs: %v", ids)
	}
}
