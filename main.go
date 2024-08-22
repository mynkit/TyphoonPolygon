package main

import (
	"fmt"
	"math"
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

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func radToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// 距離と方位角から新しい緯度経度を計算する
func getCirclePoint(centerLat float64, centerLon float64, radius float64, bearing float64) Point {
	// 緯度・経度・方位角をラジアンに変換
	centerLatRad := degToRad(centerLat)
	centerLonRad := degToRad(centerLon)
	bearingRad := degToRad(bearing + 90) // 南方向が0度扱いになっていて感覚的でないので、東を0度扱いにする

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

func main() {
	// 元の緯度経度
	centerLat := 35.0  // 例: 東京の緯度
	centerLon := 139.0 // 例: 東京の経度

	// 距離と方位角
	radius := 100.0 // 距離 (キロメートル)
	bearing := 90.0 // 方位角 (度)

	// 新しい地点の緯度経度を計算
	circlePoint := getCirclePoint(centerLat, centerLon, radius, bearing)

	// 結果を出力
	fmt.Printf("新しい地点の緯度: %f, 経度: %f\n", circlePoint.Latitude, circlePoint.Longitude)
}
