package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
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
	thetaRad := degToRad(theta + 90) // 南方向が0度扱いになっていて感覚的でないので、東を0度扱いにする

	// 新しい緯度を計算
	circleLatRad := math.Asin(math.Sin(centerLatRad)*math.Cos(radius/EarthRadius) +
		math.Cos(centerLatRad)*math.Sin(radius/EarthRadius)*math.Cos(thetaRad))

	// 新しい経度を計算
	circleLonRad := centerLonRad + math.Atan2(math.Sin(thetaRad)*math.Sin(radius/EarthRadius)*math.Cos(centerLatRad),
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

func main() {
	typhoon := calcTyphoonPolygon(31.0, 141., 500., 280., -45., 100)

	// geojsonファイルの書き出し
	geojsonPoints := make([][]float64, 0, len(typhoon.Polygon.Coordinates)+1)
	for _, coordinate := range typhoon.Polygon.Coordinates {
		geojsonPoints = append(
			geojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}

	geoJSON := GeoJSONPolygon{
		Type:        "Polygon",
		Coordinates: [][][]float64{geojsonPoints},
	}

	// GeoJSONをファイルに書き出し
	file, err := os.Create("ellipse.geojson")
	if err != nil {
		fmt.Println("ファイル作成に失敗しました:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 読みやすいフォーマットにするためインデントを設定

	err = encoder.Encode(geoJSON)
	if err != nil {
		fmt.Println("GeoJSONの書き出しに失敗しました:", err)
		return
	}

	fmt.Println("GeoJSONファイル 'ellipse.geojson' を作成しました")
}
