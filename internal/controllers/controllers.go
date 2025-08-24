package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"stormlightlabs.org/weather_api/internal/repo"
)

// HTTPForecastController implements ForecastController for HTTP requests
type HTTPForecastController struct {
	repo repo.ForecastRepository
}

// NewHTTPForecastController creates a new HTTP forecast controller
func NewHTTPForecastController(repo repo.ForecastRepository) ForecastController {
	return &HTTPForecastController{repo: repo}
}

// Create handles POST requests to create a new forecast
func (c *HTTPForecastController) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var forecast Forecast
	if err := json.NewDecoder(r.Body).Decode(&forecast); err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
	}

	repoForecast := toRepoForecast(&forecast)
	if err := c.repo.Create(ctx, repoForecast); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to create forecast", err.Error())
	}

	response := fromRepoForecast(repoForecast)
	return writeSuccess(w, http.StatusCreated, response, "Forecast created successfully")
}

// GetByID handles GET requests to retrieve a forecast by ID
func (c *HTTPForecastController) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	forecast, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return writeError(w, http.StatusNotFound, "Forecast not found", err.Error())
	}

	response := fromRepoForecast(forecast)
	return writeSuccess(w, http.StatusOK, response, "")
}

// Update handles PUT requests to update a forecast
func (c *HTTPForecastController) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	var forecast Forecast
	if err := json.NewDecoder(r.Body).Decode(&forecast); err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
	}

	forecast.ID = id
	repoForecast := toRepoForecast(&forecast)
	if err := c.repo.Update(ctx, repoForecast); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to update forecast", err.Error())
	}

	response := fromRepoForecast(repoForecast)
	return writeSuccess(w, http.StatusOK, response, "Forecast updated successfully")
}

// Delete handles DELETE requests to remove a forecast
func (c *HTTPForecastController) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to delete forecast", err.Error())
	}

	return writeSuccess(w, http.StatusOK, nil, "Forecast deleted successfully")
}

// List handles GET requests to retrieve forecasts with pagination
func (c *HTTPForecastController) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, limit := getPagination(r)
	offset := (page - 1) * limit

	forecasts, err := c.repo.List(ctx, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve forecasts", err.Error())
	}

	total, err := c.repo.Count(ctx)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to count forecasts", err.Error())
	}

	var response []*Forecast
	for _, f := range forecasts {
		response = append(response, fromRepoForecast(f))
	}

	paginated := &PaginatedResponse[Forecast]{
		Data:       response,
		Total:      total,
		Page:       page,
		PerPage:    limit,
		TotalPages: (total + limit - 1) / limit,
	}

	return writePaginated(w, paginated)
}

// GetByCityID handles requests to get forecasts for a specific city
func (c *HTTPForecastController) GetByCityID(ctx context.Context, w http.ResponseWriter, r *http.Request, cityID int) error {
	page, limit := getPagination(r)
	offset := (page - 1) * limit

	forecasts, err := c.repo.GetByCityID(ctx, cityID, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve forecasts", err.Error())
	}

	var response []*Forecast
	for _, f := range forecasts {
		response = append(response, fromRepoForecast(f))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetLatestByCityID handles requests to get the latest forecast for a city
func (c *HTTPForecastController) GetLatestByCityID(ctx context.Context, w http.ResponseWriter, r *http.Request, cityID int) error {
	forecast, err := c.repo.GetLatestByCityID(ctx, cityID)
	if err != nil {
		return writeError(w, http.StatusNotFound, "Latest forecast not found", err.Error())
	}

	response := fromRepoForecast(forecast)
	return writeSuccess(w, http.StatusOK, response, "")
}

// GetByTimeRange handles requests to get forecasts within a time range
func (c *HTTPForecastController) GetByTimeRange(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	startTime := r.URL.Query().Get("start_time")
	endTime := r.URL.Query().Get("end_time")

	if startTime == "" || endTime == "" {
		return writeError(w, http.StatusBadRequest, "Missing parameters", "start_time and end_time are required")
	}

	page, limit := getPagination(r)
	offset := (page - 1) * limit

	forecasts, err := c.repo.GetByTimeRange(ctx, startTime, endTime, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve forecasts", err.Error())
	}

	var response []*Forecast
	for _, f := range forecasts {
		response = append(response, fromRepoForecast(f))
	}

	return writeJSON(w, http.StatusOK, response)
}

// CleanupOldForecasts handles administrative requests to remove old forecasts
func (c *HTTPForecastController) CleanupOldForecasts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30 // Default to 30 days
	}

	if err := c.repo.DeleteOldForecasts(ctx, days); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to cleanup forecasts", err.Error())
	}

	return writeSuccess(w, http.StatusOK, nil, fmt.Sprintf("Cleaned up forecasts older than %d days", days))
}

// HTTPCityController implements CityController for HTTP requests
type HTTPCityController struct {
	repo repo.CityRepository
}

// NewHTTPCityController creates a new HTTP city controller
func NewHTTPCityController(repo repo.CityRepository) CityController {
	return &HTTPCityController{repo: repo}
}

// Create handles POST requests to create a new city
func (c *HTTPCityController) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var city City
	if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
	}

	repoCity := toRepoCity(&city)
	if err := c.repo.Create(ctx, repoCity); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to create city", err.Error())
	}

	response := fromRepoCity(repoCity)
	return writeSuccess(w, http.StatusCreated, response, "City created successfully")
}

// GetByID handles GET requests to retrieve a city by ID
func (c *HTTPCityController) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	city, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return writeError(w, http.StatusNotFound, "City not found", err.Error())
	}

	response := fromRepoCity(city)
	return writeSuccess(w, http.StatusOK, response, "")
}

// Update handles PUT requests to update a city
func (c *HTTPCityController) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	var city City
	if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
	}

	city.ID = id
	repoCity := toRepoCity(&city)
	if err := c.repo.Update(ctx, repoCity); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to update city", err.Error())
	}

	response := fromRepoCity(repoCity)
	return writeSuccess(w, http.StatusOK, response, "City updated successfully")
}

// Delete handles DELETE requests to remove a city
func (c *HTTPCityController) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to delete city", err.Error())
	}

	return writeSuccess(w, http.StatusOK, nil, "City deleted successfully")
}

// List handles GET requests to retrieve cities with pagination
func (c *HTTPCityController) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, limit := getPagination(r)
	offset := (page - 1) * limit

	cities, err := c.repo.List(ctx, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve cities", err.Error())
	}

	total, err := c.repo.Count(ctx)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to count cities", err.Error())
	}

	var response []*City
	for _, city := range cities {
		response = append(response, fromRepoCity(city))
	}

	paginated := &PaginatedResponse[City]{
		Data:       response,
		Total:      total,
		Page:       page,
		PerPage:    limit,
		TotalPages: (total + limit - 1) / limit,
	}

	return writePaginated(w, paginated)
}

// Search handles requests to search cities by name or other criteria
func (c *HTTPCityController) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query().Get("q")
	if query == "" {
		return writeError(w, http.StatusBadRequest, "Missing parameter", "q (query) parameter is required")
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	cities, err := c.repo.Search(ctx, query, limit)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Search failed", err.Error())
	}

	var response []*City
	for _, city := range cities {
		response = append(response, fromRepoCity(city))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetByName handles requests to get cities by name
func (c *HTTPCityController) GetByName(ctx context.Context, w http.ResponseWriter, r *http.Request, name string) error {
	cities, err := c.repo.GetByName(ctx, name)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve cities", err.Error())
	}

	var response []*City
	for _, city := range cities {
		response = append(response, fromRepoCity(city))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetByCountry handles requests to get cities in a specific country
func (c *HTTPCityController) GetByCountry(ctx context.Context, w http.ResponseWriter, r *http.Request, countryCode string) error {
	page, limit := getPagination(r)
	offset := (page - 1) * limit

	cities, err := c.repo.GetByCountry(ctx, countryCode, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve cities", err.Error())
	}

	var response []*City
	for _, city := range cities {
		response = append(response, fromRepoCity(city))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetByCoordinates handles requests to find cities near coordinates
func (c *HTTPCityController) GetByCoordinates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid parameter", "lat must be a valid float")
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid parameter", "lon must be a valid float")
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		radius = 50.0 // Default 50km radius
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	cities, err := c.repo.GetByCoordinates(ctx, lat, lon, radius, limit)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to find cities", err.Error())
	}

	var response []*City
	for _, city := range cities {
		response = append(response, fromRepoCity(city))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetByGeonameID handles requests to get a city by GeoNames ID
func (c *HTTPCityController) GetByGeonameID(ctx context.Context, w http.ResponseWriter, r *http.Request, geonameID int) error {
	city, err := c.repo.GetByGeonameID(ctx, geonameID)
	if err != nil {
		return writeError(w, http.StatusNotFound, "City not found", err.Error())
	}

	response := fromRepoCity(city)
	return writeSuccess(w, http.StatusOK, response, "")
}

// HTTPPlaceController implements PlaceController for HTTP requests
type HTTPPlaceController struct {
	repo repo.PlaceRepository
}

// NewHTTPPlaceController creates a new HTTP place controller
func NewHTTPPlaceController(repo repo.PlaceRepository) PlaceController {
	return &HTTPPlaceController{repo: repo}
}

// Create handles POST requests to create a new place
func (c *HTTPPlaceController) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var place Place
	if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
	}

	repoPlace := toRepoPlace(&place)
	if err := c.repo.Create(ctx, repoPlace); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to create place", err.Error())
	}

	response := fromRepoPlace(repoPlace)
	return writeSuccess(w, http.StatusCreated, response, "Place created successfully")
}

// GetByID handles GET requests to retrieve a place by ID
func (c *HTTPPlaceController) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	place, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return writeError(w, http.StatusNotFound, "Place not found", err.Error())
	}

	response := fromRepoPlace(place)
	return writeSuccess(w, http.StatusOK, response, "")
}

// Update handles PUT requests to update a place
func (c *HTTPPlaceController) Update(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	var place Place
	if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
	}

	place.ID = id
	repoPlace := toRepoPlace(&place)
	if err := c.repo.Update(ctx, repoPlace); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to update place", err.Error())
	}

	response := fromRepoPlace(repoPlace)
	return writeSuccess(w, http.StatusOK, response, "Place updated successfully")
}

// Delete handles DELETE requests to remove a place
func (c *HTTPPlaceController) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request, id int) error {
	if err := c.repo.Delete(ctx, id); err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to delete place", err.Error())
	}

	return writeSuccess(w, http.StatusOK, nil, "Place deleted successfully")
}

// List handles GET requests to retrieve places with pagination
func (c *HTTPPlaceController) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, limit := getPagination(r)
	offset := (page - 1) * limit

	places, err := c.repo.List(ctx, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve places", err.Error())
	}

	total, err := c.repo.Count(ctx)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to count places", err.Error())
	}

	var response []*Place
	for _, place := range places {
		response = append(response, fromRepoPlace(place))
	}

	paginated := &PaginatedResponse[Place]{
		Data:       response,
		Total:      total,
		Page:       page,
		PerPage:    limit,
		TotalPages: (total + limit - 1) / limit,
	}

	return writePaginated(w, paginated)
}

// Search handles requests to search places by address or name
func (c *HTTPPlaceController) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query().Get("q")
	if query == "" {
		return writeError(w, http.StatusBadRequest, "Missing parameter", "q (query) parameter is required")
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	places, err := c.repo.Search(ctx, query, limit)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Search failed", err.Error())
	}

	var response []*Place
	for _, place := range places {
		response = append(response, fromRepoPlace(place))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetByCoordinates handles requests to find places near coordinates
func (c *HTTPPlaceController) GetByCoordinates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	radiusStr := r.URL.Query().Get("radius")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid parameter", "lat must be a valid float")
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return writeError(w, http.StatusBadRequest, "Invalid parameter", "lon must be a valid float")
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		radius = 10.0 // Default 10km radius for places
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	places, err := c.repo.GetByCoordinates(ctx, lat, lon, radius, limit)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to find places", err.Error())
	}

	var response []*Place
	for _, place := range places {
		response = append(response, fromRepoPlace(place))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetBySource handles requests to get places from a specific geocoding source
func (c *HTTPPlaceController) GetBySource(ctx context.Context, w http.ResponseWriter, r *http.Request, source string) error {
	page, limit := getPagination(r)
	offset := (page - 1) * limit

	places, err := c.repo.GetBySource(ctx, source, limit, offset)
	if err != nil {
		return writeError(w, http.StatusInternalServerError, "Failed to retrieve places", err.Error())
	}

	var response []*Place
	for _, place := range places {
		response = append(response, fromRepoPlace(place))
	}

	return writeJSON(w, http.StatusOK, response)
}

// GetBySourcePlaceID handles requests to get a place by its source-specific ID
func (c *HTTPPlaceController) GetBySourcePlaceID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	source := r.URL.Query().Get("source")
	sourcePlaceID := r.URL.Query().Get("source_place_id")

	if source == "" || sourcePlaceID == "" {
		return writeError(w, http.StatusBadRequest, "Missing parameters", "source and source_place_id are required")
	}

	place, err := c.repo.GetBySourcePlaceID(ctx, source, sourcePlaceID)
	if err != nil {
		return writeError(w, http.StatusNotFound, "Place not found", err.Error())
	}

	response := fromRepoPlace(place)
	return writeSuccess(w, http.StatusOK, response, "")
}

// Helper functions for model conversion
func toRepoForecast(f *Forecast) *repo.Forecast {
	return &repo.Forecast{
		ID:             f.ID,
		CityID:         f.CityID,
		SourceProvider: f.SourceProvider,
		ForecastTime:   f.ForecastTime,
		ValidTime:      f.ValidTime,
		Temperature:    f.Temperature,
		FeelsLike:      f.FeelsLike,
		Humidity:       f.Humidity,
		Pressure:       f.Pressure,
		WindSpeed:      f.WindSpeed,
		WindDirection:  f.WindDirection,
		Visibility:     f.Visibility,
		CloudCover:     f.CloudCover,
		Precipitation:  f.Precipitation,
		WeatherCode:    f.WeatherCode,
		Description:    f.Description,
		UVIndex:        f.UVIndex,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}

func fromRepoForecast(f *repo.Forecast) *Forecast {
	return &Forecast{
		ID:             f.ID,
		CityID:         f.CityID,
		SourceProvider: f.SourceProvider,
		ForecastTime:   f.ForecastTime,
		ValidTime:      f.ValidTime,
		Temperature:    f.Temperature,
		FeelsLike:      f.FeelsLike,
		Humidity:       f.Humidity,
		Pressure:       f.Pressure,
		WindSpeed:      f.WindSpeed,
		WindDirection:  f.WindDirection,
		Visibility:     f.Visibility,
		CloudCover:     f.CloudCover,
		Precipitation:  f.Precipitation,
		WeatherCode:    f.WeatherCode,
		Description:    f.Description,
		UVIndex:        f.UVIndex,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}

func toRepoCity(c *City) *repo.City {
	return &repo.City{
		ID:          c.ID,
		Name:        c.Name,
		Country:     c.Country,
		CountryCode: c.CountryCode,
		Region:      c.Region,
		Latitude:    c.Latitude,
		Longitude:   c.Longitude,
		Elevation:   c.Elevation,
		Population:  c.Population,
		Timezone:    c.Timezone,
		GeonameID:   c.GeonameID,
		IsCapital:   c.IsCapital,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func fromRepoCity(c *repo.City) *City {
	return &City{
		ID:          c.ID,
		Name:        c.Name,
		Country:     c.Country,
		CountryCode: c.CountryCode,
		Region:      c.Region,
		Latitude:    c.Latitude,
		Longitude:   c.Longitude,
		Elevation:   c.Elevation,
		Population:  c.Population,
		Timezone:    c.Timezone,
		GeonameID:   c.GeonameID,
		IsCapital:   c.IsCapital,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toRepoPlace(p *Place) *repo.Place {
	return &repo.Place{
		ID:            p.ID,
		DisplayName:   p.DisplayName,
		AddressLine1:  p.AddressLine1,
		AddressLine2:  p.AddressLine2,
		City:          p.City,
		Region:        p.Region,
		PostalCode:    p.PostalCode,
		Country:       p.Country,
		CountryCode:   p.CountryCode,
		Latitude:      p.Latitude,
		Longitude:     p.Longitude,
		PlaceType:     p.PlaceType,
		Confidence:    p.Confidence,
		Source:        p.Source,
		SourcePlaceID: p.SourcePlaceID,
		BoundingBox:   p.BoundingBox,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func fromRepoPlace(p *repo.Place) *Place {
	return &Place{
		ID:            p.ID,
		DisplayName:   p.DisplayName,
		AddressLine1:  p.AddressLine1,
		AddressLine2:  p.AddressLine2,
		City:          p.City,
		Region:        p.Region,
		PostalCode:    p.PostalCode,
		Country:       p.Country,
		CountryCode:   p.CountryCode,
		Latitude:      p.Latitude,
		Longitude:     p.Longitude,
		PlaceType:     p.PlaceType,
		Confidence:    p.Confidence,
		Source:        p.Source,
		SourcePlaceID: p.SourcePlaceID,
		BoundingBox:   p.BoundingBox,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// HTTP response helper functions
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message, details string) error {
	err := &HTTPError{
		Status:  status,
		Message: message,
		Details: details,
	}
	return writeJSON(w, status, err)
}

func writeSuccess(w http.ResponseWriter, status int, data any, message string) error {
	response := map[string]any{
		"success": true,
		"data":    data,
		"message": message,
	}
	return writeJSON(w, status, response)
}

func writePaginated(w http.ResponseWriter, data any) error {
	return writeJSON(w, http.StatusOK, data)
}

func getPagination(r *http.Request) (page, limit int) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20 // Default limit with max of 100
	}

	return page, limit
}
