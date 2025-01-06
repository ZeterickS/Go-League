package apiHelper

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"discord-bot/rank"
)

func setupEnv() {
	// Mock the environment variable
	os.Setenv("ROPT_API_TOKEN", "test-api-key")
}

func TestGetSummonerByTag(t *testing.T) {
	setupEnv()

	// Mock the response for the account API
	accountResponse := `{"puuid": "test-puuid"}`
	accountServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(accountResponse))
	}))
	defer accountServer.Close()

	// Mock the response for the summoner API
	summonerResponse := `{
		"id": "test-id",
		"accountId": "test-account-id",
		"puuid": "test-puuid",
		"name": "Cedric",
		"profileIconId": 1234,
		"revisionDate": 1610000000000,
		"summonerLevel": 30
	}`
	summonerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(summonerResponse))
	}))
	defer summonerServer.Close()

	// Mock the response for the rank API
	rankResponse := `[{
		"queueType": "RANKED_SOLO_5x5",
		"tier": "GOLD",
		"rank": "IV",
		"leaguePoints": 50
	}, {
		"queueType": "RANKED_FLEX_5x5",
		"tier": "SILVER",
		"rank": "II",
		"leaguePoints": 30
	}]`
	rankServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(rankResponse))
	}))
	defer rankServer.Close()

	// Call the function
	summoner, err := GetSummonerByTag("Cedric", "0010")
	if err != nil {
		t.Fatalf("Error getting summoner by tag: %v", err)
	}

	// Check the summoner data
	if summoner.Name != "Cedric" {
		t.Errorf("Expected name 'Cedric', got '%s'", summoner.Name)
	}
	if summoner.TagLine != "0010" {
		t.Errorf("Expected tag line '0010', got '%s'", summoner.TagLine)
	}
	if summoner.PUUID != "test-puuid" {
		t.Errorf("Expected PUUID 'test-puuid', got '%s'", summoner.PUUID)
	}
	expectedSoloRank := rank.FromString("GOLD IV 50 LP")
	if summoner.SoloRank != expectedSoloRank {
		t.Errorf("Expected solo rank '%s', got '%s'", expectedSoloRank.ToString(), summoner.SoloRank.ToString())
	}
	expectedFlexRank := rank.FromString("SILVER II 30 LP")
	if summoner.FlexRank != expectedFlexRank {
		t.Errorf("Expected flex rank '%s', got '%s'", expectedFlexRank.ToString(), summoner.FlexRank.ToString())
	}
}

func TestGetSummonerByPUUID(t *testing.T) {
	setupEnv()

	// Mock the response for the summoner API
	summonerResponse := `{
		"id": "test-id",
		"accountId": "test-account-id",
		"puuid": "test-puuid",
		"name": "Cedric",
		"profileIconId": 1234,
		"revisionDate": 1610000000000,
		"summonerLevel": 30
	}`
	summonerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(summonerResponse))
	}))
	defer summonerServer.Close()

	// Mock the response for the rank API
	rankResponse := `[{
		"queueType": "RANKED_SOLO_5x5",
		"tier": "GOLD",
		"rank": "IV",
		"leaguePoints": 50
	}, {
		"queueType": "RANKED_FLEX_5x5",
		"tier": "SILVER",
		"rank": "II",
		"leaguePoints": 30
	}]`
	rankServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(rankResponse))
	}))
	defer rankServer.Close()

	// Call the function
	summoner, err := GetSummonerByPUUID("test-puuid", "Cedric", "0010")
	if err != nil {
		t.Fatalf("Error getting summoner by PUUID: %v", err)
	}

	// Check the summoner data
	if summoner.Name != "Cedric" {
		t.Errorf("Expected name 'Cedric', got '%s'", summoner.Name)
	}
	if summoner.TagLine != "0010" {
		t.Errorf("Expected tag line '0010', got '%s'", summoner.TagLine)
	}
	if summoner.PUUID != "test-puuid" {
		t.Errorf("Expected PUUID 'test-puuid', got '%s'", summoner.PUUID)
	}
	expectedSoloRank := rank.FromString("GOLD IV 50 LP")
	if summoner.SoloRank != expectedSoloRank {
		t.Errorf("Expected solo rank '%s', got '%s'", expectedSoloRank.ToString(), summoner.SoloRank.ToString())
	}
	expectedFlexRank := rank.FromString("SILVER II 30 LP")
	if summoner.FlexRank != expectedFlexRank {
		t.Errorf("Expected flex rank '%s', got '%s'", expectedFlexRank.ToString(), summoner.FlexRank.ToString())
	}
}
