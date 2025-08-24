package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/weather_api/internal/models"
)

// NWSProvider implements WeatherProvider for the National Weather Service API
type NWSProvider struct {
	BaseURL    string
	UserAgent  string
	HTTPClient *http.Client
}

// NewNWSProvider creates a new NWS weather provider
func NewNWSProvider() *NWSProvider {
	return &NWSProvider{
		BaseURL: "https://api.weather.gov",
		// TODO: Replace with actual contact
		UserAgent: "weather-api/1.0 (contact@example.com)",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (n *NWSProvider) GetName() string {
	return "NWS"
}

func (n *NWSProvider) SupportedRegions() []string {
	return []string{"US"} // NWS only covers United States
}

// NWS API Response structures
type NWSPointResponse struct {
	Properties NWSPointProperties `json:"properties"`
}

type NWSPointProperties struct {
	GridID              string `json:"gridId"`
	GridX               int    `json:"gridX"`
	GridY               int    `json:"gridY"`
	Forecast            string `json:"forecast"`
	ForecastHourly      string `json:"forecastHourly"`
	ObservationStations string `json:"observationStations"`
}

type NWSForecastResponse struct {
	Properties NWSForecastProperties `json:"properties"`
}

type NWSForecastProperties struct {
	Periods []NWSForecastPeriod `json:"periods"`
}

type NWSForecastPeriod struct {
	Number           int    `json:"number"`
	Name             string `json:"name"`
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
	IsDaytime        bool   `json:"isDaytime"`
	Temperature      int    `json:"temperature"`
	TemperatureUnit  string `json:"temperatureUnit"`
	WindSpeed        string `json:"windSpeed"`
	WindDirection    string `json:"windDirection"`
	Icon             string `json:"icon"`
	ShortForecast    string `json:"shortForecast"`
	DetailedForecast string `json:"detailedForecast"`
}

type NWSObservationResponse struct {
	Properties NWSObservationProperties `json:"properties"`
}

type NWSObservationProperties struct {
	Timestamp          string               `json:"timestamp"`
	Temperature        NWSQuantitativeValue `json:"temperature"`
	Dewpoint           NWSQuantitativeValue `json:"dewpoint"`
	WindDirection      NWSQuantitativeValue `json:"windDirection"`
	WindSpeed          NWSQuantitativeValue `json:"windSpeed"`
	BarometricPressure NWSQuantitativeValue `json:"barometricPressure"`
	RelativeHumidity   NWSQuantitativeValue `json:"relativeHumidity"`
	Visibility         NWSQuantitativeValue `json:"visibility"`
	TextDescription    string               `json:"textDescription"`
}

type NWSQuantitativeValue struct {
	Value          *float64 `json:"value"`
	UnitCode       string   `json:"unitCode"`
	QualityControl string   `json:"qualityControl"`
}

type NWSAlertsResponse struct {
	Features []NWSAlert `json:"features"`
}

type NWSAlert struct {
	Properties NWSAlertProperties `json:"properties"`
}

type NWSAlertProperties struct {
	ID          string `json:"id"`
	Event       string `json:"event"`
	Headline    string `json:"headline"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Urgency     string `json:"urgency"`
	Category    string `json:"category"`
	Onset       string `json:"onset"`
	Expires     string `json:"expires"`
	AreaDesc    string `json:"areaDesc"`
}

func (n *NWSProvider) GetCurrentWeather(ctx context.Context, lat, lon float64) (*models.Forecast, error) {
	// First get the grid point info
	point, err := n.getGridPoint(ctx, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid point: %w", err)
	}

	// Get current observation from the nearest station
	stationsURL := fmt.Sprintf("%s/gridpoints/%s/%d,%d/stations", n.BaseURL, point.Properties.GridID, point.Properties.GridX, point.Properties.GridY)
	stations, err := n.makeRequest(ctx, stationsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get observation stations: %w", err)
	}

	var stationsResp struct {
		Features []struct {
			Properties struct {
				StationIdentifier string `json:"stationIdentifier"`
			} `json:"properties"`
		} `json:"features"`
	}

	if err := json.Unmarshal(stations, &stationsResp); err != nil {
		return nil, fmt.Errorf("failed to parse stations response: %w", err)
	}

	if len(stationsResp.Features) == 0 {
		return nil, fmt.Errorf("no observation stations found")
	}

	// Get current observation from the first station
	stationID := stationsResp.Features[0].Properties.StationIdentifier
	obsURL := fmt.Sprintf("%s/stations/%s/observations/latest", n.BaseURL, stationID)

	obsData, err := n.makeRequest(ctx, obsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get current observation: %w", err)
	}

	var obsResp NWSObservationResponse
	if err := json.Unmarshal(obsData, &obsResp); err != nil {
		return nil, fmt.Errorf("failed to parse observation response: %w", err)
	}

	return n.observationToForecast(&obsResp, lat, lon)
}

func (n *NWSProvider) GetForecast(ctx context.Context, lat, lon float64, days int) ([]*models.Forecast, error) {
	// Get grid point info
	point, err := n.getGridPoint(ctx, lat, lon)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid point: %w", err)
	}

	// Get forecast data
	forecastData, err := n.makeRequest(ctx, point.Properties.Forecast)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	var forecastResp NWSForecastResponse
	if err := json.Unmarshal(forecastData, &forecastResp); err != nil {
		return nil, fmt.Errorf("failed to parse forecast response: %w", err)
	}

	// Convert NWS periods to forecast models
	var forecasts []*models.Forecast
	maxPeriods := days * 2 // NWS gives day/night periods
	if maxPeriods > len(forecastResp.Properties.Periods) {
		maxPeriods = len(forecastResp.Properties.Periods)
	}

	for i := 0; i < maxPeriods; i++ {
		period := forecastResp.Properties.Periods[i]
		forecast, err := n.periodToForecast(&period, lat, lon)
		if err != nil {
			continue // Skip invalid periods
		}
		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

func (n *NWSProvider) GetAlerts(ctx context.Context, lat, lon float64) ([]WeatherAlert, error) {
	alertsURL := fmt.Sprintf("%s/alerts/active?point=%f,%f", n.BaseURL, lat, lon)

	alertData, err := n.makeRequest(ctx, alertsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	var alertsResp NWSAlertsResponse
	if err := json.Unmarshal(alertData, &alertsResp); err != nil {
		return nil, fmt.Errorf("failed to parse alerts response: %w", err)
	}

	var alerts []WeatherAlert
	for _, nwsAlert := range alertsResp.Features {
		alert, err := n.nwsAlertToWeatherAlert(&nwsAlert)
		if err != nil {
			continue // Skip invalid alerts
		}
		alerts = append(alerts, *alert)
	}

	return alerts, nil
}

func (n *NWSProvider) getGridPoint(ctx context.Context, lat, lon float64) (*NWSPointResponse, error) {
	url := fmt.Sprintf("%s/points/%f,%f", n.BaseURL, lat, lon)

	data, err := n.makeRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var point NWSPointResponse
	if err := json.Unmarshal(data, &point); err != nil {
		return nil, fmt.Errorf("failed to parse point response: %w", err)
	}

	return &point, nil
}

func (n *NWSProvider) makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", n.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := n.HTTPClient.Do(req)
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

func (n *NWSProvider) observationToForecast(obs *NWSObservationResponse, lat, lon float64) (*models.Forecast, error) {
	var timestamp time.Time
	var err error

	if obs.Properties.Timestamp != "" {
		timestamp, err = time.Parse(time.RFC3339, obs.Properties.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
	} else {
		timestamp = time.Now() // Use current time if timestamp is empty
	}

	forecast := &models.Forecast{
		SourceProvider: n.GetName(),
		ForecastTime:   timestamp,
		ValidTime:      timestamp,
		Description:    obs.Properties.TextDescription,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Convert temperature (Celsius)
	if obs.Properties.Temperature.Value != nil {
		forecast.Temperature = *obs.Properties.Temperature.Value
	}

	// Convert humidity
	if obs.Properties.RelativeHumidity.Value != nil {
		forecast.Humidity = *obs.Properties.RelativeHumidity.Value
	}

	// Convert pressure (hPa)
	if obs.Properties.BarometricPressure.Value != nil {
		forecast.Pressure = *obs.Properties.BarometricPressure.Value / 100 // Convert Pa to hPa
	}

	// Convert wind speed (m/s)
	if obs.Properties.WindSpeed.Value != nil {
		forecast.WindSpeed = *obs.Properties.WindSpeed.Value
	}

	// Convert wind direction (degrees)
	if obs.Properties.WindDirection.Value != nil {
		forecast.WindDirection = *obs.Properties.WindDirection.Value
	}

	// Convert visibility (km)
	if obs.Properties.Visibility.Value != nil {
		forecast.Visibility = *obs.Properties.Visibility.Value / 1000 // Convert m to km
	}

	return forecast, nil
}

func (n *NWSProvider) periodToForecast(period *NWSForecastPeriod, lat, lon float64) (*models.Forecast, error) {
	startTime, err := time.Parse(time.RFC3339, period.StartTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	_, err = time.Parse(time.RFC3339, period.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	forecast := &models.Forecast{
		SourceProvider: n.GetName(),
		ForecastTime:   time.Now(),
		ValidTime:      startTime,
		Description:    period.DetailedForecast,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Convert temperature
	if period.TemperatureUnit == "F" {
		// Convert Fahrenheit to Celsius
		forecast.Temperature = (float64(period.Temperature) - 32) * 5 / 9
	} else {
		forecast.Temperature = float64(period.Temperature)
	}

	// Parse wind speed (format like "5 mph" or "5 to 10 mph")
	if period.WindSpeed != "" {
		windParts := strings.Fields(period.WindSpeed)
		if len(windParts) >= 2 {
			if speed, err := strconv.Atoi(windParts[0]); err == nil {
				// Convert mph to m/s
				forecast.WindSpeed = float64(speed) * 0.44704
			}
		}
	}

	// Parse wind direction
	forecast.WindDirection = n.parseWindDirection(period.WindDirection)

	return forecast, nil
}

func (n *NWSProvider) parseWindDirection(direction string) float64 {
	directions := map[string]float64{
		"N": 0, "NNE": 22.5, "NE": 45, "ENE": 67.5,
		"E": 90, "ESE": 112.5, "SE": 135, "SSE": 157.5,
		"S": 180, "SSW": 202.5, "SW": 225, "WSW": 247.5,
		"W": 270, "WNW": 292.5, "NW": 315, "NNW": 337.5,
	}

	if deg, ok := directions[strings.ToUpper(direction)]; ok {
		return deg
	}
	return 0 // Default to North if unknown
}

func (n *NWSProvider) nwsAlertToWeatherAlert(nwsAlert *NWSAlert) (*WeatherAlert, error) {
	alert := &WeatherAlert{
		ID:          nwsAlert.Properties.ID,
		Title:       nwsAlert.Properties.Event,
		Description: nwsAlert.Properties.Description,
		Severity:    strings.ToLower(nwsAlert.Properties.Severity),
		Urgency:     strings.ToLower(nwsAlert.Properties.Urgency),
		Category:    strings.ToLower(nwsAlert.Properties.Category),
		Areas:       []string{nwsAlert.Properties.AreaDesc},
	}

	// Parse timestamps
	if nwsAlert.Properties.Onset != "" {
		if startTime, err := time.Parse(time.RFC3339, nwsAlert.Properties.Onset); err == nil {
			alert.StartTime = startTime
		}
	}

	if nwsAlert.Properties.Expires != "" {
		if endTime, err := time.Parse(time.RFC3339, nwsAlert.Properties.Expires); err == nil {
			alert.EndTime = endTime
		}
	}

	return alert, nil
}
