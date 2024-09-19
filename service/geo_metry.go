package service

import (
	"log"
	"typhoon-polygon/model"
	"typhoon-polygon/usecase"

	"github.com/twpayne/go-geos"
)

func CalcStormAreaPolygon(stormAreaTimeSeries []model.StormArea) []model.Point {
	stormAreaPairs := [][]model.Point{}

	if len(stormAreaTimeSeries) >= 2 {
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
	} else if len(stormAreaTimeSeries) == 1 {
		// 暴風域が一つしかない場合はそれをそのままstormAreaPairsとする
		stormAreaPairs = append(
			stormAreaPairs,
			usecase.CalcTyphoonPoints(
				stormAreaTimeSeries[0].CenterPoint.Latitude,
				stormAreaTimeSeries[0].CenterPoint.Longitude,
				stormAreaTimeSeries[0].CircleLongRadius,
				stormAreaTimeSeries[0].CircleShortRadius,
				stormAreaTimeSeries[0].CircleLongDirection,
				120,
			),
		)
	} else {
		// 暴風域がない場合は空配列を返す
		return []model.Point{}
	}

	wkt := usecase.MultiPolygonToWKT(stormAreaPairs)
	geom, err := geos.NewGeomFromWKT(wkt)
	if err != nil {
		log.Fatalf("CalcStormAreaPolygon Error: %v, wkt: %v, stormAreaTimeSeries: %v", err, wkt, stormAreaTimeSeries)
	}
	buffered := geom.Buffer(0, 32)

	bufferedWKT := buffered.ToWKT()

	bufferedPoints, err := usecase.WktToPolygonPoints(bufferedWKT)
	if err != nil {
		log.Fatalf("CalcStormAreaPolygon Error: %v, polygon: %v, stormAreaTimeSeries: %v", err, bufferedWKT, stormAreaTimeSeries)
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

	if len(forecastCircles) >= 2 {
		for i := range forecastCircles[:len(forecastCircles)-1] {
			forecastCirclePairs = append(
				forecastCirclePairs,
				usecase.ConvexHull(usecase.ConcatPoints(
					forecastCircles[i],
					forecastCircles[i+1],
				)),
			)
		}
	} else if len(forecastCircles) == 1 {
		forecastCirclePairs = append(
			forecastCirclePairs,
			forecastCircles[0],
		)
	} else {
		return model.ForecastCirclePolygons{
			ForecastCircles:      forecastCircles,
			ForecastCircleBorder: []model.Point{},
			CenterLine:           centerLine,
		}
	}

	wkt := usecase.MultiPolygonToWKT(forecastCirclePairs)
	geom, err := geos.NewGeomFromWKT(wkt)
	if err != nil {
		log.Fatalf("CalcForecastCirclePolygons Error: %v, wkt: %v, forecastCircleTimeSeries: %v", err, wkt, forecastCircleTimeSeries)
	}
	buffered := geom.Buffer(0, 32)

	bufferedWKT := buffered.ToWKT()

	bufferedPoints, err := usecase.WktToPolygonPoints(bufferedWKT)
	if err != nil {
		log.Fatalf("CalcForecastCirclePolygons Error: %v, polygon: %v, forecastCircleTimeSeries: %v", err, bufferedWKT, forecastCircleTimeSeries)
	}

	forecastCircleBorder := bufferedPoints

	return model.ForecastCirclePolygons{
		ForecastCircles:      forecastCircles,
		ForecastCircleBorder: forecastCircleBorder,
		CenterLine:           centerLine,
	}
}
