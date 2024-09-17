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
			CenterPoint:          model.Point{Latitude: 17.1, Longitude: 124.4},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.1,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.1,
		},
		{
			CenterPoint:          model.Point{Latitude: 16.8, Longitude: 118.0},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.1,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.1,
		},
		{
			CenterPoint:          model.Point{Latitude: 16.4, Longitude: 114.0},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.1,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.1,
		},
		{
			CenterPoint:          model.Point{Latitude: 16.9, Longitude: 111.2},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.1,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.1,
		},
		{
			CenterPoint:          model.Point{Latitude: 18.8, Longitude: 108.9},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.1,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.1,
		},
	}

	forecastCircleTimeSeries := []model.ForecastCircle{
		{
			CenterPoint:          model.Point{Latitude: 17.1, Longitude: 124.4},
			CircleLongDirection:  0.,
			CircleLongRadius:     0.,
			CircleShortDirection: 0.,
			CircleShortRadius:    0.,
		},
		{
			CenterPoint:          model.Point{Latitude: 16.8, Longitude: 118.0},
			CircleLongDirection:  0.,
			CircleLongRadius:     150.,
			CircleShortDirection: 0.,
			CircleShortRadius:    150.,
		},
		{
			CenterPoint:          model.Point{Latitude: 16.4, Longitude: 114.0},
			CircleLongDirection:  0.,
			CircleLongRadius:     240.,
			CircleShortDirection: 0.,
			CircleShortRadius:    240.,
		},
		{
			CenterPoint:          model.Point{Latitude: 16.9, Longitude: 111.2},
			CircleLongDirection:  0.,
			CircleLongRadius:     310.,
			CircleShortDirection: 0.,
			CircleShortRadius:    310.,
		},
		{
			CenterPoint:          model.Point{Latitude: 18.8, Longitude: 108.9},
			CircleLongDirection:  0.,
			CircleLongRadius:     390.,
			CircleShortDirection: 0.,
			CircleShortRadius:    390.,
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
