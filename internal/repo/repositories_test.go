package repo

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

// MockDB implements the DB interface for testing
type MockDB struct {
	shouldError bool
	errorMsg    string
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return nil, fmt.Errorf("mock not fully implemented")
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	// Return nil to trigger errors in repository methods that try to scan
	return nil
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return &MockResult{rowsAffected: 1, lastInsertID: 123}, nil
}

// MockResult implements sql.Result for testing
type MockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r *MockResult) LastInsertId() (int64, error) {
	return r.lastInsertID, nil
}

func (r *MockResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func TestRepository(t *testing.T) {
	t.Run("Concurrency", func(t *testing.T) {
		mockDB := &MockDB{shouldError: false}
		repo := NewPostgreSQLForecastRepository(mockDB)
		ctx := context.Background()

		done := make(chan error, 3)

		for i := range 3 {
			go func(id int) {
				_, err := repo.List(ctx, 10, id*10)
				done <- err
			}(i)
		}

		for range 3 {
			err := <-done
			_ = err
		}
	})

	t.Run("Error messages", func(t *testing.T) {
		mockDB := &MockDB{shouldError: true, errorMsg: "connection refused"}
		repo := NewPostgreSQLForecastRepository(mockDB)
		ctx := context.Background()

		_, err := repo.List(ctx, 10, 0)
		if err == nil {
			t.Error("Expected error but got nil")
		}

		if err != nil {
			errStr := err.Error()
			if len(errStr) == 0 {
				t.Error("Error message is empty")
			}

			if len(errStr) < 10 {
				t.Errorf("Error message too short: %s", errStr)
			}
		}
	})

	t.Run("Interface Compliance", func(t *testing.T) {
		var _ Repository[Forecast] = (*PostgreSQLForecastRepository)(nil)
		var _ ForecastRepository = (*PostgreSQLForecastRepository)(nil)

		var _ Repository[City] = (*PostgreSQLCityRepository)(nil)
		var _ CityRepository = (*PostgreSQLCityRepository)(nil)

		var _ Repository[Place] = (*PostgreSQLPlaceRepository)(nil)
		var _ PlaceRepository = (*PostgreSQLPlaceRepository)(nil)

		t.Run("Creation", func(t *testing.T) {
			mockDB := &MockDB{}
			forecastRepo := NewPostgreSQLForecastRepository(mockDB)
			cityRepo := NewPostgreSQLCityRepository(mockDB)
			placeRepo := NewPostgreSQLPlaceRepository(mockDB)

			var _ Repository[Forecast] = forecastRepo
			var _ Repository[City] = cityRepo
			var _ Repository[Place] = placeRepo

			var _ ForecastRepository = forecastRepo
			var _ CityRepository = cityRepo
			var _ PlaceRepository = placeRepo

			if forecastRepo == nil || cityRepo == nil || placeRepo == nil {
				t.Error("One or more repositories are nil")
			}
		})
	})

	t.Run("Constructors", func(t *testing.T) {
		mockDB := &MockDB{}

		forecastRepo := NewPostgreSQLForecastRepository(mockDB)
		if forecastRepo == nil {
			t.Error("NewPostgreSQLForecastRepository returned nil")
		}

		cityRepo := NewPostgreSQLCityRepository(mockDB)
		if cityRepo == nil {
			t.Error("NewPostgreSQLCityRepository returned nil")
		}

		placeRepo := NewPostgreSQLPlaceRepository(mockDB)
		if placeRepo == nil {
			t.Error("NewPostgreSQLPlaceRepository returned nil")
		}
	})

	t.Run("Query Context", func(t *testing.T) {
		mockDB := &MockDB{shouldError: true, errorMsg: "database connection error"}

		t.Run("ForecastRepository List handles errors", func(t *testing.T) {
			repo := NewPostgreSQLForecastRepository(mockDB)
			ctx := context.Background()

			forecasts, err := repo.List(ctx, 10, 0)
			if err == nil {
				t.Error("Expected error from database, got nil")
			}
			if forecasts != nil {
				t.Error("Expected nil forecasts on error")
			}
		})

		t.Run("ForecastRepository GetByCityID handles errors", func(t *testing.T) {
			repo := NewPostgreSQLForecastRepository(mockDB)
			ctx := context.Background()

			forecasts, err := repo.GetByCityID(ctx, 123, 10, 0)
			if err == nil {
				t.Error("Expected error from database, got nil")
			}
			if forecasts != nil {
				t.Error("Expected nil forecasts on error")
			}
		})

		t.Run("CityRepository Search handles errors", func(t *testing.T) {
			repo := NewPostgreSQLCityRepository(mockDB)
			ctx := context.Background()

			cities, err := repo.Search(ctx, "San Francisco", 10)
			if err == nil {
				t.Error("Expected error from database, got nil")
			}
			if cities != nil {
				t.Error("Expected nil cities on error")
			}
		})

		t.Run("PlaceRepository Search handles errors", func(t *testing.T) {
			repo := NewPostgreSQLPlaceRepository(mockDB)
			ctx := context.Background()

			places, err := repo.Search(ctx, "Golden Gate", 5)
			if err == nil {
				t.Error("Expected error from database, got nil")
			}
			if places != nil {
				t.Error("Expected nil places on error")
			}
		})
	})

	t.Run("Exec Context", func(t *testing.T) {
		t.Run("DeleteOldForecasts succeeds with mock", func(t *testing.T) {
			mockDB := &MockDB{shouldError: false}
			repo := NewPostgreSQLForecastRepository(mockDB)
			ctx := context.Background()

			err := repo.DeleteOldForecasts(ctx, 7)
			if err != nil {
				t.Errorf("Expected successful operation, got error: %v", err)
			}
		})

		t.Run("Delete methods handle errors", func(t *testing.T) {
			mockDB := &MockDB{shouldError: true, errorMsg: "delete failed"}

			forecastRepo := NewPostgreSQLForecastRepository(mockDB)
			ctx := context.Background()

			err := forecastRepo.Delete(ctx, 1)
			if err == nil {
				t.Error("Expected error from database, got nil")
			}
		})
	})

}

func BenchmarkRepositories(b *testing.B) {
	b.Run("Forecast List", func(b *testing.B) {
		mockDB := &MockDB{shouldError: false}
		repo := NewPostgreSQLForecastRepository(mockDB)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = repo.List(ctx, 10, 0)
		}
	})

	b.Run("City Search", func(b *testing.B) {
		mockDB := &MockDB{shouldError: false}
		repo := NewPostgreSQLCityRepository(mockDB)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = repo.Search(ctx, "San Francisco", 10)
		}
	})
}
