# Geofencing & Real-time Alert System

Full-stack assessment submission for MapUp. A Go backend tracks vehicle
locations against PostGIS-backed polygonal geofences and pushes
entry/exit alerts to a React + Leaflet dashboard over WebSocket.

- **Backend** — Go 1.22, chi router, pgx, PostGIS, gorilla/websocket
- **Frontend** — React 18 + Vite + TypeScript, Tailwind CSS, Leaflet
- **Infra** — Docker Compose (PostGIS + backend + frontend)

> 📺 **Demo video**: _link will be added after recording_

---

## TL;DR — run the whole stack in one command

```bash
docker compose up --build
```

Then open **http://localhost:5173** in a browser. Backend is at
**http://localhost:8080**, WebSocket at **ws://localhost:8080/ws/alerts**.

See [SETUP.md](./SETUP.md) for full setup, env vars, and curl examples
for every endpoint.

---

## Repository layout

```
.
├── backend/                 Go service (9 REST endpoints + WebSocket)
│   ├── cmd/server/          entrypoint
│   ├── internal/
│   │   ├── api/             HTTP handlers + middleware (time_ns, CORS, ratelimit)
│   │   ├── ws/              WebSocket hub + client pumps
│   │   ├── alerts/          entry/exit dispatcher
│   │   ├── geo/             polygon validation + WKT helpers
│   │   ├── store/           PostGIS-aware data layer (pgx)
│   │   ├── models/          DTOs
│   │   ├── ids/             prefixed id generator (geo_, veh_, …)
│   │   └── config/          env loader
│   ├── migrations/          SQL schema
│   └── Dockerfile
├── frontend/                React app
│   ├── src/
│   │   ├── api/             axios client + typed DTOs
│   │   ├── components/      MapView (Leaflet + draw)
│   │   ├── hooks/           useAlertsSocket (reconnect + feed)
│   │   └── pages/           Dashboard / Vehicles / Alerts / History
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml       postgis + backend + frontend
├── SETUP.md                 prerequisites, run commands, curl examples
└── README.md                this file
```

---

## API contract

The 9 REST endpoints and the WebSocket are implemented exactly as
specified in the assessment brief. Every JSON response includes
`"time_ns"` — the handler's execution time in nanoseconds.

| # | Method | Path                                 |
|---|--------|--------------------------------------|
| 1 | POST   | `/geofences`                         |
| 2 | GET    | `/geofences`                         |
| 3 | POST   | `/vehicles`                          |
| 4 | GET    | `/vehicles`                          |
| 5 | POST   | `/vehicles/location`                 |
| 6 | GET    | `/vehicles/location/{vehicle_id}`    |
| 7 | POST   | `/alerts/configure`                  |
| 8 | GET    | `/alerts`                            |
| 9 | GET    | `/violations/history`                |
|   | GET    | `/ws/alerts` (WebSocket)             |

See [SETUP.md](./SETUP.md#api-testing-guide-curl-examples) for curl
examples of each.

---

## Bonus features included

- **Pagination** on every list endpoint (`?limit=&offset=`).
- **Rate limiting** on `POST /vehicles/location` (token bucket).
- **Unit tests** for polygon validation and entry/exit set-diff logic
  (`go test ./...`).

---

## Tech choices — why

- **chi** over net/http's mux for clean middleware chains.
- **pgx** over database/sql for native PostgreSQL types and batch ops.
- **PostGIS GEOGRAPHY(POLYGON, 4326)** + `ST_Covers` for accurate
  great-circle point-in-polygon (vs. Cartesian shortcuts that drift
  away from the equator).
- **`vehicle_geofence_state` cache table** so the location handler can
  diff against the previous state set instead of replaying history on
  every update — keeps `POST /vehicles/location` O(1) in history size.
- **WebSocket hub with non-blocking broadcast** — slow consumers are
  dropped, so one stalled client can't back-pressure the system.
- **Leaflet** over Mapbox so no API key is needed to demo.
