import * as React from 'react';
import { useEffect } from 'react';
import { MapContainer, TileLayer, Polygon, CircleMarker, useMap, useMapEvents, ZoomControl } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import * as h3 from 'h3-js';
import { LocateFixed } from 'lucide-react';

export interface MapMarker {
  id: string;
  lat: number;
  lng: number;
  hex: string;
  status: string;
  label?: string;
  type: 'request' | 'user';
}

interface MapViewProps {
  markers: MapMarker[];
  center?: [number, number];
  zoom?: number;
  highlightedId?: string | null;
  onClickMap?: (lat: number, lng: number) => void;
  onMarkerClick?: (marker: MapMarker) => void;
  onRecenter?: () => void;
  className?: string;
}

const statusColorMap: Record<string, string> = {
  Pending: '#f59e0b', // amber
  Accepted: '#22c55e', // green
  Declined: '#ef4444', // red
  Canceled: '#6b7280', // gray
  Fulfilled: '#3b82f6', // blue
  Processing: '#8b5cf6', // violet
};

const requestMarkerColor = '#f43f5e'; // rose (distinct from status colors)

function MapEvents({ onClick }: { onClick?: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) {
      onClick?.(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

function ChangeView({ center, zoom }: { center: [number, number]; zoom: number }) {
  const map = useMap();
  useEffect(() => {
    map.setView(center, zoom);
  }, [center, zoom, map]);
  return null;
}

function getBoundary(hex: string): [number, number][] {
  try {
    return h3.cellToBoundary(hex) as [number, number][];
  } catch {
    return [];
  }
}

export default function MapView({
  markers,
  center = [23.8103, 90.4125], // Default to Dhaka
  zoom = 13,
  highlightedId,
  onClickMap,
  onMarkerClick,
  onRecenter,
  className = "h-full w-full",
}: MapViewProps) {
  return (
    <div className={className}>
      <MapContainer
        center={center}
        zoom={zoom}
        scrollWheelZoom={true}
        zoomControl={false}
        className="h-full w-full outline-none"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
          url="https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png"
        />

        {markers.map((marker) => {
          const boundary = getBoundary(marker.hex);

          const isHighlighted = marker.id === highlightedId;
          const baseColor =
            marker.type === 'request'
              ? requestMarkerColor
              : (statusColorMap[marker.status] || statusColorMap['Processing']);
          const color = isHighlighted ? '#fff' : baseColor;

          return (
            <React.Fragment key={marker.id}>
              <Polygon
                positions={boundary}
                pathOptions={{
                  fillColor: color,
                  fillOpacity: isHighlighted ? 0.6 : 0.3,
                  color: color,
                  weight: isHighlighted ? 3 : 2,
                }}
                interactive={!!onMarkerClick}
                eventHandlers={{
                  click: () => onMarkerClick?.(marker),
                }}
              />
              <CircleMarker
                center={[marker.lat, marker.lng]}
                pathOptions={{
                  fillColor: color,
                  fillOpacity: 1,
                  color: marker.type === 'request' ? '#000' : (isHighlighted ? '#000' : '#fff'),
                  weight: marker.type === 'request' ? (isHighlighted ? 3 : 2) : (isHighlighted ? 2 : 1),
                }}
                interactive={!!onMarkerClick}
                className={onMarkerClick ? 'cursor-pointer' : ''}
              />
            </React.Fragment>
          );
        })}

        <ZoomControl position="bottomleft" />
        <div className="leaflet-bottom leaflet-left !mb-20 !ml-3 pointer-events-auto">
          <button
            onClick={() => onRecenter?.()}
            className="bg-white hover:bg-gray-100 text-black p-2 rounded-md shadow-md border border-black/20 transition-colors flex items-center justify-center"
            title="Recenter Map"
          >
            <LocateFixed className="h-4 w-4" />
          </button>
        </div>

        <MapEvents onClick={onClickMap} />
        <ChangeView center={center} zoom={zoom} />
      </MapContainer>
    </div>
  );
}
