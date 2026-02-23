package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

// GenerateBriefing calls OpenAI with weather data, location, and context to produce a daily briefing.
func GenerateBriefing(loc Location, weather WeatherData, stdinContext, promptText, lang string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	userMessage := buildUserMessage(loc, weather, stdinContext, lang)

	fmt.Fprintf(os.Stderr, "User Message:\n%s", userMessage)

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model:        openai.ChatModelGPT5,
		Instructions: openai.String(promptText),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(userMessage),
		},
		Reasoning: shared.ReasoningParam{
			Effort: shared.ReasoningEffortMedium,
		},
		Tools: []responses.ToolUnionParam{
			{OfWebSearch: &responses.WebSearchToolParam{
				Type:              responses.WebSearchToolTypeWebSearch,
				SearchContextSize: responses.WebSearchToolSearchContextSizeHigh,
				UserLocation: responses.WebSearchToolUserLocationParam{
					Type:    "approximate",
					City:    openai.String(loc.City),
					Region:  openai.String(loc.Region),
					Country: openai.String(strings.ToUpper(loc.CountryCode)),
				},
			}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "OpenAI API Response:\n%s", resp.OutputText())

	return resp.OutputText(), nil
}

func buildUserMessage(loc Location, weather WeatherData, stdinContext, lang string) string {
	var b strings.Builder

	b.WriteString("=== LOCATION ===\n")
	b.WriteString(fmt.Sprintf("Coordinates: %.5f, %.5f\n", loc.Latitude, loc.Longitude))
	b.WriteString(fmt.Sprintf("Place: %s\n", loc.DisplayName))
	if loc.City != "" {
		b.WriteString(fmt.Sprintf("City: %s\n", loc.City))
	}
	if loc.Region != "" {
		b.WriteString(fmt.Sprintf("Region: %s\n", loc.Region))
	}
	b.WriteString(fmt.Sprintf("Country: %s (%s)\n", loc.Country, strings.ToUpper(loc.CountryCode)))
	b.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("Language: %s\n", lang))

	b.WriteString("\n")
	b.WriteString(FormatWeatherData(weather))

	if stdinContext != "" {
		b.WriteString("\n")
		b.WriteString(stdinContext)
	}

	return b.String()
}
