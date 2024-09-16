package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	geojson "github.com/paulmach/go.geojson"
	geos "github.com/twpayne/go-geos"
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

// Helper function to determine the orientation of three points
// 0 -> p, q and r are collinear
// 1 -> Clockwise
// -1 -> Counterclockwise
func orientation(p, q, r Point) int {
	val := (q.Longitude-p.Longitude)*(r.Latitude-q.Latitude) - (q.Latitude-p.Latitude)*(r.Longitude-q.Longitude)
	if val == 0 {
		return 0
	}
	if val > 0 {
		return 1
	}
	return -1
}

// Distance between two points (Euclidean distance)
func dist(p1, p2 Point) float64 {
	return math.Sqrt((p2.Latitude-p1.Latitude)*(p2.Latitude-p1.Latitude) + (p2.Longitude-p1.Longitude)*(p2.Longitude-p1.Longitude))
}

// ConvexHull function using Graham scan algorithm
func ConvexHull(points []Point) []Point {
	n := len(points)
	if n < 3 {
		return points
	}

	// Step 1: Find the point with the lowest Latitude (in case of tie, the leftmost Longitude)
	p0 := points[0]
	for i := 1; i < n; i++ {
		if points[i].Latitude < p0.Latitude || (points[i].Latitude == p0.Latitude && points[i].Longitude < p0.Longitude) {
			p0 = points[i]
		}
	}

	// Step 2: Sort the points based on polar angle with p0
	sort.Slice(points, func(i, j int) bool {
		o := orientation(p0, points[i], points[j])
		if o == 0 {
			return dist(p0, points[i]) < dist(p0, points[j])
		}
		return o == -1
	})

	// Step 3: Build the convex hull using a stack
	hull := []Point{p0, points[1], points[2]}

	for i := 3; i < n; i++ {
		for len(hull) > 1 && orientation(hull[len(hull)-2], hull[len(hull)-1], points[i]) != -1 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, points[i])
	}

	return hull
}

func ConcatPoints(slices ...[]Point) []Point {
	var result []Point
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
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

func pointsToPolygonWKT(points []Point) string {
	// ポリゴンが閉じているかチェックし、閉じていない場合は閉じる
	if len(points) > 1 && (points[0].Latitude != points[len(points)-1].Latitude || points[0].Longitude != points[len(points)-1].Longitude) {
		points = append(points, points[0])
	}

	var coords []string
	for _, point := range points {
		coords = append(coords, fmt.Sprintf("%f %f", point.Longitude, point.Latitude))
	}

	// POLYGON WKT形式に変換
	wkt := fmt.Sprintf("((%s))", strings.Join(coords, ", "))
	return wkt
}

// [][]PointからWKT形式のMULTIPOLYGONを作成する関数
func multiPolygonToWKT(multiPolygon [][]Point) string {
	var polygons []string
	for _, polygon := range multiPolygon {
		polygons = append(polygons, pointsToPolygonWKT(polygon))
	}

	// MULTIPOLYGON WKT形式に変換
	wkt := fmt.Sprintf("MULTIPOLYGON(%s)", strings.Join(polygons, ", "))
	return wkt
}

func wktToPolygonPoints(wkt string) ([]Point, error) {
	// WKTからPOLYGONの座標部分を抽出
	wkt = strings.TrimPrefix(wkt, "POLYGON((")
	wkt = strings.TrimSuffix(wkt, "))")
	coordPairs := strings.Split(wkt, ", ")

	var polygon []Point

	// 各座標ペアを処理
	for _, pair := range coordPairs {
		coords := strings.Split(pair, " ")
		if len(coords) != 2 {
			// 詳細なエラーメッセージを出力
			fmt.Printf("無効な座標ペア: %s\n", pair)
			continue // 無効な座標ペアをスキップ
		}

		// 経度と緯度をfloat64に変換
		lon, err := strconv.ParseFloat(coords[0], 64)
		if err != nil {
			return nil, fmt.Errorf("無効な経度値: %v", err)
		}
		lat, err := strconv.ParseFloat(coords[1], 64)
		if err != nil {
			return nil, fmt.Errorf("無効な緯度値: %v", err)
		}

		// Point構造体に追加
		polygon = append(polygon, Point{Latitude: lat, Longitude: lon})
	}

	return polygon, nil
}

func wktToMultiPolygonPoints(wkt string) ([][]Point, error) {
	// WKTからMULTIPOLYGONの座標部分を抽出
	wkt = strings.TrimPrefix(wkt, "MULTIPOLYGON(")
	wkt = strings.TrimSuffix(wkt, ")")
	polygonStrs := strings.Split(wkt, ")), ((")

	var multiPolygon [][]Point

	// 各POLYGONの座標を処理
	for _, polygonStr := range polygonStrs {
		polygonStr = strings.Trim(polygonStr, "()")
		coordPairs := strings.Split(polygonStr, ", ")

		var polygon []Point
		for _, pair := range coordPairs {
			coords := strings.Split(pair, " ")
			if len(coords) != 2 {
				return nil, fmt.Errorf("無効な座標ペア: %s", pair)
			}

			// 経度と緯度をfloat64に変換
			lon, err := strconv.ParseFloat(coords[0], 64)
			if err != nil {
				return nil, err
			}
			lat, err := strconv.ParseFloat(coords[1], 64)
			if err != nil {
				return nil, err
			}

			// Point構造体に追加
			polygon = append(polygon, Point{Latitude: lat, Longitude: lon})
		}
		multiPolygon = append(multiPolygon, polygon)
	}

	return multiPolygon, nil
}

func calcTyphoonPoints(typhoonCenterLat, typhoonCenterLon, wideAreaRadius, narrowAreaRadius, wideAreaBearing float64, numPoints int) []Point {
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

	return points
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

	stormAreas := [][]Point{
		ConvexHull(ConcatPoints(
			calcTyphoonPoints(22.3, 140.9, 55., 55., 0., 120),
			calcTyphoonPoints(24.9, 139.6, 130., 130., 0., 120),
		)),
		ConvexHull(ConcatPoints(
			calcTyphoonPoints(24.9, 139.6, 130., 130., 0., 120),
			calcTyphoonPoints(26.8, 137.8, 190., 190., 0., 120),
		)),
		ConvexHull(ConcatPoints(
			calcTyphoonPoints(26.8, 137.8, 190., 190., 0., 120),
			calcTyphoonPoints(29.2, 133.7, 310., 310., 0., 120),
		)),
		ConvexHull(ConcatPoints(
			calcTyphoonPoints(29.2, 133.7, 310., 310., 0., 120),
			calcTyphoonPoints(32.2, 133.3, 360., 360., 0., 120),
		)),
	}

	wkt := multiPolygonToWKT(stormAreas)
	geom, err := geos.NewGeomFromWKT(wkt)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}
	buffered := geom.Buffer(0, 32)

	fmt.Println(buffered)

	bufferedWKT := buffered.ToWKT()

	bufferedPoints, err := wktToPolygonPoints(bufferedWKT)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}

	typhoons := []Typhoon{
		// 実況
		// calcTyphoonPolygon(22.3, 140.9, 330., 220., 0., 120), // 強風域
		// calcTyphoonPolygon(22.3, 140.9, 55., 55., 0., 120), // 暴風域
		// 予報　１２時間後
		calcTyphoonPolygon(24.9, 139.6, 75., 75., 0., 120), // 予報円
		// calcTyphoonPolygon(24.9, 139.6, 130., 130., 0., 120), // 暴風警戒域
		// 予報　２４時間後
		calcTyphoonPolygon(26.8, 137.8, 105., 105., 0., 120), // 予報円
		// calcTyphoonPolygon(26.8, 137.8, 190., 190., 0., 120), // 暴風警戒域
		// 予報　４８時間後
		calcTyphoonPolygon(29.2, 133.7, 155., 155., 0., 120), // 予報円
		// calcTyphoonPolygon(29.2, 133.7, 310., 310., 0., 120), // 暴風警戒域
		// 予報　７２時間後
		calcTyphoonPolygon(32.2, 133.3, 220., 220., 0., 120), // 予報円
		// calcTyphoonPolygon(32.2, 133.3, 360., 360., 0., 120), // 暴風警戒域
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

	geojsonPoints := make([][]float64, 0, len(bufferedPoints)+1)
	for _, coordinate := range bufferedPoints {
		geojsonPoints = append(
			geojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}
	coordinates := [][][]float64{geojsonPoints}
	polygon := geojson.NewPolygonFeature(coordinates)
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

	dis := haversineDistance(22.3, 140.9, 24.9, 139.6)
	fmt.Println("dis = ", dis)
	theta := calculateTheta(29.2, 133.7, 32.2, 133.3)
	fmt.Println("theta = ", theta)
}
