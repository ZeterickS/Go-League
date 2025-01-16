package assethelper

import (
	"discord-bot/types/match"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var runeData map[int]string
var spellData map[int]string

func init() {
	// Parse runes.json
	file, err := os.Open("../../assets/15.1.1/jsonmaps/runes.json")
	if err != nil {
		fmt.Printf("Error opening runes.json: %v\n", err)
		return
	}
	defer file.Close()

	var runes []struct {
		ID    int `json:"id"`
		Slots []struct {
			Runes []struct {
				ID   int    `json:"id"`
				Icon string `json:"icon"`
			} `json:"runes"`
		} `json:"slots"`
	}

	if err := json.NewDecoder(file).Decode(&runes); err != nil {
		fmt.Printf("Error decoding runes.json: %v\n", err)
		return
	}

	runeData = make(map[int]string)
	for _, runeCategory := range runes {
		for _, slot := range runeCategory.Slots {
			for _, rune := range slot.Runes {
				runeData[rune.ID] = rune.Icon
			}
		}
	}

	// Parse spells.json
	file, err = os.Open("../../assets/15.1.1/jsonmaps/spells.json")
	if err != nil {
		fmt.Printf("Error opening spells.json: %v\n", err)
		return
	}
	defer file.Close()

	var spells struct {
		Data map[string]struct {
			Key   string `json:"key"`
			Image struct {
				Full string `json:"full"`
			} `json:"image"`
		} `json:"data"`
	}

	if err := json.NewDecoder(file).Decode(&spells); err != nil {
		fmt.Printf("Error decoding spells.json: %v\n", err)
		return
	}

	spellData = make(map[int]string)
	for _, spell := range spells.Data {
		spellID, err := strconv.Atoi(spell.Key)
		if err != nil {
			fmt.Printf("Error converting spell ID %s to int: %v\n", spell.Key, err)
			continue
		}
		spellData[spellID] = spell.Image.Full
	}
}

// GetItemFiles takes a list of item IDs and returns a slice of os.File pointers corresponding to the assets.
func GetItemFiles(itemIDs []int) ([]*os.File, error) {
	var files []*os.File
	for _, itemID := range itemIDs {
		filePath := fmt.Sprintf("../../assets/items/%d.png", itemID)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file for item ID %d: %w", itemID, err)
		}
		files = append(files, file)
	}
	return files, nil
}

// GetPerkFiles takes a list of perk IDs and returns a slice of os.File pointers corresponding to the assets.
func GetPerkFiles(perks match.Perks) ([]*os.File, error) {
	var files []*os.File
	var styleIDs []int
	styleIDs = append(styleIDs, perks.PerkStyle, perks.PerkSubStyle)
	for _, perkID := range styleIDs {
		iconPath, ok := runeData[perkID]
		if !ok {
			return nil, fmt.Errorf("icon not found for perk ID %d", perkID)
		}
		filePath := filepath.Join("../../assets", iconPath)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file for perk ID %d: %w", perkID, err)
		}
		files = append(files, file)
	}
	return files, nil
}

// GetSpellFiles takes a list of spell IDs and returns a slice of os.File pointers corresponding to the assets.
func GetSpellFiles(spellIDs []int) ([]*os.File, error) {
	var files []*os.File
	for _, spellID := range spellIDs {
		imagePath, ok := spellData[spellID]
		if !ok {
			return nil, fmt.Errorf("image not found for spell ID %d", spellID)
		}
		filePath := filepath.Join("../../assets/spells", imagePath)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file for spell ID %d: %w", spellID, err)
		}
		files = append(files, file)
	}
	return files, nil
}
