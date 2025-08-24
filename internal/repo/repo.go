package repo

import (
	"context"
	"database/sql"
)

// Repository defines the common interface for all data repositories
type Repository[T any] interface {
	// Create inserts a new record and returns the created entity with populated ID
	Create(ctx context.Context, entity *T) error

	// GetByID retrieves a single record by its ID
	GetByID(ctx context.Context, id int) (*T, error)

	// Update modifies an existing record
	Update(ctx context.Context, entity *T) error

	// Delete removes a record by its ID
	Delete(ctx context.Context, id int) error

	// List retrieves records with pagination support
	List(ctx context.Context, limit, offset int) ([]*T, error)

	// Count returns the total number of records
	Count(ctx context.Context) (int, error)
}

// ForecastRepository extends the base repository with forecast-specific methods
type ForecastRepository interface {
	Repository[Forecast]

	// GetByCityID retrieves forecasts for a specific city
	GetByCityID(ctx context.Context, cityID int, limit, offset int) ([]*Forecast, error)

	// GetByTimeRange retrieves forecasts within a time range
	GetByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*Forecast, error)

	// GetLatestByCityID retrieves the most recent forecast for a city
	GetLatestByCityID(ctx context.Context, cityID int) (*Forecast, error)

	// DeleteOldForecasts removes forecasts older than the specified number of days
	DeleteOldForecasts(ctx context.Context, days int) error
}

// CityRepository extends the base repository with city-specific methods
type CityRepository interface {
	Repository[City]

	// GetByName retrieves cities by name
	GetByName(ctx context.Context, name string) ([]*City, error)

	// GetByCountry retrieves cities in a specific country
	GetByCountry(ctx context.Context, countryCode string, limit, offset int) ([]*City, error)

	// GetByCoordinates finds cities within a radius of given coordinates
	GetByCoordinates(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*City, error)

	// GetByGeonameID retrieves a city by its GeoNames ID
	GetByGeonameID(ctx context.Context, geonameID int) (*City, error)

	// Search performs text search on city names
	Search(ctx context.Context, query string, limit int) ([]*City, error)
}

// PlaceRepository extends the base repository with place-specific methods
type PlaceRepository interface {
	Repository[Place]

	// GetByCoordinates finds places within a radius of given coordinates
	GetByCoordinates(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*Place, error)

	// Search performs text search on place names and addresses
	Search(ctx context.Context, query string, limit int) ([]*Place, error)

	// GetBySource retrieves places by their geocoding source
	GetBySource(ctx context.Context, source string, limit, offset int) ([]*Place, error)

	// GetBySourcePlaceID retrieves a place by its source-specific ID
	GetBySourcePlaceID(ctx context.Context, source, sourcePlaceID string) (*Place, error)
}

// Forecast represents the forecast model for the repository
type Forecast struct {
	ID             int     `db:"id"`
	CityID         int     `db:"city_id"`
	SourceProvider string  `db:"source_provider"`
	ForecastTime   string  `db:"forecast_time"`
	ValidTime      string  `db:"valid_time"`
	Temperature    float64 `db:"temperature"`
	FeelsLike      float64 `db:"feels_like"`
	Humidity       float64 `db:"humidity"`
	Pressure       float64 `db:"pressure"`
	WindSpeed      float64 `db:"wind_speed"`
	WindDirection  float64 `db:"wind_direction"`
	Visibility     float64 `db:"visibility"`
	CloudCover     float64 `db:"cloud_cover"`
	Precipitation  float64 `db:"precipitation"`
	WeatherCode    string  `db:"weather_code"`
	Description    string  `db:"description"`
	UVIndex        float64 `db:"uv_index"`
	CreatedAt      string  `db:"created_at"`
	UpdatedAt      string  `db:"updated_at"`
}

// City represents the city model for the repository
type City struct {
	ID          int     `db:"id"`
	Name        string  `db:"name"`
	Country     string  `db:"country"`
	CountryCode string  `db:"country_code"`
	Region      string  `db:"region"`
	Latitude    float64 `db:"latitude"`
	Longitude   float64 `db:"longitude"`
	Elevation   float64 `db:"elevation"`
	Population  int     `db:"population"`
	Timezone    string  `db:"timezone"`
	GeonameID   int     `db:"geoname_id"`
	IsCapital   bool    `db:"is_capital"`
	IsActive    bool    `db:"is_active"`
	CreatedAt   string  `db:"created_at"`
	UpdatedAt   string  `db:"updated_at"`
}

// Place represents the place model for the repository
type Place struct {
	ID            int     `db:"id"`
	DisplayName   string  `db:"display_name"`
	AddressLine1  string  `db:"address_line1"`
	AddressLine2  string  `db:"address_line2"`
	City          string  `db:"city"`
	Region        string  `db:"region"`
	PostalCode    string  `db:"postal_code"`
	Country       string  `db:"country"`
	CountryCode   string  `db:"country_code"`
	Latitude      float64 `db:"latitude"`
	Longitude     float64 `db:"longitude"`
	PlaceType     string  `db:"place_type"`
	Confidence    float64 `db:"confidence"`
	Source        string  `db:"source"`
	SourcePlaceID string  `db:"source_place_id"`
	BoundingBox   string  `db:"bounding_box"`
	CreatedAt     string  `db:"created_at"`
	UpdatedAt     string  `db:"updated_at"`
}

// DB interface abstracts database operations
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
