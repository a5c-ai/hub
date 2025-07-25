'use client';

import { usePWAContext } from '@/components/providers/PWAProvider';
import { WifiIcon } from '@heroicons/react/24/outline';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import { cn } from '@/lib/utils';

interface OfflineIndicatorProps {
  className?: string;
  showOnlineStatus?: boolean;
}

export function OfflineIndicator({ className, showOnlineStatus = false }: OfflineIndicatorProps) {
  const { isOnline } = usePWAContext();

  // Don't show anything if online and not configured to show online status
  if (isOnline && !showOnlineStatus) {
    return null;
  }

  return (
    <div
      className={cn(
        'inline-flex items-center space-x-2 px-2 py-1 rounded-md text-xs font-medium transition-colors',
        isOnline
          ? 'bg-success/10 text-success border border-success/20'
          : 'bg-warning/10 text-warning border border-warning/20',
        className
      )}
    >
      {isOnline ? (
        <WifiIcon className="h-3 w-3" />
      ) : (
        <ExclamationTriangleIcon className="h-3 w-3" />
      )}
      <span>{isOnline ? 'Online' : 'Offline'}</span>
    </div>
  );
}

export function OfflineBanner() {
  const { isOnline } = usePWAContext();

  if (isOnline) {
    return null;
  }

  return (
    <div className="bg-warning/90 border-b border-warning text-warning-foreground">
      <div className="container mx-auto px-4 py-2">
        <div className="flex items-center justify-center space-x-2 text-sm">
          <ExclamationTriangleIcon className="h-4 w-4" />
          <span>You&apos;re offline. Some features may be limited.</span>
        </div>
      </div>
    </div>
  );
}