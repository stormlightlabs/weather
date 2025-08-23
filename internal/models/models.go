package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Model represents the base interface for all data models
type Model interface {
	Validate() error
	TableName() string
}

// Forecast represents weather forecast data from various sources
type Forecast struct {
	ID              int       `json:"id" db:"id"`
	CityID          int       `json:"city_id" db:"city_id"`
	SourceProvider  string    `json:"source_provider" db:"source_provider"` // NOAA, Met.no, etc.
	ForecastTime    time.Time `json:"forecast_time" db:"forecast_time"`
	ValidTime       time.Time `json:"valid_time" db:"valid_time"`
	Temperature     float64   `json:"temperature" db:"temperature"`         // Celsius
	FeelsLike       float64   `json:"feels_like" db:"feels_like"`           // Celsius
	Humidity        float64   `json:"humidity" db:"humidity"`               // Percentage
	Pressure        float64   `json:"pressure" db:"pressure"`               // hPa
	WindSpeed       float64   `json:"wind_speed" db:"wind_speed"`           // m/s
	WindDirection   float64   `json:"wind_direction" db:"wind_direction"`   // degrees
	Visibility      float64   `json:"visibility" db:"visibility"`           // km
	CloudCover      float64   `json:"cloud_cover" db:"cloud_cover"`         // percentage
	Precipitation   float64   `json:"precipitation" db:"precipitation"`     // mm
	WeatherCode     string    `json:"weather_code" db:"weather_code"`       // provider-specific
	Description     string    `json:"description" db:"description"`
	UVIndex         float64   `json:"uv_index" db:"uv_index"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// User represents an authenticated user
type User struct {
	ID                int       `json:"id" db:"id"`
	GitHubID          int       `json:"github_id" db:"github_id"`
	Username          string    `json:"username" db:"username"`
	Email             string    `json:"email" db:"email"`
	AvatarURL         string    `json:"avatar_url" db:"avatar_url"`
	PreferredUnits    string    `json:"preferred_units" db:"preferred_units"` // metric, imperial
	PreferredLanguage string    `json:"preferred_language" db:"preferred_language"`
	DefaultCityID     *int      `json:"default_city_id" db:"default_city_id"`
	APIKeyHash        string    `json:"-" db:"api_key_hash"` // hashed API key for programmatic access
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	LastLoginAt       time.Time `json:"last_login_at" db:"last_login_at"`
}

// City represents a city with weather data
type City struct {
	ID             int     `json:"id" db:"id"`
	Name           string  `json:"name" db:"name"`
	Country        string  `json:"country" db:"country"`
	CountryCode    string  `json:"country_code" db:"country_code"` // ISO 3166-1 alpha-2
	Region         string  `json:"region" db:"region"`             // state/province
	Latitude       float64 `json:"latitude" db:"latitude"`
	Longitude      float64 `json:"longitude" db:"longitude"`
	Elevation      float64 `json:"elevation" db:"elevation"`     // meters above sea level
	Population     int     `json:"population" db:"population"`
	Timezone       string  `json:"timezone" db:"timezone"`       // IANA timezone
	GeonameID      int     `json:"geoname_id" db:"geoname_id"`   // GeoNames.org ID
	IsCapital      bool    `json:"is_capital" db:"is_capital"`
	IsActive       bool    `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Place represents a geocoded location for address/place lookups
type Place struct {
	ID             int     `json:"id" db:"id"`
	DisplayName    string  `json:"display_name" db:"display_name"`
	AddressLine1   string  `json:"address_line1" db:"address_line1"`
	AddressLine2   string  `json:"address_line2" db:"address_line2"`
	City           string  `json:"city" db:"city"`
	Region         string  `json:"region" db:"region"`
	PostalCode     string  `json:"postal_code" db:"postal_code"`
	Country        string  `json:"country" db:"country"`
	CountryCode    string  `json:"country_code" db:"country_code"`
	Latitude       float64 `json:"latitude" db:"latitude"`
	Longitude      float64 `json:"longitude" db:"longitude"`
	PlaceType      string  `json:"place_type" db:"place_type"`     // house, building, city, etc.
	Confidence     float64 `json:"confidence" db:"confidence"`     // geocoding confidence 0-1
	Source         string  `json:"source" db:"source"`             // Nominatim, Census, etc.
	SourcePlaceID  string  `json:"source_place_id" db:"source_place_id"`
	BoundingBox    string  `json:"bounding_box" db:"bounding_box"` // JSON array of coordinates
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Forecast Model interface implementation
func (f *Forecast) Validate() error {
	if f.CityID <= 0 {
		return fmt.Errorf("city_id must be positive")
	}
	if f.SourceProvider == "" {
		return fmt.Errorf("source_provider is required")
	}
	if f.ForecastTime.IsZero() {
		return fmt.Errorf("forecast_time is required")
	}
	if f.ValidTime.IsZero() {
		return fmt.Errorf("valid_time is required")
	}
	if f.Temperature < -273.15 { // absolute zero in Celsius
		return fmt.Errorf("temperature cannot be below absolute zero")
	}
	if f.Humidity < 0 || f.Humidity > 100 {
		return fmt.Errorf("humidity must be between 0 and 100")
	}
	if f.Pressure < 0 {
		return fmt.Errorf("pressure cannot be negative")
	}
	if f.WindSpeed < 0 {
		return fmt.Errorf("wind_speed cannot be negative")
	}
	if f.WindDirection < 0 || f.WindDirection >= 360 {
		return fmt.Errorf("wind_direction must be between 0 and 359 degrees")
	}
	if f.CloudCover < 0 || f.CloudCover > 100 {
		return fmt.Errorf("cloud_cover must be between 0 and 100")
	}
	if f.Precipitation < 0 {
		return fmt.Errorf("precipitation cannot be negative")
	}
	if f.UVIndex < 0 {
		return fmt.Errorf("uv_index cannot be negative")
	}
	return nil
}

func (f *Forecast) TableName() string {
	return "forecasts"
}

// User Model interface implementation
func (u *User) Validate() error {
	if u.GitHubID <= 0 {
		return fmt.Errorf("github_id must be positive")
	}
	if u.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	// Simple email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return fmt.Errorf("invalid email format")
	}
	if u.PreferredUnits != "" && u.PreferredUnits != "metric" && u.PreferredUnits != "imperial" {
		return fmt.Errorf("preferred_units must be 'metric' or 'imperial'")
	}
	return nil
}

func (u *User) TableName() string {
	return "users"
}

// City Model interface implementation
func (c *City) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}
	if c.Country == "" {
		return fmt.Errorf("country is required")
	}
	if c.CountryCode != "" {
		if len(c.CountryCode) != 2 {
			return fmt.Errorf("country_code must be 2 characters (ISO 3166-1 alpha-2)")
		}
		c.CountryCode = strings.ToUpper(c.CountryCode)
	}
	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if c.Population < 0 {
		return fmt.Errorf("population cannot be negative")
	}
	return nil
}

func (c *City) TableName() string {
	return "cities"
}

// Place Model interface implementation
func (p *Place) Validate() error {
	if p.DisplayName == "" {
		return fmt.Errorf("display_name is required")
	}
	if len(p.DisplayName) > 500 {
		return fmt.Errorf("display_name must be 500 characters or less")
	}
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if p.Confidence < 0 || p.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	if p.CountryCode != "" {
		if len(p.CountryCode) != 2 {
			return fmt.Errorf("country_code must be 2 characters (ISO 3166-1 alpha-2)")
		}
		p.CountryCode = strings.ToUpper(p.CountryCode)
	}
	if p.Source == "" {
		return fmt.Errorf("source is required")
	}
	return nil
}

func (p *Place) TableName() string {
	return "places"
}
