package usecase

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"typhoon-polygon/model"

	geojson "github.com/paulmach/go.geojson"
)

// 定数: 地球の半径 (キロメートル)
const EarthRadius = 6378.137

// Helper function to determine the orientation of three points
// 0 -> p, q and r are collinear
// 1 -> Clockwise
// -1 -> Counterclockwise
func orientation(p, q, r model.Point) int {
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
func dist(p1, p2 model.Point) float64 {
	return math.Sqrt((p2.Latitude-p1.Latitude)*(p2.Latitude-p1.Latitude) + (p2.Longitude-p1.Longitude)*(p2.Longitude-p1.Longitude))
}

// ConvexHull function using Graham scan algorithm
func ConvexHull(points []model.Point) []model.Point {
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
	hull := []model.Point{p0, points[1], points[2]}

	for i := 3; i < n; i++ {
		for len(hull) > 1 && orientation(hull[len(hull)-2], hull[len(hull)-1], points[i]) != -1 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, points[i])
	}

	return hull
}

func ConcatPoints(slices ...[]model.Point) []model.Point {
	var result []model.Point
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
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
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
func CalculateTheta(lat1, lon1, lat2, lon2 float64) float64 {
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
func CalcCirclePoint(centerLat, centerLon, radius, theta float64) model.Point {
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

	return model.Point{
		Latitude:  circleLat,
		Longitude: circleLon,
	}
}

func PointsToPolygonWKT(points []model.Point) string {
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
func MultiPolygonToWKT(multiPolygon [][]model.Point) string {
	var polygons []string
	for _, polygon := range multiPolygon {
		polygons = append(polygons, PointsToPolygonWKT(polygon))
	}

	// MULTIPOLYGON WKT形式に変換
	wkt := fmt.Sprintf("MULTIPOLYGON(%s)", strings.Join(polygons, ", "))
	return wkt
}

func WktToPolygonPoints(wkt string) ([]model.Point, error) {
	// WKTからPOLYGONの座標部分を抽出
	wkt = strings.TrimPrefix(wkt, "POLYGON((")
	wkt = strings.TrimSuffix(wkt, "))")
	coordPairs := strings.Split(wkt, ", ")

	var polygon []model.Point

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
		polygon = append(polygon, model.Point{Latitude: lat, Longitude: lon})
	}

	return polygon, nil
}

func WktToMultiPolygonPoints(wkt string) ([][]model.Point, error) {
	// WKTからMULTIPOLYGONの座標部分を抽出
	wkt = strings.TrimPrefix(wkt, "MULTIPOLYGON(")
	wkt = strings.TrimSuffix(wkt, ")")
	polygonStrs := strings.Split(wkt, ")), ((")

	var multiPolygon [][]model.Point

	// 各POLYGONの座標を処理
	for _, polygonStr := range polygonStrs {
		polygonStr = strings.Trim(polygonStr, "()")
		coordPairs := strings.Split(polygonStr, ", ")

		var polygon []model.Point
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
			polygon = append(polygon, model.Point{Latitude: lat, Longitude: lon})
		}
		multiPolygon = append(multiPolygon, polygon)
	}

	return multiPolygon, nil
}

func CalcTyphoonPoints(typhoonCenterLat, typhoonCenterLon, wideAreaRadius, narrowAreaRadius, wideAreaBearing float64, numPoints int) []model.Point {
	points := make([]model.Point, 0, numPoints+1)

	// 円の中心は、台風の中心からwideAreaBearingの方角に、
	// 広域の半径(wideAreaRadius)から円の半径((wideAreaRadius + narrowAreaRadius) / 2.)を引いた距離
	// だけ進めば円の中心の緯度経度になる
	circleRadius := (wideAreaRadius + narrowAreaRadius) / 2.
	circleCenterPoint := CalcCirclePoint(typhoonCenterLat, typhoonCenterLon, wideAreaRadius-circleRadius, wideAreaBearing)

	for i := 0; i <= numPoints; i++ {
		angle := 360 * float64(i) / float64(numPoints)
		circlePoint := CalcCirclePoint(circleCenterPoint.Latitude, circleCenterPoint.Longitude, circleRadius, angle)
		points = append(points, circlePoint)
	}

	return points
}

func SaveGeoJSONToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

func MakeGeojsonPolygon(points []model.Point) *geojson.Feature {
	geoJsonPoints := make([][]float64, 0, len(points)+1)
	for _, coordinate := range points {
		geoJsonPoints = append(
			geoJsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}
	coordinates := [][][]float64{geoJsonPoints}
	polygon := geojson.NewPolygonFeature(coordinates)
	return polygon
}

func MakeGeojsonLineString(points []model.Point) *geojson.Feature {
	geojsonPoints := make([][]float64, 0, len(points)+1)
	for _, coordinate := range points {
		geojsonPoints = append(
			geojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}
	lineString := geojson.NewLineStringFeature(geojsonPoints)
	return lineString
}
