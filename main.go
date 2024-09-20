package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"typhoon-polygon/model"
	"typhoon-polygon/service"
	"typhoon-polygon/usecase"

	geojson "github.com/paulmach/go.geojson"
)

func main() {
	// 検索するディレクトリ
	dir := "./json"

	var jsonFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// .json拡張子のファイルを見つけたらリストに追加
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			jsonFiles = append(jsonFiles, path)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, path := range jsonFiles {
		// JSONファイルを読み込む
		fmt.Println(path)
		jsonFile, err := os.Open(path)
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
			if typhoon.TargetTimestampType == "実況" {
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
		if len(forecastCircleTimeSeries) > 1 {
			forecastCirclePolygons := service.CalcForecastCirclePolygons(forecastCircleTimeSeries)
			for _, circle := range forecastCirclePolygons.ForecastCircles {
				polygon := usecase.MakeGeojsonPolygon(circle)
				featureCollection.AddFeature(polygon)
			}
			if len(forecastCirclePolygons.ForecastCircleBorder) > 0 {
				forcastCircleBorderPolygon := usecase.MakeGeojsonPolygon(forecastCirclePolygons.ForecastCircleBorder)
				featureCollection.AddFeature(forcastCircleBorderPolygon)
			}

			centerLineLineString := usecase.MakeGeojsonLineString(forecastCirclePolygons.CenterLine)
			featureCollection.AddFeature(centerLineLineString)
		}

		// GeoJSONとしてエンコード
		geoJSON, err := json.MarshalIndent(featureCollection, "", "  ")
		if err != nil {
			fmt.Println("Error encoding GeoJSON:", err)
			return
		}

		// GeoJSONを出力
		// fmt.Println(string(geoJSON))

		// ファイルに保存する
		savePath := strings.Replace(path, ".json", ".geojson", 1)
		savePath = strings.Replace(savePath, "json/", "geojson/", 1)
		err = usecase.SaveGeoJSONToFile(savePath, geoJSON)
		if err != nil {
			fmt.Println("Error saving GeoJSON to file:", err)
			return
		}

		fmt.Printf("GeoJSON successfully written to %s\n", savePath)
	}
}
