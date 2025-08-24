package models

import (
	"strings"
	"testing"
	"time"
)

func TestForecastValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		forecast    Forecast
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid forecast",
			forecast: Forecast{
				CityID:         1,
				SourceProvider: "NOAA",
				ForecastTime:   now,
				ValidTime:      now.Add(time.Hour),
				Temperature:    20.0,
				Humidity:       60.0,
				Pressure:       1013.25,
				WindSpeed:      5.0,
				WindDirection:  180.0,
				CloudCover:     50.0,
				Precipitation:  0.0,
				UVIndex:        5.0,
			},
			expectError: false,
		},
		{
			name: "invalid city_id",
			forecast: Forecast{
				CityID:         0,
				SourceProvider: "NOAA",
				ForecastTime:   now,
				ValidTime:      now.Add(time.Hour),
			},
			expectError: true,
			errorMsg:    "city_id must be positive",
		},
		{
			name: "missing source_provider",
			forecast: Forecast{
				CityID:       1,
				ForecastTime: now,
				ValidTime:    now.Add(time.Hour),
			},
			expectError: true,
			errorMsg:    "source_provider is required",
		},
		{
			name: "invalid temperature below absolute zero",
			forecast: Forecast{
				CityID:         1,
				SourceProvider: "NOAA",
				ForecastTime:   now,
				ValidTime:      now.Add(time.Hour),
				Temperature:    -300.0,
			},
			expectError: true,
			errorMsg:    "temperature cannot be below absolute zero",
		},
		{
			name: "invalid humidity over 100",
			forecast: Forecast{
				CityID:         1,
				SourceProvider: "NOAA",
				ForecastTime:   now,
				ValidTime:      now.Add(time.Hour),
				Temperature:    20.0,
				Humidity:       150.0,
			},
			expectError: true,
			errorMsg:    "humidity must be between 0 and 100",
		},
		{
			name: "invalid wind direction",
			forecast: Forecast{
				CityID:         1,
				SourceProvider: "NOAA",
				ForecastTime:   now,
				ValidTime:      now.Add(time.Hour),
				Temperature:    20.0,
				Humidity:       60.0,
				WindDirection:  400.0,
			},
			expectError: true,
			errorMsg:    "wind_direction must be between 0 and 359 degrees",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.forecast.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestForecastTableName(t *testing.T) {
	f := &Forecast{}
	if got := f.TableName(); got != "forecasts" {
		t.Errorf("expected 'forecasts', got '%s'", got)
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name        string
		user        User
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid user",
			user: User{
				GitHubID:          12345,
				Username:          "testuser",
				Email:             "test@example.com",
				PreferredUnits:    "metric",
				PreferredLanguage: "en",
			},
			expectError: false,
		},
		{
			name: "invalid github_id",
			user: User{
				GitHubID: 0,
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectError: true,
			errorMsg:    "github_id must be positive",
		},
		{
			name: "missing username",
			user: User{
				GitHubID: 12345,
				Email:    "test@example.com",
			},
			expectError: true,
			errorMsg:    "username is required",
		},
		{
			name: "username too short",
			user: User{
				GitHubID: 12345,
				Username: "ab",
				Email:    "test@example.com",
			},
			expectError: true,
			errorMsg:    "username must be between 3 and 50 characters",
		},
		{
			name: "username too long",
			user: User{
				GitHubID: 12345,
				Username: "verylongusernamethatexceedsfiftycharacterslimitation",
				Email:    "test@example.com",
			},
			expectError: true,
			errorMsg:    "username must be between 3 and 50 characters",
		},
		{
			name: "missing email",
			user: User{
				GitHubID: 12345,
				Username: "testuser",
			},
			expectError: true,
			errorMsg:    "email is required",
		},
		{
			name: "invalid email format",
			user: User{
				GitHubID: 12345,
				Username: "testuser",
				Email:    "invalid-email",
			},
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name: "invalid preferred_units",
			user: User{
				GitHubID:       12345,
				Username:       "testuser",
				Email:          "test@example.com",
				PreferredUnits: "invalid",
			},
			expectError: true,
			errorMsg:    "preferred_units must be 'metric' or 'imperial'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestUserTableName(t *testing.T) {
	u := &User{}
	if got := u.TableName(); got != "users" {
		t.Errorf("expected 'users', got '%s'", got)
	}
}

func TestCityValidate(t *testing.T) {
	tests := []struct {
		name        string
		city        City
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid city",
			city: City{
				Name:        "New York",
				Country:     "United States",
				CountryCode: "US",
				Region:      "New York",
				Latitude:    40.7128,
				Longitude:   -74.0060,
				Population:  8000000,
				Timezone:    "America/New_York",
			},
			expectError: false,
		},
		{
			name: "missing name",
			city: City{
				Country:   "United States",
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "name too long",
			city: City{
				Name:      strings.Repeat("a", 256), // 256 characters
				Country:   "United States",
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			expectError: true,
			errorMsg:    "name must be 255 characters or less",
		},
		{
			name: "missing country",
			city: City{
				Name:      "New York",
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			expectError: true,
			errorMsg:    "country is required",
		},
		{
			name: "invalid country code length",
			city: City{
				Name:        "New York",
				Country:     "United States",
				CountryCode: "USA",
				Latitude:    40.7128,
				Longitude:   -74.0060,
			},
			expectError: true,
			errorMsg:    "country_code must be 2 characters (ISO 3166-1 alpha-2)",
		},
		{
			name: "invalid latitude",
			city: City{
				Name:      "New York",
				Country:   "United States",
				Latitude:  100.0,
				Longitude: -74.0060,
			},
			expectError: true,
			errorMsg:    "latitude must be between -90 and 90",
		},
		{
			name: "invalid longitude",
			city: City{
				Name:      "New York",
				Country:   "United States",
				Latitude:  40.7128,
				Longitude: 200.0,
			},
			expectError: true,
			errorMsg:    "longitude must be between -180 and 180",
		},
		{
			name: "negative population",
			city: City{
				Name:       "New York",
				Country:    "United States",
				Latitude:   40.7128,
				Longitude:  -74.0060,
				Population: -100,
			},
			expectError: true,
			errorMsg:    "population cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.city.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestCityTableName(t *testing.T) {
	c := &City{}
	if got := c.TableName(); got != "cities" {
		t.Errorf("expected 'cities', got '%s'", got)
	}
}

func TestPlaceValidate(t *testing.T) {
	tests := []struct {
		name        string
		place       Place
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid place",
			place: Place{
				DisplayName:  "123 Main St, New York, NY, USA",
				AddressLine1: "123 Main St",
				City:         "New York",
				Region:       "NY",
				Country:      "United States",
				CountryCode:  "US",
				Latitude:     40.7128,
				Longitude:    -74.0060,
				Confidence:   0.95,
				Source:       "Nominatim",
				PlaceType:    "house",
			},
			expectError: false,
		},
		{
			name: "missing display_name",
			place: Place{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Source:    "Nominatim",
			},
			expectError: true,
			errorMsg:    "display_name is required",
		},
		{
			name: "display_name too long",
			place: Place{
				DisplayName: strings.Repeat("a", 501), // 501 characters
				Latitude:    40.7128,
				Longitude:   -74.0060,
				Source:      "Nominatim",
			},
			expectError: true,
			errorMsg:    "display_name must be 500 characters or less",
		},
		{
			name: "invalid latitude",
			place: Place{
				DisplayName: "123 Main St",
				Latitude:    100.0,
				Longitude:   -74.0060,
				Source:      "Nominatim",
			},
			expectError: true,
			errorMsg:    "latitude must be between -90 and 90",
		},
		{
			name: "invalid longitude",
			place: Place{
				DisplayName: "123 Main St",
				Latitude:    40.7128,
				Longitude:   200.0,
				Source:      "Nominatim",
			},
			expectError: true,
			errorMsg:    "longitude must be between -180 and 180",
		},
		{
			name: "invalid confidence",
			place: Place{
				DisplayName: "123 Main St",
				Latitude:    40.7128,
				Longitude:   -74.0060,
				Confidence:  1.5,
				Source:      "Nominatim",
			},
			expectError: true,
			errorMsg:    "confidence must be between 0 and 1",
		},
		{
			name: "invalid country code length",
			place: Place{
				DisplayName: "123 Main St",
				CountryCode: "USA",
				Latitude:    40.7128,
				Longitude:   -74.0060,
				Source:      "Nominatim",
			},
			expectError: true,
			errorMsg:    "country_code must be 2 characters (ISO 3166-1 alpha-2)",
		},
		{
			name: "missing source",
			place: Place{
				DisplayName: "123 Main St",
				Latitude:    40.7128,
				Longitude:   -74.0060,
			},
			expectError: true,
			errorMsg:    "source is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.place.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPlaceTableName(t *testing.T) {
	p := &Place{}
	if got := p.TableName(); got != "places" {
		t.Errorf("expected 'places', got '%s'", got)
	}
}

func TestModelInterface(t *testing.T) {
	var _ Model = &Forecast{}
	var _ Model = &User{}
	var _ Model = &City{}
	var _ Model = &Place{}
}

func TestCountryCodeNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase us", "us", "US"},
		{"mixed case ca", "Ca", "CA"},
		{"uppercase gb", "GB", "GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test City
			city := City{
				Name:        "Test City",
				Country:     "Test Country",
				CountryCode: tt.input,
				Latitude:    40.0,
				Longitude:   -74.0,
			}
			city.Validate()
			if city.CountryCode != tt.expected {
				t.Errorf("City CountryCode: expected '%s', got '%s'", tt.expected, city.CountryCode)
			}

			// Test Place
			place := Place{
				DisplayName: "Test Place",
				CountryCode: tt.input,
				Latitude:    40.0,
				Longitude:   -74.0,
				Source:      "Test",
			}
			place.Validate()
			if place.CountryCode != tt.expected {
				t.Errorf("Place CountryCode: expected '%s', got '%s'", tt.expected, place.CountryCode)
			}
		})
	}
}
