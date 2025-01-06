package databaseHelper

import (
	"os"
	"testing"
	"time"

	"discord-bot/rank"
	"discord-bot/types"
)

const testFilename = "test_summoners.json"

func setup() {
	// Remove the test file if it exists
	os.Remove(testFilename)
}

func teardown() {
	// Remove the test file after tests
	os.Remove(testFilename)
}

func TestSaveAndLoadSummoners(t *testing.T) {
	setup()
	defer teardown()

	summoners := make(map[string]*types.Summoner)
	summoner := types.NewSummoner("Cedric", "0010", "accountID", "id", "puuid", rank.FromString("GOLD IV 50 LP"), rank.FromString("SILVER I 0 LP"), time.Now())
	summoners[summoner.Name] = summoner

	// Save the summoners to a file
	err := SaveSummonersToFile(summoners)
	if err != nil {
		t.Fatalf("Error saving summoners: %v", err)
	}

	// Load the summoners from the file
	loadedSummoners, err := LoadSummonersFromFile()
	if err != nil {
		t.Fatalf("Error loading summoners: %v", err)
	}

	// Check if the loaded summoner matches the original summoner
	loadedSummoner, exists := loadedSummoners[summoner.Name]
	if !exists {
		t.Fatalf("Summoner %s not found in loaded summoners", summoner.Name)
	}

	if loadedSummoner.Name != summoner.Name || loadedSummoner.TagLine != summoner.TagLine || loadedSummoner.AccountID != summoner.AccountID || loadedSummoner.ID != summoner.ID || loadedSummoner.PUUID != summoner.PUUID || loadedSummoner.Rank.ToString() != summoner.Rank.ToString() || loadedSummoner.LastRank.ToString() != summoner.LastRank.ToString() {
		t.Fatalf("Loaded summoner does not match original summoner")
	}
}
