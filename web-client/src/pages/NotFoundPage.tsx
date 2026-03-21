import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Frown, Home } from 'lucide-react';

export default function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6 text-center">
      <div className="h-24 w-24 rounded-full bg-destructive/10 flex items-center justify-center">
        <Frown className="h-12 w-12 text-destructive" />
      </div>
      <div className="space-y-2">
        <h1 className="text-4xl font-bold tracking-tighter">404 - Not Found</h1>
        <p className="text-muted-foreground max-w-[400px]">
          The page you are looking for doesn't exist or has been moved to another location.
        </p>
      </div>
      <Link to="/">
        <Button className="gap-2">
          <Home className="h-4 w-4" /> Go Home
        </Button>
      </Link>
    </div>
  );
}
