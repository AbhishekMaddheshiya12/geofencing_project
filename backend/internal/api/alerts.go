package api

import (
	"net/http"

	"github.com/mapup/geofence/internal/models"
)

type configureAlertReq struct {
	GeofenceID string  `json:"geofence_id"`
	VehicleID  *string `json:"vehicle_id"`
	EventType  string  `json:"event_type"`
}

func (s *Server) handleConfigureAlert(w http.ResponseWriter, r *http.Request) {
	var req configureAlertReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := requireNonEmpty(req.GeofenceID, "geofence_id"); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !models.ValidEventType(req.EventType) {
		WriteError(w, r, http.StatusBadRequest, "event_type must be entry, exit, or both")
		return
	}
	if _, err := s.DB.GetGeofenceLite(r.Context(), req.GeofenceID); err != nil {
		WriteError(w, r, http.StatusBadRequest, "geofence_id not found")
		return
	}
	if req.VehicleID != nil && *req.VehicleID != "" {
		if _, err := s.DB.GetVehicleLite(r.Context(), *req.VehicleID); err != nil {
			WriteError(w, r, http.StatusBadRequest, "vehicle_id not found")
			return
		}
	} else {
		req.VehicleID = nil
	}
	a := &models.AlertConfig{
		GeofenceID: req.GeofenceID,
		VehicleID:  req.VehicleID,
		EventType:  req.EventType,
	}
	if err := s.DB.CreateAlertConfig(r.Context(), a); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	resp := map[string]any{
		"alert_id":    a.AlertID,
		"geofence_id": a.GeofenceID,
		"event_type":  a.EventType,
		"status":      a.Status,
	}
	if a.VehicleID != nil {
		resp["vehicle_id"] = *a.VehicleID
	}
	WriteJSON(w, r, http.StatusCreated, resp)
}

func (s *Server) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	geofenceID := r.URL.Query().Get("geofence_id")
	vehicleID := r.URL.Query().Get("vehicle_id")
	limit, offset := parsePagination(r, s.Cfg.DefaultPageSize, s.Cfg.MaxPageSize)
	alerts, err := s.DB.ListAlertConfigs(r.Context(), geofenceID, vehicleID, limit, offset)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if alerts == nil {
		alerts = []models.AlertConfig{}
	}
	WriteJSON(w, r, http.StatusOK, map[string]any{
		"alerts": alerts,
	})
}
