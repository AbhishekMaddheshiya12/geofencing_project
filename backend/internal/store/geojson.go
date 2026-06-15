package store

import (
	"encoding/json"
	"fmt"
)

// geoJSONPolygonToLatLng decodes a PostGIS ST_AsGeoJSON polygon (which is
// in [lng, lat] order) back into [lat, lng] pairs for the API response.
func geoJSONPolygonToLatLng(s string) ([][]float64, error) {
	var feature struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(s), &feature); err != nil {
		return nil, err
	}
	// Polygon coordinates: [[[lng,lat], [lng,lat], ...]]
	var rings [][][]float64
	if err := json.Unmarshal(feature.Coordinates, &rings); err != nil {
		return nil, err
	}
	if len(rings) == 0 {
		return nil, fmt.Errorf("empty polygon")
	}
	ring := rings[0]
	out := make([][]float64, len(ring))
	for i, p := range ring {
		if len(p) < 2 {
			return nil, fmt.Errorf("bad point %d", i)
		}
		out[i] = []float64{p[1], p[0]} // flip back to [lat,lng]
	}
	return out, nil
}
