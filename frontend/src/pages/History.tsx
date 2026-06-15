import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { listGeofences, listVehicles, listViolations } from '../api/client'
import type { Geofence, Vehicle, Violation } from '../api/types'

export default function History() {
  const [violations, setViolations] = useState<Violation[]>([])
  const [total, setTotal] = useState(0)
  const [geofences, setGeofences] = useState<Geofence[]>([])
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [filters, setFilters] = useState({
    vehicle_id: '',
    geofence_id: '',
    start_date: '',
    end_date: '',
    limit: 50,
  })

  async function refresh() {
    const params: any = { limit: filters.limit }
    if (filters.vehicle_id) params.vehicle_id = filters.vehicle_id
    if (filters.geofence_id) params.geofence_id = filters.geofence_id
    if (filters.start_date) params.start_date = new Date(filters.start_date).toISOString()
    if (filters.end_date) params.end_date = new Date(filters.end_date).toISOString()
    const r = await listViolations(params)
    setViolations(r.violations)
    setTotal(r.total_count)
  }

  useEffect(() => {
    Promise.all([listGeofences(), listVehicles()])
      .then(([g, v]) => {
        setGeofences(g)
        setVehicles(v)
      })
      .catch((e) => toast.error(e.message))
  }, [])

  useEffect(() => {
    refresh().catch((e) => toast.error(e.message))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filters])

  return (
    <div className="p-6 overflow-y-auto h-full">
      <h1 className="text-xl font-semibold mb-4">Violation History</h1>

      <div className="bg-white rounded-lg shadow p-4 mb-6 grid grid-cols-5 gap-3">
        <select
          value={filters.vehicle_id}
          onChange={(e) => setFilters({ ...filters, vehicle_id: e.target.value })}
          className="border rounded px-2 py-1"
        >
          <option value="">All vehicles</option>
          {vehicles.map((v) => (
            <option key={v.id} value={v.id}>{v.vehicle_number}</option>
          ))}
        </select>
        <select
          value={filters.geofence_id}
          onChange={(e) => setFilters({ ...filters, geofence_id: e.target.value })}
          className="border rounded px-2 py-1"
        >
          <option value="">All geofences</option>
          {geofences.map((g) => (
            <option key={g.id} value={g.id}>{g.name}</option>
          ))}
        </select>
        <input
          type="datetime-local"
          value={filters.start_date}
          onChange={(e) => setFilters({ ...filters, start_date: e.target.value })}
          className="border rounded px-2 py-1"
        />
        <input
          type="datetime-local"
          value={filters.end_date}
          onChange={(e) => setFilters({ ...filters, end_date: e.target.value })}
          className="border rounded px-2 py-1"
        />
        <select
          value={filters.limit}
          onChange={(e) => setFilters({ ...filters, limit: parseInt(e.target.value, 10) })}
          className="border rounded px-2 py-1"
        >
          {[25, 50, 100, 200, 500].map((n) => (
            <option key={n} value={n}>{n} / page</option>
          ))}
        </select>
      </div>

      <div className="mb-2 text-sm text-slate-600">
        Showing {violations.length} of {total} events
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left">
            <tr>
              <th className="px-3 py-2">When</th>
              <th className="px-3 py-2">Vehicle</th>
              <th className="px-3 py-2">Geofence</th>
              <th className="px-3 py-2">Event</th>
              <th className="px-3 py-2">Location</th>
            </tr>
          </thead>
          <tbody>
            {violations.map((v) => (
              <tr key={v.id} className="border-t">
                <td className="px-3 py-2 text-slate-600 whitespace-nowrap">
                  {new Date(v.timestamp).toLocaleString()}
                </td>
                <td className="px-3 py-2">{v.vehicle_number}</td>
                <td className="px-3 py-2">{v.geofence_name}</td>
                <td className="px-3 py-2">
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full ${
                      v.event_type === 'entry'
                        ? 'bg-green-100 text-green-700'
                        : 'bg-amber-100 text-amber-700'
                    }`}
                  >
                    {v.event_type}
                  </span>
                </td>
                <td className="px-3 py-2 text-slate-600">
                  {v.latitude.toFixed(4)}, {v.longitude.toFixed(4)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
