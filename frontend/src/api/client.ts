import axios from 'axios'
import type {
  AlertConfig,
  Category,
  CurrentGeofence,
  EventType,
  Geofence,
  Vehicle,
  Violation,
} from './types'

const baseURL =
  (import.meta as any).env?.VITE_API_BASE || 'http://localhost:8080'

export const http = axios.create({ baseURL })

// --- Geofences ---
export async function createGeofence(input: {
  name: string
  description?: string
  category: Category
  coordinates: [number, number][]
}) {
  const { data } = await http.post('/geofences', input)
  return data as { id: string; name: string; status: string; time_ns: string }
}

export async function listGeofences(category?: Category) {
  const { data } = await http.get('/geofences', {
    params: category ? { category } : {},
  })
  return data.geofences as Geofence[]
}

// --- Vehicles ---
export async function createVehicle(input: {
  vehicle_number: string
  driver_name: string
  vehicle_type: string
  phone: string
}) {
  const { data } = await http.post('/vehicles', input)
  return data as { id: string; vehicle_number: string; status: string }
}

export async function listVehicles() {
  const { data } = await http.get('/vehicles')
  return data.vehicles as Vehicle[]
}

// --- Locations ---
export async function updateLocation(input: {
  vehicle_id: string
  latitude: number
  longitude: number
  timestamp: string
}) {
  const { data } = await http.post('/vehicles/location', input)
  return data as {
    vehicle_id: string
    location_updated: boolean
    current_geofences: CurrentGeofence[]
  }
}

export async function getVehicleLocation(vehicle_id: string) {
  const { data } = await http.get(`/vehicles/location/${vehicle_id}`)
  return data as {
    vehicle_id: string
    vehicle_number: string
    current_location: { latitude: number; longitude: number; timestamp: string } | null
    current_geofences: CurrentGeofence[]
  }
}

// --- Alerts ---
export async function configureAlert(input: {
  geofence_id: string
  vehicle_id?: string
  event_type: EventType
}) {
  const body: any = {
    geofence_id: input.geofence_id,
    event_type: input.event_type,
  }
  if (input.vehicle_id) body.vehicle_id = input.vehicle_id
  const { data } = await http.post('/alerts/configure', body)
  return data
}

export async function listAlerts(filters?: {
  geofence_id?: string
  vehicle_id?: string
}) {
  const { data } = await http.get('/alerts', { params: filters || {} })
  return data.alerts as AlertConfig[]
}

// --- Violations ---
export async function listViolations(filters?: {
  vehicle_id?: string
  geofence_id?: string
  start_date?: string
  end_date?: string
  limit?: number
}) {
  const { data } = await http.get('/violations/history', { params: filters || {} })
  return {
    violations: data.violations as Violation[],
    total_count: data.total_count as number,
  }
}
