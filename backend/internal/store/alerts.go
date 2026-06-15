package store

import (
	"context"

	"github.com/mapup/geofence/internal/ids"
	"github.com/mapup/geofence/internal/models"
)

func (db *DB) CreateAlertConfig(ctx context.Context, a *models.AlertConfig) error {
	a.AlertID = ids.Alert()
	a.Status = "active"
	const q = `
		INSERT INTO alert_configs (id, geofence_id, vehicle_id, event_type, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`
	return db.Pool.QueryRow(ctx, q, a.AlertID, a.GeofenceID, a.VehicleID, a.EventType, a.Status).
		Scan(&a.CreatedAt)
}

func (db *DB) ListAlertConfigs(ctx context.Context, geofenceID, vehicleID string, limit, offset int) ([]models.AlertConfig, error) {
	const q = `
		SELECT ac.id, ac.geofence_id, g.name,
		       ac.vehicle_id, v.vehicle_number,
		       ac.event_type, ac.status, ac.created_at
		FROM alert_configs ac
		JOIN geofences g ON g.id = ac.geofence_id
		LEFT JOIN vehicles v ON v.id = ac.vehicle_id
		WHERE ($1 = '' OR ac.geofence_id = $1)
		  AND ($2 = '' OR ac.vehicle_id = $2)
		ORDER BY ac.created_at DESC
		LIMIT $3 OFFSET $4`
	rows, err := db.Pool.Query(ctx, q, geofenceID, vehicleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.AlertConfig
	for rows.Next() {
		var a models.AlertConfig
		if err := rows.Scan(&a.AlertID, &a.GeofenceID, &a.GeofenceName,
			&a.VehicleID, &a.VehicleNumber, &a.EventType, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// MatchingConfigs returns alert configs that should fire for (geofenceID, vehicleID, eventType).
// A config matches when:
//   - geofence_id equals geofenceID
//   - vehicle_id is NULL (applies to all vehicles) or equals vehicleID
//   - event_type is 'both' or equals eventType
//   - status is 'active'
func (db *DB) MatchingConfigs(ctx context.Context, geofenceID, vehicleID, eventType string) (int, error) {
	const q = `
		SELECT count(*) FROM alert_configs
		WHERE geofence_id = $1
		  AND (vehicle_id IS NULL OR vehicle_id = $2)
		  AND (event_type = 'both' OR event_type = $3)
		  AND status = 'active'`
	var n int
	err := db.Pool.QueryRow(ctx, q, geofenceID, vehicleID, eventType).Scan(&n)
	return n, err
}
