export type Category =
  | 'delivery_zone'
  | 'restricted_zone'
  | 'toll_zone'
  | 'customer_area'

export type EventType = 'entry' | 'exit' | 'both'

export interface Geofence {
  id: string
  name: string
  description?: string
  coordinates: [number, number][]
  category: Category
  created_at: string
  status?: string
}

export interface Vehicle {
  id: string
  vehicle_number: string
  driver_name: string
  vehicle_type: string
  phone: string
  status: string
  created_at: string
}

export interface AlertConfig {
  alert_id: string
  geofence_id: string
  geofence_name?: string
  vehicle_id?: string
  vehicle_number?: string
  event_type: EventType
  status: string
  created_at: string
}

export interface Violation {
  id: string
  vehicle_id: string
  vehicle_number: string
  geofence_id: string
  geofence_name: string
  event_type: 'entry' | 'exit'
  latitude: number
  longitude: number
  timestamp: string
}

export interface AlertEnvelope {
  event_id: string
  event_type: 'entry' | 'exit'
  timestamp: string
  vehicle: { vehicle_id: string; vehicle_number: string; driver_name: string }
  geofence: { geofence_id: string; geofence_name: string; category: Category }
  location: { latitude: number; longitude: number }
}

export interface CurrentGeofence {
  geofence_id: string
  geofence_name: string
  status?: string
  category?: Category
}
