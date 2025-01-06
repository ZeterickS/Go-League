package common

import (
	"testing"
)

func TestRankToString(t *testing.T) {
	tests := []struct {
		rank     Rank
		expected string
	}{
		{Rank(0), "IRON IV 00 LP"},
		{Rank(1), "IRON IV 01 LP"},
		{Rank(100), "IRON III 00 LP"},
		{Rank(200), "IRON II 00 LP"},
		{Rank(300), "IRON I 00 LP"},
		{Rank(400), "BRONZE IV 00 LP"},
		{Rank(500), "BRONZE III 00 LP"},
		{Rank(600), "BRONZE II 00 LP"},
		{Rank(700), "BRONZE I 00 LP"},
		{Rank(800), "SILVER IV 00 LP"},
		{Rank(900), "SILVER III 00 LP"},
		{Rank(1000), "SILVER II 00 LP"},
		{Rank(1100), "SILVER I 00 LP"},
		{Rank(1200), "GOLD IV 00 LP"},
		{Rank(1300), "GOLD III 00 LP"},
		{Rank(1400), "GOLD II 00 LP"},
		{Rank(1500), "GOLD I 00 LP"},
		{Rank(1600), "PLATINUM IV 00 LP"},
		{Rank(1700), "PLATINUM III 00 LP"},
		{Rank(1800), "PLATINUM II 00 LP"},
		{Rank(1900), "PLATINUM I 00 LP"},
		{Rank(2000), "EMERALD IV 00 LP"},
		{Rank(2100), "EMERALD III 00 LP"},
		{Rank(2200), "EMERALD II 00 LP"},
		{Rank(2300), "EMERALD I 00 LP"},
		{Rank(2400), "DIAMOND IV 00 LP"},
		{Rank(2500), "DIAMOND III 00 LP"},
		{Rank(2600), "DIAMOND II 00 LP"},
		{Rank(2700), "DIAMOND I 00 LP"},
		{Rank(2800), "MASTER I 00 LP"},
		{Rank(2900), "GRANDMASTER I 00 LP"},
		{Rank(3000), "CHALLENGER I 00 LP"},
	}

	for _, test := range tests {
		result := test.rank.ToString()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		rankStr  string
		expected Rank
	}{
		{"IRON IV 00 LP", Rank(0)},
		{"IRON IV 01 LP", Rank(1)},
		{"IRON III 00 LP", Rank(100)},
		{"IRON II 00 LP", Rank(200)},
		{"IRON I 00 LP", Rank(300)},
		{"BRONZE IV 00 LP", Rank(400)},
		{"BRONZE III 00 LP", Rank(500)},
		{"BRONZE II 00 LP", Rank(600)},
		{"BRONZE I 00 LP", Rank(700)},
		{"SILVER IV 00 LP", Rank(800)},
		{"SILVER III 00 LP", Rank(900)},
		{"SILVER II 00 LP", Rank(1000)},
		{"SILVER I 00 LP", Rank(1100)},
		{"GOLD IV 00 LP", Rank(1200)},
		{"GOLD III 00 LP", Rank(1300)},
		{"GOLD II 00 LP", Rank(1400)},
		{"GOLD I 00 LP", Rank(1500)},
		{"PLATINUM IV 00 LP", Rank(1600)},
		{"PLATINUM III 00 LP", Rank(1700)},
		{"PLATINUM II 00 LP", Rank(1800)},
		{"PLATINUM I 00 LP", Rank(1900)},
		{"EMERALD IV 00 LP", Rank(2000)},
		{"EMERALD III 00 LP", Rank(2100)},
		{"EMERALD II 00 LP", Rank(2200)},
		{"EMERALD I 00 LP", Rank(2300)},
		{"DIAMOND IV 00 LP", Rank(2400)},
		{"DIAMOND III 00 LP", Rank(2500)},
		{"DIAMOND II 00 LP", Rank(2600)},
		{"DIAMOND I 00 LP", Rank(2700)},
		{"MASTER I 00 LP", Rank(2800)},
		{"GRANDMASTER I 00 LP", Rank(2900)},
		{"CHALLENGER I 00 LP", Rank(3000)},
	}

	for _, test := range tests {
		result := FromString(test.rankStr)
		if result != test.expected {
			t.Errorf("Expected %d, got %d", test.expected, result)
		}
	}
}
