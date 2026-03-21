import { useEffect, useState, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import MapView from '@/components/MapView';
import type { ExtendedDonationRequest, User } from '@/lib/types';
import type { MapMarker } from '@/components/MapView';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Droplet, MapPin, Calendar, Users, ArrowLeft, Loader2, Check, X, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';

export default function RequestDetailsPage() {
  const { id } = useParams();
  const { token, userId } = useAuth();
  const [data, setData] = useState<ExtendedDonationRequest | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState<'Accepted' | 'Declined' | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (token) {
      Promise.all([
        id ? api.requests.get(token, id) : Promise.resolve(null),
        api.auth.getMe(token)
      ])
        .then(([requestData, userData]) => {
          if (requestData) setData(requestData);
          setUser(userData);
        })
        .catch(err => {
          console.error(err);
          toast.error("Failed to load request or user details");
        })
        .finally(() => setLoading(false));
    }
  }, [token, id]);

  const { isEligible, ineligibilityReason } = useMemo(() => {
    if (!user || !data) return { isEligible: true };
    
    // Blood type check
    const userBloodType = user.health?.find(h => h.info_type === 'blood_type')?.details;
    if (userBloodType && userBloodType !== data.request.blood_type) {
      // In a real app, only compatible types would be notified, but we check here too.
      // But for now, let's say only exact matches are eligible for simplicity if that's the rule.
      // However, usually any compatible type is okay. The request says "disable if ineligible".
      // Let's stick to blood type if it's explicitly different and not compatible.
      // For now, let's just use the last donation date as the primary ineligibility reason mentioned in the request.
    }

    // Last donation date check (90 days)
    const lastDonationStr = user.health?.find(h => h.info_type === 'last_donation_date')?.details;
    if (lastDonationStr) {
      const lastDonation = new Date(lastDonationStr);
      const ninetyDaysAgo = new Date();
      ninetyDaysAgo.setDate(ninetyDaysAgo.getDate() - 90);
      
      if (lastDonation > ninetyDaysAgo) {
        return { 
          isEligible: false, 
          ineligibilityReason: `Last donation was on ${lastDonation.toLocaleDateString()}. You must wait 90 days between donations.`
        };
      }
    }

    return { isEligible: true };
  }, [user, data]);

  const handleAction = async (action: 'Accepted' | 'Declined') => {
    if (!token || !id) return;
    
    if (action === 'Accepted' && !isEligible) {
      toast.error(ineligibilityReason || "You are not eligible to donate at this time.");
      return;
    }

    setSubmitting(action);
    try {
      await api.requests.respond(token, id, action);
      const updated = await api.requests.get(token, id);
      setData(updated);
      toast.success(`Request ${action.toLowerCase()} successfully`);
    } catch (err: any) {
      console.error(err);
      toast.error(err.message || `Failed to ${action.toLowerCase()} request`);
    } finally {
      setSubmitting(null);
    }
  };

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!data) return <div className="p-8 text-center text-destructive font-bold">Request not found</div>;

  const { request, notified_users } = data;

  const markers: MapMarker[] = [
    {
      id: request.id,
      lat: request.location_lat,
      lng: request.location_lng,
      hex: request.location_hex,
      status: request.status,
      type: 'request',
    },
    ...notified_users.map(u => ({
      id: u.user_id,
      lat: u.lat,
      lng: u.lng,
      hex: u.h3_hex,
      status: u.action,
      track_id: u.user_id,
      type: 'user' as const,
    }))
  ];

  const myResponse = notified_users.find(u => u.user_id === userId)?.action;
  const hasResponded = myResponse === 'Accepted' || myResponse === 'Declined';

  return (
    <div className="h-[calc(100vh-64px)] flex flex-col md:flex-row relative">
      <div className="absolute top-4 left-4 z-40">
        <Button variant="outline" size="sm" className="bg-background/80 backdrop-blur-md gap-2 shadow-lg border-white/20" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
      </div>

      <div className="flex-1 h-1/2 md:h-full relative z-0">
        <MapView
          markers={markers}
          center={[request.location_lat, request.location_lng]}
          zoom={14}
        />
      </div>

      <div className="w-full md:w-[420px] h-1/2 md:h-full bg-background border-l flex flex-col z-10 shadow-2xl shrink-0 overflow-hidden">
        <div className="p-4 flex flex-col gap-4 overflow-y-auto flex-1 scrollbar-hide">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
              <Droplet className="h-6 w-6 text-destructive" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between gap-2">
                <h2 className="text-2xl font-black tracking-tight truncate">{request.blood_type}</h2>
                <Badge variant={request.status === 'Pending' ? 'default' : 'outline'} className="capitalize">
                  {request.status}
                </Badge>
              </div>
            </div>
          </div>

          <Card className="bg-muted/30 border-none shadow-none">
            <CardContent className="p-3 flex flex-col gap-3">
              <div className="flex items-start gap-2 text-sm">
                <MapPin className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                <span className="font-medium leading-tight">{request.location_name}</span>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Users className="h-3.5 w-3.5" />
                  <span>{request.bag_count} bags required</span>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Calendar className="h-3.5 w-3.5" />
                  <span>By {new Date(request.required_by_date).toLocaleDateString()}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="flex flex-col gap-1.5 pt-1">
            <h3 className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground/70">Description</h3>
            <p className="text-sm leading-relaxed text-foreground/90">{request.description || "No description provided."}</p>
          </div>

          <div className="flex flex-col gap-3 py-4 border-y border-white/5">
            <div className="flex items-center justify-between">
              <h3 className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground/70">Your Response</h3>
              {hasResponded && (
                <Badge variant="secondary" className="gap-1 text-[10px] h-5 px-1.5 font-medium">
                  {myResponse === 'Accepted' ? <Check className="h-3 w-3" /> : <X className="h-3 w-3" />}
                  {myResponse}
                </Badge>
              )}
            </div>
            
            {!isEligible && !hasResponded && (
              <div className="bg-amber-500/10 border border-amber-500/20 rounded-md p-2 flex items-start gap-2 text-[10px] text-amber-500">
                <AlertTriangle className="h-3 w-3 shrink-0 mt-0.5" />
                <p>{ineligibilityReason}</p>
              </div>
            )}

            <div className="flex gap-3">
              <Button
                className="flex-1 gap-2 h-10 text-sm font-bold"
                variant={myResponse === 'Accepted' ? 'default' : 'outline'}
                disabled={submitting !== null || request.status !== 'Pending' || hasResponded || !isEligible}
                onClick={() => handleAction('Accepted')}
              >
                {submitting === 'Accepted' ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
                Accept
              </Button>
              <Button
                className="flex-1 gap-2 h-10 text-sm font-bold"
                variant={myResponse === 'Declined' ? 'destructive' : 'outline'}
                disabled={submitting !== null || request.status !== 'Pending' || hasResponded}
                onClick={() => handleAction('Declined')}
              >
                {submitting === 'Declined' ? <Loader2 className="h-4 w-4 animate-spin" /> : <X className="h-4 w-4" />}
                Decline
              </Button>
            </div>
          </div>

          <div className="flex flex-col gap-3 flex-1 min-h-0">
            <h3 className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground/70">Engagement Monitor</h3>
            <div className="flex flex-col gap-2 overflow-y-auto pr-1">
              {notified_users.map((u) => (
                <div key={u.user_id} className={`flex items-center justify-between p-2.5 rounded-lg border text-xs transition-colors ${u.user_id === userId ? 'bg-primary/5 border-primary/20' : 'bg-card/40 border-white/5'}`}>
                  <div className="flex items-center gap-2.5">
                    <div className={`h-1.5 w-1.5 rounded-full ${u.action === 'Accepted' ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]' : u.action === 'Declined' ? 'bg-red-500' : 'bg-amber-500 animate-pulse'}`} />
                    <span className="font-medium">{u.user_id === userId ? 'You' : `User ${u.user_id.slice(-6)}`}</span>
                  </div>
                  <Badge variant="ghost" className="text-[9px] h-4 px-1 capitalize font-normal opacity-70">{u.action}</Badge>
                </div>
              ))}
              {notified_users.length === 0 && (
                <div className="h-24 flex flex-col items-center justify-center gap-2 text-muted-foreground/40 border-2 border-dashed border-white/5 rounded-xl">
                  <Users className="h-8 w-8" />
                  <p className="text-[10px] uppercase tracking-widest font-bold">No engagements yet</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
