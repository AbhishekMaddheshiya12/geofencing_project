# Quickstart — for whoever just unzipped this project

You need exactly **one** thing installed: **Docker Desktop** (24+).
Everything else (Go, Node, PostgreSQL, PostGIS) runs inside containers.

## Run it

```bash
# 1. Open a terminal inside the unzipped folder
cd path/to/project

# 2. Bring up the whole stack (first run takes ~3 min to download images)
docker compose up --build
```

When you see this in the backend logs, you're ready:
```
{"level":"INFO","msg":"migrations applied"}
{"level":"INFO","msg":"listening","addr":":8080"}
```

## Open the app

| What       | URL                              |
|------------|----------------------------------|
| Frontend   | http://localhost:5173            |
| Backend    | http://localhost:8080            |
| Health     | http://localhost:8080/healthz    |
| WebSocket  | ws://localhost:8080/ws/alerts    |

## Try the flow

1. Open http://localhost:5173 → **Dashboard**.
2. Click the polygon tool (top-right of the map) and draw a shape.
3. A modal opens — give it a name, pick `restricted_zone`, **Save**.
4. Go to **Vehicles**, register a new vehicle.
5. Go to **Alerts**, configure an alert: pick your geofence,
   "All vehicles", `both`.
6. Back on **Dashboard**, pick your vehicle in the sidebar dropdown.
7. Click anywhere **inside** the polygon → you get an **entry** toast.
8. Click **outside** → you get an **exit** toast.
9. **History** page shows both events.

## Stop & clean up

```bash
# Stop the containers (keeps data)
docker compose down

# Stop + delete everything (containers, volumes, downloaded images)
docker compose down -v --rmi all
```

## Need to develop on it instead?

See [SETUP.md](./SETUP.md) — it has Go/Node native setup instructions
and curl examples for every endpoint.
