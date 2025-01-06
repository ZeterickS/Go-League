package types

import (
	"discord-bot/common"
	"time"
)

// Summoner represents a user with various attributes
type Summoner struct {
	Name         string
	TagLine      string
	AccountID    string
	ID           string
	PUUID        string
	SoloRank     common.Rank
	LastSoloRank common.Rank
	FlexRank     common.Rank
	LastFlexRank common.Rank
	Updated      time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields name, tagLine, accountID, ID, puuid, and Rank
func NewSummoner(name string, tagLine string, accountID string, id string, puuid string, soloRank common.Rank, lastSoloRank common.Rank, flexRank common.Rank, lastFlexRank common.Rank, updated time.Time) *Summoner {
	summoner := &Summoner{
		Name:         name,
		TagLine:      tagLine,
		AccountID:    accountID,
		ID:           id,
		PUUID:        puuid,
		SoloRank:         soloRank,
		LastSoloRank:     lastSoloRank,
		FlexRank:     flexRank,
		LastFlexRank: lastFlexRank,
		Updated:      updated,
	}

	return summoner
}

// GetNameTag returns the name and tagline of the summoner in <name>#<tagline> format
func (s *Summoner) GetNameTag() string {
	return s.Name + "#" + s.TagLine
}

// UpdateRank updates the summoner's rank and last rank
func (s *Summoner) UpdateSoloRank(newSoloRank common.Rank) {
	s.LastSoloRank = s.SoloRank
	s.SoloRank = newSoloRank
	s.Updated = time.Now()
}

func (s *Summoner) UpdateFlexRank(newFlexRank common.Rank) {
	s.LastFlexRank = s.FlexRank
	s.FlexRank = newFlexRank
	s.Updated = time.Now()
}
