package common

import (
	"fmt"
)

type Rank int 

// Division mapping for League of Legends
var divisions = map[int]string{
	0:  "Iron IV",
	1:  "Iron III",
	2:  "Iron II",
	3:  "Iron I",
	4:  "Bronze IV",
	5:  "Bronze III",
	6:  "Bronze II",
	7:  "Bronze I",
	8:  "Silver IV",
	9:  "Silver III",
	10: "Silver II",
	11: "Silver I",
	12: "Gold IV",
	13: "Gold III",
	14: "Gold II",
	15: "Gold I",
	16: "Platinum IV",
	17: "Platinum III",
	18: "Platinum II",
	19: "Platinum I",
	20: "Emerald IV",
	21: "Emerald III",
	22: "Emerald II",
	23: "Emerald I",
	24: "Diamond IV",
	25: "Diamond III",
	26: "Diamond II",
	27: "Diamond I",
	28: "Master I",
	29: "Grandmaster I",
	30: "Challenger I",
}

// String method to format the Rank value
func (r Rank) ToString() string {
	divisionInt := int(r) / 100
	levelPoints := int(r) % 100

	division, exists := divisions[divisionInt]
	if !exists {
		division = "Unknown"
	}

	return fmt.Sprintf("%s %02d LP", division, levelPoints)
}

func FromString(rankStr string) (Rank, error) {
	var division string
	var levelPoints int

	_, err := fmt.Sscanf(rankStr, "%s %d LP", &division, &levelPoints)
	if err != nil {
		return 0, err
	}

	var divisionInt int
	for key, value := range divisions {
		if value == division {
			divisionInt = key
			break
		}
	}

	if divisionInt == 0 && division != divisions[0] {
		return 0, fmt.Errorf("invalid division: %s", division)
	}

	return Rank(divisionInt*100 + levelPoints), nil
}