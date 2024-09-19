package model

type LinearRing struct {
	Coordinates []Point
}

type Point struct {
	Latitude  float64
	Longitude float64
}

type GeoJSONPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type StormArea struct {
	CenterPoint          Point // NOTE: 台風の中心であって、円の中心ではない
	CircleLongDirection  float64
	CircleLongRadius     float64
	CircleShortDirection float64
	CircleShortRadius    float64
}

type ForecastCircle struct {
	CenterPoint          Point // NOTE: 台風の中心であって、円の中心ではない
	CircleLongDirection  float64
	CircleLongRadius     float64
	CircleShortDirection float64
	CircleShortRadius    float64
}

type ForecastCirclePolygons struct {
	ForecastCircles      [][]Point // MultiPolygon
	ForecastCircleBorder []Point   // Polygon
	CenterLine           []Point   // LineString
}

type TyphoonWarningArea struct {
	WarningAreaType      string `json:"warning_area_type"`
	WindSpeed            int    `json:"wind_speed"`
	CircleLongDirection  string `json:"circle_long_direction"`
	CircleLongRadius     int    `json:"circle_long_radius"`
	CircleShortDirection string `json:"circle_short_direction"`
	CircleShortRadius    int    `json:"circle_short_radius"`
}

type Typhoon struct {
	TargetTimestamp           string               `json:"target_timestamp"`
	TargetTimestampType       string               `json:"target_timestamp_type"`
	Latitude                  float64              `json:"latitude"`
	Longitude                 float64              `json:"longitude"`
	Location                  string               `json:"location"`
	Direction                 string               `json:"direction"`
	Velocity                  int                  `json:"velocity"`
	CentralPressure           int                  `json:"central_pressure"`
	MaxWindSpeedNearTheCenter int                  `json:"max_wind_speed_near_the_center"`
	InstantaneousMaxWindSpeed int                  `json:"instantaneous_max_wind_speed"`
	WarningAreas              []TyphoonWarningArea `json:"warning_areas"`
}
