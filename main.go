package main

import (
	"encoding/json"
	"fmt"
	"typhoon-polygon/model"
	"typhoon-polygon/service"
	"typhoon-polygon/usecase"

	geojson "github.com/paulmach/go.geojson"
)

func main() {

	stormAreaTimeSeries := []model.StormArea{
		model.StormArea{
			CenterPoint:          model.Point{Latitude: 22.3, Longitude: 140.9},
			CircleLongDirection:  0.,
			CircleLongRadius:     55.,
			CircleShortDirection: 0.,
			CircleShortRadius:    55.,
		},
		model.StormArea{
			CenterPoint:          model.Point{Latitude: 24.9, Longitude: 139.6},
			CircleLongDirection:  0.,
			CircleLongRadius:     130.,
			CircleShortDirection: 0.,
			CircleShortRadius:    130.,
		},
		model.StormArea{
			CenterPoint:          model.Point{Latitude: 26.8, Longitude: 137.8},
			CircleLongDirection:  0.,
			CircleLongRadius:     190.,
			CircleShortDirection: 0.,
			CircleShortRadius:    190.,
		},
		model.StormArea{
			CenterPoint:          model.Point{Latitude: 29.2, Longitude: 133.7},
			CircleLongDirection:  0.,
			CircleLongRadius:     310.,
			CircleShortDirection: 0.,
			CircleShortRadius:    310.,
		},
		model.StormArea{
			CenterPoint:          model.Point{Latitude: 32.2, Longitude: 133.3},
			CircleLongDirection:  0.,
			CircleLongRadius:     360.,
			CircleShortDirection: 0.,
			CircleShortRadius:    360.,
		},
	}

	stormAreaBorderPoints := service.CalcStormAreaPolygon(stormAreaTimeSeries)

	typhoons := []model.TyphoonPolygon{
		// 実況
		// CalcTyphoonPolygon(22.3, 140.9, 330., 220., 0., 120), // 強風域
		usecase.CalcTyphoonPolygon(22.3, 140.9, 55., 55., 0., 120), // 暴風域
		// 予報　１２時間後
		usecase.CalcTyphoonPolygon(24.9, 139.6, 75., 75., 0., 120), // 予報円
		// CalcTyphoonPolygon(24.9, 139.6, 130., 130., 0., 120), // 暴風警戒域
		// 予報　２４時間後
		usecase.CalcTyphoonPolygon(26.8, 137.8, 105., 105., 0., 120), // 予報円
		// CalcTyphoonPolygon(26.8, 137.8, 190., 190., 0., 120), // 暴風警戒域
		// 予報　４８時間後
		usecase.CalcTyphoonPolygon(29.2, 133.7, 155., 155., 0., 120), // 予報円
		// CalcTyphoonPolygon(29.2, 133.7, 310., 310., 0., 120), // 暴風警戒域
		// 予報　７２時間後
		usecase.CalcTyphoonPolygon(32.2, 133.3, 220., 220., 0., 120), // 予報円
		// CalcTyphoonPolygon(32.2, 133.3, 360., 360., 0., 120), // 暴風警戒域
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

	geojsonPoints := make([][]float64, 0, len(stormAreaBorderPoints)+1)
	for _, coordinate := range stormAreaBorderPoints {
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
	err = usecase.SaveGeoJSONToFile("output.geojson", geoJSON)
	if err != nil {
		fmt.Println("Error saving GeoJSON to file:", err)
		return
	}

	fmt.Println("GeoJSON successfully written to output.geojson")
}
