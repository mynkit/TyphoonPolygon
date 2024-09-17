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
