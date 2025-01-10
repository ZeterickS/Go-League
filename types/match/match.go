package match

import (
	"discord-bot/types/rank"
	"discord-bot/types/summoner"
)

type Participant struct {
	Summoner   summoner.Summoner
	Perks      Perks
	ChampionID int
}

type Team struct {
	TeamID       int
	Participants []Participant
}

type OngoingMatch struct {
	GameID int64
	Teams  [2]Team
}

type Perks struct {
	PerkIDs      []int
	PerkStyle    int
	PerkSubStyle int
}

// AverageRank calculates the average rank of the team
func (t *Team) AverageRank() rank.Rank {
	if len(t.Participants) == 0 {
		return 0
	}

	var totalRank int
	var count int
	for _, participant := range t.Participants {
		if participant.Summoner.SoloRank != 0 {
			totalRank += int(participant.Summoner.SoloRank)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return rank.Rank(totalRank / count)
}
