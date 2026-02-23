package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	lat := flag.Float64("lat", 0, "Latitude of the current position (required)")
	lon := flag.Float64("lon", 0, "Longitude of the current position (required)")
	lang := flag.String("lang", "de", "Language for the briefing (e.g. de, en, fr)")
	promptPath := flag.String("prompt", "", "Path to the system prompt markdown file (default: prompt.md next to binary)")
	flag.Parse()

	if *lat == 0 && *lon == 0 {
		fmt.Fprintln(os.Stderr, "Error: --lat and --lon are required")
		fmt.Fprintln(os.Stderr, "Usage: briefing --lat <latitude> --lon <longitude> [--lang <language>] [--prompt <prompt.md>]")
		os.Exit(1)
	}

	promptFile := resolvePromptPath(*promptPath)
	promptText, err := os.ReadFile(promptFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading prompt file %s: %v\n", promptFile, err)
		os.Exit(1)
	}

	stdinContext, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Reverse geocoding position...")
	loc, err := ReverseGeocode(*lat, *lon)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error geocoding: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Location: %s\n", loc.DisplayName)

	fmt.Fprintln(os.Stderr, "Fetching weather data...")
	weather, err := FetchWeather(*lat, *lon)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching weather: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Weather: %.1f°C, %s\n", weather.Current.Temperature, weatherCodeToText(weather.Current.WeatherCode))

	fmt.Fprintln(os.Stderr, "Generating briefing via OpenAI...")
	briefing, err := GenerateBriefing(loc, weather, stdinContext, string(promptText), *lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating briefing: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(briefing)
}

// resolvePromptPath finds the prompt.md file, checking the explicit path first,
// then falling back to the directory of the running executable.
func resolvePromptPath(explicit string) string {
	if explicit != "" {
		return explicit
	}

	exe, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "prompt.md")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Fall back to current working directory
	return "prompt.md"
}

func readStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// No piped input
		return "", nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
