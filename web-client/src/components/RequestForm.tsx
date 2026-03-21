import { useState } from 'react';
import { useForm, type SubmitHandler } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { api } from '@/lib/api';
import { useAuth } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormDescription,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import MapView from './MapView';
import { MapPin, Droplet, Loader2 } from 'lucide-react';

const formSchema = z.object({
  location_name: z.string().min(3),
  blood_type: z.string(),
  bag_count: z.number().min(1, 'At least 1 bag is required'),
  required_by_date: z.string(),
  description: z.string().optional(),
  requester_info: z.string().min(5),
  location_lat: z.number(),
  location_lng: z.number(),
});

type FormValues = z.infer<typeof formSchema>;

interface RequestFormProps {
  onSuccess: () => void;
  onCancel: () => void;
}

export default function RequestForm({ onSuccess, onCancel }: RequestFormProps) {
  const { token } = useAuth();
  const [loading, setLoading] = useState(false);
  const [latLng, setLatLng] = useState<[number, number] | null>(null);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      location_name: "",
      blood_type: "O+",
      bag_count: 1,
      required_by_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      description: "",
      requester_info: "",
      location_lat: 0,
      location_lng: 0,
    },
  });

  const onSubmit: SubmitHandler<FormValues> = async (values) => {
    if (!token) return;
    setLoading(true);
    try {
      await api.requests.create(token, {
        ...values,
        required_by_date: new Date(values.required_by_date).toISOString(),
      });
      onSuccess();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleMapClick = (lat: number, lng: number) => {
    setLatLng([lat, lng]);
    form.setValue('location_lat', lat);
    form.setValue('location_lng', lng);
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="flex flex-col gap-6">
            <FormField
              control={form.control}
              name="location_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Location Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Dhaka Medical College..." {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="blood_type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Blood Type</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {['A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-'].map(t => (
                          <SelectItem key={t} value={t}>{t}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="bag_count"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Bag Count</FormLabel>
                    <FormControl>
                      <Input 
                        type="number" 
                        min="1" 
                        {...field} 
                        onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="required_by_date"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Required By</FormLabel>
                  <FormControl>
                    <Input type="date" {...field} />
                  </FormControl>
                  <FormDescription>At least 2 days from today</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="requester_info"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Requester Info</FormLabel>
                  <FormControl>
                    <Input placeholder="John Doe - 017..." {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className="flex flex-col gap-4">
            <FormLabel className="flex items-center gap-2">
              <MapPin className="h-4 w-4 text-primary" />
              Location on Map
            </FormLabel>
            <div className="h-[300px] border rounded-lg overflow-hidden bg-card/20 border-dashed border-primary/30 relative">
              <MapView 
                markers={latLng ? [{
                  id: 'temp',
                  lat: latLng[0],
                  lng: latLng[1],
                  hex: "temp",
                  status: 'Pending',
                  type: 'request'
                }] : []}
                onClickMap={handleMapClick}
              />
              {!latLng && (
                <div className="absolute inset-0 flex items-center justify-center bg-background/50 pointer-events-none text-xs font-bold uppercase tracking-widest text-primary/60">
                  Click on map to set location
                </div>
              )}
            </div>
            {latLng && (
              <p className="text-[10px] text-muted-foreground font-mono">
                Lat: {latLng[0].toFixed(5)}, Lng: {latLng[1].toFixed(5)}
              </p>
            )}
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Additional Description</FormLabel>
                  <FormControl>
                    <Textarea 
                      placeholder="Any specific requirement or instruction..." 
                      className="resize-none h-24"
                      {...field} 
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-6 border-t">
          <Button variant="outline" type="button" onClick={onCancel}>Cancel</Button>
          <Button type="submit" disabled={loading || !latLng} className="gap-2">
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Droplet className="h-4 w-4" />}
            Submit Request
          </Button>
        </div>
      </form>
    </Form>
  );
}
