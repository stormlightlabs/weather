package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCensusProvider_GetName(t *testing.T) {
	census := NewCensusProvider()
	if census.GetName() != "Census" {
		t.Errorf("expected name 'Census', got '%s'", census.GetName())
	}
}

func TestCensusProvider_SupportedRegions(t *testing.T) {
	census := NewCensusProvider()
	regions := census.SupportedRegions()
	if len(regions) != 1 || regions[0] != "US" {
		t.Errorf("expected regions ['US'], got %v", regions)
	}
}

func TestCensusProvider_buildAddressLine1(t *testing.T) {
	census := NewCensusProvider()
	
	tests := []struct {
		name       string
		components CensusAddressComponents
		expected   string
	}{
		{
			name: "full address",
			components: CensusAddressComponents{
				FromAddress:   "123",
				PreDirection:  "N",
				StreetName:    "Main",
				SuffixType:    "St",
				SuffixDirection: "SW",
			},
			expected: "123 N Main St SW",
		},
		{
			name: "address with range",
			components: CensusAddressComponents{
				FromAddress: "100",
				ToAddress:   "199",
				StreetName:  "Oak",
				SuffixType:  "Ave",
			},
			expected: "100-199 Oak Ave",
		},
		{
			name: "minimal address",
			components: CensusAddressComponents{
				StreetName: "Broadway",
			},
			expected: "Broadway",
		},
		{
			name: "with qualifiers",
			components: CensusAddressComponents{
				FromAddress:     "500",
				StreetName:      "First",
				SuffixType:      "St",
				PreQualifier:    "Old",
				SuffixQualifier: "Ext",
			},
			expected: "500 First St Old Ext",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := census.buildAddressLine1(&test.components)
			if result != test.expected {
				t.Errorf("buildAddressLine1() = '%s', expected '%s'", result, test.expected)
			}
		})
	}
}

func TestCensusProvider_calculateConfidence(t *testing.T) {
	census := NewCensusProvider()
	
	tests := []struct {
		name      string
		original  string
		matched   string
		expectMin float64
		expectMax float64
	}{
		{
			name:      "exact match",
			original:  "123 Main St",
			matched:   "123 Main St",
			expectMin: 1.0,
			expectMax: 1.0,
		},
		{
			name:      "partial match",
			original:  "123 Main Street",
			matched:   "123 Main St",
			expectMin: 0.5,
			expectMax: 0.95,
		},
		{
			name:      "no match",
			original:  "123 Oak Ave",
			matched:   "456 Pine St",
			expectMin: 0.1,
			expectMax: 0.5,
		},
		{
			name:      "empty original",
			original:  "",
			matched:   "123 Main St",
			expectMin: 0.4,
			expectMax: 0.6,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := census.calculateConfidence(test.original, test.matched)
			if result < test.expectMin || result > test.expectMax {
				t.Errorf("calculateConfidence('%s', '%s') = %f, expected between %f and %f", 
					test.original, test.matched, result, test.expectMin, test.expectMax)
			}
		})
	}
}

func TestCensusProvider_GeocodeAddress_MockServer(t *testing.T) {
	geocodeResponse := CensusGeocodeResponse{
		Result: CensusResult{
			Input: CensusInput{
				Address: CensusInputAddress{
					Address: "123 Main St, Anytown, ST",
				},
			},
			AddressMatches: []CensusAddressMatch{
				{
					MatchedAddress: "123 Main St, Anytown, ST, 12345",
					Coordinates: CensusCoordinates{
						X: -76.6413,
						Y: 39.0458,
					},
					TigerLine: CensusTigerLine{
						TigerLineId: "12345678",
					},
					AddressComponents: CensusAddressComponents{
						FromAddress: "123",
						StreetName:  "Main",
						SuffixType:  "St",
						City:        "Anytown",
						State:       "ST",
						Zip:         "12345",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/locations/onelineaddress") {
			json.NewEncoder(w).Encode(geocodeResponse)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	census := NewCensusProvider()
	census.BaseURL = server.URL

	ctx := context.Background()
	places, err := census.GeocodeAddress(ctx, "123 Main St, Anytown, ST")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(places) != 1 {
		t.Errorf("expected 1 place, got %d", len(places))
	}

	place := places[0]
	if place.DisplayName != "123 Main St, Anytown, ST, 12345" {
		t.Errorf("expected display name '123 Main St, Anytown, ST, 12345', got '%s'", place.DisplayName)
	}
	if place.Latitude != 39.0458 {
		t.Errorf("expected latitude 39.0458, got %f", place.Latitude)
	}
	if place.Longitude != -76.6413 {
		t.Errorf("expected longitude -76.6413, got %f", place.Longitude)
	}
	if place.AddressLine1 != "123 Main St" {
		t.Errorf("expected address line 1 '123 Main St', got '%s'", place.AddressLine1)
	}
	if place.City != "Anytown" {
		t.Errorf("expected city 'Anytown', got '%s'", place.City)
	}
	if place.Region != "ST" {
		t.Errorf("expected region 'ST', got '%s'", place.Region)
	}
	if place.PostalCode != "12345" {
		t.Errorf("expected postal code '12345', got '%s'", place.PostalCode)
	}
	if place.Country != "United States" {
		t.Errorf("expected country 'United States', got '%s'", place.Country)
	}
	if place.CountryCode != "US" {
		t.Errorf("expected country code 'US', got '%s'", place.CountryCode)
	}
	if place.PlaceType != "address" {
		t.Errorf("expected place type 'address', got '%s'", place.PlaceType)
	}
	if place.Source != "Census" {
		t.Errorf("expected source 'Census', got '%s'", place.Source)
	}
	if place.SourcePlaceID != "12345678" {
		t.Errorf("expected source place ID '12345678', got '%s'", place.SourcePlaceID)
	}
	if place.Confidence <= 0 || place.Confidence > 1 {
		t.Errorf("expected confidence between 0 and 1, got %f", place.Confidence)
	}
}

func TestCensusProvider_ReverseGeocode_MockServer(t *testing.T) {
	reverseResponse := CensusReverseGeocodeResponse{
		Result: CensusReverseResult{
			Input: CensusReverseInput{
				Location: CensusLocation{
					X: -76.6413,
					Y: 39.0458,
				},
			},
			AddressMatches: []CensusReverseMatch{
				{
					MatchedAddress: "456 Oak Ave, Somewhere, ST, 54321",
					AddressComponents: CensusAddressComponents{
						FromAddress: "456",
						StreetName:  "Oak",
						SuffixType:  "Ave",
						City:        "Somewhere",
						State:       "ST",
						Zip:         "54321",
					},
					TigerLine: CensusTigerLine{
						TigerLineId: "87654321",
					},
					Coordinates: CensusCoordinates{
						X: -76.6413,
						Y: 39.0458,
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/locations/reversegeocoding") {
			json.NewEncoder(w).Encode(reverseResponse)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	census := NewCensusProvider()
	census.BaseURL = server.URL

	ctx := context.Background()
	place, err := census.ReverseGeocode(ctx, 39.0458, -76.6413)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if place.DisplayName != "456 Oak Ave, Somewhere, ST, 54321" {
		t.Errorf("expected display name '456 Oak Ave, Somewhere, ST, 54321', got '%s'", place.DisplayName)
	}
	if place.Latitude != 39.0458 {
		t.Errorf("expected latitude 39.0458, got %f", place.Latitude)
	}
	if place.Longitude != -76.6413 {
		t.Errorf("expected longitude -76.6413, got %f", place.Longitude)
	}
	if place.AddressLine1 != "456 Oak Ave" {
		t.Errorf("expected address line 1 '456 Oak Ave', got '%s'", place.AddressLine1)
	}
	if place.City != "Somewhere" {
		t.Errorf("expected city 'Somewhere', got '%s'", place.City)
	}
	if place.Region != "ST" {
		t.Errorf("expected region 'ST', got '%s'", place.Region)
	}
	if place.PostalCode != "54321" {
		t.Errorf("expected postal code '54321', got '%s'", place.PostalCode)
	}
	if place.PlaceType != "address" {
		t.Errorf("expected place type 'address', got '%s'", place.PlaceType)
	}
	if place.Source != "Census" {
		t.Errorf("expected source 'Census', got '%s'", place.Source)
	}
	if place.SourcePlaceID != "87654321" {
		t.Errorf("expected source place ID '87654321', got '%s'", place.SourcePlaceID)
	}
	if place.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %f", place.Confidence)
	}
}

func TestCensusProvider_ErrorHandling(t *testing.T) {
	// Test with server that returns empty results
	emptyResponse := CensusGeocodeResponse{
		Result: CensusResult{
			AddressMatches: []CensusAddressMatch{},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/locations/onelineaddress"):
			json.NewEncoder(w).Encode(emptyResponse)
		case strings.Contains(r.URL.Path, "/locations/reversegeocoding"):
			json.NewEncoder(w).Encode(CensusReverseGeocodeResponse{
				Result: CensusReverseResult{
					AddressMatches: []CensusReverseMatch{},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	census := NewCensusProvider()
	census.BaseURL = server.URL

	ctx := context.Background()

	// Test GeocodeAddress with empty results
	_, err := census.GeocodeAddress(ctx, "NonExistent Address")
	if err == nil {
		t.Error("expected error for empty geocoding results, got nil")
	}
	if !strings.Contains(err.Error(), "no geocoding results found") {
		t.Errorf("expected 'no geocoding results found' in error, got: %v", err)
	}

	// Test ReverseGeocode with empty results
	_, err = census.ReverseGeocode(ctx, 0.0, 0.0)
	if err == nil {
		t.Error("expected error for empty reverse geocoding results, got nil")
	}
	if !strings.Contains(err.Error(), "no reverse geocoding results found") {
		t.Errorf("expected 'no reverse geocoding results found' in error, got: %v", err)
	}
}

func TestCensusProvider_ErrorHandling_HTTPError(t *testing.T) {
	// Test with server that returns 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	census := NewCensusProvider()
	census.BaseURL = server.URL

	ctx := context.Background()

	// Test GeocodeAddress with HTTP error
	_, err := census.GeocodeAddress(ctx, "Test Address")
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "geocoding request failed") {
		t.Errorf("expected 'geocoding request failed' in error, got: %v", err)
	}

	// Test ReverseGeocode with HTTP error
	_, err = census.ReverseGeocode(ctx, 39.0458, -76.6413)
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "reverse geocoding request failed") {
		t.Errorf("expected 'reverse geocoding request failed' in error, got: %v", err)
	}
}

func TestCensusProvider_ErrorHandling_InvalidJSON(t *testing.T) {
	// Test with server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	census := NewCensusProvider()
	census.BaseURL = server.URL

	ctx := context.Background()

	// Test GeocodeAddress with invalid JSON
	_, err := census.GeocodeAddress(ctx, "Test Address")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "geocoding request failed") {
		t.Errorf("expected 'geocoding request failed' in error, got: %v", err)
	}

	// Test ReverseGeocode with invalid JSON
	_, err = census.ReverseGeocode(ctx, 39.0458, -76.6413)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "reverse geocoding request failed") {
		t.Errorf("expected 'reverse geocoding request failed' in error, got: %v", err)
	}
}