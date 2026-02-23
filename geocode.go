package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type nominatimResponse struct {
	DisplayName string            `json:"display_name"`
	Address     nominatimAddress  `json:"address"`
}

type nominatimAddress struct {
	City        string `json:"city"`
	Town        string `json:"town"`
	Village     string `json:"village"`
	Municipality string `json:"municipality"`
	County      string `json:"county"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

// ReverseGeocode resolves a GPS position to a human-readable location via Nominatim.
func ReverseGeocode(lat, lon float64) (Location, error) {
	url := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=json&accept-language=en",
		lat, lon,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Location{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "sailingnomads-briefing/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("nominatim request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Location{}, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	var result nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Location{}, fmt.Errorf("decoding nominatim response: %w", err)
	}

	city := result.Address.City
	if city == "" {
		city = result.Address.Town
	}
	if city == "" {
		city = result.Address.Village
	}
	if city == "" {
		city = result.Address.Municipality
	}

	region := result.Address.State
	if region == "" {
		region = result.Address.County
	}

	return Location{
		Latitude:    lat,
		Longitude:   lon,
		City:        city,
		Region:      region,
		Country:     result.Address.Country,
		CountryCode: result.Address.CountryCode,
		DisplayName: result.DisplayName,
	}, nil
}
