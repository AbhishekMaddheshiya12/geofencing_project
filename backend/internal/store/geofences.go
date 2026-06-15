package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/mapup/geofence/internal/geo"
	"github.com/mapup/geofence/internal/ids"
	"github.com/mapup/geofence/internal/models"
)

func (db *DB) CreateGeofence(ctx context.Context, g *models.Geofence) error {
	g.ID = ids.Geofence()
	g.Status = "active"
	wkt := geo.ToWKT(g.Coordinates)
	const q = `
		INSERT INTO geofences (id, name, description, category, geom, status)
		VALUES ($1, $2, $3, $4, ST_GeogFromText('SRID=4326;' || $5), $6)
		RETURNING created_at`
	return db.Pool.QueryRow(ctx, q, g.ID, g.Name, g.Description, g.Category, wkt, g.Status).
		Scan(&g.CreatedAt)
}

// ListGeofences returns all geofences, optionally filtered by category.
// The coordinates are reconstructed from the PostGIS geometry as [lat,lng] pairs.
func (db *DB) ListGeofences(ctx context.Context, category string, limit, offset int) ([]models.Geofence, error) {
	q := `
		SELECT id, name, COALESCE(description, ''), category, status, created_at,
		       ST_AsGeoJSON(geom::geometry)
		FROM geofences
		WHERE ($1 = '' OR category = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := db.Pool.Query(ctx, q, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Geofence
	for rows.Next() {
		var g models.Geofence
		var geoJSON string
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Category, &g.Status, &g.CreatedAt, &geoJSON); err != nil {
			return nil, err
		}
		coords, err := geoJSONPolygonToLatLng(geoJSON)
		if err != nil {
			return nil, fmt.Errorf("decode geofence %s: %w", g.ID, err)
		}
		g.Coordinates = coords
		out = append(out, g)
	}
	return out, rows.Err()
}

// GetGeofenceByID is used by the alerts dispatcher to fill the alert envelope.
type GeofenceLite struct {
	ID       string
	Name     string
	Category string
}

func (db *DB) GetGeofenceLite(ctx context.Context, id string) (GeofenceLite, error) {
	var g GeofenceLite
	err := db.Pool.QueryRow(ctx,
		`SELECT id, name, category FROM geofences WHERE id = $1`, id).
		Scan(&g.ID, &g.Name, &g.Category)
	if errors.Is(err, pgx.ErrNoRows) {
		return g, ErrNotFound
	}
	return g, err
}

// GeofencesContainingPoint returns the IDs of all geofences whose polygon
// covers the given (lat, lng) point.
func (db *DB) GeofencesContainingPoint(ctx context.Context, lat, lng float64) ([]GeofenceLite, error) {
	const q = `
		SELECT id, name, category
		FROM geofences
		WHERE ST_Covers(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography)
		  AND status = 'active'`
	rows, err := db.Pool.Query(ctx, q, lng, lat) // lng first for ST_MakePoint(x, y)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GeofenceLite
	for rows.Next() {
		var g GeofenceLite
		if err := rows.Scan(&g.ID, &g.Name, &g.Category); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
