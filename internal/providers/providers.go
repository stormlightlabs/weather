package providers

import (
	"context"
	"time"

	"stormlightlabs.org/weather_api/internal/models"
)

// WeatherProvider defines the interface for weather data providers
type WeatherProvider interface {
	// GetName returns the provider name (e.g., "NWS", "Met.no")
	GetName() string

	// GetCurrentWeather retrieves current weather conditions for a location
	GetCurrentWeather(ctx context.Context, lat, lon float64) (*models.Forecast, error)

	// GetForecast retrieves weather forecast for a location
	GetForecast(ctx context.Context, lat, lon float64, days int) ([]*models.Forecast, error)

	// GetAlerts retrieves weather alerts for a location (if supported)
	GetAlerts(ctx context.Context, lat, lon float64) ([]WeatherAlert, error)

	// SupportedRegions returns the geographic regions this provider supports
	SupportedRegions() []string
}

// GeocodeProvider defines the interface for geocoding providers
type GeocodeProvider interface {
	// GetName returns the provider name (e.g., "Census", "Nominatim")
	GetName() string

	// GeocodeAddress converts an address string to coordinates and place info
	GeocodeAddress(ctx context.Context, address string) ([]*models.Place, error)

	// ReverseGeocode converts coordinates to address information
	ReverseGeocode(ctx context.Context, lat, lon float64) (*models.Place, error)

	// SupportedRegions returns the geographic regions this provider supports
	SupportedRegions() []string
}

// WeatherAlert represents a weather alert/warning
type WeatherAlert struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"` // Minor, Moderate, Severe, Extreme
	Urgency     string    `json:"urgency"`  // Immediate, Expected, Future
	Category    string    `json:"category"` // Geo, Met, Safety, Security, Rescue, Fire, Health, Env, Transport, Infra, CBRNE, Other
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Areas       []string  `json:"areas"` // Affected geographic areas
}

// ProviderResponse wraps provider responses with metadata
type ProviderResponse struct {
	Provider  string        `json:"provider"`
	Timestamp time.Time     `json:"timestamp"`
	Data      interface{}   `json:"data"`
	Error     error         `json:"error,omitempty"`
	Cached    bool          `json:"cached"`
	TTL       time.Duration `json:"ttl,omitempty"`
}

// ProviderManager manages multiple providers
type ProviderManager struct {
	weatherProviders []WeatherProvider
	geocodeProviders []GeocodeProvider
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		weatherProviders: make([]WeatherProvider, 0),
		geocodeProviders: make([]GeocodeProvider, 0),
	}
}

// RegisterWeatherProvider adds a weather provider
func (pm *ProviderManager) RegisterWeatherProvider(provider WeatherProvider) {
	pm.weatherProviders = append(pm.weatherProviders, provider)
}

// RegisterGeocodeProvider adds a geocode provider
func (pm *ProviderManager) RegisterGeocodeProvider(provider GeocodeProvider) {
	pm.geocodeProviders = append(pm.geocodeProviders, provider)
}

// GetWeatherProviders returns all registered weather providers
func (pm *ProviderManager) GetWeatherProviders() []WeatherProvider {
	return pm.weatherProviders
}

// GetGeocodeProviders returns all registered geocode providers
func (pm *ProviderManager) GetGeocodeProviders() []GeocodeProvider {
	return pm.geocodeProviders
}

// GetWeatherProviderByName returns a weather provider by name
func (pm *ProviderManager) GetWeatherProviderByName(name string) WeatherProvider {
	for _, provider := range pm.weatherProviders {
		if provider.GetName() == name {
			return provider
		}
	}
	return nil
}

// GetGeocodeProviderByName returns a geocode provider by name
func (pm *ProviderManager) GetGeocodeProviderByName(name string) GeocodeProvider {
	for _, provider := range pm.geocodeProviders {
		if provider.GetName() == name {
			return provider
		}
	}
	return nil
}
