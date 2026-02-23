package main

// Location holds reverse-geocoded information about a GPS position.
type Location struct {
	Latitude    float64
	Longitude   float64
	City        string
	Region      string
	Country     string
	CountryCode string
	DisplayName string
}

// CurrentWeather holds the current weather conditions.
type CurrentWeather struct {
	Temperature   float64 // °C
	WindSpeed     float64 // km/h
	WindDirection float64 // degrees
	WeatherCode   int
	Humidity      int     // %
	Pressure      float64 // hPa
	CloudCover    int     // %
	Precipitation float64 // mm
}

// DailyForecast holds a single day's forecast.
type DailyForecast struct {
	Date              string
	TempMax           float64
	TempMin           float64
	PrecipitationSum  float64
	PrecipitationProb int
	WindSpeedMax      float64
	WindDirection     float64
	WeatherCode       int
}

// HourlyForecast holds a single hour's forecast.
type HourlyForecast struct {
	Time          string
	Temperature   float64
	WindSpeed     float64
	WindDirection float64
	Precipitation float64
	WeatherCode   int
}

// MarineData holds marine/wave conditions.
type MarineData struct {
	WaveHeight       float64 // meters
	WaveDirection    float64 // degrees
	WavePeriod       float64 // seconds
	WindWaveHeight   float64
	SwellWaveHeight  float64
	SwellWaveDir     float64
	SwellWavePeriod  float64
}

// HourlyMarine holds hourly marine forecast data.
type HourlyMarine struct {
	Time             string
	WaveHeight       float64
	WaveDirection    float64
	WavePeriod       float64
	WindWaveHeight   float64
	SwellWaveHeight  float64
	SwellWaveDir     float64
	SwellWavePeriod  float64
}

// WeatherData holds all weather information for a location.
type WeatherData struct {
	Current       CurrentWeather
	Daily         []DailyForecast
	Hourly        []HourlyForecast
	Marine        MarineData
	HourlyMarine  []HourlyMarine
	Timezone      string
}
