import { useEffect, useState, useMemo } from 'react';
import { Droplets } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import MapView from '@/components/MapView';
import type { DonationRequest } from '@/lib/types';
import type { MapMarker } from '@/components/MapView';
import RequestList from '@/components/RequestList';
import RequestForm from '@/components/RequestForm';
import { useNavigate } from 'react-router-dom';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';

export default function RequestsPage() {
  const { token } = useAuth();
  const [requests, setRequests] = useState<DonationRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastId, setLastId] = useState<string | undefined>();
  const [hasMore, setHasMore] = useState(false);
  const [isNewRequestOpen, setIsNewRequestOpen] = useState(false);
  const [hoveredRequestId, setHoveredRequestId] = useState<string | null>(null);
  const [mapCenter, setMapCenter] = useState<[number, number]>([23.8103, 90.4125]);
  const [mapZoom, setMapZoom] = useState(13);
  const navigate = useNavigate();

  const loadRequests = async (isLoadMore = false) => {
    if (!token) return;
    setLoading(true);
    try {
      const filters = { status: 'Pending' };
      if (isLoadMore && lastId) filters.last_request_id = lastId;
      
      const res = await api.requests.list(token, filters);
      if (isLoadMore) {
        setRequests(prev => [...prev, ...res.requests]);
      } else {
        setRequests(res.requests);
      }
      setLastId(res.last_request_id);
      setHasMore(!!res.last_request_id);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRequests();
  }, [token]);

  const markers: MapMarker[] = useMemo(() => {
    return requests.map(req => ({
      id: req.id,
      lat: req.location_lat,
      lng: req.location_lng,
      hex: req.location_hex,
      status: req.status,
      type: 'request',
    }));
  }, [requests]);

  const handleMarkerClick = (marker: MapMarker) => {
    navigate(`/requests/${marker.id}`);
  };

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
        requests={requests}
        loading={loading}
        onSelect={(req) => navigate(`/requests/${req.id}`)}
        onHover={handleHover}
        onLoadMore={() => loadRequests(true)}
        hasMore={hasMore}
        onCreateClick={() => setIsNewRequestOpen(true)}
        selectedRequestId={hoveredRequestId || undefined}
      />
      
      <div className="flex-1 h-full z-0">
        <MapView 
          markers={markers} 
          onMarkerClick={handleMarkerClick}
          onRecenter={() => {
            setMapCenter([23.8103, 90.4125]);
            setMapZoom(13);
          }}
          highlightedId={hoveredRequestId}
          center={mapCenter}
          zoom={mapZoom}
        />
      </div>

      <Dialog open={isNewRequestOpen} onOpenChange={setIsNewRequestOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-2xl font-black flex items-center gap-2">
              <Droplets className="h-6 w-6 text-destructive" />
              New Blood Request
            </DialogTitle>
          </DialogHeader>
          <RequestForm 
            onSuccess={() => {
              setIsNewRequestOpen(false);
              loadRequests();
            }} 
            onCancel={() => setIsNewRequestOpen(false)} 
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}
