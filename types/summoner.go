package types

import (
    "time"
    "discord-bot/common"
)

// Summoner represents a user with various attributes
type Summoner struct {
    name      string
    tagLine   string
	accountID string
	id		  string
	puuid	  string
    Rank      common.Rank
    lastRank  common.Rank
    updated   time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields name, tagLine, accountID, ID, puuid, and Rank
func NewSummoner(name string, tagLine string, accountID string, id string, puuid string, rank int) *Summoner {

    summoner := &Summoner{
        name: name,
		tagLine: tagLine,
		accountID: accountID,
		id : id,
		puuid: puuid,
		Rank: common.Rank(rank),
		lastRank: common.Rank(rank),
        updated: time.Now(),
    }

    return summoner
}

// toNameTag returns the name and tagline of the summoner in <name>#<tagline> format
func (s Summoner) ToNameTag() string {
    return s.name + "#" + s.tagLine
}

// UpdateRank updates the summoner's rank and last rank
func (s *Summoner) UpdateRank(newRank int) {
    s.lastRank = s.Rank
    s.Rank = common.Rank(newRank)
    s.updated = time.Now()
}