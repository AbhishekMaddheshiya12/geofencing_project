CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS geofences (
  id           TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  description  TEXT,
  category     TEXT NOT NULL CHECK (category IN
                 ('delivery_zone','restricted_zone','toll_zone','customer_area')),
  geom         GEOGRAPHY(POLYGON, 4326) NOT NULL,
  status       TEXT NOT NULL DEFAULT 'active',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS geofences_geom_gix ON geofences USING GIST (geom);
CREATE INDEX IF NOT EXISTS geofences_category_idx ON geofences (category);

CREATE TABLE IF NOT EXISTS vehicles (
  id              TEXT PRIMARY KEY,
  vehicle_number  TEXT UNIQUE NOT NULL,
  driver_name     TEXT NOT NULL,
  vehicle_type    TEXT NOT NULL,
  phone           TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'active',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vehicle_locations (
  id          BIGSERIAL PRIMARY KEY,
  vehicle_id  TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  point       GEOGRAPHY(POINT, 4326) NOT NULL,
  ts          TIMESTAMPTZ NOT NULL,
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS vloc_vehicle_ts_idx ON vehicle_locations (vehicle_id, ts DESC);
CREATE INDEX IF NOT EXISTS vloc_point_gix ON vehicle_locations USING GIST (point);

CREATE TABLE IF NOT EXISTS vehicle_geofence_state (
  vehicle_id   TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  geofence_id  TEXT NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
  entered_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (vehicle_id, geofence_id)
);

CREATE TABLE IF NOT EXISTS alert_configs (
  id           TEXT PRIMARY KEY,
  geofence_id  TEXT NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
  vehicle_id   TEXT REFERENCES vehicles(id) ON DELETE CASCADE,
  event_type   TEXT NOT NULL CHECK (event_type IN ('entry','exit','both')),
  status       TEXT NOT NULL DEFAULT 'active',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS alert_configs_geo_idx ON alert_configs (geofence_id);
CREATE INDEX IF NOT EXISTS alert_configs_veh_idx ON alert_configs (vehicle_id);

CREATE TABLE IF NOT EXISTS violations (
  id           TEXT PRIMARY KEY,
  vehicle_id   TEXT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
  geofence_id  TEXT NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
  event_type   TEXT NOT NULL CHECK (event_type IN ('entry','exit')),
  latitude     DOUBLE PRECISION NOT NULL,
  longitude    DOUBLE PRECISION NOT NULL,
  ts           TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS viol_vehicle_ts_idx ON violations (vehicle_id, ts DESC);
CREATE INDEX IF NOT EXISTS viol_geofence_ts_idx ON violations (geofence_id, ts DESC);
