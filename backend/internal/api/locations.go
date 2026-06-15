package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mapup/geofence/internal/store"
)

type locationUpdateReq struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}

type currentGeofenceOut struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Status       string `json:"status"`
}

func (s *Server) handleLocationUpdate(w http.ResponseWriter, r *http.Request) {
	var req locationUpdateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := requireNonEmpty(req.VehicleID, "vehicle_id"); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		WriteError(w, r, http.StatusBadRequest, "latitude out of range")
		return
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		WriteError(w, r, http.StatusBadRequest, "longitude out of range")
		return
	}
	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		WriteError(w, r, http.StatusBadRequest, "timestamp must be RFC3339")
		return
	}
	veh, err := s.DB.GetVehicleLite(r.Context(), req.VehicleID)
	if errors.Is(err, store.ErrNotFound) {
		WriteError(w, r, http.StatusNotFound, "vehicle not found")
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.DB.InsertLocation(r.Context(), req.VehicleID, req.Latitude, req.Longitude, ts); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	containing, err := s.DB.GeofencesContainingPoint(r.Context(), req.Latitude, req.Longitude)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	previous, err := s.DB.CurrentGeofenceState(r.Context(), req.VehicleID)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	current := map[string]struct{}{}
	currentList := make([]currentGeofenceOut, 0, len(containing))
	for _, g := range containing {
		current[g.ID] = struct{}{}
		currentList = append(currentList, currentGeofenceOut{
			GeofenceID:   g.ID,
			GeofenceName: g.Name,
			Status:       "inside",
		})
	}
	entered, exited := diffSets(previous, current)

	if err := s.DB.AddGeofenceState(r.Context(), req.VehicleID, entered); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.DB.RemoveGeofenceState(r.Context(), req.VehicleID, exited); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Fire alerts asynchronously so the HTTP response is not blocked.
	if len(entered)+len(exited) > 0 {
		go func(veh store.VehicleLite, lat, lng float64, ts time.Time, entered, exited []string) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			s.Dispatcher.HandleEvents(ctx, veh, lat, lng, ts, entered, exited)
		}(veh, req.Latitude, req.Longitude, ts, entered, exited)
	}

	WriteJSON(w, r, http.StatusOK, map[string]any{
		"vehicle_id":        req.VehicleID,
		"location_updated":  true,
		"current_geofences": currentList,
	})
}

func diffSets(prev, curr map[string]struct{}) (entered, exited []string) {
	for id := range curr {
		if _, ok := prev[id]; !ok {
			entered = append(entered, id)
		}
	}
	for id := range prev {
		if _, ok := curr[id]; !ok {
			exited = append(exited, id)
		}
	}
	return
}

type latestLocationCurrentGeofence struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Category     string `json:"category"`
}

func (s *Server) handleGetVehicleLocation(w http.ResponseWriter, r *http.Request) {
	vehicleID := chi.URLParam(r, "vehicle_id")
	if vehicleID == "" {
		WriteError(w, r, http.StatusBadRequest, "vehicle_id is required")
		return
	}
	veh, err := s.DB.GetVehicleLite(r.Context(), vehicleID)
	if errors.Is(err, store.ErrNotFound) {
		WriteError(w, r, http.StatusNotFound, "vehicle not found")
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	loc, err := s.DB.LatestLocation(r.Context(), vehicleID)
	if errors.Is(err, store.ErrNotFound) {
		WriteJSON(w, r, http.StatusOK, map[string]any{
			"vehicle_id":        veh.ID,
			"vehicle_number":    veh.VehicleNumber,
			"current_location":  nil,
			"current_geofences": []any{},
		})
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	containing, err := s.DB.GeofencesContainingPoint(r.Context(), loc.Latitude, loc.Longitude)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]latestLocationCurrentGeofence, 0, len(containing))
	for _, g := range containing {
		out = append(out, latestLocationCurrentGeofence{
			GeofenceID:   g.ID,
			GeofenceName: g.Name,
			Category:     g.Category,
		})
	}
	WriteJSON(w, r, http.StatusOK, map[string]any{
		"vehicle_id":     veh.ID,
		"vehicle_number": veh.VehicleNumber,
		"current_location": map[string]any{
			"latitude":  loc.Latitude,
			"longitude": loc.Longitude,
			"timestamp": loc.Timestamp.UTC().Format(time.RFC3339),
		},
		"current_geofences": out,
	})
}
