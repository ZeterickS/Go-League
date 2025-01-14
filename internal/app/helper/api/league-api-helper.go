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

// LoadEnv loads environment variables from a .env file if they are not already set.
//
// Synopsis:
//
//	err := LoadEnv()
//
// Parameters:
//
//	None.
//
// Returns:
//   - error: An error if the .env file could not be loaded, otherwise nil.
//
// Notes:
//
//	This function is used to ensure that environment variables are loaded before making API requests.
func LoadEnv() error {
	if os.Getenv("ROPT_API_TOKEN") != "" {
		return nil
	}
	return godotenv.Load()
}

// waitForRateLimiters waits until both rate limiters allow a request.
//
// Synopsis:
//
//	waitForRateLimiters()
//
// Parameters:
//
//	None.
//
// Returns:
//
//	None.
//
// Notes:
//
//	This function is used to handle rate limiting for API requests.
func waitForRateLimiters() {
	for !rateLimiterPerSecond.Allow() || !rateLimiterPer2Minutes.Allow() {
		time.Sleep(time.Second)
	}
}

// makeRequest makes an HTTP GET request and handles rate limit errors.
//
// Synopsis:
//
//	resp, err := makeRequest(url)
//
// Parameters:
//   - url: string - The URL to make the request to.
//
// Returns:
//   - *http.Response: The HTTP response.
//   - error: An error if the request failed.
//
// Notes:
//
//	This function makes an HTTP GET request and waits for 10 seconds if the rate limit is exceeded.
func makeRequest(url string) (*http.Response, error) {
	waitForRateLimiters()
	for retries := 0; retries < 2; retries++ {
		resp, err := http.Get(url)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				time.Sleep(10 * time.Second)
				continue
			}
			if retries == 1 {
				return nil, fmt.Errorf("failed to make request after retries: %w", err)
			}
		} else {
			return resp, nil
		}
	}
	return nil, fmt.Errorf("failed to make request after retries")
}

// GetSummonerByTag fetches summoner data by tag from the League of Legends API.
//
// Synopsis:
//
//	summoner, err := GetSummonerByTag("summonerName", "tagLine")
//
// Parameters:
//   - name: string - The summoner's name.
//   - tagLine: string - The summoner's tag line.
//
// Returns:
//   - *summoner.Summoner: The summoner data.
//   - error: An error if the summoner data could not be fetched.
//
// Notes:
//
//	This function first fetches the PUUID using the summoner's name and tag line, then fetches the summoner data using the PUUID.
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
	resp, err := makeRequest(url)
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

// GetSummonerByPUUID fetches summoner data by PUUID from the League of Legends API.
//
// Synopsis:
//
//	summoner, err := GetSummonerByPUUID("puuid")
//
// Parameters:
//   - puuid: string - The summoner's PUUID.
//
// Returns:
//   - *summoner.Summoner: The summoner data.
//   - error: An error if the summoner data could not be fetched.
//
// Notes:
//
//	This function fetches the summoner data using the PUUID and also retrieves the summoner's rank.
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
	resp, err := makeRequest(url)
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

// GetSummonerRank fetches the rank and division of a summoner by their ID from the League of Legends API.
//
// Synopsis:
//
//	soloRank, flexRank, err := GetSummonerRank("summonerID")
//
// Parameters:
//   - summonerID: string - The summoner's ID.
//
// Returns:
//   - rank.Rank: The solo rank of the summoner.
//   - rank.Rank: The flex rank of the summoner.
//   - error: An error if the rank data could not be fetched.
//
// Notes:
//
//	This function fetches the rank data for both solo and flex queues.
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
	resp, err := makeRequest(url)
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

// GetLastRankedMatch fetches the last ranked match of a summoner by their PUUID from the League of Legends API.
//
// Synopsis:
//
//	matchID, err := GetLastRankedMatch("puuid")
//
// Parameters:
//   - puuid: string - The summoner's PUUID.
//
// Returns:
//   - string: The ID of the last ranked match.
//   - error: An error if the match data could not be fetched.
//
// Notes:
//
//	This function fetches the ID of the last ranked match played by the summoner.
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
	resp, err := makeRequest(url)
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

// GetOngoingMatchByPUUID checks if there is an ongoing match for the given summoner's PUUID.
//
// Synopsis:
//
//	ongoingMatch, err := GetOngoingMatchByPUUID("puuid", "apiKey")
//
// Parameters:
//   - puuid: string - The summoner's PUUID.
//   - apiKey: string - The API key for authentication.
//
// Returns:
//   - *match.Match: The ongoing match data, or nil if no match is found.
//   - error: An error if the match data could not be fetched.
//
// Notes:
//
//	This function checks if there is an ongoing match for the summoner and returns the match data if found.
func GetOngoingMatchByPUUID(puuid, apiKey string) (*match.Match, error) {
	waitForRateLimiters()

	url := fmt.Sprintf("%s/by-summoner/%s?api_key=%s", riotSpectatorBaseURL, puuid, apiKey)
	resp, err := makeRequest(url)
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

	return ongoingMatch, nil
}

// GetNameTagByPUUID fetches the name and tag from a PUUID and returns them as two separate strings.
//
// Synopsis:
//
//	name, tagLine, err := GetNameTagByPUUID("puuid")
//
// Parameters:
//   - puuid: string - The summoner's PUUID.
//
// Returns:
//   - string: The summoner's name.
//   - string: The summoner's tag line.
//   - error: An error if the name and tag line could not be fetched.
//
// Notes:
//
//	This function fetches the summoner's name and tag line using the PUUID.
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
	resp, err := makeRequest(url)
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
