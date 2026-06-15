package api

import (
	"net/http"
	"time"

	"github.com/mapup/geofence/internal/models"
	"github.com/mapup/geofence/internal/store"
)

func (s *Server) handleListViolations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ViolationFilters{
		VehicleID:  q.Get("vehicle_id"),
		GeofenceID: q.Get("geofence_id"),
		Limit:      s.Cfg.DefaultPageSize,
	}
	if v := q.Get("start_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			WriteError(w, r, http.StatusBadRequest, "start_date must be RFC3339")
			return
		}
		f.StartDate = &t
	}
	if v := q.Get("end_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			WriteError(w, r, http.StatusBadRequest, "end_date must be RFC3339")
			return
		}
		f.EndDate = &t
	}
	limit, _ := parsePagination(r, s.Cfg.DefaultPageSize, s.Cfg.MaxPageSize)
	f.Limit = limit
	vs, total, err := s.DB.ListViolations(r.Context(), f)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if vs == nil {
		vs = []models.Violation{}
	}
	WriteJSON(w, r, http.StatusOK, map[string]any{
		"violations":  vs,
		"total_count": total,
	})
}
