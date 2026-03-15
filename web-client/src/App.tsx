import { useState, useEffect } from 'react';
import { MapContainer, TileLayer, Marker, Popup, Polygon } from 'react-leaflet';
import { cellToBoundary } from 'h3-js';
import { Search, Droplets } from 'lucide-react';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

// Fix leaflet icon paths
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

// Mocked Data structures based on backend responses
interface RequestData {
  id: string;
  bloodType: string;
  bagCount: number;
  lat: number;
  lng: number;
  h3Hex: string;
  status: string;
  locationName: string;
}

interface NotifiedUserData {
  user_id: string;
  lat: number;
  lng: number;
  h3_hex: string;
  action: string;
}

function App() {
  const [requestId, setRequestId] = useState(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get('id') || '';
  });
  const [requestData, setRequestData] = useState<RequestData | null>(null);
  const [notifiedUsers, setNotifiedUsers] = useState<NotifiedUserData[]>([]);
  const [isTracking, setIsTracking] = useState(!!requestId);
  
  // Bangladesh center default
  const defaultCenter: [number, number] = [23.8103, 90.4125];

  const fetchRequestData = async (reqId: string) => {
    try {
      const res = await fetch(`http://localhost:8080/requests/${reqId}`);
      if (!res.ok) {
        throw new Error(`Error fetching request: ${res.statusText}`);
      }
      const data = await res.json();
      
      const req = data.request;
      setRequestData({
        id: req.ID,
        bloodType: req.BloodType,
        bagCount: req.BagCount,
        lat: req.LocationLat,
        lng: req.LocationLng,
        h3Hex: req.LocationHex,
        status: req.Status,
        locationName: req.LocationName
      });
      
      setNotifiedUsers(data.notified_users || []);
    } catch (err) {
      console.error('Failed to fetch request data', err);
    }
  };

  useEffect(() => {
    let interval: number;
    if (isTracking && requestId) {
      // Fetch immediately once
      fetchRequestData(requestId);
      // Then poll every 5 seconds
      interval = window.setInterval(() => {
        fetchRequestData(requestId);
      }, 5000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isTracking, requestId]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!requestId) return;
    
    // Update URL without reloading the page
    const newUrl = new URL(window.location.href);
    newUrl.searchParams.set('id', requestId);
    window.history.pushState({}, '', newUrl);

    setIsTracking(true);
  };
  
  // Helper to draw a hex cell as Leaflet LatLng tuples
  const getHexPolygon = (hex: string): [number, number][] => {
    try {
      const boundary = cellToBoundary(hex);
      // h3 returns [lat, lng], Leaflet Polygon expects [lat, lng] tuples
      return boundary as [number, number][];
    } catch (e) {
      return [];
    }
  };

  return (
    <>
      <header className="header">
        <h1><Droplets /> BloodConnect Monitor</h1>
        <form className="search-bar" onSubmit={handleSearch}>
          <input 
            type="text" 
            placeholder="Enter Request ID (e.g. request_01KKS...)" 
            value={requestId}
            onChange={(e) => setRequestId(e.target.value)}
          />
          <button type="submit"><Search size={18} style={{ verticalAlign: 'middle', marginRight: '4px' }}/> Track</button>
        </form>
      </header>

      {requestData && (
        <div className="request-info-panel">
          <h3>Request Information</h3>
          <div className="info-row">
            <span className="info-label">Blood Type:</span>
            <span>{requestData.bloodType}</span>
          </div>
          <div className="info-row">
            <span className="info-label">Bags Required:</span>
            <span>{requestData.bagCount}</span>
          </div>
          <div className="info-row">
            <span className="info-label">Status:</span>
            <span>{requestData.status}</span>
          </div>
          <div className="info-row">
            <span className="info-label">Location:</span>
            <span>{requestData.locationName}</span>
          </div>
          <div className="info-row">
            <span className="info-label">Notified Donors:</span>
            <span>{notifiedUsers.length}</span>
          </div>
        </div>
      )}

      <div className="map-container">
        <MapContainer 
          center={requestData ? [requestData.lat, requestData.lng] : defaultCenter} 
          zoom={requestData ? 12 : 7} 
          style={{ height: '100%', width: '100%' }}
        >
          <TileLayer
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />

           {requestData && (
             <>
               {/* Request Origin Marker & Hex */}
               {requestData.lat !== undefined && requestData.lng !== undefined && (
                 <>
                   <Marker position={[requestData.lat, requestData.lng]}>
                     <Popup>
                       <strong>Blood Request Origin</strong><br/>
                       Needed: {requestData.bloodType} ({requestData.bagCount} bags)
                     </Popup>
                   </Marker>
                   
                   {requestData.h3Hex && (
                     <Polygon 
                       positions={getHexPolygon(requestData.h3Hex)} 
                       pathOptions={{ color: '#e53935', fillColor: '#e53935', fillOpacity: 0.2, weight: 2 }} 
                     />
                   )}
                 </>
               )}

               {/* Notified Users Hexagons */}
               {notifiedUsers
                 .filter(u => u.h3_hex)
                 .map((u, idx) => {
                   let color = '#2b82e2'; // pending color (blue)
                   let fillOpacity = 0.2;
                   
                   if (u.action === 'Accepted' || u.action === 'Donated') {
                     color = '#43a047'; // accepted/donated color (green)
                     fillOpacity = 0.5;
                   } else if (u.action === 'Declined') {
                     color = '#757575'; // declined color (grey)
                   }
                   
                   return (
                     <Polygon 
                       key={idx}
                       positions={getHexPolygon(u.h3_hex)} 
                       pathOptions={{ color, fillColor: color, fillOpacity, weight: 1 }} 
                     >
                       <Popup>Notified Donor #{idx+1} ({u.action || 'Pending'})</Popup>
                     </Polygon>
                   );
               })}
             </>
          )}

        </MapContainer>
      </div>
    </>
  )
}

export default App
