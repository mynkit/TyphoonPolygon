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

// 2つの緯度・経度の間の距離を計算する関数
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := degToRad(lat1)
	lon1Rad := degToRad(lon1)
	lat2Rad := degToRad(lat2)
	lon2Rad := degToRad(lon2)

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := EarthRadius * c
	return distance
}

// 方位角を計算する関数
func calculateTheta(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := degToRad(lat1)
	lon1Rad := degToRad(lon1)
	lat2Rad := degToRad(lat2)
	lon2Rad := degToRad(lon2)

	dlon := lon2Rad - lon1Rad

	x := math.Sin(dlon) * math.Cos(lat2Rad)
	y := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dlon)

	initialBearing := math.Atan2(x, y)
	initialBearing = initialBearing * 180 / math.Pi
	bearing := math.Mod(initialBearing+360, 360) // 方位角を0〜360の範囲に正規化
	// 方位を表すbearingは北が0度の時計回りなので、感覚に合うように補正
	// (thetaは東が0度で反時計回りと捉えたい)
	theta := math.Mod(90-bearing+360, 360) // 角度を0〜360の範囲に正規化

	return theta
}

// 距離と方位角から新しい緯度経度を計算する
func calcCirclePoint(centerLat, centerLon, radius, theta float64) Point {
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

func calcTyphoonPolygon(typhoonCenterLat, typhoonCenterLon, wideAreaRadius, narrowAreaRadius, wideAreaBearing float64, numPoints int) Typhoon {
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
	typhoons := []Typhoon{
		// 実況
		// calcTyphoonPolygon(22.3, 140.9, 330., 220., 0., 120), // 強風域
		calcTyphoonPolygon(22.3, 140.9, 55., 55., 0., 120), // 暴風域
		// 予報　１２時間後
		calcTyphoonPolygon(24.9, 139.6, 75., 75., 0., 120),   // 予報円
		calcTyphoonPolygon(24.9, 139.6, 130., 130., 0., 120), // 暴風警戒域
		// 予報　２４時間後
		calcTyphoonPolygon(26.8, 137.8, 105., 105., 0., 120), // 予報円
		calcTyphoonPolygon(26.8, 137.8, 190., 190., 0., 120), // 暴風警戒域
		// 予報　４８時間後
		calcTyphoonPolygon(29.2, 133.7, 155., 155., 0., 120), // 予報円
		calcTyphoonPolygon(29.2, 133.7, 310., 310., 0., 120), // 暴風警戒域
		// 予報　７２時間後
		calcTyphoonPolygon(32.2, 133.3, 220., 220., 0., 120), // 予報円
		calcTyphoonPolygon(32.2, 133.3, 360., 360., 0., 120), // 暴風警戒域
	}

	// GeoJsonファイルの作成

	// FeatureCollectionにPointとPolygonを追加
	featureCollection := geojson.NewFeatureCollection()

	for _, typhoon := range typhoons {
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

		featureCollection.AddFeature(point)
		featureCollection.AddFeature(polygon)
	}

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

	dis := haversineDistance(22.3, 140.9, 24.9, 139.6)
	fmt.Println("dis = ", dis)
	theta := calculateTheta(29.2, 133.7, 32.2, 133.3)
	fmt.Println("theta = ", theta)
}
