package main

import (
	"fmt"
	"math"
)

// 定数: 地球の半径 (キロメートル)
const EarthRadius = 6378.137 // 地球の半径 (キロメートル)

// 度からラジアンへの変換
func degToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// ラジアンから度への変換
func radToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// 距離と方位角から新しい緯度経度を計算する
func calculateNewLatLon(lat, lon, distance, bearing float64) (float64, float64) {
	// 緯度・経度・方位角をラジアンに変換
	latRad := degToRad(lat)
	lonRad := degToRad(lon)
	bearingRad := degToRad(bearing + 90) // 南方向が0度扱いになっていて感覚的でないので、東を0度扱いにする

	// 新しい緯度を計算
	newLatRad := math.Asin(math.Sin(latRad)*math.Cos(distance/EarthRadius) +
		math.Cos(latRad)*math.Sin(distance/EarthRadius)*math.Cos(bearingRad))

	// 新しい経度を計算
	newLonRad := lonRad + math.Atan2(math.Sin(bearingRad)*math.Sin(distance/EarthRadius)*math.Cos(latRad),
		math.Cos(distance/EarthRadius)-math.Sin(latRad)*math.Sin(newLatRad))

	// ラジアンから度に戻す
	newLat := radToDeg(newLatRad)
	newLon := radToDeg(newLonRad)

	return newLat, newLon
}

func main() {
	// 元の緯度経度
	lat := 35.0  // 例: 東京の緯度
	lon := 139.0 // 例: 東京の経度

	// 距離と方位角
	distance := 100.0 // 距離 (キロメートル)
	bearing := 90.0   // 方位角 (度)

	// 新しい地点の緯度経度を計算
	newLat, newLon := calculateNewLatLon(lat, lon, distance, bearing)

	// 結果を出力
	fmt.Printf("新しい地点の緯度: %f, 経度: %f\n", newLat, newLon)
}
