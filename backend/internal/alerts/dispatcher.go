package alerts

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/mapup/geofence/internal/ids"
	"github.com/mapup/geofence/internal/models"
	"github.com/mapup/geofence/internal/store"
	"github.com/mapup/geofence/internal/ws"
)

// Diff computes entered/exited geofence sets given current and previous sets.
// Entered = current ∖ previous, Exited = previous ∖ current.
func Diff(previous, current map[string]struct{}) (entered, exited []string) {
	for id := range current {
		if _, ok := previous[id]; !ok {
			entered = append(entered, id)
		}
	}
	for id := range previous {
		if _, ok := current[id]; !ok {
			exited = append(exited, id)
		}
	}
	return entered, exited
}

type Dispatcher struct {
	DB  *store.DB
	Hub *ws.Hub
}

// HandleEvents records every entry/exit event in `violations` and pushes
// a WebSocket alert for those matched by an active alert_configs row.
// Runs async — the caller (location handler) should call this in a goroutine.
func (d *Dispatcher) HandleEvents(ctx context.Context, veh store.VehicleLite, lat, lng float64, ts time.Time, entered, exited []string) {
	process := func(geofenceID, eventType string) {
		g, err := d.DB.GetGeofenceLite(ctx, geofenceID)
		if err != nil {
			slog.Error("dispatch: get geofence", "id", geofenceID, "err", err)
			return
		}
		if _, err := d.DB.InsertViolation(ctx, store.ViolationInput{
			VehicleID:  veh.ID,
			GeofenceID: geofenceID,
			EventType:  eventType,
			Latitude:   lat,
			Longitude:  lng,
			Timestamp:  ts,
		}); err != nil {
			slog.Error("dispatch: insert violation", "err", err)
		}

		matches, err := d.DB.MatchingConfigs(ctx, geofenceID, veh.ID, eventType)
		if err != nil {
			slog.Error("dispatch: match configs", "err", err)
			return
		}
		if matches == 0 {
			return
		}

		env := models.AlertEnvelope{
			EventID:   ids.Event(),
			EventType: eventType,
			Timestamp: ts,
		}
		env.Vehicle.VehicleID = veh.ID
		env.Vehicle.VehicleNumber = veh.VehicleNumber
		env.Vehicle.DriverName = veh.DriverName
		env.Geofence.GeofenceID = g.ID
		env.Geofence.GeofenceName = g.Name
		env.Geofence.Category = g.Category
		env.Location.Latitude = lat
		env.Location.Longitude = lng

		payload, err := json.Marshal(env)
		if err != nil {
			slog.Error("dispatch: marshal", "err", err)
			return
		}
		d.Hub.Broadcast(payload)
	}

	for _, id := range entered {
		process(id, "entry")
	}
	for _, id := range exited {
		process(id, "exit")
	}
}
