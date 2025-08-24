package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stormlightlabs.org/weather_api/internal/repo"
)

// MockForecastRepository implements repo.ForecastRepository for testing
type MockForecastRepository struct {
	shouldError bool
	errorMsg    string
	forecasts   []*repo.Forecast
	forecast    *repo.Forecast
	count       int
}

func (m *MockForecastRepository) Create(ctx context.Context, forecast *repo.Forecast) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	forecast.ID = 123
	return nil
}

func (m *MockForecastRepository) GetByID(ctx context.Context, id int) (*repo.Forecast, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.forecast, nil
}

func (m *MockForecastRepository) Update(ctx context.Context, forecast *repo.Forecast) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

func (m *MockForecastRepository) Delete(ctx context.Context, id int) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

func (m *MockForecastRepository) List(ctx context.Context, limit, offset int) ([]*repo.Forecast, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.forecasts, nil
}

func (m *MockForecastRepository) Count(ctx context.Context) (int, error) {
	if m.shouldError {
		return 0, &repoError{msg: m.errorMsg}
	}
	return m.count, nil
}

func (m *MockForecastRepository) GetByCityID(ctx context.Context, cityID int, limit, offset int) ([]*repo.Forecast, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.forecasts, nil
}

func (m *MockForecastRepository) GetByTimeRange(ctx context.Context, startTime, endTime string, limit, offset int) ([]*repo.Forecast, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.forecasts, nil
}

func (m *MockForecastRepository) GetLatestByCityID(ctx context.Context, cityID int) (*repo.Forecast, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.forecast, nil
}

func (m *MockForecastRepository) DeleteOldForecasts(ctx context.Context, days int) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

// MockCityRepository implements repo.CityRepository for testing
type MockCityRepository struct {
	shouldError bool
	errorMsg    string
	cities      []*repo.City
	city        *repo.City
	count       int
}

func (m *MockCityRepository) Create(ctx context.Context, city *repo.City) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	city.ID = 456
	return nil
}

func (m *MockCityRepository) GetByID(ctx context.Context, id int) (*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.city, nil
}

func (m *MockCityRepository) Update(ctx context.Context, city *repo.City) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

func (m *MockCityRepository) Delete(ctx context.Context, id int) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

func (m *MockCityRepository) List(ctx context.Context, limit, offset int) ([]*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.cities, nil
}

func (m *MockCityRepository) Count(ctx context.Context) (int, error) {
	if m.shouldError {
		return 0, &repoError{msg: m.errorMsg}
	}
	return m.count, nil
}

func (m *MockCityRepository) GetByName(ctx context.Context, name string) ([]*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.cities, nil
}

func (m *MockCityRepository) GetByCountry(ctx context.Context, countryCode string, limit, offset int) ([]*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.cities, nil
}

func (m *MockCityRepository) GetByCoordinates(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.cities, nil
}

func (m *MockCityRepository) GetByGeonameID(ctx context.Context, geonameID int) (*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.city, nil
}

func (m *MockCityRepository) Search(ctx context.Context, query string, limit int) ([]*repo.City, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.cities, nil
}

// MockPlaceRepository implements repo.PlaceRepository for testing
type MockPlaceRepository struct {
	shouldError bool
	errorMsg    string
	places      []*repo.Place
	place       *repo.Place
	count       int
}

func (m *MockPlaceRepository) Create(ctx context.Context, place *repo.Place) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	place.ID = 789
	return nil
}

func (m *MockPlaceRepository) GetByID(ctx context.Context, id int) (*repo.Place, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.place, nil
}

func (m *MockPlaceRepository) Update(ctx context.Context, place *repo.Place) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

func (m *MockPlaceRepository) Delete(ctx context.Context, id int) error {
	if m.shouldError {
		return &repoError{msg: m.errorMsg}
	}
	return nil
}

func (m *MockPlaceRepository) List(ctx context.Context, limit, offset int) ([]*repo.Place, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.places, nil
}

func (m *MockPlaceRepository) Count(ctx context.Context) (int, error) {
	if m.shouldError {
		return 0, &repoError{msg: m.errorMsg}
	}
	return m.count, nil
}

func (m *MockPlaceRepository) GetByCoordinates(ctx context.Context, lat, lon, radiusKm float64, limit int) ([]*repo.Place, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.places, nil
}

func (m *MockPlaceRepository) Search(ctx context.Context, query string, limit int) ([]*repo.Place, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.places, nil
}

func (m *MockPlaceRepository) GetBySource(ctx context.Context, source string, limit, offset int) ([]*repo.Place, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.places, nil
}

func (m *MockPlaceRepository) GetBySourcePlaceID(ctx context.Context, source, sourcePlaceID string) (*repo.Place, error) {
	if m.shouldError {
		return nil, &repoError{msg: m.errorMsg}
	}
	return m.place, nil
}

type repoError struct {
	msg string
}

func (e *repoError) Error() string {
	return e.msg
}

func createTestRepoForecast() *repo.Forecast {
	return &repo.Forecast{
		ID:             1,
		CityID:         123,
		SourceProvider: "NOAA",
		ForecastTime:   "2024-01-15T12:00:00Z",
		ValidTime:      "2024-01-15T15:00:00Z",
		Temperature:    20.5,
		Humidity:       65.0,
		Pressure:       1013.25,
		WindSpeed:      5.5,
		WindDirection:  180.0,
		CloudCover:     25.0,
		Precipitation:  0.0,
		WeatherCode:    "partly_cloudy",
		Description:    "Partly cloudy",
		UVIndex:        3.0,
		CreatedAt:      "2024-01-15T12:00:00Z",
		UpdatedAt:      "2024-01-15T12:00:00Z",
	}
}

func createTestControllerForecast() *Forecast {
	return &Forecast{
		CityID:         123,
		SourceProvider: "NOAA",
		ForecastTime:   "2024-01-15T12:00:00Z",
		ValidTime:      "2024-01-15T15:00:00Z",
		Temperature:    20.5,
		Humidity:       65.0,
		Pressure:       1013.25,
		WindSpeed:      5.5,
		WindDirection:  180.0,
		CloudCover:     25.0,
		Precipitation:  0.0,
		WeatherCode:    "partly_cloudy",
		Description:    "Partly cloudy",
		UVIndex:        3.0,
	}
}

func createTestRepoCity() *repo.City {
	return &repo.City{
		ID:          1,
		Name:        "San Francisco",
		Country:     "United States",
		CountryCode: "US",
		Region:      "California",
		Latitude:    37.7749,
		Longitude:   -122.4194,
		Population:  884363,
		Timezone:    "America/Los_Angeles",
		IsActive:    true,
		CreatedAt:   "2024-01-15T12:00:00Z",
		UpdatedAt:   "2024-01-15T12:00:00Z",
	}
}

func createTestRepoPlace() *repo.Place {
	return &repo.Place{
		ID:           1,
		DisplayName:  "Golden Gate Bridge",
		AddressLine1: "Golden Gate Bridge",
		City:         "San Francisco",
		Region:       "California",
		Country:      "United States",
		CountryCode:  "US",
		Latitude:     37.8199,
		Longitude:    -122.4783,
		Confidence:   0.95,
		Source:       "Nominatim",
		CreatedAt:    "2024-01-15T12:00:00Z",
		UpdatedAt:    "2024-01-15T12:00:00Z",
	}
}

func TestControllers(t *testing.T) {
	t.Run("ForecastController", func(t *testing.T) {
		t.Run("interface compliance", func(t *testing.T) {
			mockRepo := &MockForecastRepository{}
			controller := NewHTTPForecastController(mockRepo)

			var _ ForecastController = controller
			var _ Controller[Forecast] = controller
		})

		t.Run("Create success", func(t *testing.T) {
			mockRepo := &MockForecastRepository{}
			controller := NewHTTPForecastController(mockRepo)

			forecast := createTestControllerForecast()
			body, _ := json.Marshal(forecast)

			req := httptest.NewRequest("POST", "/forecasts", bytes.NewReader(body))
			w := httptest.NewRecorder()

			err := controller.Create(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusCreated {
				t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
			}
		})

		t.Run("Create error", func(t *testing.T) {
			mockRepo := &MockForecastRepository{shouldError: true, errorMsg: "database error"}
			controller := NewHTTPForecastController(mockRepo)

			forecast := createTestControllerForecast()
			body, _ := json.Marshal(forecast)

			req := httptest.NewRequest("POST", "/forecasts", bytes.NewReader(body))
			w := httptest.NewRecorder()

			_ = controller.Create(context.Background(), w, req)

			if w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
			}
		})

		t.Run("GetByID success", func(t *testing.T) {
			mockRepo := &MockForecastRepository{forecast: createTestRepoForecast()}
			controller := NewHTTPForecastController(mockRepo)

			req := httptest.NewRequest("GET", "/forecasts/1", nil)
			w := httptest.NewRecorder()

			err := controller.GetByID(context.Background(), w, req, 1)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("List with pagination", func(t *testing.T) {
			forecasts := []*repo.Forecast{createTestRepoForecast()}
			mockRepo := &MockForecastRepository{forecasts: forecasts, count: 1}
			controller := NewHTTPForecastController(mockRepo)

			req := httptest.NewRequest("GET", "/forecasts?page=1&limit=10", nil)
			w := httptest.NewRecorder()

			err := controller.List(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("GetByCityID", func(t *testing.T) {
			forecasts := []*repo.Forecast{createTestRepoForecast()}
			mockRepo := &MockForecastRepository{forecasts: forecasts}
			controller := NewHTTPForecastController(mockRepo)

			req := httptest.NewRequest("GET", "/cities/123/forecasts", nil)
			w := httptest.NewRecorder()

			err := controller.GetByCityID(context.Background(), w, req, 123)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("CleanupOldForecasts", func(t *testing.T) {
			mockRepo := &MockForecastRepository{}
			controller := NewHTTPForecastController(mockRepo)

			req := httptest.NewRequest("DELETE", "/forecasts/cleanup?days=30", nil)
			w := httptest.NewRecorder()

			err := controller.CleanupOldForecasts(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	})

	t.Run("CityController", func(t *testing.T) {
		t.Run("interface compliance", func(t *testing.T) {
			mockRepo := &MockCityRepository{}
			controller := NewHTTPCityController(mockRepo)

			var _ CityController = controller
			var _ Controller[City] = controller
		})

		t.Run("Search", func(t *testing.T) {
			cities := []*repo.City{createTestRepoCity()}
			mockRepo := &MockCityRepository{cities: cities}
			controller := NewHTTPCityController(mockRepo)

			req := httptest.NewRequest("GET", "/cities/search?q=San+Francisco", nil)
			w := httptest.NewRecorder()

			err := controller.Search(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("Search missing query", func(t *testing.T) {
			mockRepo := &MockCityRepository{}
			controller := NewHTTPCityController(mockRepo)

			req := httptest.NewRequest("GET", "/cities/search", nil)
			w := httptest.NewRecorder()

			_ = controller.Search(context.Background(), w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})

		t.Run("GetByCoordinates", func(t *testing.T) {
			cities := []*repo.City{createTestRepoCity()}
			mockRepo := &MockCityRepository{cities: cities}
			controller := NewHTTPCityController(mockRepo)

			req := httptest.NewRequest("GET", "/cities/coordinates?lat=37.7749&lon=-122.4194&radius=50", nil)
			w := httptest.NewRecorder()

			err := controller.GetByCoordinates(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("GetByCoordinates invalid lat", func(t *testing.T) {
			mockRepo := &MockCityRepository{}
			controller := NewHTTPCityController(mockRepo)

			req := httptest.NewRequest("GET", "/cities/coordinates?lat=invalid&lon=-122.4194", nil)
			w := httptest.NewRecorder()

			_ = controller.GetByCoordinates(context.Background(), w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	})

	t.Run("PlaceController", func(t *testing.T) {
		t.Run("interface compliance", func(t *testing.T) {
			mockRepo := &MockPlaceRepository{}
			controller := NewHTTPPlaceController(mockRepo)

			var _ PlaceController = controller
			var _ Controller[Place] = controller
		})

		t.Run("Search", func(t *testing.T) {
			places := []*repo.Place{createTestRepoPlace()}
			mockRepo := &MockPlaceRepository{places: places}
			controller := NewHTTPPlaceController(mockRepo)

			req := httptest.NewRequest("GET", "/places/search?q=Golden+Gate", nil)
			w := httptest.NewRecorder()

			err := controller.Search(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("GetBySourcePlaceID", func(t *testing.T) {
			mockRepo := &MockPlaceRepository{place: createTestRepoPlace()}
			controller := NewHTTPPlaceController(mockRepo)

			req := httptest.NewRequest("GET", "/places/source?source=Nominatim&source_place_id=123", nil)
			w := httptest.NewRecorder()

			err := controller.GetBySourcePlaceID(context.Background(), w, req)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})

		t.Run("GetBySourcePlaceID missing parameters", func(t *testing.T) {
			mockRepo := &MockPlaceRepository{}
			controller := NewHTTPPlaceController(mockRepo)

			req := httptest.NewRequest("GET", "/places/source?source=Nominatim", nil)
			w := httptest.NewRecorder()

			_ = controller.GetBySourcePlaceID(context.Background(), w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	})
}

// Benchmark tests
func BenchmarkControllers(b *testing.B) {
	b.Run("ForecastController Create", func(b *testing.B) {
		mockRepo := &MockForecastRepository{}
		controller := NewHTTPForecastController(mockRepo)
		forecast := createTestControllerForecast()
		body, _ := json.Marshal(forecast)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("POST", "/forecasts", bytes.NewReader(body))
			w := httptest.NewRecorder()
			_ = controller.Create(context.Background(), w, req)
		}
	})

	b.Run("CityController Search", func(b *testing.B) {
		cities := []*repo.City{createTestRepoCity()}
		mockRepo := &MockCityRepository{cities: cities}
		controller := NewHTTPCityController(mockRepo)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/cities/search?q=San+Francisco", nil)
			w := httptest.NewRecorder()
			_ = controller.Search(context.Background(), w, req)
		}
	})
}
