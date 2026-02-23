package main

import (
	"testing"
)

func TestDegToCompass(t *testing.T) {
	tests := []struct {
		deg  float64
		want string
	}{
		{0, "N"},
		{45, "NE"},
		{90, "E"},
		{135, "SE"},
		{180, "S"},
		{225, "SW"},
		{270, "W"},
		{315, "NW"},
		{360, "N"},
		{22.5, "NNE"},
		{350, "N"},
	}

	for _, tt := range tests {
		got := degToCompass(tt.deg)
		if got != tt.want {
			t.Errorf("degToCompass(%v) = %q, want %q", tt.deg, got, tt.want)
		}
	}
}

func TestWeatherCodeToText(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{0, "Clear sky"},
		{3, "Overcast"},
		{61, "Slight rain"},
		{95, "Thunderstorm"},
		{999, "Unknown (999)"},
	}

	for _, tt := range tests {
		got := weatherCodeToText(tt.code)
		if got != tt.want {
			t.Errorf("weatherCodeToText(%d) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestFormatWeatherData(t *testing.T) {
	data := WeatherData{
		Timezone: "Europe/Berlin",
		Current: CurrentWeather{
			Temperature:   22.5,
			WindSpeed:     15.0,
			WindDirection: 180,
			WeatherCode:   1,
			Humidity:      65,
			Pressure:      1013,
			CloudCover:    30,
			Precipitation: 0,
		},
		Daily: []DailyForecast{
			{
				Date:              "2026-03-15",
				TempMax:           25,
				TempMin:           15,
				PrecipitationSum:  0,
				PrecipitationProb: 10,
				WindSpeedMax:      20,
				WindDirection:     180,
				WeatherCode:       1,
			},
		},
	}

	result := FormatWeatherData(data)

	if result == "" {
		t.Error("FormatWeatherData returned empty string")
	}

	checks := []string{
		"22.5°C",
		"Europe/Berlin",
		"15.0 km/h",
		"S",
		"Mainly clear",
		"2026-03-15",
	}
	for _, check := range checks {
		if !contains(result, check) {
			t.Errorf("FormatWeatherData output missing %q", check)
		}
	}
}

func TestResolvePromptPath(t *testing.T) {
	// Explicit path always wins
	got := resolvePromptPath("/some/explicit/path.md")
	if got != "/some/explicit/path.md" {
		t.Errorf("resolvePromptPath with explicit path = %q, want /some/explicit/path.md", got)
	}
}

func TestSafeIndex(t *testing.T) {
	s := []float64{1.0, 2.0, 3.0}
	if got := safeIndex(s, 0); got != 1.0 {
		t.Errorf("safeIndex(s, 0) = %v, want 1.0", got)
	}
	if got := safeIndex(s, 5); got != 0 {
		t.Errorf("safeIndex(s, 5) = %v, want 0", got)
	}
	if got := safeIndex(nil, 0); got != 0 {
		t.Errorf("safeIndex(nil, 0) = %v, want 0", got)
	}
}

func TestSafeIndexInt(t *testing.T) {
	s := []int{10, 20, 30}
	if got := safeIndexInt(s, 1); got != 20 {
		t.Errorf("safeIndexInt(s, 1) = %v, want 20", got)
	}
	if got := safeIndexInt(s, 10); got != 0 {
		t.Errorf("safeIndexInt(s, 10) = %v, want 0", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
