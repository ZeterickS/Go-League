package main

import (
    "time"
)

// Summoner represents a user with various attributes
type Summoner struct {
    name      string
    tagLine   string
	accountID string
	id		  string
	puuid	  string
    rank      int
    lastRank  int
    updated   time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields name, tagLine, accountID, ID, puuid, and Rank
func NewSummoner(name string, tagLine string, accountID string, id string, puuid string, rank int) *Summoner {#

    summoner := &Summoner{
        name:    name,
		tagLine: tagLine,
		accountID: accountID,
		id : id,
		puuid: puuid,
		rank:    rank,
		lastRank: rank,
        updated: time.Now(),
    }

    return summoner
}

// GetNameTag returns the name and tagline of the summoner in <name>#<tagline> format
func GetNameTag() string {
	return Name + "#" + TagLine
}

// UpdateRank updates the summoner's rank and last rank
func (s *Summoner) UpdateRank(newRank int) {
    s.LastRank = s.Rank
    s.Rank = newRank
    s.Updated = time.Now()
}