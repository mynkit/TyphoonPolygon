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

type TyphoonPolygon struct {
	CenterPoint Point // NOTE: 台風の中心であって、円の中心ではない
	Polygon     LinearRing
}
