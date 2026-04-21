import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import MapView from '@/components/MapView';
import type { DonationRequest } from '@/lib/types';
import type { MapMarker } from '@/components/MapView';
import RequestList from '@/components/RequestList';

export default function MyRequestsPage() {
  const { token, userId } = useAuth();
  const [requests, setRequests] = useState<DonationRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [hoveredRequestId, setHoveredRequestId] = useState<string | null>(null);
  const [mapCenter, setMapCenter] = useState<[number, number]>([23.8103, 90.4125]);
  const [mapZoom, setMapZoom] = useState(13);
  const navigate = useNavigate();

  useEffect(() => {
    if (!token) {
      Promise.resolve().then(() => {
        setLoading(false);
        setRequests([]);
      });
      return;
    }

    Promise.resolve().then(() => setLoading(true));
    const resolveId = async () => {
      if (userId) return userId;
      const me = await api.auth.getMe(token);
      return me.id;
    };

    Promise.all([resolveId(), api.requests.list(token, {})])
      .then(([uid, res]) => {
        const mine = (res.requests || []).filter((r) => r.user_id === uid);
        setRequests(mine);
      })
      .catch((err) => {
        console.error(err);
        setRequests([]);
      })
      .finally(() => setLoading(false));
  }, [token, userId]);

  const markers: MapMarker[] = useMemo(() => {
    return requests.map((req) => ({
      id: req.id,
      lat: req.location_lat,
      lng: req.location_lng,
      hex: req.location_hex,
      status: req.status,
      type: 'request',
    }));
  }, [requests]);

  const handleHover = (request: DonationRequest | null) => {
    setHoveredRequestId(request?.id || null);
    if (request) {
      setMapCenter([request.location_lat, request.location_lng]);
      setMapZoom(15);
    }
  };

  return (
    <div className="h-[calc(100vh-64px)] flex">
      <RequestList
        title="My Requests"
        showCreate={false}
        requests={requests}
        loading={loading}
        onSelect={(req) => navigate(`/requests/${req.id}`)}
        onHover={handleHover}
        selectedRequestId={hoveredRequestId || undefined}
      />

      <div className="flex-1 h-full z-0">
        <MapView
          markers={markers}
          onMarkerClick={(marker) => navigate(`/requests/${marker.id}`)}
          onRecenter={() => {
            setMapCenter([23.8103, 90.4125]);
            setMapZoom(13);
          }}
          highlightedId={hoveredRequestId}
          center={mapCenter}
          zoom={mapZoom}
        />
      </div>
    </div>
  );
}

