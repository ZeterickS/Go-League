package rank

import (
	"fmt"
	"strconv"
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
		for divisionInt >= 100 {
			divisionInt /= 10
		}
	}

	// Extract all numbers after the first two digits
	// I was not able to make this mathematically correct, so I had to convert it to a string and cut the first two digits
	levelPoints = CutFirstTwoDigits(int(r))

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
	var tier, rank, divisionIntString, leaguePointsString string
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

	// Convert the Integer values to Strings to add them to eachother
	leaguePointsString = fmt.Sprintf("%02d", leaguePoints)
	divisionIntString = fmt.Sprintf("%d", divisionInt)

	rankInt, err = strconv.Atoi(divisionIntString + leaguePointsString)

	if err != nil {
		return 0
	}

	return Rank(rankInt)
}

func RankDifference(rank1 Rank, rank2 Rank) int {
	if rank1/100 <= 28 && rank2/100 <= 28 {
		return int(rank1 - rank2)
	}

	// If the division is higher than 28, we need to adjust the division and level points
	// We get the first two Digits from the Rank value as the division
	var divisionInt1, divisionInt2, levelPoints1, levelPoints2 int
	divisionInt1 = int(rank1) / 100
	divisionInt2 = int(rank2) / 100

	i := 0
	for divisionInt1 >= 100 {
		divisionInt1 /= 10
		i++
	}

	j := 0
	for divisionInt2 >= 100 {
		divisionInt2 /= 10
		j++
	}

	if i == j {
		return int(rank1 - rank2)
	}

	levelPoints1 = CutFirstTwoDigits(int(rank1))
	levelPoints2 = CutFirstTwoDigits(int(rank2))

	if i != j {
		return levelPoints1 - levelPoints2
	}

	// This should never be reached
	return 0
}

func CutFirstTwoDigits(value int) int {
	valueStr := fmt.Sprintf("%d", value)
	if len(valueStr) <= 2 {
		return 0
	}
	cutValue, _ := strconv.Atoi(valueStr[2:])
	return cutValue
}
