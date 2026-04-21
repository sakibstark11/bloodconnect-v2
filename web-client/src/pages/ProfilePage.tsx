import { useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import MapView from '@/components/MapView';
import type { User } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { MapPin, Heart, ShieldPlus, Loader2, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

export default function ProfilePage() {
  const { token } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (token) {
      api.auth.getMe(token)
        .then((u) => {
          setUser(u);
        })
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [token]);

  const handleAddLocation = async (lat: number, lng: number) => {
    if (!token) return;
    setSaving(true);
    try {
      await api.user.addLocation(token, { lat, lng });
      const updatedUser = await api.auth.getMe(token);
      setUser(updatedUser);
      toast.success("Location added successfully");
    } catch (err) {
      toast.error(err.message || "Failed to add location");
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteLocation = async (h3hex: string) => {
    if (!token) return;
    try {
      await api.user.deleteLocation(token, h3hex);
      const updatedUser = await api.auth.getMe(token);
      setUser(updatedUser);
      toast.success("Location deleted successfully");
    } catch (err) {
      toast.error(err.message || "Failed to delete location");
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
                <div className="p-3 bg-destructive/10 rounded-md border border-destructive/20 text-destructive font-black text-center text-4xl">
                  {user.health?.find(h => h.info_type === 'blood_type')?.details || '??'}
                </div>
                <p className="text-xs text-muted-foreground text-center italic">Blood type can only be confirmed once.</p>
              </div>
            </CardContent>
          </Card>

          <Card className="bg-card/50 backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <MapPin className="h-5 w-5 text-primary" />
                Your Locations
              </CardTitle>
              <CardDescription>At least one is required</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {user.locations?.map((loc) => (
                <div key={loc.h3_hex} className="flex items-center justify-between p-2 bg-background/50 rounded-md text-sm border border-border/50">
                  <div className="flex flex-col">
                    <span className="font-mono text-[10px] text-muted-foreground">{loc.h3_hex}</span>
                    <span>{loc.lat.toFixed(4)}, {loc.lng.toFixed(4)}</span>
                  </div>
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-8 w-8 text-destructive hover:bg-destructive/10"
                    onClick={() => handleDeleteLocation(loc.h3_hex)}
                    disabled={(user.locations?.length || 0) <= 1}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
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
                    Preferred Donation Locations
                  </CardTitle>
                  <CardDescription>Click on the map to add a new donation area</CardDescription>
                </div>
                {saving && <Loader2 className="h-4 w-4 animate-spin text-primary" />}
              </div>
            </CardHeader>
            <CardContent className="p-0 flex-1 relative">
              <MapView 
                markers={user.locations?.map(loc => ({
                  id: loc.h3_hex,
                  lat: loc.lat,
                  lng: loc.lng,
                  hex: loc.h3_hex,
                  status: 'Fulfilled',
                  type: 'user'
                })) || []}
                onClickMap={handleAddLocation}
                center={user.locations && user.locations.length > 0 ? [user.locations[0].lat, user.locations[0].lng] : undefined}
                zoom={14}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
