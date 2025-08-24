package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PostgreSQLForecastRepository implements ForecastRepository for PostgreSQL
type PostgreSQLForecastRepository struct {
	db DB
}

// NewPostgreSQLForecastRepository creates a new PostgreSQL forecast repository
func NewPostgreSQLForecastRepository(db DB) ForecastRepository {
	return &PostgreSQLForecastRepository{db: db}
}

// Create inserts a new forecast record
func (r *PostgreSQLForecastRepository) Create(ctx context.Context, forecast *Forecast) error {
	query := `
		INSERT INTO forecasts (
			city_id, source_provider, forecast_time, valid_time, temperature,
			feels_like, humidity, pressure, wind_speed, wind_direction,
			visibility, cloud_cover, precipitation, weather_code, description,
			uv_index, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id`

	now := time.Now().UTC().Format(time.RFC3339)
	err := r.db.QueryRowContext(ctx, query,
		forecast.CityID, forecast.SourceProvider, forecast.ForecastTime, forecast.ValidTime,
		forecast.Temperature, forecast.FeelsLike, forecast.Humidity, forecast.Pressure,
		forecast.WindSpeed, forecast.WindDirection, forecast.Visibility, forecast.CloudCover,
		forecast.Precipitation, forecast.WeatherCode, forecast.Description, forecast.UVIndex,
		now, now,
	).Scan(&forecast.ID)

	if err != nil {
		return fmt.Errorf("failed to create forecast: %w", err)
	}

	forecast.CreatedAt = now
	forecast.UpdatedAt = now
	return nil
}

// GetByID retrieves a forecast by its ID
func (r *PostgreSQLForecastRepository) GetByID(ctx context.Context, id int) (*Forecast, error) {
	query := `
		SELECT id, city_id, source_provider, forecast_time, valid_time, temperature,
			   feels_like, humidity, pressure, wind_speed, wind_direction, visibility,
			   cloud_cover, precipitation, weather_code, description, uv_index,
			   created_at, updated_at
		FROM forecasts WHERE id = $1`

	forecast := &Forecast{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&forecast.ID, &forecast.CityID, &forecast.SourceProvider, &forecast.ForecastTime,
		&forecast.ValidTime, &forecast.Temperature, &forecast.FeelsLike, &forecast.Humidity,
		&forecast.Pressure, &forecast.WindSpeed, &forecast.WindDirection, &forecast.Visibility,
		&forecast.CloudCover, &forecast.Precipitation, &forecast.WeatherCode, &forecast.Description,
		&forecast.UVIndex, &forecast.CreatedAt, &forecast.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("forecast with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	return forecast, nil
}

// Update modifies an existing forecast record
func (r *PostgreSQLForecastRepository) Update(ctx context.Context, forecast *Forecast) error {
	query := `
		UPDATE forecasts SET
			city_id = $2, source_provider = $3, forecast_time = $4, valid_time = $5,
			temperature = $6, feels_like = $7, humidity = $8, pressure = $9,
			wind_speed = $10, wind_direction = $11, visibility = $12, cloud_cover = $13,
			precipitation = $14, weather_code = $15, description = $16, uv_index = $17,
			updated_at = $18
		WHERE id = $1`

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, query,
		forecast.ID, forecast.CityID, forecast.SourceProvider, forecast.ForecastTime,
		forecast.ValidTime, forecast.Temperature, forecast.FeelsLike, forecast.Humidity,
		forecast.Pressure, forecast.WindSpeed, forecast.WindDirection, forecast.Visibility,
		forecast.CloudCover, forecast.Precipitation, forecast.WeatherCode, forecast.Description,
		forecast.UVIndex, now,
	)

	if err != nil {
		return fmt.Errorf("failed to update forecast: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("forecast with id %d not found", forecast.ID)
	}

	forecast.UpdatedAt = now
	return nil
}

// Delete removes a forecast record by its ID
func (r *PostgreSQLForecastRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM forecasts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete forecast: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("forecast with id %d not found", id)
	}

	return nil
}

// List retrieves forecasts with pagination
func (r *PostgreSQLForecastRepository) List(ctx context.Context, limit, offset int) ([]*Forecast, error) {
	query := `
		SELECT id, city_id, source_provider, forecast_time, valid_time, temperature,
			   feels_like, humidity, pressure, wind_speed, wind_direction, visibility,
			   cloud_cover, precipitation, weather_code, description, uv_index,
			   created_at, updated_at
		FROM forecasts ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list forecasts: %w", err)
	}
	defer rows.Close()

	var forecasts []*Forecast
	for rows.Next() {
		forecast := &Forecast{}
		err := rows.Scan(
			&forecast.ID, &forecast.CityID, &forecast.SourceProvider, &forecast.ForecastTime,
			&forecast.ValidTime, &forecast.Temperature, &forecast.FeelsLike, &forecast.Humidity,
			&forecast.Pressure, &forecast.WindSpeed, &forecast.WindDirection, &forecast.Visibility,
			&forecast.CloudCover, &forecast.Precipitation, &forecast.WeatherCode, &forecast.Description,
			&forecast.UVIndex, &forecast.CreatedAt, &forecast.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan forecast: %w", err)
		}
		forecasts = append(forecasts, forecast)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating forecasts: %w", err)
	}

	return forecasts, nil
}

// Count returns the total number of forecast records
func (r *PostgreSQLForecastRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM forecasts`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count forecasts: %w", err)
	}
	return count, nil
}

// GetByCityID retrieves forecasts for a specific city
func (r *PostgreSQLForecastRepository) GetByCityID(ctx context.Context, cityID int, limit, offset int) ([]*Forecast, error) {
	query := `
		SELECT id, city_id, source_provider, forecast_time, valid_time, temperature,
			   feels_like, humidity, pressure, wind_speed, wind_direction, visibility,
			   cloud_cover, precipitation, weather_code, description, uv_index,
			   created_at, updated_at
		FROM forecasts WHERE city_id = $1 ORDER BY valid_time DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, cityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecasts by city: %w", err)
	}
	defer rows.Close()

	var forecasts []*Forecast
	for rows.Next() {
		forecast := &Forecast{}
		err := rows.Scan(
			&forecast.ID, &forecast.CityID, &forecast.SourceProvider, &forecast.ForecastTime,
			&forecast.ValidTime, &forecast.Temperature, &forecast.FeelsLike, &forecast.Humidity,
			&forecast.Pressure, &forecast.WindSpeed, &forecast.WindDirection, &forecast.Visibility,
			&forecast.CloudCover, &forecast.Precipitation, &forecast.WeatherCode, &forecast.Description,
			&forecast.UVIndex, &forecast.CreatedAt, &forecast.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan forecast: %w", err)
		}
		forecasts = append(forecasts, forecast)
	}

	return forecasts, rows.Err()
}

// GetByTimeRange retrieves forecasts within a time range
func (r *PostgreSQLForecastRepository) GetByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*Forecast, error) {
	query := `
		SELECT id, city_id, source_provider, forecast_time, valid_time, temperature,
			   feels_like, humidity, pressure, wind_speed, wind_direction, visibility,
			   cloud_cover, precipitation, weather_code, description, uv_index,
			   created_at, updated_at
		FROM forecasts
		WHERE valid_time >= $1 AND valid_time <= $2
		ORDER BY valid_time ASC LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecasts by time range: %w", err)
	}
	defer rows.Close()

	var forecasts []*Forecast
	for rows.Next() {
		forecast := &Forecast{}
		err := rows.Scan(
			&forecast.ID, &forecast.CityID, &forecast.SourceProvider, &forecast.ForecastTime,
			&forecast.ValidTime, &forecast.Temperature, &forecast.FeelsLike, &forecast.Humidity,
			&forecast.Pressure, &forecast.WindSpeed, &forecast.WindDirection, &forecast.Visibility,
			&forecast.CloudCover, &forecast.Precipitation, &forecast.WeatherCode, &forecast.Description,
			&forecast.UVIndex, &forecast.CreatedAt, &forecast.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan forecast: %w", err)
		}
		forecasts = append(forecasts, forecast)
	}

	return forecasts, rows.Err()
}

// GetLatestByCityID retrieves the most recent forecast for a city
func (r *PostgreSQLForecastRepository) GetLatestByCityID(ctx context.Context, cityID int) (*Forecast, error) {
	query := `
		SELECT id, city_id, source_provider, forecast_time, valid_time, temperature,
			   feels_like, humidity, pressure, wind_speed, wind_direction, visibility,
			   cloud_cover, precipitation, weather_code, description, uv_index,
			   created_at, updated_at
		FROM forecasts WHERE city_id = $1 ORDER BY valid_time DESC LIMIT 1`

	forecast := &Forecast{}
	err := r.db.QueryRowContext(ctx, query, cityID).Scan(
		&forecast.ID, &forecast.CityID, &forecast.SourceProvider, &forecast.ForecastTime,
		&forecast.ValidTime, &forecast.Temperature, &forecast.FeelsLike, &forecast.Humidity,
		&forecast.Pressure, &forecast.WindSpeed, &forecast.WindDirection, &forecast.Visibility,
		&forecast.CloudCover, &forecast.Precipitation, &forecast.WeatherCode, &forecast.Description,
		&forecast.UVIndex, &forecast.CreatedAt, &forecast.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no forecasts found for city %d", cityID)
		}
		return nil, fmt.Errorf("failed to get latest forecast: %w", err)
	}

	return forecast, nil
}

// DeleteOldForecasts removes forecasts older than the specified number of days
func (r *PostgreSQLForecastRepository) DeleteOldForecasts(ctx context.Context, days int) error {
	query := `DELETE FROM forecasts WHERE valid_time < NOW() - INTERVAL '%d days'`
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return fmt.Errorf("failed to delete old forecasts: %w", err)
	}
	return nil
}

// PostgreSQLCityRepository implements CityRepository for PostgreSQL
type PostgreSQLCityRepository struct {
	db DB
}

// NewPostgreSQLCityRepository creates a new PostgreSQL city repository
func NewPostgreSQLCityRepository(db DB) CityRepository {
	return &PostgreSQLCityRepository{db: db}
}

// Create inserts a new city record
func (r *PostgreSQLCityRepository) Create(ctx context.Context, city *City) error {
	query := `
		INSERT INTO cities (
			name, country, country_code, region, latitude, longitude,
			elevation, population, timezone, geoname_id, is_capital,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id`

	now := time.Now().UTC().Format(time.RFC3339)
	err := r.db.QueryRowContext(ctx, query,
		city.Name, city.Country, city.CountryCode, city.Region,
		city.Latitude, city.Longitude, city.Elevation, city.Population,
		city.Timezone, city.GeonameID, city.IsCapital, city.IsActive,
		now, now,
	).Scan(&city.ID)

	if err != nil {
		return fmt.Errorf("failed to create city: %w", err)
	}

	city.CreatedAt = now
	city.UpdatedAt = now
	return nil
}

// GetByID retrieves a city by its ID
func (r *PostgreSQLCityRepository) GetByID(ctx context.Context, id int) (*City, error) {
	query := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at
		FROM cities WHERE id = $1`

	city := &City{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
		&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
		&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
		&city.CreatedAt, &city.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("city with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get city: %w", err)
	}

	return city, nil
}

// Update modifies an existing city record
func (r *PostgreSQLCityRepository) Update(ctx context.Context, city *City) error {
	query := `
		UPDATE cities SET
			name = $2, country = $3, country_code = $4, region = $5,
			latitude = $6, longitude = $7, elevation = $8, population = $9,
			timezone = $10, geoname_id = $11, is_capital = $12, is_active = $13,
			updated_at = $14
		WHERE id = $1`

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, query,
		city.ID, city.Name, city.Country, city.CountryCode, city.Region,
		city.Latitude, city.Longitude, city.Elevation, city.Population,
		city.Timezone, city.GeonameID, city.IsCapital, city.IsActive, now,
	)

	if err != nil {
		return fmt.Errorf("failed to update city: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("city with id %d not found", city.ID)
	}

	city.UpdatedAt = now
	return nil
}

// Delete removes a city record by its ID
func (r *PostgreSQLCityRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM cities WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete city: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("city with id %d not found", id)
	}

	return nil
}

// List retrieves cities with pagination
func (r *PostgreSQLCityRepository) List(ctx context.Context, limit, offset int) ([]*City, error) {
	query := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at
		FROM cities ORDER BY name ASC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list cities: %w", err)
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		city := &City{}
		err := rows.Scan(
			&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
			&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
			&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
			&city.CreatedAt, &city.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, rows.Err()
}

// Count returns the total number of city records
func (r *PostgreSQLCityRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM cities`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cities: %w", err)
	}
	return count, nil
}

// GetByName retrieves cities by name
func (r *PostgreSQLCityRepository) GetByName(ctx context.Context, name string) ([]*City, error) {
	query := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at
		FROM cities WHERE LOWER(name) = LOWER($1) ORDER BY population DESC`

	rows, err := r.db.QueryContext(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities by name: %w", err)
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		city := &City{}
		err := rows.Scan(
			&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
			&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
			&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
			&city.CreatedAt, &city.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, rows.Err()
}

// GetByCountry retrieves cities in a specific country
func (r *PostgreSQLCityRepository) GetByCountry(ctx context.Context, countryCode string, limit, offset int) ([]*City, error) {
	query := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at
		FROM cities WHERE country_code = $1 ORDER BY population DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, countryCode, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities by country: %w", err)
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		city := &City{}
		err := rows.Scan(
			&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
			&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
			&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
			&city.CreatedAt, &city.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, rows.Err()
}

// GetByCoordinates finds cities within a radius of given coordinates
//
//	Uses the haversine formula to calculate distance
func (r *PostgreSQLCityRepository) GetByCoordinates(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*City, error) {
	query := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at,
			   (6371 * acos(cos(radians($1)) * cos(radians(latitude)) *
			   cos(radians(longitude) - radians($2)) + sin(radians($1)) *
			   sin(radians(latitude)))) AS distance
		FROM cities
		WHERE (6371 * acos(cos(radians($1)) * cos(radians(latitude)) *
			  cos(radians(longitude) - radians($2)) + sin(radians($1)) *
			  sin(radians(latitude)))) <= $3
		ORDER BY distance ASC LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, lat, lon, radiusKm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities by coordinates: %w", err)
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		city := &City{}
		var distance float64
		err := rows.Scan(
			&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
			&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
			&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
			&city.CreatedAt, &city.UpdatedAt, &distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, rows.Err()
}

// GetByGeonameID retrieves a city by its GeoNames ID
func (r *PostgreSQLCityRepository) GetByGeonameID(ctx context.Context, geonameID int) (*City, error) {
	query := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at
		FROM cities WHERE geoname_id = $1`

	city := &City{}
	err := r.db.QueryRowContext(ctx, query, geonameID).Scan(
		&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
		&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
		&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
		&city.CreatedAt, &city.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("city with geoname_id %d not found", geonameID)
		}
		return nil, fmt.Errorf("failed to get city by geoname_id: %w", err)
	}

	return city, nil
}

// Search performs text search on city names
func (r *PostgreSQLCityRepository) Search(ctx context.Context, query string, limit int) ([]*City, error) {
	searchQuery := `
		SELECT id, name, country, country_code, region, latitude, longitude,
			   elevation, population, timezone, geoname_id, is_capital,
			   is_active, created_at, updated_at
		FROM cities
		WHERE LOWER(name) LIKE LOWER($1) OR LOWER(country) LIKE LOWER($1)
		ORDER BY population DESC LIMIT $2`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search cities: %w", err)
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		city := &City{}
		err := rows.Scan(
			&city.ID, &city.Name, &city.Country, &city.CountryCode, &city.Region,
			&city.Latitude, &city.Longitude, &city.Elevation, &city.Population,
			&city.Timezone, &city.GeonameID, &city.IsCapital, &city.IsActive,
			&city.CreatedAt, &city.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, rows.Err()
}

// PostgreSQLPlaceRepository implements PlaceRepository for PostgreSQL
type PostgreSQLPlaceRepository struct {
	db DB
}

// NewPostgreSQLPlaceRepository creates a new PostgreSQL place repository
func NewPostgreSQLPlaceRepository(db DB) PlaceRepository {
	return &PostgreSQLPlaceRepository{db: db}
}

// Create inserts a new place record
func (r *PostgreSQLPlaceRepository) Create(ctx context.Context, place *Place) error {
	query := `
		INSERT INTO places (
			display_name, address_line1, address_line2, city, region,
			postal_code, country, country_code, latitude, longitude,
			place_type, confidence, source, source_place_id, bounding_box,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id`

	now := time.Now().UTC().Format(time.RFC3339)
	err := r.db.QueryRowContext(ctx, query,
		place.DisplayName, place.AddressLine1, place.AddressLine2, place.City,
		place.Region, place.PostalCode, place.Country, place.CountryCode,
		place.Latitude, place.Longitude, place.PlaceType, place.Confidence,
		place.Source, place.SourcePlaceID, place.BoundingBox, now, now,
	).Scan(&place.ID)

	if err != nil {
		return fmt.Errorf("failed to create place: %w", err)
	}

	place.CreatedAt = now
	place.UpdatedAt = now
	return nil
}

// GetByID retrieves a place by its ID
func (r *PostgreSQLPlaceRepository) GetByID(ctx context.Context, id int) (*Place, error) {
	query := `
		SELECT id, display_name, address_line1, address_line2, city, region,
			   postal_code, country, country_code, latitude, longitude,
			   place_type, confidence, source, source_place_id, bounding_box,
			   created_at, updated_at
		FROM places WHERE id = $1`

	place := &Place{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&place.ID, &place.DisplayName, &place.AddressLine1, &place.AddressLine2,
		&place.City, &place.Region, &place.PostalCode, &place.Country,
		&place.CountryCode, &place.Latitude, &place.Longitude, &place.PlaceType,
		&place.Confidence, &place.Source, &place.SourcePlaceID, &place.BoundingBox,
		&place.CreatedAt, &place.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("place with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get place: %w", err)
	}

	return place, nil
}

// Update modifies an existing place record
func (r *PostgreSQLPlaceRepository) Update(ctx context.Context, place *Place) error {
	query := `
		UPDATE places SET
			display_name = $2, address_line1 = $3, address_line2 = $4, city = $5,
			region = $6, postal_code = $7, country = $8, country_code = $9,
			latitude = $10, longitude = $11, place_type = $12, confidence = $13,
			source = $14, source_place_id = $15, bounding_box = $16, updated_at = $17
		WHERE id = $1`

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(ctx, query,
		place.ID, place.DisplayName, place.AddressLine1, place.AddressLine2,
		place.City, place.Region, place.PostalCode, place.Country,
		place.CountryCode, place.Latitude, place.Longitude, place.PlaceType,
		place.Confidence, place.Source, place.SourcePlaceID, place.BoundingBox, now,
	)

	if err != nil {
		return fmt.Errorf("failed to update place: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("place with id %d not found", place.ID)
	}

	place.UpdatedAt = now
	return nil
}

// Delete removes a place record by its ID
func (r *PostgreSQLPlaceRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM places WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete place: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("place with id %d not found", id)
	}

	return nil
}

// List retrieves places with pagination
func (r *PostgreSQLPlaceRepository) List(ctx context.Context, limit, offset int) ([]*Place, error) {
	query := `
		SELECT id, display_name, address_line1, address_line2, city, region,
			   postal_code, country, country_code, latitude, longitude,
			   place_type, confidence, source, source_place_id, bounding_box,
			   created_at, updated_at
		FROM places ORDER BY confidence DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list places: %w", err)
	}
	defer rows.Close()

	var places []*Place
	for rows.Next() {
		place := &Place{}
		err := rows.Scan(
			&place.ID, &place.DisplayName, &place.AddressLine1, &place.AddressLine2,
			&place.City, &place.Region, &place.PostalCode, &place.Country,
			&place.CountryCode, &place.Latitude, &place.Longitude, &place.PlaceType,
			&place.Confidence, &place.Source, &place.SourcePlaceID, &place.BoundingBox,
			&place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}
		places = append(places, place)
	}

	return places, rows.Err()
}

// Count returns the total number of place records
func (r *PostgreSQLPlaceRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM places`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count places: %w", err)
	}
	return count, nil
}

// GetByCoordinates finds places within a radius of given coordinates
func (r *PostgreSQLPlaceRepository) GetByCoordinates(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*Place, error) {
	query := `
		SELECT id, display_name, address_line1, address_line2, city, region,
			   postal_code, country, country_code, latitude, longitude,
			   place_type, confidence, source, source_place_id, bounding_box,
			   created_at, updated_at,
			   (6371 * acos(cos(radians($1)) * cos(radians(latitude)) *
			   cos(radians(longitude) - radians($2)) + sin(radians($1)) *
			   sin(radians(latitude)))) AS distance
		FROM places
		WHERE (6371 * acos(cos(radians($1)) * cos(radians(latitude)) *
			  cos(radians(longitude) - radians($2)) + sin(radians($1)) *
			  sin(radians(latitude)))) <= $3
		ORDER BY distance ASC LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, lat, lon, radiusKm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get places by coordinates: %w", err)
	}
	defer rows.Close()

	var places []*Place
	for rows.Next() {
		place := &Place{}
		var distance float64
		err := rows.Scan(
			&place.ID, &place.DisplayName, &place.AddressLine1, &place.AddressLine2,
			&place.City, &place.Region, &place.PostalCode, &place.Country,
			&place.CountryCode, &place.Latitude, &place.Longitude, &place.PlaceType,
			&place.Confidence, &place.Source, &place.SourcePlaceID, &place.BoundingBox,
			&place.CreatedAt, &place.UpdatedAt, &distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}
		places = append(places, place)
	}

	return places, rows.Err()
}

// Search performs text search on place names and addresses
func (r *PostgreSQLPlaceRepository) Search(ctx context.Context, query string, limit int) ([]*Place, error) {
	searchQuery := `
		SELECT id, display_name, address_line1, address_line2, city, region,
			   postal_code, country, country_code, latitude, longitude,
			   place_type, confidence, source, source_place_id, bounding_box,
			   created_at, updated_at
		FROM places
		WHERE LOWER(display_name) LIKE LOWER($1)
		   OR LOWER(address_line1) LIKE LOWER($1)
		   OR LOWER(city) LIKE LOWER($1)
		ORDER BY confidence DESC LIMIT $2`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search places: %w", err)
	}
	defer rows.Close()

	var places []*Place
	for rows.Next() {
		place := &Place{}
		err := rows.Scan(
			&place.ID, &place.DisplayName, &place.AddressLine1, &place.AddressLine2,
			&place.City, &place.Region, &place.PostalCode, &place.Country,
			&place.CountryCode, &place.Latitude, &place.Longitude, &place.PlaceType,
			&place.Confidence, &place.Source, &place.SourcePlaceID, &place.BoundingBox,
			&place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}
		places = append(places, place)
	}

	return places, rows.Err()
}

// GetBySource retrieves places by their geocoding source
func (r *PostgreSQLPlaceRepository) GetBySource(ctx context.Context, source string, limit, offset int) ([]*Place, error) {
	query := `
		SELECT id, display_name, address_line1, address_line2, city, region,
			   postal_code, country, country_code, latitude, longitude,
			   place_type, confidence, source, source_place_id, bounding_box,
			   created_at, updated_at
		FROM places WHERE source = $1 ORDER BY confidence DESC LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, source, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get places by source: %w", err)
	}
	defer rows.Close()

	var places []*Place
	for rows.Next() {
		place := &Place{}
		err := rows.Scan(
			&place.ID, &place.DisplayName, &place.AddressLine1, &place.AddressLine2,
			&place.City, &place.Region, &place.PostalCode, &place.Country,
			&place.CountryCode, &place.Latitude, &place.Longitude, &place.PlaceType,
			&place.Confidence, &place.Source, &place.SourcePlaceID, &place.BoundingBox,
			&place.CreatedAt, &place.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan place: %w", err)
		}
		places = append(places, place)
	}

	return places, rows.Err()
}

// GetBySourcePlaceID retrieves a place by its source-specific ID
func (r *PostgreSQLPlaceRepository) GetBySourcePlaceID(ctx context.Context, source, sourcePlaceID string) (*Place, error) {
	query := `
		SELECT id, display_name, address_line1, address_line2, city, region,
			   postal_code, country, country_code, latitude, longitude,
			   place_type, confidence, source, source_place_id, bounding_box,
			   created_at, updated_at
		FROM places WHERE source = $1 AND source_place_id = $2`

	place := &Place{}
	err := r.db.QueryRowContext(ctx, query, source, sourcePlaceID).Scan(
		&place.ID, &place.DisplayName, &place.AddressLine1, &place.AddressLine2,
		&place.City, &place.Region, &place.PostalCode, &place.Country,
		&place.CountryCode, &place.Latitude, &place.Longitude, &place.PlaceType,
		&place.Confidence, &place.Source, &place.SourcePlaceID, &place.BoundingBox,
		&place.CreatedAt, &place.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("place with source %s and source_place_id %s not found", source, sourcePlaceID)
		}
		return nil, fmt.Errorf("failed to get place by source place id: %w", err)
	}

	return place, nil
}
