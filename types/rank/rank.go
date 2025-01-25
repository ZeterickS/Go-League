package rank

import (
	"fmt"
)

type Rank int

// Division mapping for League of Legends
var divisions = map[int]string{
	0:  "UNRANKED",
	1:  "IRON IV",
	2:  "IRON III",
	3:  "IRON II",
	4:  "IRON I",
	5:  "BRONZE IV",
	6:  "BRONZE III",
	7:  "BRONZE II",
	8:  "BRONZE I",
	9:  "SILVER IV",
	10: "SILVER III",
	11: "SILVER II",
	12: "SILVER I",
	13: "GOLD IV",
	14: "GOLD III",
	15: "GOLD II",
	16: "GOLD I",
	17: "PLATINUM IV",
	18: "PLATINUM III",
	19: "PLATINUM II",
	20: "PLATINUM I",
	21: "EMERALD IV",
	22: "EMERALD III",
	23: "EMERALD II",
	24: "EMERALD I",
	25: "DIAMOND IV",
	26: "DIAMOND III",
	27: "DIAMOND II",
	28: "DIAMOND I",
	29: "MASTER I",
	30: "GRANDMASTER I",
	31: "CHALLENGER I",
}

// String method to format the Rank value
func (r Rank) ToString() string {
	var divisionInt, levelPoints int
	divisionInt = int(r) / 100

	// If the division is higher than 28, we need to adjust the division and level points
	// We get the first two Digits from the Rank value as the division
	if divisionInt > 28 {
		divisionInt = int(r) / 10000
		levelPoints = int(r) % 10000
	}

	division, exists := divisions[divisionInt]
	if !exists {
		division = "Unknown"
	}

	if divisionInt == 0 {
		return "UNRANKED"
	}

	return fmt.Sprintf("%s %02d LP", division, levelPoints)
}

func FromString(rankStr string) Rank {
	var tier, rank string
	var leaguePoints, divisionInt, rankInt int

	_, err := fmt.Sscanf(rankStr, "%s %s %d LP", &tier, &rank, &leaguePoints)
	if err != nil {
		return 0
	}

	fullDivision := fmt.Sprintf("%s %s", tier, rank)
	for key, value := range divisions {
		if value == fullDivision {
			divisionInt = key
			break
		}
	}

	if divisionInt == 0 && fullDivision != divisions[0] {
		return 0
	}

	if divisionInt > 28 {
		rankInt = divisionInt*10000 + leaguePoints
	} else {
		rankInt = divisionInt*100 + leaguePoints
	}

	return Rank(rankInt)
}

func RankDifference(rank1 Rank, rank2 Rank) int {
	rankChange := 0

	rankChange = int(rank1 - rank2)

	if abs(rankChange) > 1000 {
		rankChange = 0
	}

	// This should never be reached
	return rankChange
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
