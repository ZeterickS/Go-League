package cdragon

import "fmt"

const baseCDNURL = "https://cdn.communitydragon.org/15.1.1"
const baseRankedURL = "https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/images/ranked-mini-crests"

// GetProfileIconURL generates the URL for a given profile icon ID
func GetProfileIconURL(profileIconID int) string {
	return fmt.Sprintf("%s/profile-icon/%d", baseCDNURL, profileIconID)
}

// GetRankedPictureURL generates the URL for a given rank
func GetRankedPictureURL(rank string) string {
	return fmt.Sprintf("%s/%s.png", baseRankedURL, rank)
}

// Add more functions for other resources as needed
