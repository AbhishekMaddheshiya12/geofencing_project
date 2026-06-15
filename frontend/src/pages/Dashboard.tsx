import { useEffect, useMemo, useState } from 'react'
import toast from 'react-hot-toast'
import MapView from '../components/MapView'
import {
  createGeofence,
  getVehicleLocation,
  listGeofences,
  listVehicles,
  updateLocation,
} from '../api/client'
import type { Category, Geofence, Vehicle } from '../api/types'
import { useAlertsSocket } from '../hooks/useAlertsSocket'

const CATEGORIES: Category[] = [
  'delivery_zone',
  'restricted_zone',
  'toll_zone',
  'customer_area',
]

export default function Dashboard() {
  const [geofences, setGeofences] = useState<Geofence[]>([])
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [locations, setLocations] = useState<
    Record<string, { lat: number; lng: number }>
  >({})
  const [pendingPolygon, setPendingPolygon] = useState<[number, number][] | null>(null)
  const [selectedVehicleId, setSelectedVehicleId] = useState<string>('')

  const { connected, feed } = useAlertsSocket((a) => {
    const icon = a.event_type === 'entry' ? '➡️' : '⬅️'
    toast(
      `${icon} ${a.vehicle.vehicle_number} ${a.event_type} ${a.geofence.geofence_name}`,
      { duration: 6000 }
    )
  })

  async function refresh() {
    const [g, v] = await Promise.all([listGeofences(), listVehicles()])
    setGeofences(g)
    setVehicles(v)
    const locs: Record<string, { lat: number; lng: number }> = {}
    for (const veh of v) {
      try {
        const r = await getVehicleLocation(veh.id)
        if (r.current_location) {
          locs[veh.id] = { lat: r.current_location.latitude, lng: r.current_location.longitude }
        }
      } catch {
        // ignore vehicles with no location yet
      }
    }
    setLocations(locs)
  }

  useEffect(() => {
    refresh().catch((e) => toast.error(`Load failed: ${e.message}`))
  }, [])

  const vehicleMarkers = useMemo(
    () =>
      vehicles
        .filter((v) => locations[v.id])
        .map((v) => ({ vehicle: v, lat: locations[v.id].lat, lng: locations[v.id].lng })),
    [vehicles, locations]
  )

  async function handleMapClick(lat: number, lng: number) {
    if (!selectedVehicleId) return
    try {
      const r = await updateLocation({
        vehicle_id: selectedVehicleId,
        latitude: lat,
        longitude: lng,
        timestamp: new Date().toISOString(),
      })
      toast.success(`Location updated (${r.current_geofences.length} geofences)`)
      setLocations((prev) => ({ ...prev, [selectedVehicleId]: { lat, lng } }))
    } catch (e: any) {
      toast.error(e.response?.data?.message || e.message)
    }
  }

  return (
    <div className="h-full flex">
      <div className="flex-1 relative">
        <MapView
          geofences={geofences}
          vehicleLocations={vehicleMarkers}
          onPolygonDrawn={(latlngs) => setPendingPolygon(latlngs)}
          onMapClick={handleMapClick}
        />
        {pendingPolygon && (
          <NewGeofenceModal
            coordinates={pendingPolygon}
            onClose={() => setPendingPolygon(null)}
            onSaved={async () => {
              setPendingPolygon(null)
              await refresh()
            }}
          />
        )}
      </div>

      <aside className="w-96 bg-white border-l border-slate-200 flex flex-col">
        <div className="p-4 border-b border-slate-200">
          <div className="flex items-center justify-between mb-2">
            <h2 className="font-semibold">Live Alerts</h2>
            <span
              className={`text-xs px-2 py-0.5 rounded-full ${
                connected ? 'bg-green-100 text-green-700' : 'bg-slate-200 text-slate-600'
              }`}
            >
              {connected ? 'connected' : 'disconnected'}
            </span>
          </div>
          <label className="block text-xs text-slate-600">Update location for vehicle (click map)</label>
          <select
            value={selectedVehicleId}
            onChange={(e) => setSelectedVehicleId(e.target.value)}
            className="mt-1 w-full border rounded px-2 py-1 text-sm"
          >
            <option value="">— Select a vehicle —</option>
            {vehicles.map((v) => (
              <option key={v.id} value={v.id}>
                {v.vehicle_number} ({v.driver_name})
              </option>
            ))}
          </select>
        </div>
        <div className="flex-1 overflow-y-auto p-3 space-y-2">
          {feed.length === 0 && (
            <div className="text-sm text-slate-500 text-center py-8">No alerts yet</div>
          )}
          {feed.map((a) => (
            <div
              key={a.event_id}
              className={`rounded border px-3 py-2 text-sm ${
                a.event_type === 'entry'
                  ? 'border-green-200 bg-green-50'
                  : 'border-amber-200 bg-amber-50'
              }`}
            >
              <div className="flex justify-between">
                <span className="font-medium">{a.vehicle.vehicle_number}</span>
                <span className="text-xs text-slate-500">
                  {new Date(a.timestamp).toLocaleTimeString()}
                </span>
              </div>
              <div className="text-xs">
                <span className="font-semibold uppercase">{a.event_type}</span>{' '}
                {a.geofence.geofence_name}{' '}
                <span className="text-slate-500">({a.geofence.category})</span>
              </div>
            </div>
          ))}
        </div>
      </aside>
    </div>
  )
}

function NewGeofenceModal({
  coordinates,
  onClose,
  onSaved,
}: {
  coordinates: [number, number][]
  onClose: () => void
  onSaved: () => void
}) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState<Category>('delivery_zone')
  const [busy, setBusy] = useState(false)

  async function save() {
    if (!name) return toast.error('Name is required')
    setBusy(true)
    try {
      await createGeofence({ name, description, category, coordinates })
      toast.success('Geofence created')
      onSaved()
    } catch (e: any) {
      toast.error(e.response?.data?.message || e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="absolute inset-0 bg-black/40 z-[1000] flex items-center justify-center">
      <div className="bg-white rounded-lg p-5 w-96 shadow-xl">
        <h3 className="font-semibold mb-3">New Geofence</h3>
        <input
          className="w-full border rounded px-2 py-1 mb-2"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input
          className="w-full border rounded px-2 py-1 mb-2"
          placeholder="Description (optional)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value as Category)}
          className="w-full border rounded px-2 py-1 mb-3"
        >
          {CATEGORIES.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </select>
        <div className="text-xs text-slate-500 mb-3">{coordinates.length} points</div>
        <div className="flex justify-end gap-2">
          <button onClick={onClose} className="px-3 py-1 rounded border">
            Cancel
          </button>
          <button
            onClick={save}
            disabled={busy}
            className="px-3 py-1 rounded bg-slate-900 text-white disabled:opacity-50"
          >
            {busy ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}
