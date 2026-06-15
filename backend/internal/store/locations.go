package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Location struct {
	Latitude  float64
	Longitude float64
	Timestamp time.Time
}

func (db *DB) InsertLocation(ctx context.Context, vehicleID string, lat, lng float64, ts time.Time) error {
	const q = `
		INSERT INTO vehicle_locations (vehicle_id, point, ts)
		VALUES ($1,
		        ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
		        $4)`
	_, err := db.Pool.Exec(ctx, q, vehicleID, lng, lat, ts)
	return err
}

func (db *DB) LatestLocation(ctx context.Context, vehicleID string) (Location, error) {
	var loc Location
	const q = `
		SELECT ST_Y(point::geometry) AS lat,
		       ST_X(point::geometry) AS lng,
		       ts
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY ts DESC
		LIMIT 1`
	err := db.Pool.QueryRow(ctx, q, vehicleID).Scan(&loc.Latitude, &loc.Longitude, &loc.Timestamp)
	if errors.Is(err, pgx.ErrNoRows) {
		return loc, ErrNotFound
	}
	return loc, err
}

// GeofenceState mutation helpers ------------------------------------------

func (db *DB) CurrentGeofenceState(ctx context.Context, vehicleID string) (map[string]struct{}, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT geofence_id FROM vehicle_geofence_state WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]struct{}{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = struct{}{}
	}
	return out, rows.Err()
}

func (db *DB) AddGeofenceState(ctx context.Context, vehicleID string, geofenceIDs []string) error {
	if len(geofenceIDs) == 0 {
		return nil
	}
	const q = `
		INSERT INTO vehicle_geofence_state (vehicle_id, geofence_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`
	batch := &pgx.Batch{}
	for _, gid := range geofenceIDs {
		batch.Queue(q, vehicleID, gid)
	}
	br := db.Pool.SendBatch(ctx, batch)
	defer br.Close()
	for range geofenceIDs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) RemoveGeofenceState(ctx context.Context, vehicleID string, geofenceIDs []string) error {
	if len(geofenceIDs) == 0 {
		return nil
	}
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM vehicle_geofence_state
		 WHERE vehicle_id = $1 AND geofence_id = ANY($2)`,
		vehicleID, geofenceIDs)
	return err
}
