package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNWSProvider_GetName(t *testing.T) {
	nws := NewNWSProvider()
	if nws.GetName() != "NWS" {
		t.Errorf("expected name 'NWS', got '%s'", nws.GetName())
	}
}

func TestNWSProvider_SupportedRegions(t *testing.T) {
	nws := NewNWSProvider()
	regions := nws.SupportedRegions()
	if len(regions) != 1 || regions[0] != "US" {
		t.Errorf("expected regions ['US'], got %v", regions)
	}
}

func TestNWSProvider_parseWindDirection(t *testing.T) {
	nws := NewNWSProvider()
	
	tests := []struct {
		input    string
		expected float64
	}{
		{"N", 0},
		{"NE", 45},
		{"E", 90},
		{"SE", 135},
		{"S", 180},
		{"SW", 225},
		{"W", 270},
		{"NW", 315},
		{"n", 0},    // Test case insensitive
		{"ne", 45},  // Test case insensitive
		{"Unknown", 0}, // Test unknown direction defaults to North
		{"", 0},     // Test empty string
	}

	for _, test := range tests {
		result := nws.parseWindDirection(test.input)
		if result != test.expected {
			t.Errorf("parseWindDirection(%q) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestNWSProvider_GetCurrentWeather_MockServer(t *testing.T) {
	// Create mock responses
	pointResponse := NWSPointResponse{
		Properties: NWSPointProperties{
			GridID:              "TOP",
			GridX:               31,
			GridY:               80,
			ObservationStations: "/gridpoints/TOP/31,80/stations",
		},
	}

	stationsResponse := struct {
		Features []struct {
			Properties struct {
				StationIdentifier string `json:"stationIdentifier"`
			} `json:"properties"`
		} `json:"features"`
	}{
		Features: []struct {
			Properties struct {
				StationIdentifier string `json:"stationIdentifier"`
			} `json:"properties"`
		}{
			{
				Properties: struct {
					StationIdentifier string `json:"stationIdentifier"`
				}{
					StationIdentifier: "KTOP",
				},
			},
		},
	}

	temp := 20.5
	humidity := 65.0
	pressure := 101325.0 // Pa
	windSpeed := 5.2
	windDir := 180.0
	visibility := 16000.0 // meters

	observationResponse := NWSObservationResponse{
		Properties: NWSObservationProperties{
			Timestamp: "2024-01-15T12:00:00-05:00",
			Temperature: NWSQuantitativeValue{
				Value: &temp,
			},
			RelativeHumidity: NWSQuantitativeValue{
				Value: &humidity,
			},
			BarometricPressure: NWSQuantitativeValue{
				Value: &pressure,
			},
			WindSpeed: NWSQuantitativeValue{
				Value: &windSpeed,
			},
			WindDirection: NWSQuantitativeValue{
				Value: &windDir,
			},
			Visibility: NWSQuantitativeValue{
				Value: &visibility,
			},
			TextDescription: "Clear skies",
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case strings.Contains(r.URL.Path, "/points/"):
			json.NewEncoder(w).Encode(pointResponse)
		case strings.Contains(r.URL.Path, "/stations") && !strings.Contains(r.URL.Path, "/observations"):
			json.NewEncoder(w).Encode(stationsResponse)
		case strings.Contains(r.URL.Path, "/observations/latest"):
			json.NewEncoder(w).Encode(observationResponse)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create NWS provider with test server
	nws := NewNWSProvider()
	nws.BaseURL = server.URL

	ctx := context.Background()
	forecast, err := nws.GetCurrentWeather(ctx, 39.0458, -76.6413) // Baltimore coordinates

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if forecast.SourceProvider != "NWS" {
		t.Errorf("expected source provider 'NWS', got '%s'", forecast.SourceProvider)
	}
	if forecast.Temperature != 20.5 {
		t.Errorf("expected temperature 20.5, got %f", forecast.Temperature)
	}
	if forecast.Humidity != 65.0 {
		t.Errorf("expected humidity 65.0, got %f", forecast.Humidity)
	}
	if forecast.Pressure != 1013.25 { // Converted from Pa to hPa
		t.Errorf("expected pressure 1013.25, got %f", forecast.Pressure)
	}
	if forecast.WindSpeed != 5.2 {
		t.Errorf("expected wind speed 5.2, got %f", forecast.WindSpeed)
	}
	if forecast.WindDirection != 180.0 {
		t.Errorf("expected wind direction 180.0, got %f", forecast.WindDirection)
	}
	if forecast.Visibility != 16.0 { // Converted from m to km
		t.Errorf("expected visibility 16.0, got %f", forecast.Visibility)
	}
	if forecast.Description != "Clear skies" {
		t.Errorf("expected description 'Clear skies', got '%s'", forecast.Description)
	}
}

func TestNWSProvider_GetForecast_MockServer(t *testing.T) {
	// Create test server first to get URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Will be replaced below
	}))
	defer server.Close()

	pointResponse := NWSPointResponse{
		Properties: NWSPointProperties{
			GridID:   "TOP",
			GridX:    31,
			GridY:    80,
			Forecast: server.URL + "/gridpoints/TOP/31,80/forecast",
		},
	}

	forecastResponse := NWSForecastResponse{
		Properties: NWSForecastProperties{
			Periods: []NWSForecastPeriod{
				{
					Number:           1,
					Name:             "Today",
					StartTime:        "2024-01-15T06:00:00-05:00",
					EndTime:          "2024-01-15T18:00:00-05:00",
					IsDaytime:        true,
					Temperature:      75,
					TemperatureUnit:  "F",
					WindSpeed:        "10 mph",
					WindDirection:    "SW",
					ShortForecast:    "Sunny",
					DetailedForecast: "Sunny skies with light winds",
				},
				{
					Number:           2,
					Name:             "Tonight",
					StartTime:        "2024-01-15T18:00:00-05:00",
					EndTime:          "2024-01-16T06:00:00-05:00",
					IsDaytime:        false,
					Temperature:      60,
					TemperatureUnit:  "F",
					WindSpeed:        "5 mph",
					WindDirection:    "W",
					ShortForecast:    "Clear",
					DetailedForecast: "Clear skies overnight",
				},
			},
		},
	}

	// Replace the server handler
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case strings.Contains(r.URL.Path, "/points/"):
			json.NewEncoder(w).Encode(pointResponse)
		case strings.Contains(r.URL.Path, "/forecast"):
			json.NewEncoder(w).Encode(forecastResponse)
		default:
			http.NotFound(w, r)
		}
	})

	nws := NewNWSProvider()
	nws.BaseURL = server.URL

	ctx := context.Background()
	forecasts, err := nws.GetForecast(ctx, 39.0458, -76.6413, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(forecasts) != 2 {
		t.Errorf("expected 2 forecast periods, got %d", len(forecasts))
	}

	// Test first period (daytime)
	first := forecasts[0]
	if first.SourceProvider != "NWS" {
		t.Errorf("expected source provider 'NWS', got '%s'", first.SourceProvider)
	}
	expectedTemp := (75.0 - 32) * 5 / 9 // Convert F to C
	if abs(first.Temperature-expectedTemp) > 0.1 {
		t.Errorf("expected temperature ~%f, got %f", expectedTemp, first.Temperature)
	}
	expectedWindSpeed := 10 * 0.44704 // Convert mph to m/s
	if abs(first.WindSpeed-expectedWindSpeed) > 0.1 {
		t.Errorf("expected wind speed ~%f, got %f", expectedWindSpeed, first.WindSpeed)
	}
	if first.WindDirection != 225.0 { // SW = 225 degrees
		t.Errorf("expected wind direction 225.0, got %f", first.WindDirection)
	}

	// Test second period (nighttime)
	second := forecasts[1]
	expectedTemp2 := (60.0 - 32) * 5 / 9 // Convert F to C
	if abs(second.Temperature-expectedTemp2) > 0.1 {
		t.Errorf("expected temperature ~%f, got %f", expectedTemp2, second.Temperature)
	}
}

func TestNWSProvider_GetAlerts_MockServer(t *testing.T) {
	alertsResponse := NWSAlertsResponse{
		Features: []NWSAlert{
			{
				Properties: NWSAlertProperties{
					ID:          "test-alert-1",
					Event:       "Severe Thunderstorm Warning",
					Headline:    "Severe Thunderstorm Warning in effect",
					Description: "Severe thunderstorms are expected",
					Severity:    "Severe",
					Urgency:     "Immediate",
					Category:    "Met",
					Onset:       time.Now().Format(time.RFC3339),
					Expires:     time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					AreaDesc:    "Test County",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alertsResponse)
	}))
	defer server.Close()

	nws := NewNWSProvider()
	nws.BaseURL = server.URL

	ctx := context.Background()
	alerts, err := nws.GetAlerts(ctx, 39.0458, -76.6413)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.ID != "test-alert-1" {
		t.Errorf("expected alert ID 'test-alert-1', got '%s'", alert.ID)
	}
	if alert.Title != "Severe Thunderstorm Warning" {
		t.Errorf("expected title 'Severe Thunderstorm Warning', got '%s'", alert.Title)
	}
	if alert.Severity != "severe" {
		t.Errorf("expected severity 'severe', got '%s'", alert.Severity)
	}
	if alert.Urgency != "immediate" {
		t.Errorf("expected urgency 'immediate', got '%s'", alert.Urgency)
	}
	if len(alert.Areas) != 1 || alert.Areas[0] != "Test County" {
		t.Errorf("expected areas ['Test County'], got %v", alert.Areas)
	}
}

func TestNWSProvider_ErrorHandling(t *testing.T) {
	// Test with server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	nws := NewNWSProvider()
	nws.BaseURL = server.URL

	ctx := context.Background()
	
	// Test GetCurrentWeather error handling
	_, err := nws.GetCurrentWeather(ctx, 39.0458, -76.6413)
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get grid point") {
		t.Errorf("expected 'failed to get grid point' in error, got: %v", err)
	}

	// Test GetForecast error handling
	_, err = nws.GetForecast(ctx, 39.0458, -76.6413, 1)
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}

	// Test GetAlerts error handling
	_, err = nws.GetAlerts(ctx, 39.0458, -76.6413)
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}