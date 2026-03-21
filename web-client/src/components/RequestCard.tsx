import type { DonationRequest } from '@/lib/types';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { MapPin, Droplet, Users, Calendar } from 'lucide-react';

interface RequestCardProps {
  request: DonationRequest;
  onClick?: (request: DonationRequest) => void;
  onHover?: (request: DonationRequest | null) => void;
  isActive?: boolean;
}

const statusColorMap: Record<string, string> = {
  Pending: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
  Accepted: 'bg-green-500/10 text-green-500 border-green-500/20',
  Canceled: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
  Fulfilled: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
};

export default function RequestCard({ request, onClick, onHover, isActive }: RequestCardProps) {
  return (
    <Card
      className={`cursor-pointer transition-all hover:border-primary/50 group ${isActive ? 'border-primary ring-1 ring-primary/50' : 'bg-card/50 backdrop-blur-sm'}`}
      onClick={() => onClick?.(request)}
      onMouseEnter={() => onHover?.(request)}
      onMouseLeave={() => onHover?.(null)}
    >
      <CardContent className="flex flex-col gap-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2">
            <div className="w-10 rounded-full bg-destructive/10 flex items-center justify-center shrink-0">
              <Droplet className="w-6 text-destructive" />
            </div>
            <div>
              <h3 className="font-bold text-lg leading-tight">{request.blood_type}</h3>
              <p className="text-xs text-muted-foreground break-words">{request.location_name}</p>
            </div>
          </div>
          <Badge variant="outline" className={statusColorMap[request.status] || ''}>
            {request.status}
          </Badge>
        </div>

        <div className="flex flex-wrap gap-x-4 gap-y-2 text-xs text-muted-foreground">
          <div className="flex items-center gap-1">
            <Users className="w-3" />
            <span>{request.bag_count} bags needed</span>
          </div>
          <div className="flex items-center gap-1">
            <Calendar className="w-3" />
            <span>{new Date(request.required_by_date).toLocaleDateString()}</span>
          </div>
        </div>

        <div className="flex items-center gap-1 text-[10px] uppercase font-bold tracking-wider text-muted-foreground/60 border-t pt-2 mt-1">
          <MapPin className="w-3" />
          <span className="truncate">Hex: {request.location_hex}</span>
        </div>
      </CardContent>
    </Card>
  );
}
