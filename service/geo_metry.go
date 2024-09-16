package service

import (
	"log"
	"typhoon-polygon/model"
	"typhoon-polygon/usecase"

	"github.com/twpayne/go-geos"
)

func CalcStormAreaPolygon(stormAreaTimeSeries []model.StormArea) []model.Point {
	stormAreas := [][]model.Point{}

	for i := range stormAreaTimeSeries[:len(stormAreaTimeSeries)-1] {
		stormAreas = append(
			stormAreas,
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

	wkt := usecase.MultiPolygonToWKT(stormAreas)
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
