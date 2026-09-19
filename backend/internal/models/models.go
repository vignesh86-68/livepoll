// Package models holds the domain types shared across the application.
//
// These structs describe what a user, a poll and a vote *are*. They carry bson
// tags because they are persisted to MongoDB, but they deliberately do not
// carry json tags for internal fields: what goes over the wire is decided by
// the response types in the httpapi package, not by the storage layer. Keeping
// those separate means a change to the database schema cannot accidentally
// leak a new field into the public API.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Poll status values.
const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)

// User is an account that can create and manage polls.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"` // always stored lowercased and trimmed
	Name      string             `bson:"name"`
	PassHash  string             `bson:"passHash"` // bcrypt; never leaves the server
	CreatedAt time.Time          `bson:"createdAt"`
}

// Option is one choice within a poll. IDs are short, stable and generated
// server-side, so a client can never invent one or renumber the set.
type Option struct {
	ID   string `bson:"id"`
	Text string `bson:"text"`
}

// Poll is a question with a fixed set of options, reachable by its share Code.
//
// Totals and TotalVotes are the durable tally. Redis holds the live copy that
// serves reads and drives broadcasts, but these fields are what it is rebuilt
// from if Redis is ever emptied, so they must be kept accurate.
type Poll struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Code       string             `bson:"code"`
	Question   string             `bson:"question"`
	Options    []Option           `bson:"options"`
	OwnerID    primitive.ObjectID `bson:"ownerId"`
	OwnerName  string             `bson:"ownerName"` // denormalised so the public page needs one read
	Status     string             `bson:"status"`
	Multi      bool               `bson:"multi"` // may a voter pick more than one option
	ClosesAt   *time.Time         `bson:"closesAt,omitempty"`
	Totals     map[string]int64   `bson:"totals"`
	TotalVotes int64              `bson:"totalVotes"`
	CreatedAt  time.Time          `bson:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt"`
}

// IsOpen reports whether the poll currently accepts votes.
//
// A poll can be shut either explicitly by its owner or implicitly by passing
// ClosesAt. Both are checked here, in one place, so no caller can accidentally
// honour one rule and forget the other.
func (p *Poll) IsOpen(now time.Time) bool {
	if p.Status != StatusOpen {
		return false
	}
	if p.ClosesAt != nil && now.After(*p.ClosesAt) {
		return false
	}
	return true
}

// EffectiveStatus is the status as a client should see it, accounting for a
// deadline that has passed but not yet been written back to the database.
func (p *Poll) EffectiveStatus(now time.Time) string {
	if p.IsOpen(now) {
		return StatusOpen
	}
	return StatusClosed
}

// HasOption reports whether id belongs to this poll. Used to reject votes that
// reference an option from some other poll, or one simply made up by a client.
func (p *Poll) HasOption(id string) bool {
	for _, o := range p.Options {
		if o.ID == id {
			return true
		}
	}
	return false
}

// OptionIDs returns the option identifiers in display order.
func (p *Poll) OptionIDs() []string {
	ids := make([]string, 0, len(p.Options))
	for _, o := range p.Options {
		ids = append(ids, o.ID)
	}
	return ids
}

// Vote records one ballot. The unique index on (pollId, voterKey) is what
// actually enforces one vote per person; the Redis check in front of it is only
// a fast filter.
type Vote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	PollID    primitive.ObjectID `bson:"pollId"`
	OptionIDs []string           `bson:"optionIds"`
	VoterKey  string             `bson:"voterKey"`
	CreatedAt time.Time          `bson:"createdAt"`
}

// Tally is a point-in-time count for a poll. This one *does* carry json tags:
// it is sent to clients over both HTTP and the WebSocket, and both must agree
// on its shape.
//
// Seq is a per-poll sequence number that increases with every change. Clients
// use it to spot a missed message and trigger a resync, which turns a dropped
// broadcast from a silently wrong number into a self-healing refresh.
type Tally struct {
	Counts map[string]int64 `json:"counts"`
	Total  int64            `json:"total"`
	Seq    int64            `json:"seq"`
}

// Event types pushed to connected clients.
const (
	EventSnapshot = "snapshot" // full state, sent immediately on connect
	EventTally    = "tally"    // a vote landed
	EventViewers  = "viewers"  // someone opened or closed the page
	EventClosed   = "closed"   // the poll stopped accepting votes
)

// Event is the envelope for everything sent over the WebSocket. A single
// tagged envelope, rather than several bare message shapes, means the client
// has exactly one thing to parse and can ignore event types it does not know
// about, which keeps old clients working when new ones are added.
type Event struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Seq     int64  `json:"seq,omitempty"`
	Tally   *Tally `json:"tally,omitempty"`
	Viewers int64  `json:"viewers,omitempty"`
	Status  string `json:"status,omitempty"`
	At      int64  `json:"at"` // unix milliseconds
}
