package providers

import (
	"context"
	"testing"
	"time"

	"stormlightlabs.org/weather_api/internal/models"
)

func TestProviderManager(t *testing.T) {
	pm := NewProviderManager()

	// Test initial state
	if len(pm.GetWeatherProviders()) != 0 {
		t.Errorf("expected 0 weather providers, got %d", len(pm.GetWeatherProviders()))
	}
	if len(pm.GetGeocodeProviders()) != 0 {
		t.Errorf("expected 0 geocode providers, got %d", len(pm.GetGeocodeProviders()))
	}

	// Create test providers
	nws := NewNWSProvider()
	census := NewCensusProvider()

	// Register providers
	pm.RegisterWeatherProvider(nws)
	pm.RegisterGeocodeProvider(census)

	// Test registered providers
	if len(pm.GetWeatherProviders()) != 1 {
		t.Errorf("expected 1 weather provider, got %d", len(pm.GetWeatherProviders()))
	}
	if len(pm.GetGeocodeProviders()) != 1 {
		t.Errorf("expected 1 geocode provider, got %d", len(pm.GetGeocodeProviders()))
	}

	// Test get by name
	weatherProvider := pm.GetWeatherProviderByName("NWS")
	if weatherProvider == nil {
		t.Error("expected to find NWS weather provider")
	}
	if weatherProvider.GetName() != "NWS" {
		t.Errorf("expected provider name 'NWS', got '%s'", weatherProvider.GetName())
	}

	geocodeProvider := pm.GetGeocodeProviderByName("Census")
	if geocodeProvider == nil {
		t.Error("expected to find Census geocode provider")
	}
	if geocodeProvider.GetName() != "Census" {
		t.Errorf("expected provider name 'Census', got '%s'", geocodeProvider.GetName())
	}

	// Test non-existent provider
	if pm.GetWeatherProviderByName("NonExistent") != nil {
		t.Error("expected nil for non-existent weather provider")
	}
	if pm.GetGeocodeProviderByName("NonExistent") != nil {
		t.Error("expected nil for non-existent geocode provider")
	}
}

func TestWeatherAlert(t *testing.T) {
	alert := WeatherAlert{
		ID:          "test-alert-1",
		Title:       "Severe Thunderstorm Warning",
		Description: "Severe thunderstorms expected",
		Severity:    "severe",
		Urgency:     "immediate",
		Category:    "met",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(2 * time.Hour),
		Areas:       []string{"Test County", "Another County"},
	}

	if alert.ID != "test-alert-1" {
		t.Errorf("expected ID 'test-alert-1', got '%s'", alert.ID)
	}
	if alert.Severity != "severe" {
		t.Errorf("expected severity 'severe', got '%s'", alert.Severity)
	}
	if len(alert.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(alert.Areas))
	}
}

func TestProviderResponse(t *testing.T) {
	now := time.Now()
	response := ProviderResponse{
		Provider:  "TestProvider",
		Timestamp: now,
		Data:      "test data",
		Cached:    true,
		TTL:       time.Hour,
	}

	if response.Provider != "TestProvider" {
		t.Errorf("expected provider 'TestProvider', got '%s'", response.Provider)
	}
	if response.Cached != true {
		t.Errorf("expected cached to be true, got %v", response.Cached)
	}
	if response.TTL != time.Hour {
		t.Errorf("expected TTL 1 hour, got %v", response.TTL)
	}
}

// Mock providers for testing interface compliance
type MockWeatherProvider struct {
	name string
}

func (m *MockWeatherProvider) GetName() string {
	return m.name
}

func (m *MockWeatherProvider) GetCurrentWeather(ctx context.Context, lat, lon float64) (*models.Forecast, error) {
	return &models.Forecast{
		SourceProvider: m.name,
		Temperature:    20.0,
		Humidity:       60.0,
	}, nil
}

func (m *MockWeatherProvider) GetForecast(ctx context.Context, lat, lon float64, days int) ([]*models.Forecast, error) {
	forecasts := make([]*models.Forecast, days)
	for i := 0; i < days; i++ {
		forecasts[i] = &models.Forecast{
			SourceProvider: m.name,
			Temperature:    20.0 + float64(i),
			ValidTime:      time.Now().Add(time.Duration(i) * 24 * time.Hour),
		}
	}
	return forecasts, nil
}

func (m *MockWeatherProvider) GetAlerts(ctx context.Context, lat, lon float64) ([]WeatherAlert, error) {
	return []WeatherAlert{
		{
			ID:       "mock-alert-1",
			Title:    "Test Alert",
			Severity: "minor",
		},
	}, nil
}

func (m *MockWeatherProvider) SupportedRegions() []string {
	return []string{"TEST"}
}

type MockGeocodeProvider struct {
	name string
}

func (m *MockGeocodeProvider) GetName() string {
	return m.name
}

func (m *MockGeocodeProvider) GeocodeAddress(ctx context.Context, address string) ([]*models.Place, error) {
	return []*models.Place{
		{
			DisplayName: address,
			Latitude:    40.7128,
			Longitude:   -74.0060,
			Source:      m.name,
			Confidence:  0.95,
		},
	}, nil
}

func (m *MockGeocodeProvider) ReverseGeocode(ctx context.Context, lat, lon float64) (*models.Place, error) {
	return &models.Place{
		DisplayName: "Test Address",
		Latitude:    lat,
		Longitude:   lon,
		Source:      m.name,
		Confidence:  0.90,
	}, nil
}

func (m *MockGeocodeProvider) SupportedRegions() []string {
	return []string{"TEST"}
}

func TestProviderInterfaces(t *testing.T) {
	// Test that mock providers implement the interfaces
	var _ WeatherProvider = &MockWeatherProvider{}
	var _ GeocodeProvider = &MockGeocodeProvider{}

	// Test that real providers implement the interfaces
	var _ WeatherProvider = &NWSProvider{}
	var _ GeocodeProvider = &CensusProvider{}
}

func TestMockProviders(t *testing.T) {
	ctx := context.Background()

	// Test mock weather provider
	mockWeather := &MockWeatherProvider{name: "MockWeather"}

	if mockWeather.GetName() != "MockWeather" {
		t.Errorf("expected name 'MockWeather', got '%s'", mockWeather.GetName())
	}

	regions := mockWeather.SupportedRegions()
	if len(regions) != 1 || regions[0] != "TEST" {
		t.Errorf("expected regions ['TEST'], got %v", regions)
	}

	currentWeather, err := mockWeather.GetCurrentWeather(ctx, 40.7128, -74.0060)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if currentWeather.Temperature != 20.0 {
		t.Errorf("expected temperature 20.0, got %f", currentWeather.Temperature)
	}
	if currentWeather.SourceProvider != "MockWeather" {
		t.Errorf("expected source provider 'MockWeather', got '%s'", currentWeather.SourceProvider)
	}

	forecast, err := mockWeather.GetForecast(ctx, 40.7128, -74.0060, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forecast) != 3 {
		t.Errorf("expected 3 forecast periods, got %d", len(forecast))
	}
	if forecast[0].Temperature != 20.0 {
		t.Errorf("expected first forecast temperature 20.0, got %f", forecast[0].Temperature)
	}
	if forecast[2].Temperature != 22.0 {
		t.Errorf("expected third forecast temperature 22.0, got %f", forecast[2].Temperature)
	}

	alerts, err := mockWeather.GetAlerts(ctx, 40.7128, -74.0060)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].ID != "mock-alert-1" {
		t.Errorf("expected alert ID 'mock-alert-1', got '%s'", alerts[0].ID)
	}

	// Test mock geocode provider
	mockGeocode := &MockGeocodeProvider{name: "MockGeocode"}

	if mockGeocode.GetName() != "MockGeocode" {
		t.Errorf("expected name 'MockGeocode', got '%s'", mockGeocode.GetName())
	}

	geocodeRegions := mockGeocode.SupportedRegions()
	if len(geocodeRegions) != 1 || geocodeRegions[0] != "TEST" {
		t.Errorf("expected regions ['TEST'], got %v", geocodeRegions)
	}

	places, err := mockGeocode.GeocodeAddress(ctx, "123 Test St, Test City, TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(places) != 1 {
		t.Errorf("expected 1 place, got %d", len(places))
	}
	if places[0].Latitude != 40.7128 {
		t.Errorf("expected latitude 40.7128, got %f", places[0].Latitude)
	}
	if places[0].Source != "MockGeocode" {
		t.Errorf("expected source 'MockGeocode', got '%s'", places[0].Source)
	}

	place, err := mockGeocode.ReverseGeocode(ctx, 40.7128, -74.0060)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if place.DisplayName != "Test Address" {
		t.Errorf("expected display name 'Test Address', got '%s'", place.DisplayName)
	}
	if place.Latitude != 40.7128 {
		t.Errorf("expected latitude 40.7128, got %f", place.Latitude)
	}
	if place.Longitude != -74.0060 {
		t.Errorf("expected longitude -74.0060, got %f", place.Longitude)
	}
}

func TestProviderManagerWithMocks(t *testing.T) {
	pm := NewProviderManager()

	// Register mock providers
	mockWeather := &MockWeatherProvider{name: "MockWeather"}
	mockGeocode := &MockGeocodeProvider{name: "MockGeocode"}

	pm.RegisterWeatherProvider(mockWeather)
	pm.RegisterGeocodeProvider(mockGeocode)

	// Test provider retrieval
	weatherProvider := pm.GetWeatherProviderByName("MockWeather")
	if weatherProvider == nil {
		t.Error("expected to find MockWeather provider")
	}

	geocodeProvider := pm.GetGeocodeProviderByName("MockGeocode")
	if geocodeProvider == nil {
		t.Error("expected to find MockGeocode provider")
	}

	// Test functionality through manager
	ctx := context.Background()

	forecast, err := weatherProvider.GetCurrentWeather(ctx, 40.7128, -74.0060)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if forecast.SourceProvider != "MockWeather" {
		t.Errorf("expected source provider 'MockWeather', got '%s'", forecast.SourceProvider)
	}

	places, err := geocodeProvider.GeocodeAddress(ctx, "test address")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(places) != 1 {
		t.Errorf("expected 1 place, got %d", len(places))
	}
}
