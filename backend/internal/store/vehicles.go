package store

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mapup/geofence/internal/ids"
	"github.com/mapup/geofence/internal/models"
)

func (db *DB) CreateVehicle(ctx context.Context, v *models.Vehicle) error {
	v.ID = ids.Vehicle()
	v.Status = "active"
	const q = `
		INSERT INTO vehicles (id, vehicle_number, driver_name, vehicle_type, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`
	err := db.Pool.QueryRow(ctx, q, v.ID, v.VehicleNumber, v.DriverName, v.VehicleType, v.Phone, v.Status).
		Scan(&v.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return ErrConflict
	}
	return err
}

func (db *DB) ListVehicles(ctx context.Context, limit, offset int) ([]models.Vehicle, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at
		 FROM vehicles
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.VehicleNumber, &v.DriverName, &v.VehicleType, &v.Phone, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

type VehicleLite struct {
	ID            string
	VehicleNumber string
	DriverName    string
}

func (db *DB) GetVehicleLite(ctx context.Context, id string) (VehicleLite, error) {
	var v VehicleLite
	err := db.Pool.QueryRow(ctx,
		`SELECT id, vehicle_number, driver_name FROM vehicles WHERE id = $1`, id).
		Scan(&v.ID, &v.VehicleNumber, &v.DriverName)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}
