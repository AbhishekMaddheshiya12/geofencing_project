import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { createVehicle, getVehicleLocation, listVehicles, updateLocation } from '../api/client'
import type { CurrentGeofence, Vehicle } from '../api/types'

interface Row extends Vehicle {
  location?: { lat: number; lng: number; ts: string } | null
  geofences?: CurrentGeofence[]
}

export default function Vehicles() {
  const [rows, setRows] = useState<Row[]>([])
  const [form, setForm] = useState({
    vehicle_number: '',
    driver_name: '',
    vehicle_type: 'truck',
    phone: '',
  })

  async function refresh() {
    const vs = await listVehicles()
    const enriched: Row[] = await Promise.all(
      vs.map(async (v) => {
        try {
          const r = await getVehicleLocation(v.id)
          return {
            ...v,
            location: r.current_location
              ? {
                  lat: r.current_location.latitude,
                  lng: r.current_location.longitude,
                  ts: r.current_location.timestamp,
                }
              : null,
            geofences: r.current_geofences,
          }
        } catch {
          return v
        }
      })
    )
    setRows(enriched)
  }

  useEffect(() => {
    refresh().catch((e) => toast.error(e.message))
  }, [])

  async function register() {
    if (!form.vehicle_number || !form.driver_name || !form.phone) {
      return toast.error('Fill all required fields')
    }
    try {
      await createVehicle(form)
      toast.success('Vehicle registered')
      setForm({ vehicle_number: '', driver_name: '', vehicle_type: 'truck', phone: '' })
      await refresh()
    } catch (e: any) {
      toast.error(e.response?.data?.message || e.message)
    }
  }

  async function quickLocation(vid: string) {
    const latStr = prompt('Latitude?')
    if (!latStr) return
    const lngStr = prompt('Longitude?')
    if (!lngStr) return
    try {
      await updateLocation({
        vehicle_id: vid,
        latitude: parseFloat(latStr),
        longitude: parseFloat(lngStr),
        timestamp: new Date().toISOString(),
      })
      toast.success('Location updated')
      await refresh()
    } catch (e: any) {
      toast.error(e.response?.data?.message || e.message)
    }
  }

  return (
    <div className="p-6 overflow-y-auto h-full">
      <h1 className="text-xl font-semibold mb-4">Vehicles</h1>

      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <h2 className="font-medium mb-3">Register Vehicle</h2>
        <div className="grid grid-cols-2 gap-3">
          <input
            className="border rounded px-2 py-1"
            placeholder="Vehicle number"
            value={form.vehicle_number}
            onChange={(e) => setForm({ ...form, vehicle_number: e.target.value })}
          />
          <input
            className="border rounded px-2 py-1"
            placeholder="Driver name"
            value={form.driver_name}
            onChange={(e) => setForm({ ...form, driver_name: e.target.value })}
          />
          <select
            className="border rounded px-2 py-1"
            value={form.vehicle_type}
            onChange={(e) => setForm({ ...form, vehicle_type: e.target.value })}
          >
            {['truck', 'car', 'van', 'motorcycle'].map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
          <input
            className="border rounded px-2 py-1"
            placeholder="Phone"
            value={form.phone}
            onChange={(e) => setForm({ ...form, phone: e.target.value })}
          />
        </div>
        <button
          onClick={register}
          className="mt-3 bg-slate-900 text-white px-3 py-1.5 rounded"
        >
          Register
        </button>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left">
            <tr>
              <th className="px-3 py-2">Vehicle</th>
              <th className="px-3 py-2">Driver</th>
              <th className="px-3 py-2">Type</th>
              <th className="px-3 py-2">Current location</th>
              <th className="px-3 py-2">Inside geofences</th>
              <th className="px-3 py-2"></th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => (
              <tr key={r.id} className="border-t">
                <td className="px-3 py-2 font-medium">{r.vehicle_number}</td>
                <td className="px-3 py-2">{r.driver_name}</td>
                <td className="px-3 py-2">{r.vehicle_type}</td>
                <td className="px-3 py-2 text-slate-600">
                  {r.location
                    ? `${r.location.lat.toFixed(4)}, ${r.location.lng.toFixed(4)}`
                    : '—'}
                </td>
                <td className="px-3 py-2">
                  {r.geofences && r.geofences.length > 0
                    ? r.geofences.map((g) => g.geofence_name).join(', ')
                    : '—'}
                </td>
                <td className="px-3 py-2">
                  <button
                    onClick={() => quickLocation(r.id)}
                    className="text-xs text-blue-600 hover:underline"
                  >
                    Update location
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
