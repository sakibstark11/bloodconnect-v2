import type { DonationRequest } from '@/lib/types';
import RequestCard from './RequestCard';
import { Button } from '@/components/ui/button';
import { Loader2, Plus, Search } from 'lucide-react';
import { Input } from '@/components/ui/input';

interface RequestListProps {
  requests: DonationRequest[];
  loading?: boolean;
  onSelect?: (request: DonationRequest) => void;
  onHover?: (request: DonationRequest | null) => void;
  onLoadMore?: () => void;
  selectedRequestId?: string;
  hasMore?: boolean;
  onCreateClick?: () => void;
}

export default function RequestList({
  requests,
  loading,
  onSelect,
  onHover,
  onLoadMore,
  selectedRequestId,
  hasMore,
  onCreateClick,
}: RequestListProps) {
  return (
    /* Changed h-screen to h-full so it inherits the wrapper's calc height.
       Added overflow-hidden to prevent the entire sidebar from scrolling. */
    <div className="flex flex-col h-full bg-background/80 backdrop-blur-md border-r w-[420px] shrink-0 shadow-2xl relative z-40 overflow-hidden">

      {/* Header: shrink-0 is vital so the search bar doesn't disappear when cards push up */}
      <div className="p-4 border-b flex flex-col gap-4 shrink-0">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-bold tracking-tight">Active Requests</h2>
          <Button size="sm" onClick={onCreateClick} className="gap-2">
            <Plus className="h-4 w-4" />
            New
          </Button>
        </div>
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by location or blood type..."
            className="pl-9 bg-muted/50 border-none"
          />
        </div>
      </div>

      {/* 1. flex-1: Takes up all remaining space.
          2. min-h-0: Allows the div to be smaller than its children (enables scroll).
          3. overflow-y-auto: Shows scrollbar only when content exceeds height.
      */}
      <div className="p-4 flex-1 overflow-y-auto min-h-0 flex flex-col gap-4">
        {requests.length === 0 && !loading ? (
          <div className="h-full flex flex-col items-center justify-center text-center text-muted-foreground">
            <Droplets className="w-12 mb-4 opacity-20" />
            <p className="text-sm">No matching requests found near you.</p>
          </div>
        ) : (
          <>
            {requests.map((req) => (
              <div key={req.id} className="shrink-0">
                <RequestCard
                  request={req}
                  onClick={onSelect}
                  onHover={onHover}
                  isActive={selectedRequestId === req.id}
                />
              </div>
            ))}

            {loading && (
              <div className="py-8 flex justify-center shrink-0">
                <Loader2 className="h-6 w-6 animate-spin text-primary" />
              </div>
            )}

            {hasMore && !loading && (
              <Button
                variant="ghost"
                className="mt-2 w-full text-xs text-muted-foreground hover:text-primary shrink-0"
                onClick={onLoadMore}
              >
                Load more results
              </Button>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function Droplets(props: any) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M7 16.3c2.2 0 4-1.8 4-4 0-3.3-4-6-4-6s-4 2.7-4 6c0 2.2 1.8 4 4 4Z" />
      <path d="M17 15.8c1.7 0 3-1.3 3-3 0-2.5-3-4.5-3-4.5s-3 2-3 4.5c0 1.7 1.3 3 3 3Z" />
    </svg>
  );
}
