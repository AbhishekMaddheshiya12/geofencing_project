import { useEffect } from 'react'
import { MapContainer, TileLayer, Polygon, Marker, Popup, FeatureGroup, useMap, useMapEvents } from 'react-leaflet'
import { EditControl } from 'react-leaflet-draw'
import L from 'leaflet'
import type { Geofence, Vehicle } from '../api/types'

// Default Leaflet marker icons don't ship in the bundle; patch the URLs.
delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
})

const categoryColor: Record<string, string> = {
  delivery_zone: '#16a34a',
  restricted_zone: '#dc2626',
  toll_zone: '#ca8a04',
  customer_area: '#2563eb',
}

interface Props {
  geofences: Geofence[]
  vehicleLocations: { vehicle: Vehicle; lat: number; lng: number }[]
  onPolygonDrawn?: (latlngs: [number, number][]) => void
  onMapClick?: (lat: number, lng: number) => void
  center?: [number, number]
}

function ClickHandler({ onMapClick }: { onMapClick?: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) {
      onMapClick?.(e.latlng.lat, e.latlng.lng)
    },
  })
  return null
}

function Recenter({ center }: { center?: [number, number] }) {
  const map = useMap()
  useEffect(() => {
    if (center) map.flyTo(center, map.getZoom())
  }, [center, map])
  return null
}

export default function MapView({
  geofences,
  vehicleLocations,
  onPolygonDrawn,
  onMapClick,
  center,
}: Props) {
  return (
    <MapContainer center={center || [37.7749, -122.4194]} zoom={13} className="h-full w-full">
      <TileLayer
        attribution='&copy; OpenStreetMap'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <Recenter center={center} />
      <ClickHandler onMapClick={onMapClick} />

      {onPolygonDrawn && (
        <FeatureGroup>
          <EditControl
            position="topright"
            onCreated={(e: any) => {
              if (e.layerType === 'polygon') {
                const latlngs = e.layer
                  .getLatLngs()[0]
                  .map((p: any) => [p.lat, p.lng] as [number, number])
                // close polygon
                latlngs.push(latlngs[0])
                onPolygonDrawn(latlngs)
                e.layer.remove()
              }
            }}
            draw={{
              polygon: { allowIntersection: false, showArea: false },
              rectangle: false,
              circle: false,
              circlemarker: false,
              marker: false,
              polyline: false,
            }}
            edit={{ edit: false, remove: false }}
          />
        </FeatureGroup>
      )}

      {geofences.map((g) => (
        <Polygon
          key={g.id}
          positions={g.coordinates as [number, number][]}
          pathOptions={{
            color: categoryColor[g.category] || '#64748b',
            weight: 2,
            fillOpacity: 0.15,
          }}
        >
          <Popup>
            <div>
              <div className="font-semibold">{g.name}</div>
              <div className="text-xs text-slate-600">{g.category}</div>
              {g.description && <div className="text-xs mt-1">{g.description}</div>}
            </div>
          </Popup>
        </Polygon>
      ))}

      {vehicleLocations.map(({ vehicle, lat, lng }) => (
        <Marker key={vehicle.id} position={[lat, lng]}>
          <Popup>
            <div>
              <div className="font-semibold">{vehicle.vehicle_number}</div>
              <div className="text-xs">{vehicle.driver_name} ({vehicle.vehicle_type})</div>
              <div className="text-xs text-slate-600">
                {lat.toFixed(5)}, {lng.toFixed(5)}
              </div>
            </div>
          </Popup>
        </Marker>
      ))}
    </MapContainer>
  )
}
