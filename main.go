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

	stormAreaBorderPolygon := usecase.MakeGeojsonPolygon(stormAreaBorderPoints)
	featureCollection.AddFeature(stormAreaBorderPolygon)

	for _, circle := range forecastCirclePolygons.ForecastCircles {
		polygon := usecase.MakeGeojsonPolygon(circle)
		featureCollection.AddFeature(polygon)
	}

	forcastCircleBorderPolygon := usecase.MakeGeojsonPolygon(forecastCirclePolygons.ForecastCircleBorder)
	featureCollection.AddFeature(forcastCircleBorderPolygon)

	centerLineLineString := usecase.MakeGeojsonLineString(forecastCirclePolygons.CenterLine)
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
