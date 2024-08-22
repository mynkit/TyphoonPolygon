package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	geojson "github.com/paulmach/go.geojson"
)

// 定数: 地球の半径 (キロメートル)
const EarthRadius = 6378.137

type LinearRing struct {
	Coordinates []Point
}

type Point struct {
	Latitude  float64
	Longitude float64
}

type Typhoon struct {
	CenterPoint Point // NOTE: 台風の中心であって、円の中心ではない
	Polygon     LinearRing
}

type GeoJSONPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func radToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// 距離と方位角から新しい緯度経度を計算する
func calcCirclePoint(centerLat float64, centerLon float64, radius float64, theta float64) Point {
	// 緯度・経度・方位角をラジアンに変換
	centerLatRad := degToRad(centerLat)
	centerLonRad := degToRad(centerLon)
	// 方位を表すbearingは北が0度の時計回りなので、感覚に合うように補正
	// (thetaは東が0度で反時計回りと捉えたい)
	bearingRad := degToRad(90 - theta)

	// 新しい緯度を計算
	circleLatRad := math.Asin(math.Sin(centerLatRad)*math.Cos(radius/EarthRadius) +
		math.Cos(centerLatRad)*math.Sin(radius/EarthRadius)*math.Cos(bearingRad))

	// 新しい経度を計算
	circleLonRad := centerLonRad + math.Atan2(math.Sin(bearingRad)*math.Sin(radius/EarthRadius)*math.Cos(centerLatRad),
		math.Cos(radius/EarthRadius)-math.Sin(centerLatRad)*math.Sin(circleLatRad))

	// ラジアンから度に戻す
	circleLat := radToDeg(circleLatRad)
	circleLon := radToDeg(circleLonRad)

	return Point{
		Latitude:  circleLat,
		Longitude: circleLon,
	}
}

func calcTyphoonPolygon(typhoonCenterLat float64, typhoonCenterLon float64, wideAreaRadius float64, narrowAreaRadius float64, wideAreaBearing float64, numPoints int) Typhoon {
	points := make([]Point, 0, numPoints+1)

	// 円の中心は、台風の中心からwideAreaBearingの方角に、
	// 広域の半径(wideAreaRadius)から円の半径((wideAreaRadius + narrowAreaRadius) / 2.)を引いた距離
	// だけ進めば円の中心の緯度経度になる
	circleRadius := (wideAreaRadius + narrowAreaRadius) / 2.
	circleCenterPoint := calcCirclePoint(typhoonCenterLat, typhoonCenterLon, wideAreaRadius-circleRadius, wideAreaBearing)

	for i := 0; i <= numPoints; i++ {
		angle := 360 * float64(i) / float64(numPoints)
		circlePoint := calcCirclePoint(circleCenterPoint.Latitude, circleCenterPoint.Longitude, circleRadius, angle)
		points = append(points, circlePoint)
	}

	return Typhoon{
		CenterPoint: Point{Latitude: typhoonCenterLat, Longitude: typhoonCenterLon},
		Polygon: LinearRing{
			Coordinates: points,
		},
	}
}

func saveGeoJSONToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

func main() {
	typhoon := calcTyphoonPolygon(31.0, 141., 500., 280., -45., 100)

	// GeoJsonファイルの作成

	// Polygonの作成
	geojsonPoints := make([][]float64, 0, len(typhoon.Polygon.Coordinates)+1)
	for _, coordinate := range typhoon.Polygon.Coordinates {
		geojsonPoints = append(
			geojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}
	coordinates := [][][]float64{geojsonPoints}
	polygon := geojson.NewPolygonFeature(coordinates)

	// Pointの作成
	point := geojson.NewPointFeature([]float64{typhoon.CenterPoint.Longitude, typhoon.CenterPoint.Latitude})

	// FeatureCollectionにPointとPolygonを追加
	featureCollection := geojson.NewFeatureCollection()
	featureCollection.AddFeature(point)
	featureCollection.AddFeature(polygon)

	// GeoJSONとしてエンコード
	geoJSON, err := json.MarshalIndent(featureCollection, "", "  ")
	if err != nil {
		fmt.Println("Error encoding GeoJSON:", err)
		return
	}

	// GeoJSONを出力
	fmt.Println(string(geoJSON))

	// ファイルに保存する場合
	err = saveGeoJSONToFile("output.geojson", geoJSON)
	if err != nil {
		fmt.Println("Error saving GeoJSON to file:", err)
		return
	}

	fmt.Println("GeoJSON successfully written to output.geojson")
}
