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
	Rank         common.Rank
	LastRank     common.Rank
	RankFlex     common.Rank
	LastFlexRank common.Rank
	Updated      time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields name, tagLine, accountID, ID, puuid, and Rank
func NewSummoner(name string, tagLine string, accountID string, id string, puuid string, rank common.Rank, lastRank common.Rank, rankFlex common.Rank, lastFlexRank common.Rank, updated time.Time) *Summoner {
	summoner := &Summoner{
		Name:         name,
		TagLine:      tagLine,
		AccountID:    accountID,
		ID:           id,
		PUUID:        puuid,
		Rank:         rank,
		LastRank:     lastRank,
		RankFlex:     rankFlex,
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
func (s *Summoner) UpdateRank(newRank common.Rank) {
	s.LastRank = s.Rank
	s.Rank = newRank
	s.Updated = time.Now()
}
