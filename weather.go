package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FetchWeather retrieves current conditions and forecasts from Open-Meteo.
func FetchWeather(lat, lon float64) (WeatherData, error) {
	hourlyParams := []string{
		"temperature_2m", "wind_speed_10m", "wind_direction_10m",
		"precipitation", "weather_code",
	}
	dailyParams := []string{
		"temperature_2m_max", "temperature_2m_min",
		"precipitation_sum", "precipitation_probability_max",
		"wind_speed_10m_max", "wind_direction_10m_dominant",
		"weather_code",
	}
	currentParams := []string{
		"temperature_2m", "wind_speed_10m", "wind_direction_10m",
		"relative_humidity_2m", "surface_pressure", "cloud_cover",
		"precipitation", "weather_code",
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f"+
			"&current=%s&hourly=%s&daily=%s"+
			"&timezone=auto&forecast_days=7&forecast_hours=48",
		lat, lon,
		strings.Join(currentParams, ","),
		strings.Join(hourlyParams, ","),
		strings.Join(dailyParams, ","),
	)

	weatherResp, err := fetchJSON[openMeteoWeatherResponse](url)
	if err != nil {
		return WeatherData{}, fmt.Errorf("fetching weather: %w", err)
	}

	data := WeatherData{
		Timezone: weatherResp.Timezone,
		Current: CurrentWeather{
			Temperature:   weatherResp.Current.Temperature2m,
			WindSpeed:     weatherResp.Current.WindSpeed10m,
			WindDirection: weatherResp.Current.WindDirection10m,
			WeatherCode:   weatherResp.Current.WeatherCode,
			Humidity:      weatherResp.Current.RelativeHumidity2m,
			Pressure:      weatherResp.Current.SurfacePressure,
			CloudCover:    weatherResp.Current.CloudCover,
			Precipitation: weatherResp.Current.Precipitation,
		},
	}

	for i, t := range weatherResp.Daily.Time {
		data.Daily = append(data.Daily, DailyForecast{
			Date:              t,
			TempMax:           safeIndex(weatherResp.Daily.Temperature2mMax, i),
			TempMin:           safeIndex(weatherResp.Daily.Temperature2mMin, i),
			PrecipitationSum:  safeIndex(weatherResp.Daily.PrecipitationSum, i),
			PrecipitationProb: safeIndexInt(weatherResp.Daily.PrecipitationProbMax, i),
			WindSpeedMax:      safeIndex(weatherResp.Daily.WindSpeed10mMax, i),
			WindDirection:     safeIndex(weatherResp.Daily.WindDirection10mDom, i),
			WeatherCode:       safeIndexInt(weatherResp.Daily.WeatherCode, i),
		})
	}

	for i, t := range weatherResp.Hourly.Time {
		data.Hourly = append(data.Hourly, HourlyForecast{
			Time:          t,
			Temperature:   safeIndex(weatherResp.Hourly.Temperature2m, i),
			WindSpeed:     safeIndex(weatherResp.Hourly.WindSpeed10m, i),
			WindDirection: safeIndex(weatherResp.Hourly.WindDirection10m, i),
			Precipitation: safeIndex(weatherResp.Hourly.Precipitation, i),
			WeatherCode:   safeIndexInt(weatherResp.Hourly.WeatherCode, i),
		})
	}

	marine, err := fetchMarine(lat, lon)
	if err != nil {
		fmt.Printf("Warning: could not fetch marine data: %v\n", err)
	} else {
		data.Marine = marine.Current
		data.HourlyMarine = marine.Hourly
	}

	return data, nil
}

type marineResult struct {
	Current MarineData
	Hourly  []HourlyMarine
}

func fetchMarine(lat, lon float64) (marineResult, error) {
	hourlyParams := []string{
		"wave_height", "wave_direction", "wave_period",
		"wind_wave_height",
		"swell_wave_height", "swell_wave_direction", "swell_wave_period",
	}
	currentParams := []string{
		"wave_height", "wave_direction", "wave_period",
		"wind_wave_height",
		"swell_wave_height", "swell_wave_direction", "swell_wave_period",
	}

	url := fmt.Sprintf(
		"https://marine-api.open-meteo.com/v1/marine?latitude=%f&longitude=%f"+
			"&current=%s&hourly=%s&timezone=auto&forecast_hours=48",
		lat, lon,
		strings.Join(currentParams, ","),
		strings.Join(hourlyParams, ","),
	)

	resp, err := fetchJSON[openMeteoMarineResponse](url)
	if err != nil {
		return marineResult{}, err
	}

	result := marineResult{
		Current: MarineData{
			WaveHeight:      resp.Current.WaveHeight,
			WaveDirection:   resp.Current.WaveDirection,
			WavePeriod:      resp.Current.WavePeriod,
			WindWaveHeight:  resp.Current.WindWaveHeight,
			SwellWaveHeight: resp.Current.SwellWaveHeight,
			SwellWaveDir:    resp.Current.SwellWaveDirection,
			SwellWavePeriod: resp.Current.SwellWavePeriod,
		},
	}

	for i, t := range resp.Hourly.Time {
		result.Hourly = append(result.Hourly, HourlyMarine{
			Time:            t,
			WaveHeight:      safeIndex(resp.Hourly.WaveHeight, i),
			WaveDirection:   safeIndex(resp.Hourly.WaveDirection, i),
			WavePeriod:      safeIndex(resp.Hourly.WavePeriod, i),
			WindWaveHeight:  safeIndex(resp.Hourly.WindWaveHeight, i),
			SwellWaveHeight: safeIndex(resp.Hourly.SwellWaveHeight, i),
			SwellWaveDir:    safeIndex(resp.Hourly.SwellWaveDirection, i),
			SwellWavePeriod: safeIndex(resp.Hourly.SwellWavePeriod, i),
		})
	}

	return result, nil
}

func fetchJSON[T any](url string) (T, error) {
	var zero T
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("decoding response: %w", err)
	}
	return result, nil
}

func safeIndex(s []float64, i int) float64 {
	if i < len(s) {
		return s[i]
	}
	return 0
}

func safeIndexInt(s []int, i int) int {
	if i < len(s) {
		return s[i]
	}
	return 0
}

// FormatWeatherData produces a human-readable summary of all weather data for the LLM prompt.
func FormatWeatherData(w WeatherData) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== CURRENT WEATHER (Timezone: %s) ===\n", w.Timezone))
	b.WriteString(fmt.Sprintf("Temperature: %.1f°C\n", w.Current.Temperature))
	b.WriteString(fmt.Sprintf("Wind: %.1f km/h from %s (%d°)\n", w.Current.WindSpeed, degToCompass(w.Current.WindDirection), int(w.Current.WindDirection)))
	b.WriteString(fmt.Sprintf("Humidity: %d%%\n", w.Current.Humidity))
	b.WriteString(fmt.Sprintf("Pressure: %.0f hPa\n", w.Current.Pressure))
	b.WriteString(fmt.Sprintf("Cloud cover: %d%%\n", w.Current.CloudCover))
	b.WriteString(fmt.Sprintf("Precipitation: %.1f mm\n", w.Current.Precipitation))
	b.WriteString(fmt.Sprintf("Conditions: %s\n", weatherCodeToText(w.Current.WeatherCode)))

	b.WriteString("\n=== 7-DAY FORECAST ===\n")
	for _, d := range w.Daily {
		b.WriteString(fmt.Sprintf("%s: %s, %.0f–%.0f°C, wind up to %.0f km/h from %s, precip %.1fmm (prob %d%%)\n",
			d.Date, weatherCodeToText(d.WeatherCode),
			d.TempMin, d.TempMax, d.WindSpeedMax,
			degToCompass(d.WindDirection), d.PrecipitationSum, d.PrecipitationProb))
	}

	b.WriteString("\n=== HOURLY FORECAST (next 48h) ===\n")
	for _, h := range w.Hourly {
		b.WriteString(fmt.Sprintf("%s: %.1f°C, wind %.0f km/h %s, precip %.1fmm, %s\n",
			h.Time, h.Temperature, h.WindSpeed,
			degToCompass(h.WindDirection), h.Precipitation, weatherCodeToText(h.WeatherCode)))
	}

	if w.Marine.WaveHeight > 0 {
		b.WriteString("\n=== CURRENT MARINE CONDITIONS ===\n")
		b.WriteString(fmt.Sprintf("Wave height: %.1fm, direction %s (%d°), period %.1fs\n",
			w.Marine.WaveHeight, degToCompass(w.Marine.WaveDirection), int(w.Marine.WaveDirection), w.Marine.WavePeriod))
		b.WriteString(fmt.Sprintf("Wind waves: %.1fm\n", w.Marine.WindWaveHeight))
		b.WriteString(fmt.Sprintf("Swell: %.1fm from %s, period %.1fs\n",
			w.Marine.SwellWaveHeight, degToCompass(w.Marine.SwellWaveDir), w.Marine.SwellWavePeriod))

		b.WriteString("\n=== HOURLY MARINE FORECAST (next 48h) ===\n")
		for _, m := range w.HourlyMarine {
			b.WriteString(fmt.Sprintf("%s: waves %.1fm %s period %.1fs, swell %.1fm %s\n",
				m.Time, m.WaveHeight, degToCompass(m.WaveDirection), m.WavePeriod,
				m.SwellWaveHeight, degToCompass(m.SwellWaveDir)))
		}
	}

	return b.String()
}

func degToCompass(deg float64) string {
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := int((deg + 11.25) / 22.5) % 16
	if idx < 0 {
		idx += 16
	}
	return dirs[idx]
}

func weatherCodeToText(code int) string {
	codes := map[int]string{
		0:  "Clear sky",
		1:  "Mainly clear", 2: "Partly cloudy", 3: "Overcast",
		45: "Fog", 48: "Depositing rime fog",
		51: "Light drizzle", 53: "Moderate drizzle", 55: "Dense drizzle",
		56: "Light freezing drizzle", 57: "Dense freezing drizzle",
		61: "Slight rain", 63: "Moderate rain", 65: "Heavy rain",
		66: "Light freezing rain", 67: "Heavy freezing rain",
		71: "Slight snowfall", 73: "Moderate snowfall", 75: "Heavy snowfall",
		77: "Snow grains",
		80: "Slight rain showers", 81: "Moderate rain showers", 82: "Violent rain showers",
		85: "Slight snow showers", 86: "Heavy snow showers",
		95: "Thunderstorm", 96: "Thunderstorm with slight hail", 99: "Thunderstorm with heavy hail",
	}
	if text, ok := codes[code]; ok {
		return text
	}
	return fmt.Sprintf("Unknown (%d)", code)
}

// Open-Meteo JSON response structures

type openMeteoWeatherResponse struct {
	Timezone string `json:"timezone"`
	Current  struct {
		Temperature2m      float64 `json:"temperature_2m"`
		WindSpeed10m       float64 `json:"wind_speed_10m"`
		WindDirection10m   float64 `json:"wind_direction_10m"`
		RelativeHumidity2m int     `json:"relative_humidity_2m"`
		SurfacePressure    float64 `json:"surface_pressure"`
		CloudCover         int     `json:"cloud_cover"`
		Precipitation      float64 `json:"precipitation"`
		WeatherCode        int     `json:"weather_code"`
	} `json:"current"`
	Daily struct {
		Time                 []string  `json:"time"`
		Temperature2mMax     []float64 `json:"temperature_2m_max"`
		Temperature2mMin     []float64 `json:"temperature_2m_min"`
		PrecipitationSum     []float64 `json:"precipitation_sum"`
		PrecipitationProbMax []int     `json:"precipitation_probability_max"`
		WindSpeed10mMax      []float64 `json:"wind_speed_10m_max"`
		WindDirection10mDom  []float64 `json:"wind_direction_10m_dominant"`
		WeatherCode          []int     `json:"weather_code"`
	} `json:"daily"`
	Hourly struct {
		Time            []string  `json:"time"`
		Temperature2m   []float64 `json:"temperature_2m"`
		WindSpeed10m    []float64 `json:"wind_speed_10m"`
		WindDirection10m []float64 `json:"wind_direction_10m"`
		Precipitation   []float64 `json:"precipitation"`
		WeatherCode     []int     `json:"weather_code"`
	} `json:"hourly"`
}

type openMeteoMarineResponse struct {
	Current struct {
		WaveHeight         float64 `json:"wave_height"`
		WaveDirection      float64 `json:"wave_direction"`
		WavePeriod         float64 `json:"wave_period"`
		WindWaveHeight     float64 `json:"wind_wave_height"`
		SwellWaveHeight    float64 `json:"swell_wave_height"`
		SwellWaveDirection float64 `json:"swell_wave_direction"`
		SwellWavePeriod    float64 `json:"swell_wave_period"`
	} `json:"current"`
	Hourly struct {
		Time               []string  `json:"time"`
		WaveHeight         []float64 `json:"wave_height"`
		WaveDirection      []float64 `json:"wave_direction"`
		WavePeriod         []float64 `json:"wave_period"`
		WindWaveHeight     []float64 `json:"wind_wave_height"`
		SwellWaveHeight    []float64 `json:"swell_wave_height"`
		SwellWaveDirection []float64 `json:"swell_wave_direction"`
		SwellWavePeriod    []float64 `json:"swell_wave_period"`
	} `json:"hourly"`
}
