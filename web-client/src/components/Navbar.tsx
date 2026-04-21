import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { Bell, LogOut, Droplets } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import NotificationPanel from '@/components/NotificationPanel';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

export default function Navbar() {
  const { logout, token, userId } = useAuth();
  const navigate = useNavigate();

  // Basic JWT decode for demo (should use a proper lib or handle in AuthProvider)
  const user = { name: "Sakib", id: "me" }; 

  return (
    <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 h-16 sticky top-0 z-50">
      <div className="container mx-auto px-4 h-full flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
          <Droplets className="text-destructive h-8 w-8" />
          <span className="text-xl font-bold tracking-tight">BloodConnect</span>
        </Link>

        {token && (
          <div className="flex items-center gap-4">
            <div className="hidden sm:flex items-center gap-1">
              <Button asChild variant="ghost" size="sm" className="text-xs">
                <Link to="/requests">Browse</Link>
              </Button>
              <Button asChild variant="ghost" size="sm" className="text-xs">
                <Link to="/my-requests">My Requests</Link>
              </Button>
            </div>

            <Popover>
              <PopoverTrigger asChild>
                <Button variant="ghost" size="icon" className="relative">
                  <Bell className="h-5 w-5" />
                  <span className="absolute top-2 right-2 h-2 w-2 bg-destructive rounded-full" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-80 p-0" align="end" sideOffset={10}>
                <NotificationPanel />
              </PopoverContent>
            </Popover>

            <Link to={`/users/${userId || user.id}`}>
              <Avatar className="h-8 w-8 cursor-pointer hover:ring-2 ring-primary transition-all">
                <AvatarFallback className="bg-primary/10 text-primary">{user.name[0]}</AvatarFallback>
              </Avatar>
            </Link>

            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                logout();
                navigate('/login');
              }}
              title="Logout"
            >
              <LogOut className="h-5 w-5" />
            </Button>
          </div>
        )}
      </div>
    </nav>
  );
}
