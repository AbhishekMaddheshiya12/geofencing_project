package geo

import (
	"errors"
	"fmt"
	"strings"
)

// ValidatePolygon enforces the spec's polygon rules:
//   - at least 4 points (3 unique + closing duplicate)
//   - first and last coordinates are identical
//   - latitude in [-90, 90], longitude in [-180, 180]
//
// Coordinates are expected as [latitude, longitude] pairs (per the spec).
func ValidatePolygon(coords [][]float64) error {
	if len(coords) < 4 {
		return fmt.Errorf("polygon requires at least 4 points (got %d)", len(coords))
	}
	for i, p := range coords {
		if len(p) != 2 {
			return fmt.Errorf("point %d must have exactly 2 elements [lat,lng]", i)
		}
		lat, lng := p[0], p[1]
		if lat < -90 || lat > 90 {
			return fmt.Errorf("point %d latitude out of range: %v", i, lat)
		}
		if lng < -180 || lng > 180 {
			return fmt.Errorf("point %d longitude out of range: %v", i, lng)
		}
	}
	first, last := coords[0], coords[len(coords)-1]
	if first[0] != last[0] || first[1] != last[1] {
		return errors.New("first and last coordinates must be identical (closed polygon)")
	}
	return nil
}

// ToWKT converts [[lat,lng], ...] coordinates into a PostGIS WKT string,
// flipping to [lng lat] order which PostGIS expects.
func ToWKT(coords [][]float64) string {
	var b strings.Builder
	b.WriteString("POLYGON((")
	for i, p := range coords {
		if i > 0 {
			b.WriteString(", ")
		}
		// lng first, then lat
		fmt.Fprintf(&b, "%.8f %.8f", p[1], p[0])
	}
	b.WriteString("))")
	return b.String()
}
