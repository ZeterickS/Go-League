package main

import (
    "time"
)

// Summoner represents a user with various attributes
type Summoner struct {
    Name      string
    Rank      int
    LastRank  int
    Profile   string
    TagLine   string
    Updated   time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields and optional fields
func NewSummoner(name string, rank int, optionalFields ...func(*Summoner)) *Summoner {
    summoner := &Summoner{
        Name:    name,
        Rank:    rank,
        Updated: time.Now(),
    }

    // Apply optional fields
    for _, opt := range optionalFields {
        opt(summoner)
    }

    return summoner
}

// WithLastRank sets the LastRank field
func WithLastRank(lastRank int) func(*Summoner) {
    return func(s *Summoner) {
        s.LastRank = lastRank
    }
}

// WithProfile sets the Profile field
func WithProfile(profile string) func(*Summoner) {
    return func(s *Summoner) {
        s.Profile = profile
    }
}

// WithTagLine sets the TagLine field
func WithTagLine(tagLine string) func(*Summoner) {
    return func(s *Summoner) {
        s.TagLine = tagLine
    }
}

// UpdateRank updates the summoner's rank and last rank
func (s *Summoner) UpdateRank(newRank int) {
    s.LastRank = s.Rank
    s.Rank = newRank
    s.Updated = time.Now()
}