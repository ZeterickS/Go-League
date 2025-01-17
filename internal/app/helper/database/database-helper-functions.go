package databaseHelper

import (
	"discord-bot/types/match"
	"discord-bot/types/summoner"
	"encoding/json"
	"fmt"
	"os"
)

// TODO - To ENV VARS
const filename = "summoners.json"
const ongoingFilename = "ongoing_matches.json"

// SaveSummonersToFile saves a map of Summoner instances to a JSON file
func SaveSummonersToFile(summoners map[string]*summoner.Summoner) error {
	// Marshal the summoners to JSON
	data, err := json.MarshalIndent(summoners, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summoners: %v", err)
	}

	// Write the summoners to the file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// LoadSummonersFromFile loads a map of Summoner instances from a JSON file
func LoadSummonersFromFile() (map[string]*summoner.Summoner, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*summoner.Summoner), nil
		}
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	if len(data) == 0 {
		return make(map[string]*summoner.Summoner), nil
	}

	var summoners map[string]*summoner.Summoner
	err = json.Unmarshal(data, &summoners)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summoners: %v", err)
	}

	return summoners, nil
}

// GetSummonerByName retrieves a Summoner instance by name from the map
func GetSummonerByName(summoners map[string]*summoner.Summoner, name string) (*summoner.Summoner, error) {
	summoner, exists := summoners[name]
	if !exists {
		return nil, fmt.Errorf("summoner with name %s not found", name)
	}
	return summoner, nil
}

// SaveOngoingMatchToFile saves an OngoingMatch instance to a JSON file
func SaveOngoingMatchToFile(ongoingMatch *match.Match) error {
	data, err := json.MarshalIndent(ongoingMatch, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ongoing match: %v", err)
	}

	err = os.WriteFile(ongoingFilename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// LoadOngoingMatchFromFile loads an array of Matches instances from a JSON file
func LoadOngoingMatchFromFile() (map[int64]*match.Match, error) {
	var ongoingMatches map[int64]*match.Match

	data, err := os.ReadFile(ongoingFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No ongoing match file found
		}
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	if len(data) == 0 {
		return make(map[int64]*match.Match), nil
	}

	err = json.Unmarshal(data, &ongoingMatches)
	if err != nil {
		return make(map[int64]*match.Match), fmt.Errorf("failed to unmarshal ongoing matches: %v", err)
	}

	return ongoingMatches, nil
}
