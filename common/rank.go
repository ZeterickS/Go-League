package common

import (
	"fmt"
)

type Rank int 

// Division mapping for League of Legends
var divisions = map[int]string{
	0:  "IRON IV",
	1:  "IRON III",
	2:  "IRON II",
	3:  "IRON I",
	4:  "BRONZE IV",
	5:  "BRONZE III",
	6:  "BRONZE II",
	7:  "BRONZE I",
	8:  "SILVER IV",
	9:  "SILVER III",
	10: "SILVER II",
	11: "SILVER I",
	12: "GOLD IV",
	13: "GOLD III",
	14: "GOLD II",
	15: "GOLD I",
	16: "PLATINUM IV",
	17: "PLATINUM III",
	18: "PLATINUM II",
	19: "PLATINUM I",
	20: "EMERALD IV",
	21: "EMERALD III",
	22: "EMERALD II",
	23: "EMERALD I",
	24: "DIAMOND IV",
	25: "DIAMOND III",
	26: "DIAMOND II",
	27: "DIAMOND I",
	28: "MASTER I",
	29: "GRANDMASTER I",
	30: "CHALLENGER I",
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

func FromString(rankStr string) (Rank) {
    var tier, rank string
    var leaguePoints int

    _, err := fmt.Sscanf(rankStr, "%s %s %d LP", &tier, &rank, &leaguePoints)
    if err != nil {
        return 0
    }

    fullDivision := fmt.Sprintf("%s %s", tier, rank)
    var divisionInt int
    for key, value := range divisions {
        if value == fullDivision {
            divisionInt = key
            break
        }
    }

    if divisionInt == 0 && fullDivision != divisions[0] {
        return 0
    }

    return Rank(divisionInt*100 + leaguePoints)
}