# SETUP.md — Geofencing & Real-time Alert System

## Prerequisites

- **Docker Desktop** 24+ (with Docker Compose v2)
- **Node.js** 20+ and **npm** 10+ (only needed if running the frontend outside Docker)
- **Go** 1.22+ (only needed if running the backend outside Docker or running tests)
- A free port range: **5432** (Postgres), **8080** (backend), **5173** (frontend)

## Architecture

```
┌──────────────┐  HTTP/WS   ┌──────────────┐   SQL    ┌────────────┐
│  React (UI)  │ ─────────► │  Go backend  │ ───────► │  PostGIS   │
│  Leaflet map │ ◄───────── │  chi + ws    │          │   16       │
└──────────────┘   alerts   └──────────────┘          └────────────┘
```

- Backend embeds a polling-free real-time push via WebSocket (`/ws/alerts`).
- PostGIS provides point-in-polygon (`ST_Covers`) for geofence detection.
- Every JSON response includes `"time_ns"` (handler execution time).

---

## Quick start (Docker Compose)

```bash
# From the repo root
docker compose up --build
```

Wait for `migrations applied` in the backend logs, then:

- **Frontend**: http://localhost:5173
- **Backend**: http://localhost:8080
- **Health check**: http://localhost:8080/healthz
- **WebSocket**: ws://localhost:8080/ws/alerts

To stop and clean up everything (containers + volume + images):
```bash
docker compose down -v --rmi all
```

---

## Local development (without Docker)

### 1. Start PostGIS only

If you want to run the Go server natively:
```bash
docker compose up postgis -d
```
Or install Postgres + PostGIS locally and apply `backend/migrations/0001_init.sql`.

### 2. Run the backend
```bash
cd backend
export DATABASE_URL="postgres://geofence:geofence@localhost:5432/geofence?sslmode=disable"
go run ./cmd/server
```

### 3. Run the frontend
```bash
cd frontend
npm install
npm run dev
```
Visit http://localhost:5173.

### 4. Run unit tests
```bash
cd backend
go test ./...
```

Tests cover polygon validation (`internal/geo`) and the entry/exit set-diff
(`internal/alerts`).

---

## API testing guide (curl examples)

> Every successful response includes `"time_ns": "<int as string>"` —
> handler execution time in nanoseconds.

### 1. Create a geofence (POST /geofences)
```bash
curl -X POST http://localhost:8080/geofences \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Downtown Delivery Zone",
    "description": "Main delivery area",
    "category": "delivery_zone",
    "coordinates": [
      [37.7749, -122.4194],
      [37.7849, -122.4194],
      [37.7849, -122.4094],
      [37.7749, -122.4094],
      [37.7749, -122.4194]
    ]
  }'
```

### 2. List geofences (GET /geofences)
```bash
curl 'http://localhost:8080/geofences'
curl 'http://localhost:8080/geofences?category=delivery_zone'
```

### 3. Register a vehicle (POST /vehicles)
```bash
curl -X POST http://localhost:8080/vehicles \
  -H 'Content-Type: application/json' \
  -d '{
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "John Doe",
    "vehicle_type": "truck",
    "phone": "+11234567890"
  }'
```

### 4. List vehicles (GET /vehicles)
```bash
curl http://localhost:8080/vehicles
```

### 5. Update vehicle location (POST /vehicles/location)
> Substitute `veh_xxx` with the ID returned by step 3.
```bash
curl -X POST http://localhost:8080/vehicles/location \
  -H 'Content-Type: application/json' \
  -d '{
    "vehicle_id": "veh_xxx",
    "latitude": 37.7800,
    "longitude": -122.4150,
    "timestamp": "2026-06-12T10:35:00Z"
  }'
```

### 6. Get current location (GET /vehicles/location/{id})
```bash
curl http://localhost:8080/vehicles/location/veh_xxx
```

### 7. Configure an alert (POST /alerts/configure)
```bash
curl -X POST http://localhost:8080/alerts/configure \
  -H 'Content-Type: application/json' \
  -d '{
    "geofence_id": "geo_xxx",
    "vehicle_id": "veh_xxx",
    "event_type": "both"
  }'
```
> Omit `vehicle_id` to apply the alert to all vehicles.

### 8. List configured alerts (GET /alerts)
```bash
curl 'http://localhost:8080/alerts'
curl 'http://localhost:8080/alerts?geofence_id=geo_xxx'
curl 'http://localhost:8080/alerts?vehicle_id=veh_xxx'
```

### 9. Violation history (GET /violations/history)
```bash
curl 'http://localhost:8080/violations/history?vehicle_id=veh_xxx&limit=100'
curl 'http://localhost:8080/violations/history?start_date=2026-06-01T00:00:00Z&end_date=2026-06-30T23:59:59Z'
```

### WebSocket: listen for live alerts
Install `wscat` once (`npm i -g wscat`) and run:
```bash
wscat -c ws://localhost:8080/ws/alerts
```
Then send a location update that crosses a geofence boundary — the alert
will arrive on this socket in real time.

---

## Frontend usage guide

1. Open http://localhost:5173.
2. **Dashboard** (default):
   - Use the polygon tool (top-right of the map) to draw a geofence; a
     modal asks for name/category and saves it.
   - Pick a vehicle in the sidebar dropdown, then click anywhere on the
     map to update its location. A toast appears for each alert.
3. **Vehicles** — register vehicles and manually update their location.
4. **Alerts** — configure rules (per-vehicle or all vehicles).
5. **History** — filter violation events by vehicle, geofence, date range.

---

## Environment variables

| Name | Default | Purpose |
|------|---------|---------|
| `HTTP_ADDR` | `:8080` | Bind address for the HTTP server |
| `DATABASE_URL` | *(required)* | Postgres connection string |
| `CORS_ALLOWED_ORIGIN` | `*` | CORS allowlist (set to your frontend origin in prod) |
| `LOCATION_RPS` | `30` | Rate limit (req/s) for `POST /vehicles/location` |
| `LOCATION_BURST` | `60` | Burst budget for the same limiter |
| `DEFAULT_PAGE_SIZE` | `50` | Default `limit` on list endpoints |
| `MAX_PAGE_SIZE` | `500` | Hard cap on `limit` |
| `VITE_API_BASE` | `http://localhost:8080` | Backend URL used by the frontend |
| `VITE_WS_URL` | `ws://localhost:8080/ws/alerts` | WebSocket URL used by the frontend |

---

## Bonus features included

- ✅ **Pagination** on `GET /geofences`, `GET /vehicles`, `GET /alerts`,
  `GET /violations/history` (`?limit=&offset=`).
- ✅ **Rate limiting** on `POST /vehicles/location` (token bucket).
- ✅ **Unit tests** for polygon validation and entry/exit set-diff.
- ⏭ JWT/API-key auth — not included; can be layered as chi middleware.
