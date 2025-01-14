package apiHelper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"discord-bot/types/match"
	"discord-bot/types/rank"
	"discord-bot/types/summoner"

	"github.com/joho/godotenv"
)

const (
	riotAccountBaseURL   = "https://europe.api.riotgames.com/riot/account/v1/accounts"
	riotSummonerBaseURL  = "https://euw1.api.riotgames.com/lol/summoner/v4/summoners"
	riotLeagueBaseURL    = "https://euw1.api.riotgames.com/lol/league/v4/entries"
	riotMatchBaseURL     = "https://euw1.api.riotgames.com/lol/match/v5/matches"
	riotSpectatorBaseURL = "https://euw1.api.riotgames.com/lol/spectator/v5/active-games"
)

// LoadEnv loads environment variables from a .env file if they are not already set
func LoadEnv() error {
	if os.Getenv("ROPT_API_TOKEN") != "" {
		return nil
	}
	return godotenv.Load()
}

// waitForRateLimiters waits until both rate limiters allow a request
func waitForRateLimiters() {
	for !rateLimiterPerSecond.Allow() || !rateLimiterPer2Minutes.Allow() {
		time.Sleep(time.Second)
	}
}

// GetSummonerByTag fetches summoner data by tag from the League of Legends API
func GetSummonerByTag(name, tagLine string) (*summoner.Summoner, error) {
	waitForRateLimiters()

	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("%s/by-riot-id/%s/%s?api_key=%s", riotAccountBaseURL, name, tagLine, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accountData struct {
		PUUID   string `json:"puuid"`
		Name    string `json:"gameName"`
		TagLine string `json:"tagLine"`
	}

	err = json.Unmarshal(body, &accountData)
	if err != nil {
		return nil, err
	}

	return GetSummonerByPUUID(accountData.PUUID)
}

// GetSummonerByPUUID fetches summoner data by PUUID from the League of Legends API
func GetSummonerByPUUID(puuid string) (*summoner.Summoner, error) {
	waitForRateLimiters()

	err := LoadEnv()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("API token not found in environment variables")
	}

	name, tagLine, err := GetNameTagByPUUID(puuid)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/by-puuid/%s?api_key=%s", riotSummonerBaseURL, puuid, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch summoner data: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var summonerData struct {
		ID            string `json:"id"`
		AccountID     string `json:"accountId"`
		PUUID         string `json:"puuid"`
		ProfileIconID int    `json:"profileIconId"`
		RevisionDate  int64  `json:"revisionDate"`
		SummonerLevel int    `json:"summonerLevel"`
	}

	err = json.Unmarshal(body, &summonerData)
	if err != nil {
		return nil, err
	}

	solorank, rankFlex, err := GetSummonerRank(summonerData.ID)

	summoner := summoner.NewSummoner(
		name,
		tagLine, // TagLine
		summonerData.AccountID,
		summonerData.ID,
		puuid,
		summonerData.ProfileIconID,
		solorank,
		rankFlex,   // FlexRank
		time.Now(), // Updated
	)

	if err != nil {
		return summoner, err
	}

	return summoner, nil
}

// GetSummonerRank fetches the rank and division of a summoner by their ID from the League of Legends API
func GetSummonerRank(summonerID string) (rank.Rank, rank.Rank, error) {
	waitForRateLimiters()

	err := LoadEnv()
	if err != nil {
		return 0, 0, fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("%s/by-summoner/%s?api_key=%s", riotLeagueBaseURL, summonerID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("failed to fetch summoner rank: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var rankData []struct {
		QueueType    string `json:"queueType"`
		Tier         string `json:"tier"`
		Rank         string `json:"rank"`
		LeaguePoints int    `json:"leaguePoints"`
	}

	err = json.Unmarshal(body, &rankData)
	if err != nil {
		return 0, 0, err
	}

	if len(rankData) == 0 {
		return 0, 0, nil
	}

	var soloRank, flexRank rank.Rank = 0, 0

	for _, entry := range rankData {
		rankStr := fmt.Sprintf("%s %s %d LP", entry.Tier, entry.Rank, entry.LeaguePoints)
		if entry.QueueType == "RANKED_SOLO_5x5" {
			soloRank = rank.FromString(rankStr)
		} else if entry.QueueType == "RANKED_FLEX_SR" {
			flexRank = rank.FromString(rankStr)
		}
	}

	return soloRank, flexRank, nil
}

// GetLastRankedMatch fetches the last ranked match of a summoner by their PUUID from the League of Legends API
func GetLastRankedMatch(puuid string) (string, error) {
	waitForRateLimiters()

	err := LoadEnv()
	if err != nil {
		return "", fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return "", fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("%s/by-puuid/%s/ids?type=ranked&start=0&count=1&api_key=%s", riotMatchBaseURL, puuid, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch last ranked match: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var matchIDs []string
	err = json.Unmarshal(body, &matchIDs)
	if err != nil {
		return "", err
	}

	if len(matchIDs) == 0 {
		return "", fmt.Errorf("no ranked matches found for summoner")
	}

	return matchIDs[0], nil
}

// GetOngoingMatchByPUUID checks if there is an ongoing match for the given summoner's PUUID
func GetOngoingMatchByPUUID(puuid, apiKey string) (*match.Match, error) {
	waitForRateLimiters()

	url := fmt.Sprintf("%s/by-summoner/%s?api_key=%s", riotSpectatorBaseURL, puuid, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Riot Games API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No ongoing match found
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %v", resp.Status)
	}

	var apiResponse struct {
		GameID       int64 `json:"gameId"`
		QueueID      int   `json:"gameQueueConfigId"`
		Participants []struct {
			PUUID      string      `json:"puuid"`
			TeamID     int         `json:"teamId"`
			ChampionID int         `json:"championId"`
			Perks      match.Perks `json:"perks"`
			SummonerID string      `json:"summonerId"`
			Spell1ID   int         `json:"spell1Id"`
			Spell2ID   int         `json:"spell2Id"`
		} `json:"participants"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Println("Response Body:", string(body))

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	gameType := "Flex"
	if apiResponse.QueueID == 440 {
		gameType = "Flex"
	} else if apiResponse.QueueID == 420 {
		gameType = "Solo/Duo"
	} else {
		gameType = "UNRANKED"
	}

	ongoingMatch := &match.Match{
		GameID:   apiResponse.GameID,
		Teams:    [2]match.Team{{TeamID: 100}, {TeamID: 200}},
		GameType: gameType,
	}

	// Populate teams and participants
	for _, participant := range apiResponse.Participants {
		teamIndex := 0
		if participant.TeamID == 200 {
			teamIndex = 1
		}

		summoner, err := GetSummonerByPUUID(participant.PUUID)

		fmt.Printf("Summoner: %+v\n", summoner)
		if err != nil {
			return nil, fmt.Errorf("failed to get summoner by PUUID: %v", err)
		}

		ongoingMatch.Teams[teamIndex].Participants = append(ongoingMatch.Teams[teamIndex].Participants, match.Participant{
			Summoner:   *summoner,
			Perks:      participant.Perks,
			ChampionID: participant.ChampionID,
			Items:      match.Items{},
			Spells: match.Spells{
				SpellIDs: []int{participant.Spell1ID, participant.Spell2ID},
			},
		})
	}

	fmt.Printf("Ongoing Match: %+v\n", ongoingMatch)

	return ongoingMatch, nil
}

// GetNameTagByPUUID fetches the name and tag from a PUUID and returns them as two separate strings
func GetNameTagByPUUID(puuid string) (string, string, error) {
	waitForRateLimiters()

	err := LoadEnv()
	if err != nil {
		return "", "", fmt.Errorf("error loading .env file")
	}

	apiKey := os.Getenv("ROPT_API_TOKEN")
	if apiKey == "" {
		return "", "", fmt.Errorf("API token not found in environment variables")
	}

	url := fmt.Sprintf("%s/by-puuid/%s?api_key=%s", riotAccountBaseURL, puuid, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request to Riot Games API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to fetch account data: %s", resp.Status)
	}

	var account struct {
		GameName string `json:"gameName"`
		TagLine  string `json:"tagLine"`
	}

	err = json.NewDecoder(resp.Body).Decode(&account)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	return account.GameName, account.TagLine, nil
}
