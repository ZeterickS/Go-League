package cdragon

import "fmt"

const baseCDNURL = "https://cdn.communitydragon.org/15.1.1"

// GetProfileIconURL generates the URL for a given profile icon ID
func GetProfileIconURL(profileIconID int) string {
	return fmt.Sprintf("%s/profile-icon/%d", baseCDNURL, profileIconID)
}

// Add more functions for other resources as needed
