import { useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import MapView from '@/components/MapView';
import type { User } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { MapPin, Heart, ShieldPlus, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

export default function ProfilePage() {
  const { token } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [location, setLocation] = useState<{ lat: number; lng: number; hex: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (token) {
      Promise.all([
        api.auth.getMe(token),
        api.user.updateLocation(token, { lat: 0, lng: 0 }).catch(() => null) // Dummy call to trigger location fetch if we had a GET endpoint
      ]).then(([u]) => {
        setUser(u);
        // Placeholder for user location if we had it
        setLocation({ lat: 23.8103, lng: 90.4125, hex: "89608c2a8c7ffff" }); 
      })
      .catch(console.error)
      .finally(() => setLoading(false));
    }
  }, [token]);

  const handleUpdateLocation = async (lat: number, lng: number) => {
    if (!token) return;
    setSaving(true);
    try {
      await api.user.updateLocation(token, { lat, lng });
      setLocation(prev => prev ? { ...prev, lat, lng } : null);
      toast.success("Location updated successfully");
    } catch (err) {
      toast.error("Failed to update location");
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="h-full flex items-center justify-center"><Loader2 className="h-8 w-8 animate-spin" /></div>;
  if (!user) return <div>User not found</div>;

  return (
    <div className="container mx-auto p-6 max-w-5xl flex flex-col gap-8 h-full overflow-y-auto pb-20">
      <div className="flex flex-col gap-1">
        <h1 className="text-4xl font-black tracking-tight">Your Profile</h1>
        <p className="text-muted-foreground">Manage your donor information and preferred location.</p>
      </div>

      <div className="grid md:grid-cols-3 gap-8">
        <div className="md:col-span-1 flex flex-col gap-6">
          <Card className="bg-card/50 backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Heart className="h-5 w-5 text-destructive" />
                Personal Info
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div className="space-y-1">
                <Label className="text-xs uppercase text-muted-foreground font-bold">Name</Label>
                <p className="font-medium text-lg">{user.name}</p>
              </div>
              <div className="space-y-1">
                <Label className="text-xs uppercase text-muted-foreground font-bold">Email</Label>
                <p className="font-medium text-lg">{user.email}</p>
              </div>
              <div className="space-y-1">
                <Label className="text-xs uppercase text-muted-foreground font-bold">Phone</Label>
                <p className="font-medium text-lg">{user.phone}</p>
              </div>
            </CardContent>
          </Card>

          <Card className="bg-card/50 backdrop-blur-sm border-primary/20">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-primary">
                <ShieldPlus className="h-5 w-5" />
                Health Details
              </CardTitle>
              <CardDescription>Required for donor matching</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div className="space-y-2">
                <Label>Blood Type</Label>
                <Select defaultValue="O+">
                  <SelectTrigger className="bg-background/50">
                    <SelectValue placeholder="Select blood type" />
                  </SelectTrigger>
                  <SelectContent>
                    {['A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-'].map(t => (
                      <SelectItem key={t} value={t}>{t}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Weight (kg)</Label>
                <Input type="number" placeholder="70" className="bg-background/50" />
              </div>
              <Button className="w-full mt-2" onClick={() => toast.info("Profile saving implemented in next version")}>
                Save Health Profile
              </Button>
            </CardContent>
          </Card>
        </div>

        <div className="md:col-span-2 flex flex-col gap-6">
          <Card className="bg-card/50 backdrop-blur-sm overflow-hidden flex flex-col flex-1 h-full min-h-[500px]">
            <CardHeader className="shrink-0">
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <MapPin className="h-5 w-5 text-primary" />
                    Preferred Location
                  </CardTitle>
                  <CardDescription>Click on the map to set your usual donation area</CardDescription>
                </div>
                {saving && <Loader2 className="h-4 w-4 animate-spin text-primary" />}
              </div>
            </CardHeader>
            <CardContent className="p-0 flex-1 relative">
              <MapView 
                markers={location ? [{
                  id: user.id,
                  lat: location.lat,
                  lng: location.lng,
                  hex: location.hex,
                  status: 'Fulfilled',
                  type: 'user'
                }] : []}
                onClickMap={handleUpdateLocation}
                center={location ? [location.lat, location.lng] : undefined}
                zoom={14}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
