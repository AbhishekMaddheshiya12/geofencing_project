package store

import (
	"context"
	"time"

	"github.com/mapup/geofence/internal/ids"
	"github.com/mapup/geofence/internal/models"
)

type ViolationInput struct {
	VehicleID  string
	GeofenceID string
	EventType  string
	Latitude   float64
	Longitude  float64
	Timestamp  time.Time
}

func (db *DB) InsertViolation(ctx context.Context, v ViolationInput) (string, error) {
	id := ids.Violation()
	const q = `
		INSERT INTO violations (id, vehicle_id, geofence_id, event_type, latitude, longitude, ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.Pool.Exec(ctx, q, id, v.VehicleID, v.GeofenceID, v.EventType, v.Latitude, v.Longitude, v.Timestamp)
	return id, err
}

type ViolationFilters struct {
	VehicleID  string
	GeofenceID string
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
}

func (db *DB) ListViolations(ctx context.Context, f ViolationFilters) ([]models.Violation, int, error) {
	const list = `
		SELECT vio.id, vio.vehicle_id, v.vehicle_number,
		       vio.geofence_id, g.name,
		       vio.event_type, vio.latitude, vio.longitude, vio.ts
		FROM violations vio
		JOIN vehicles v ON v.id = vio.vehicle_id
		JOIN geofences g ON g.id = vio.geofence_id
		WHERE ($1 = '' OR vio.vehicle_id = $1)
		  AND ($2 = '' OR vio.geofence_id = $2)
		  AND ($3::timestamptz IS NULL OR vio.ts >= $3)
		  AND ($4::timestamptz IS NULL OR vio.ts <= $4)
		ORDER BY vio.ts DESC
		LIMIT $5`
	const count = `
		SELECT count(*) FROM violations vio
		WHERE ($1 = '' OR vio.vehicle_id = $1)
		  AND ($2 = '' OR vio.geofence_id = $2)
		  AND ($3::timestamptz IS NULL OR vio.ts >= $3)
		  AND ($4::timestamptz IS NULL OR vio.ts <= $4)`

	rows, err := db.Pool.Query(ctx, list,
		f.VehicleID, f.GeofenceID, f.StartDate, f.EndDate, f.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []models.Violation
	for rows.Next() {
		var v models.Violation
		if err := rows.Scan(&v.ID, &v.VehicleID, &v.VehicleNumber,
			&v.GeofenceID, &v.GeofenceName,
			&v.EventType, &v.Latitude, &v.Longitude, &v.Timestamp); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	if err := db.Pool.QueryRow(ctx, count,
		f.VehicleID, f.GeofenceID, f.StartDate, f.EndDate).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
