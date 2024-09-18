package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"typhoon-polygon/model"
	"typhoon-polygon/service"
	"typhoon-polygon/usecase"

	geojson "github.com/paulmach/go.geojson"
)

func main() {

	// JSONファイルを読み込む
	jsonFile, err := os.Open("json/20240918130513_0_VPTW62_010000.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	// JSONファイルの内容を読み込む
	byteValue, _ := io.ReadAll(jsonFile)

	// データを構造体にデコード
	var typhoons []model.Typhoon
	err = json.Unmarshal(byteValue, &typhoons)
	if err != nil {
		log.Fatal(err)
	}

	stormAreaTimeSeries := []model.StormArea{}
	forecastCircleTimeSeries := []model.ForecastCircle{}

	for _, typhoon := range typhoons {
		for _, warningArea := range typhoon.WarningAreas {
			if warningArea.WarningAreaType == "暴風域" || warningArea.WarningAreaType == "暴風警戒域" {
				if warningArea.CircleLongRadius == 0 {
					continue
				}
				stormAreaTimeSeries = append(
					stormAreaTimeSeries,
					model.StormArea{
						CenterPoint:          model.Point{Latitude: typhoon.Latitude, Longitude: typhoon.Longitude},
						CircleLongDirection:  usecase.DirectionToDegrees(warningArea.CircleLongDirection),
						CircleLongRadius:     float64(warningArea.CircleLongRadius),
						CircleShortDirection: usecase.DirectionToDegrees(warningArea.CircleShortDirection),
						CircleShortRadius:    float64(warningArea.CircleShortRadius),
					},
				)
			}
			if warningArea.WarningAreaType == "強風域" {
				forecastCircleTimeSeries = append(
					forecastCircleTimeSeries,
					model.ForecastCircle{
						CenterPoint:          model.Point{Latitude: typhoon.Latitude, Longitude: typhoon.Longitude},
						CircleLongDirection:  float64(0),
						CircleLongRadius:     float64(0),
						CircleShortDirection: float64(0),
						CircleShortRadius:    float64(0),
					},
				)
			}
			if warningArea.WarningAreaType == "予報円" {
				if warningArea.CircleLongRadius == 0 {
					continue
				}
				forecastCircleTimeSeries = append(
					forecastCircleTimeSeries,
					model.ForecastCircle{
						CenterPoint:          model.Point{Latitude: typhoon.Latitude, Longitude: typhoon.Longitude},
						CircleLongDirection:  usecase.DirectionToDegrees(warningArea.CircleLongDirection),
						CircleLongRadius:     float64(warningArea.CircleLongRadius),
						CircleShortDirection: usecase.DirectionToDegrees(warningArea.CircleShortDirection),
						CircleShortRadius:    float64(warningArea.CircleShortRadius),
					},
				)
			}
		}
	}

	featureCollection := geojson.NewFeatureCollection()

	// 暴風域のGeoJson追加
	if len(stormAreaTimeSeries) > 0 {
		stormAreaBorderPoints := service.CalcStormAreaPolygon(stormAreaTimeSeries)
		stormAreaBorderPolygon := usecase.MakeGeojsonPolygon(stormAreaBorderPoints)
		featureCollection.AddFeature(stormAreaBorderPolygon)
	}

	// 予報円のGeoJson追加
	forecastCirclePolygons := service.CalcForecastCirclePolygons(forecastCircleTimeSeries)
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
