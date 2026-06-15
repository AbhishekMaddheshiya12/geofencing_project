import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import {
  configureAlert,
  listAlerts,
  listGeofences,
  listVehicles,
} from '../api/client'
import type { AlertConfig, EventType, Geofence, Vehicle } from '../api/types'

export default function Alerts() {
  const [alerts, setAlerts] = useState<AlertConfig[]>([])
  const [geofences, setGeofences] = useState<Geofence[]>([])
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [form, setForm] = useState<{ geofence_id: string; vehicle_id: string; event_type: EventType }>({
    geofence_id: '',
    vehicle_id: '',
    event_type: 'both',
  })

  async function refresh() {
    const [a, g, v] = await Promise.all([listAlerts(), listGeofences(), listVehicles()])
    setAlerts(a)
    setGeofences(g)
    setVehicles(v)
  }

  useEffect(() => {
    refresh().catch((e) => toast.error(e.message))
  }, [])

  async function save() {
    if (!form.geofence_id) return toast.error('Pick a geofence')
    try {
      await configureAlert({
        geofence_id: form.geofence_id,
        vehicle_id: form.vehicle_id || undefined,
        event_type: form.event_type,
      })
      toast.success('Alert configured')
      setForm({ geofence_id: '', vehicle_id: '', event_type: 'both' })
      await refresh()
    } catch (e: any) {
      toast.error(e.response?.data?.message || e.message)
    }
  }

  return (
    <div className="p-6 overflow-y-auto h-full">
      <h1 className="text-xl font-semibold mb-4">Alert Configuration</h1>

      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <h2 className="font-medium mb-3">Configure new alert</h2>
        <div className="grid grid-cols-3 gap-3">
          <select
            value={form.geofence_id}
            onChange={(e) => setForm({ ...form, geofence_id: e.target.value })}
            className="border rounded px-2 py-1"
          >
            <option value="">— Geofence —</option>
            {geofences.map((g) => (
              <option key={g.id} value={g.id}>
                {g.name} ({g.category})
              </option>
            ))}
          </select>
          <select
            value={form.vehicle_id}
            onChange={(e) => setForm({ ...form, vehicle_id: e.target.value })}
            className="border rounded px-2 py-1"
          >
            <option value="">All vehicles</option>
            {vehicles.map((v) => (
              <option key={v.id} value={v.id}>
                {v.vehicle_number} ({v.driver_name})
              </option>
            ))}
          </select>
          <select
            value={form.event_type}
            onChange={(e) => setForm({ ...form, event_type: e.target.value as EventType })}
            className="border rounded px-2 py-1"
          >
            <option value="entry">entry</option>
            <option value="exit">exit</option>
            <option value="both">both</option>
          </select>
        </div>
        <button onClick={save} className="mt-3 bg-slate-900 text-white px-3 py-1.5 rounded">
          Configure
        </button>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left">
            <tr>
              <th className="px-3 py-2">Alert ID</th>
              <th className="px-3 py-2">Geofence</th>
              <th className="px-3 py-2">Vehicle</th>
              <th className="px-3 py-2">Event</th>
              <th className="px-3 py-2">Status</th>
              <th className="px-3 py-2">Created</th>
            </tr>
          </thead>
          <tbody>
            {alerts.map((a) => (
              <tr key={a.alert_id} className="border-t">
                <td className="px-3 py-2 font-mono text-xs">{a.alert_id}</td>
                <td className="px-3 py-2">{a.geofence_name || a.geofence_id}</td>
                <td className="px-3 py-2">{a.vehicle_number || (a.vehicle_id || 'All vehicles')}</td>
                <td className="px-3 py-2">{a.event_type}</td>
                <td className="px-3 py-2">{a.status}</td>
                <td className="px-3 py-2 text-slate-500">
                  {new Date(a.created_at).toLocaleString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
