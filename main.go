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
		{
			CenterPoint:          model.Point{Latitude: 22.3, Longitude: 140.9},
			CircleLongDirection:  0.,
			CircleLongRadius:     55.,
			CircleShortDirection: 0.,
			CircleShortRadius:    55.,
		},
		{
			CenterPoint:          model.Point{Latitude: 24.9, Longitude: 139.6},
			CircleLongDirection:  0.,
			CircleLongRadius:     130.,
			CircleShortDirection: 0.,
			CircleShortRadius:    130.,
		},
		{
			CenterPoint:          model.Point{Latitude: 26.8, Longitude: 137.8},
			CircleLongDirection:  0.,
			CircleLongRadius:     190.,
			CircleShortDirection: 0.,
			CircleShortRadius:    190.,
		},
		{
			CenterPoint:          model.Point{Latitude: 29.2, Longitude: 133.7},
			CircleLongDirection:  0.,
			CircleLongRadius:     310.,
			CircleShortDirection: 0.,
			CircleShortRadius:    310.,
		},
		{
			CenterPoint:          model.Point{Latitude: 32.2, Longitude: 133.3},
			CircleLongDirection:  0.,
			CircleLongRadius:     360.,
			CircleShortDirection: 0.,
			CircleShortRadius:    360.,
		},
	}

	forecastCircleTimeSeries := []model.ForecastCircle{
		{
			CenterPoint:          model.Point{Latitude: 22.3, Longitude: 140.9},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.,
		},
		{
			CenterPoint:          model.Point{Latitude: 24.9, Longitude: 139.6},
			CircleLongDirection:  0.,
			CircleLongRadius:     75.,
			CircleShortDirection: 0.,
			CircleShortRadius:    75.,
		},
		{
			CenterPoint:          model.Point{Latitude: 26.8, Longitude: 137.8},
			CircleLongDirection:  0.,
			CircleLongRadius:     105.,
			CircleShortDirection: 0.,
			CircleShortRadius:    105.,
		},
		{
			CenterPoint:          model.Point{Latitude: 29.2, Longitude: 133.7},
			CircleLongDirection:  0.,
			CircleLongRadius:     155.,
			CircleShortDirection: 0.,
			CircleShortRadius:    155.,
		},
		{
			CenterPoint:          model.Point{Latitude: 32.2, Longitude: 133.3},
			CircleLongDirection:  0.,
			CircleLongRadius:     220.,
			CircleShortDirection: 0.,
			CircleShortRadius:    220.,
		},
	}

	stormAreaBorderPoints := service.CalcStormAreaPolygon(stormAreaTimeSeries)

	forecastCirclePolygons := service.CalcForecastCirclePolygons(forecastCircleTimeSeries)

	// GeoJsonファイルの作成

	// FeatureCollectionにPointとPolygonを追加
	featureCollection := geojson.NewFeatureCollection()

	stormAreaBorderGeojsonPoints := make([][]float64, 0, len(stormAreaBorderPoints)+1)
	for _, coordinate := range stormAreaBorderPoints {
		stormAreaBorderGeojsonPoints = append(
			stormAreaBorderGeojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}
	stormAreaBorderCoordinates := [][][]float64{stormAreaBorderGeojsonPoints}
	stormAreaBorderPolygon := geojson.NewPolygonFeature(stormAreaBorderCoordinates)
	featureCollection.AddFeature(stormAreaBorderPolygon)

	for _, circle := range forecastCirclePolygons.ForecastCircles {
		// Polygonの作成
		geojsonPoints := make([][]float64, 0, len(circle)+1)
		for _, coordinate := range circle {
			geojsonPoints = append(
				geojsonPoints,
				[]float64{coordinate.Longitude, coordinate.Latitude},
			)
		}
		coordinates := [][][]float64{geojsonPoints}
		polygon := geojson.NewPolygonFeature(coordinates)

		// // Pointの作成
		// point := geojson.NewPointFeature([]float64{typhoon.CenterPoint.Longitude, typhoon.CenterPoint.Latitude})

		// featureCollection.AddFeature(point)
		featureCollection.AddFeature(polygon)
	}

	forcastCircleBorderGeojsonPoints := make([][]float64, 0, len(forecastCirclePolygons.ForecastCircleBorder)+1)
	for _, coordinate := range forecastCirclePolygons.ForecastCircleBorder {
		forcastCircleBorderGeojsonPoints = append(
			forcastCircleBorderGeojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}
	forcastCircleBorderCoordinates := [][][]float64{forcastCircleBorderGeojsonPoints}
	forcastCircleBorderPolygon := geojson.NewPolygonFeature(forcastCircleBorderCoordinates)
	featureCollection.AddFeature(forcastCircleBorderPolygon)

	centerLineGeojsonPoints := make([][]float64, 0, len(forecastCirclePolygons.CenterLine)+1)
	for _, coordinate := range forecastCirclePolygons.CenterLine {
		centerLineGeojsonPoints = append(
			centerLineGeojsonPoints,
			[]float64{coordinate.Longitude, coordinate.Latitude},
		)
	}

	centerLineLineString := geojson.NewLineStringFeature(centerLineGeojsonPoints)
	featureCollection.AddFeature(centerLineLineString)

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
