import { useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import type { Notification } from '@/lib/types';
import { useNavigate } from 'react-router-dom';
import { Loader2, MessageSquare, AlertTriangle, CheckCircle } from 'lucide-react';

export default function NotificationPanel() {
  const { token } = useAuth();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    if (token) {
      api.notifications.list(token)
        .then(res => setNotifications(res.notifications))
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [token]);

  const handleNotificationClick = (notif: Notification) => {
    if (notif.metadata?.request_id) {
      navigate(`/requests/${notif.metadata.request_id}`);
    }
  };

  const getIcon = (type: string) => {
    switch (type) {
      case 'blood_donation_request': return <AlertTriangle className="h-5 w-5 text-amber-500" />;
      case 'blood_donation_request_acceptance': return <CheckCircle className="h-5 w-5 text-green-500" />;
      default: return <MessageSquare className="h-5 w-5 text-primary" />;
    }
  };

  if (loading) {
    return (
      <div className="p-8 flex justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="flex flex-col max-h-[400px]">
      <div className="p-4 border-b">
        <h3 className="font-semibold">Notifications</h3>
      </div>
      <div className="overflow-y-auto">
        {notifications.length === 0 ? (
          <div className="p-8 text-center text-sm text-muted-foreground">
            No notifications yet
          </div>
        ) : (
          notifications.map((n) => (
            <div
              key={n.id}
              onClick={() => handleNotificationClick(n)}
              className="p-4 border-b last:border-0 hover:bg-accent cursor-pointer transition-colors flex gap-3"
            >
              <div className="shrink-0 mt-1">{getIcon(n.type)}</div>
              <div className="flex flex-col gap-1">
                <p className="text-sm font-medium leading-none">{n.title}</p>
                <p className="text-sm text-muted-foreground line-clamp-2">{n.content}</p>
                <span className="text-[10px] text-muted-foreground uppercase font-bold tracking-wider">
                  {new Date(n.created_at).toLocaleDateString()}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
      {notifications.length > 0 && (
        <div className="p-2 border-t text-center">
          <button className="text-xs text-primary hover:underline font-medium">
            View all
          </button>
        </div>
      )}
    </div>
  );
}
