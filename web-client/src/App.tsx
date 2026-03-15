import { useState } from 'react';
import { MapContainer, TileLayer, Marker, Popup, Circle } from 'react-leaflet';
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
  status: string;
  locationName: string;
}

function App() {
  const [requestId, setRequestId] = useState('');
  const [requestData, setRequestData] = useState<RequestData | null>(null);
  const [searchRadiusKm, setSearchRadiusKm] = useState<number>(0);
  const [notifiedUsers, setNotifiedUsers] = useState<Array<{lat: number, lng: number}>>([]);
  
  // Bangladesh center default
  const defaultCenter: [number, number] = [23.8103, 90.4125];

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!requestId) return;

    try {
      // In a real scenario, this would fetch from our Go backend:
      // const res = await fetch(`http://localhost:8080/requests/${requestId}/visualization-data`);
      // const data = await res.json();
      
      // For now, let's mock the response visualization data since we don't have
      // an explicit visualization endpoint returning users and ring radius yet
      console.log('Fetching visualization data for:', requestId);
      
      // MOCK DATA based on DB seeder
      setRequestData({
        id: requestId,
        bloodType: 'A+',
        bagCount: 2,
        lat: 23.8103 + (Math.random() - 0.5) * 0.1,
        lng: 90.4125 + (Math.random() - 0.5) * 0.1,
        status: 'Pending',
        locationName: 'Dhaka Hospital'
      });
      
      // Simulate ring 2 radius (approx 10km radius per ring, so 20km)
      setSearchRadiusKm(20);
      
      // Simulate 3 notified users nearby
      setNotifiedUsers([
        { lat: 23.8153, lng: 90.4175 },
        { lat: 23.8013, lng: 90.4025 },
        { lat: 23.8203, lng: 90.4225 },
      ]);

    } catch (err) {
      console.error('Failed to fetch request data', err);
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
               {/* Request Origin Marker */}
               <Marker position={[requestData.lat, requestData.lng]}>
                 <Popup>
                   <strong>Blood Request Origin</strong><br/>
                   Needed: {requestData.bloodType} ({requestData.bagCount} bags)
                 </Popup>
               </Marker>

               {/* Search Radius Representation */}
               <Circle 
                 center={[requestData.lat, requestData.lng]} 
                 radius={searchRadiusKm * 1000} // meters
                 pathOptions={{ color: '#e53935', fillColor: '#e53935', fillOpacity: 0.1 }}
               />

               {/* Notified Users Markers */}
               {notifiedUsers.map((u, idx) => (
                 <Marker key={idx} position={[u.lat, u.lng]} opacity={0.6}>
                   <Popup>Notified Donor #{idx+1}</Popup>
                 </Marker>
               ))}
             </>
          )}

        </MapContainer>
      </div>
    </>
  )
}

export default App
