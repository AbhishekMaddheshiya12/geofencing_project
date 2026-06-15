package api

import (
	"net/http"

	"github.com/mapup/geofence/internal/geo"
	"github.com/mapup/geofence/internal/models"
)

type createGeofenceReq struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
	Category    string      `json:"category"`
}

func (s *Server) handleCreateGeofence(w http.ResponseWriter, r *http.Request) {
	var req createGeofenceReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := requireNonEmpty(req.Name, "name"); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if !models.ValidCategory(req.Category) {
		WriteError(w, r, http.StatusBadRequest, "category must be one of delivery_zone, restricted_zone, toll_zone, customer_area")
		return
	}
	if err := geo.ValidatePolygon(req.Coordinates); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	g := &models.Geofence{
		Name:        req.Name,
		Description: req.Description,
		Coordinates: req.Coordinates,
		Category:    req.Category,
	}
	if err := s.DB.CreateGeofence(r.Context(), g); err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, r, http.StatusCreated, map[string]any{
		"id":     g.ID,
		"name":   g.Name,
		"status": g.Status,
	})
}

func (s *Server) handleListGeofences(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	limit, offset := parsePagination(r, s.Cfg.DefaultPageSize, s.Cfg.MaxPageSize)
	gs, err := s.DB.ListGeofences(r.Context(), category, limit, offset)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if gs == nil {
		gs = []models.Geofence{}
	}
	WriteJSON(w, r, http.StatusOK, map[string]any{
		"geofences": gs,
	})
}
