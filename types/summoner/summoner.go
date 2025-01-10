package summoner

import (
	"discord-bot/types/rank"
	"time"
)

// Summoner represents a user with various attributes
type Summoner struct {
	Name          string
	TagLine       string
	AccountID     string
	ID            string
	PUUID         string
	ProfileIconID int
	SoloRank      rank.Rank
	FlexRank      rank.Rank
	Updated       time.Time
}

// NewSummoner creates a new Summoner instance with mandatory fields name, tagLine, accountID, ID, puuid, and Rank
func NewSummoner(name string, tagLine string, accountID string, id string, puuid string, profileIconId int, soloRank rank.Rank, flexRank rank.Rank, updated time.Time) *Summoner {
	summoner := &Summoner{
		Name:          name,
		TagLine:       tagLine,
		AccountID:     accountID,
		ID:            id,
		PUUID:         puuid,
		ProfileIconID: profileIconId,
		SoloRank:      soloRank,
		FlexRank:      flexRank,
		Updated:       updated,
	}

	return summoner
}

// GetNameTag returns the name and tagline of the summoner in <name>#<tagline> format
func (s *Summoner) GetNameTag() string {
	return s.Name + "#" + s.TagLine
}

// UpdateRank updates the summoner's rank and last rank
func (s *Summoner) UpdateSoloRank(newSoloRank rank.Rank) {
	s.SoloRank = newSoloRank
	s.Updated = time.Now()
}

func (s *Summoner) UpdateFlexRank(newFlexRank rank.Rank) {
	s.FlexRank = newFlexRank
	s.Updated = time.Now()
}
