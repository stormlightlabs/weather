package controllers

import (
	"context"
	"net/http"
)

// Controller defines the base interface for all HTTP controllers
type Controller[T any] interface {
	// Create handles POST requests to create a new resource
	Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error

	// GetByID handles GET requests to retrieve a resource by ID
	GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error

	// Update handles PUT/PATCH requests to update a resource
	Update(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error

	// Delete handles DELETE requests to remove a resource
	Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error

	// List handles GET requests to retrieve multiple resources with pagination
	List(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// ForecastController extends the base controller with forecast-specific methods
type ForecastController interface {
	Controller[Forecast]

	// GetByCityID handles requests to get forecasts for a specific city
	GetByCityID(ctx context.Context, w http.ResponseWriter, r *http.Request, cityID int) error

	// GetLatestByCityID handles requests to get the latest forecast for a city
	GetLatestByCityID(ctx context.Context, w http.ResponseWriter, r *http.Request, cityID int) error

	// GetByTimeRange handles requests to get forecasts within a time range
	GetByTimeRange(ctx context.Context, w http.ResponseWriter, r *http.Request) error

	// CleanupOldForecasts handles administrative requests to remove old forecasts
	CleanupOldForecasts(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// CityController extends the base controller with city-specific methods
type CityController interface {
	Controller[City]

	// Search handles requests to search cities by name or other criteria
	Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error

	// GetByName handles requests to get cities by name
	GetByName(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) error

	// GetByCountry handles requests to get cities in a specific country
	GetByCountry(ctx context.Context, w http.ResponseWriter, r *http.Request, countryCode string) error

	// GetByCoordinates handles requests to find cities near coordinates
	GetByCoordinates(ctx context.Context, w http.ResponseWriter, r *http.Request) error

	// GetByGeonameID handles requests to get a city by GeoNames ID
	GetByGeonameID(ctx context.Context, w http.ResponseWriter, r *http.Request, geonameID int) error
}

// PlaceController extends the base controller with place-specific methods
type PlaceController interface {
	Controller[Place]

	// Search handles requests to search places by address or name
	Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error

	// GetByCoordinates handles requests to find places near coordinates
	GetByCoordinates(ctx context.Context, w http.ResponseWriter, r *http.Request) error

	// GetBySource handles requests to get places from a specific geocoding source
	GetBySource(ctx context.Context, w http.ResponseWriter, r *http.Request, source string) error

	// GetBySourcePlaceID handles requests to get a place by its source-specific ID
	GetBySourcePlaceID(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// Forecast represents the forecast model for controllers
type Forecast struct {
	ID             int     `json:"id"`
	CityID         int     `json:"city_id"`
	SourceProvider string  `json:"source_provider"`
	ForecastTime   string  `json:"forecast_time"`
	ValidTime      string  `json:"valid_time"`
	Temperature    float64 `json:"temperature"`
	FeelsLike      float64 `json:"feels_like"`
	Humidity       float64 `json:"humidity"`
	Pressure       float64 `json:"pressure"`
	WindSpeed      float64 `json:"wind_speed"`
	WindDirection  float64 `json:"wind_direction"`
	Visibility     float64 `json:"visibility"`
	CloudCover     float64 `json:"cloud_cover"`
	Precipitation  float64 `json:"precipitation"`
	WeatherCode    string  `json:"weather_code"`
	Description    string  `json:"description"`
	UVIndex        float64 `json:"uv_index"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// City represents the city model for controllers
type City struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Elevation   float64 `json:"elevation"`
	Population  int     `json:"population"`
	Timezone    string  `json:"timezone"`
	GeonameID   int     `json:"geoname_id"`
	IsCapital   bool    `json:"is_capital"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// Place represents the place model for controllers
type Place struct {
	ID            int     `json:"id"`
	DisplayName   string  `json:"display_name"`
	AddressLine1  string  `json:"address_line1"`
	AddressLine2  string  `json:"address_line2"`
	City          string  `json:"city"`
	Region        string  `json:"region"`
	PostalCode    string  `json:"postal_code"`
	Country       string  `json:"country"`
	CountryCode   string  `json:"country_code"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	PlaceType     string  `json:"place_type"`
	Confidence    float64 `json:"confidence"`
	Source        string  `json:"source"`
	SourcePlaceID string  `json:"source_place_id"`
	BoundingBox   string  `json:"bounding_box"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// HTTPError represents a structured HTTP error response
type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response structure
type PaginatedResponse[T any] struct {
	Data       []*T `json:"data"`
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	TotalPages int  `json:"total_pages"`
}

// SuccessResponse represents a standard success response
type SuccessResponse[T any] struct {
	Success bool   `json:"success"`
	Data    *T     `json:"data"`
	Message string `json:"message,omitempty"`
}
