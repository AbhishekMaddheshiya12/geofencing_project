package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/time/rate"

	"github.com/mapup/geofence/internal/alerts"
	"github.com/mapup/geofence/internal/config"
	"github.com/mapup/geofence/internal/store"
	"github.com/mapup/geofence/internal/ws"
)

type Server struct {
	Cfg        config.Config
	DB         *store.DB
	Hub        *ws.Hub
	Dispatcher *alerts.Dispatcher
}

func New(cfg config.Config, db *store.DB, hub *ws.Hub, dispatcher *alerts.Dispatcher) *Server {
	return &Server{Cfg: cfg, DB: db, Hub: hub, Dispatcher: dispatcher}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.Cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(Logger)
	r.Use(Recover)
	r.Use(Timing)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, r, http.StatusOK, map[string]any{"status": "ok"})
	})

	// Resource endpoints (strict paths from the spec)
	r.Post("/geofences", s.handleCreateGeofence)
	r.Get("/geofences", s.handleListGeofences)

	r.Post("/vehicles", s.handleCreateVehicle)
	r.Get("/vehicles", s.handleListVehicles)

	// Rate-limit the high-traffic location endpoint.
	limiter := rate.NewLimiter(rate.Limit(s.Cfg.LocationRPS), s.Cfg.LocationBurst)
	r.With(rateLimit(limiter)).Post("/vehicles/location", s.handleLocationUpdate)
	r.Get("/vehicles/location/{vehicle_id}", s.handleGetVehicleLocation)

	r.Post("/alerts/configure", s.handleConfigureAlert)
	r.Get("/alerts", s.handleListAlerts)

	r.Get("/violations/history", s.handleListViolations)

	// WebSocket — outside the JSON middleware chain; nothing to embed time_ns into.
	r.Get("/ws/alerts", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(s.Hub, w, r)
	})

	return r
}

// rateLimit applies a shared token-bucket limiter.
func rateLimit(l *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !l.Allow() {
				WriteError(w, r, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
