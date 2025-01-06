package types

import (
	"time"
)

// Summoner represents a user with various attributes
type Summoner struct {
	Name      string
	TagLine   string
	AccountID string
	ID        string
	PUUID     string
	Rank      string
	LastRank  string
	Updated   time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields name, tagLine, accountID, ID, puuid, and Rank
func NewSummoner(name string, tagLine string, accountID string, id string, puuid string, rank string, lastRank string, updated time.Time) *Summoner {
	summoner := &Summoner{
		Name:      name,
		TagLine:   tagLine,
		AccountID: accountID,
		ID:        id,
		PUUID:     puuid,
		Rank:      rank,
		LastRank:  lastRank,
		Updated:   updated,
	}

	return summoner
}

// GetNameTag returns the name and tagline of the summoner in <name>#<tagline> format
func (s *Summoner) GetNameTag() string {
	return s.Name + "#" + s.TagLine
}

// UpdateRank updates the summoner's rank and last rank
func (s *Summoner) UpdateRank(newRank string) {
	s.LastRank = s.Rank
	s.Rank = newRank
	s.Updated = time.Now()
}
