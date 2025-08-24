package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/weather_api/internal/models"
)

// CensusProvider implements GeocodeProvider for the US Census Geocoding API
type CensusProvider struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewCensusProvider creates a new US Census geocoding provider
func NewCensusProvider() *CensusProvider {
	return &CensusProvider{
		BaseURL: "https://geocoding.geo.census.gov/geocoder",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *CensusProvider) GetName() string {
	return "Census"
}

func (c *CensusProvider) SupportedRegions() []string {
	return []string{"US"} // US Census only covers United States
}

// Census API Response structures
type CensusGeocodeResponse struct {
	Result CensusResult `json:"result"`
}

type CensusResult struct {
	Input       CensusInput           `json:"input"`
	AddressMatches []CensusAddressMatch `json:"addressMatches"`
}

type CensusInput struct {
	Address CensusInputAddress `json:"address"`
}

type CensusInputAddress struct {
	Address string `json:"address"`
}

type CensusAddressMatch struct {
	MatchedAddress string                `json:"matchedAddress"`
	Coordinates    CensusCoordinates     `json:"coordinates"`
	TigerLine      CensusTigerLine       `json:"tigerLine"`
	AddressComponents CensusAddressComponents `json:"addressComponents"`
}

type CensusCoordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CensusTigerLine struct {
	Side     string `json:"side"`
	TigerLineId string `json:"tigerLineId"`
}

type CensusAddressComponents struct {
	Zip            string `json:"zip"`
	StreetName     string `json:"streetName"`
	PreType        string `json:"preType"`
	City           string `json:"city"`
	PreDirection   string `json:"preDirection"`
	SuffixDirection string `json:"suffixDirection"`
	FromAddress    string `json:"fromAddress"`
	State          string `json:"state"`
	SuffixType     string `json:"suffixType"`
	ToAddress      string `json:"toAddress"`
	SuffixQualifier string `json:"suffixQualifier"`
	PreQualifier   string `json:"preQualifier"`
}

// Census Reverse Geocode Response structures
type CensusReverseGeocodeResponse struct {
	Result CensusReverseResult `json:"result"`
}

type CensusReverseResult struct {
	Input         CensusReverseInput      `json:"input"`
	AddressMatches []CensusReverseMatch   `json:"addressMatches"`
}

type CensusReverseInput struct {
	Location CensusLocation `json:"location"`
}

type CensusLocation struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CensusReverseMatch struct {
	MatchedAddress    string                    `json:"matchedAddress"`
	AddressComponents CensusAddressComponents   `json:"addressComponents"`
	TigerLine         CensusTigerLine          `json:"tigerLine"`
	Coordinates       CensusCoordinates         `json:"coordinates"`
}

func (c *CensusProvider) GeocodeAddress(ctx context.Context, address string) ([]*models.Place, error) {
	// Build the geocoding request URL
	params := url.Values{
		"address": {address},
		"format":  {"json"},
		"benchmark": {"2020"}, // Use 2020 Census benchmark
		"vintage":   {"Current_Current"}, // Current address range and current TIGER
	}

	requestURL := fmt.Sprintf("%s/locations/onelineaddress?%s", c.BaseURL, params.Encode())

	data, err := c.makeRequest(ctx, requestURL)
	if err != nil {
		return nil, fmt.Errorf("geocoding request failed: %w", err)
	}

	var response CensusGeocodeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	var places []*models.Place
	for _, match := range response.Result.AddressMatches {
		place, err := c.addressMatchToPlace(&match, address)
		if err != nil {
			continue // Skip invalid matches
		}
		places = append(places, place)
	}

	if len(places) == 0 {
		return nil, fmt.Errorf("no geocoding results found for address: %s", address)
	}

	return places, nil
}

func (c *CensusProvider) ReverseGeocode(ctx context.Context, lat, lon float64) (*models.Place, error) {
	// Build the reverse geocoding request URL
	params := url.Values{
		"x":         {fmt.Sprintf("%.6f", lon)},
		"y":         {fmt.Sprintf("%.6f", lat)},
		"format":    {"json"},
		"benchmark": {"2020"}, // Use 2020 Census benchmark
		"vintage":   {"Current_Current"}, // Current address range and current TIGER
	}

	requestURL := fmt.Sprintf("%s/locations/reversegeocoding?%s", c.BaseURL, params.Encode())

	data, err := c.makeRequest(ctx, requestURL)
	if err != nil {
		return nil, fmt.Errorf("reverse geocoding request failed: %w", err)
	}

	var response CensusReverseGeocodeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse reverse geocoding response: %w", err)
	}

	if len(response.Result.AddressMatches) == 0 {
		return nil, fmt.Errorf("no reverse geocoding results found for coordinates: %f, %f", lat, lon)
	}

	// Use the first (best) match
	match := response.Result.AddressMatches[0]
	return c.reverseMatchToPlace(&match, lat, lon)
}

func (c *CensusProvider) makeRequest(ctx context.Context, requestURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "weather-api/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (c *CensusProvider) addressMatchToPlace(match *CensusAddressMatch, originalAddress string) (*models.Place, error) {
	place := &models.Place{
		DisplayName:   match.MatchedAddress,
		AddressLine1:  c.buildAddressLine1(&match.AddressComponents),
		City:          match.AddressComponents.City,
		Region:        match.AddressComponents.State,
		PostalCode:    match.AddressComponents.Zip,
		Country:       "United States",
		CountryCode:   "US",
		Latitude:      match.Coordinates.Y,
		Longitude:     match.Coordinates.X,
		PlaceType:     "address",
		Confidence:    c.calculateConfidence(originalAddress, match.MatchedAddress),
		Source:        c.GetName(),
		SourcePlaceID: match.TigerLine.TigerLineId,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return place, nil
}

func (c *CensusProvider) reverseMatchToPlace(match *CensusReverseMatch, lat, lon float64) (*models.Place, error) {
	place := &models.Place{
		DisplayName:   match.MatchedAddress,
		AddressLine1:  c.buildAddressLine1(&match.AddressComponents),
		City:          match.AddressComponents.City,
		Region:        match.AddressComponents.State,
		PostalCode:    match.AddressComponents.Zip,
		Country:       "United States",
		CountryCode:   "US",
		Latitude:      lat,
		Longitude:     lon,
		PlaceType:     "address",
		Confidence:    0.9, // High confidence for reverse geocoding
		Source:        c.GetName(),
		SourcePlaceID: match.TigerLine.TigerLineId,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return place, nil
}

func (c *CensusProvider) buildAddressLine1(components *CensusAddressComponents) string {
	var parts []string

	// Add house number range if available
	if components.FromAddress != "" {
		if components.ToAddress != "" && components.ToAddress != components.FromAddress {
			parts = append(parts, fmt.Sprintf("%s-%s", components.FromAddress, components.ToAddress))
		} else {
			parts = append(parts, components.FromAddress)
		}
	}

	// Add pre-direction
	if components.PreDirection != "" {
		parts = append(parts, components.PreDirection)
	}

	// Add pre-type
	if components.PreType != "" {
		parts = append(parts, components.PreType)
	}

	// Add street name
	if components.StreetName != "" {
		parts = append(parts, components.StreetName)
	}

	// Add suffix type
	if components.SuffixType != "" {
		parts = append(parts, components.SuffixType)
	}

	// Add suffix direction
	if components.SuffixDirection != "" {
		parts = append(parts, components.SuffixDirection)
	}

	// Add qualifiers
	if components.PreQualifier != "" {
		parts = append(parts, components.PreQualifier)
	}
	
	if components.SuffixQualifier != "" {
		parts = append(parts, components.SuffixQualifier)
	}

	return strings.Join(parts, " ")
}

func (c *CensusProvider) calculateConfidence(original, matched string) float64 {
	// Simple confidence calculation based on string similarity
	original = strings.ToLower(strings.TrimSpace(original))
	matched = strings.ToLower(strings.TrimSpace(matched))

	if original == matched {
		return 1.0
	}

	// Calculate a simple similarity score
	originalWords := strings.Fields(original)
	matchedWords := strings.Fields(matched)
	
	commonWords := 0
	for _, originalWord := range originalWords {
		for _, matchedWord := range matchedWords {
			if originalWord == matchedWord {
				commonWords++
				break
			}
		}
	}

	if len(originalWords) == 0 {
		return 0.5 // Default confidence
	}

	similarity := float64(commonWords) / float64(len(originalWords))
	
	// Ensure confidence is between 0.1 and 0.95 for geocoding results
	if similarity < 0.1 {
		return 0.1
	}
	if similarity > 0.95 {
		return 0.95
	}
	
	return similarity
}

// Helper function to parse numeric strings safely
func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0.0
}