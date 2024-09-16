package service

import (
	"log"
	"typhoon-polygon/model"
	"typhoon-polygon/usecase"

	"github.com/twpayne/go-geos"
)

func CalcStormAreaPolygon(stormAreaTimeSeries []model.StormArea) []model.Point {
	stormAreaPairs := [][]model.Point{}

	for i := range stormAreaTimeSeries[:len(stormAreaTimeSeries)-1] {
		stormAreaPairs = append(
			stormAreaPairs,
			usecase.ConvexHull(usecase.ConcatPoints(
				usecase.CalcTyphoonPoints(
					stormAreaTimeSeries[i].CenterPoint.Latitude,
					stormAreaTimeSeries[i].CenterPoint.Longitude,
					stormAreaTimeSeries[i].CircleLongRadius,
					stormAreaTimeSeries[i].CircleShortRadius,
					stormAreaTimeSeries[i].CircleLongDirection,
					120,
				),
				usecase.CalcTyphoonPoints(
					stormAreaTimeSeries[i+1].CenterPoint.Latitude,
					stormAreaTimeSeries[i+1].CenterPoint.Longitude,
					stormAreaTimeSeries[i+1].CircleLongRadius,
					stormAreaTimeSeries[i+1].CircleShortRadius,
					stormAreaTimeSeries[i+1].CircleLongDirection,
					120,
				),
			)),
		)
	}

	wkt := usecase.MultiPolygonToWKT(stormAreaPairs)
	geom, err := geos.NewGeomFromWKT(wkt)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}
	buffered := geom.Buffer(0, 32)

	bufferedWKT := buffered.ToWKT()

	bufferedPoints, err := usecase.WktToPolygonPoints(bufferedWKT)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}

	return bufferedPoints
}

func CalcForecastCirclePolygons(forecastCircleTimeSeries []model.ForecastCircle) model.ForecastCirclePolygons {
	forecastCircles := [][]model.Point{}
	centerLine := []model.Point{}

	for _, v := range forecastCircleTimeSeries {
		forecastCircles = append(
			forecastCircles,
			usecase.CalcTyphoonPoints(
				v.CenterPoint.Latitude,
				v.CenterPoint.Longitude,
				v.CircleLongRadius,
				v.CircleShortRadius,
				v.CircleLongDirection,
				120,
			),
		)
		centerLine = append(
			centerLine,
			model.Point{
				Latitude:  v.CenterPoint.Latitude,
				Longitude: v.CenterPoint.Longitude,
			},
		)
	}

	forecastCirclePairs := [][]model.Point{}

	for i := range forecastCircles[:len(forecastCircles)-1] {
		forecastCirclePairs = append(
			forecastCirclePairs,
			usecase.ConvexHull(usecase.ConcatPoints(
				forecastCircles[i],
				forecastCircles[i+1],
			)),
		)
	}

	wkt := usecase.MultiPolygonToWKT(forecastCirclePairs)
	geom, err := geos.NewGeomFromWKT(wkt)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}
	buffered := geom.Buffer(0, 32)

	bufferedWKT := buffered.ToWKT()

	bufferedPoints, err := usecase.WktToPolygonPoints(bufferedWKT)
	if err != nil {
		log.Fatalf("エラー: %v", err)
	}

	return model.ForecastCirclePolygons{
		ForecastCircles:      forecastCircles,
		ForecastCircleBorder: bufferedPoints,
		CenterLine:           centerLine,
	}
}
