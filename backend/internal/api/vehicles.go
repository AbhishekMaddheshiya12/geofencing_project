package api

import (
	"errors"
	"net/http"

	"github.com/mapup/geofence/internal/models"
	"github.com/mapup/geofence/internal/store"
)

type createVehicleReq struct {
	VehicleNumber string `json:"vehicle_number"`
	DriverName    string `json:"driver_name"`
	VehicleType   string `json:"vehicle_type"`
	Phone         string `json:"phone"`
}

func (s *Server) handleCreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req createVehicleReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	for field, val := range map[string]string{
		"vehicle_number": req.VehicleNumber,
		"driver_name":    req.DriverName,
		"vehicle_type":   req.VehicleType,
		"phone":          req.Phone,
	} {
		if err := requireNonEmpty(val, field); err != nil {
			WriteError(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}
	v := &models.Vehicle{
		VehicleNumber: req.VehicleNumber,
		DriverName:    req.DriverName,
		VehicleType:   req.VehicleType,
		Phone:         req.Phone,
	}
	err := s.DB.CreateVehicle(r.Context(), v)
	if errors.Is(err, store.ErrConflict) {
		WriteError(w, r, http.StatusConflict, "vehicle_number already exists")
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, r, http.StatusCreated, map[string]any{
		"id":             v.ID,
		"vehicle_number": v.VehicleNumber,
		"status":         v.Status,
	})
}

func (s *Server) handleListVehicles(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r, s.Cfg.DefaultPageSize, s.Cfg.MaxPageSize)
	vs, err := s.DB.ListVehicles(r.Context(), limit, offset)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if vs == nil {
		vs = []models.Vehicle{}
	}
	WriteJSON(w, r, http.StatusOK, map[string]any{
		"vehicles": vs,
	})
}
