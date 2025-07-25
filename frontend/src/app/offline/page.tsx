'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { WifiIcon } from '@heroicons/react/24/outline';

export default function OfflinePage() {
  const router = useRouter();
  const [isOnline, setIsOnline] = useState(true);

  useEffect(() => {
    const updateOnlineStatus = () => {
      setIsOnline(navigator.onLine);
    };

    // Check initial status
    updateOnlineStatus();

    // Listen for online/offline events
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);

    return () => {
      window.removeEventListener('online', updateOnlineStatus);
      window.removeEventListener('offline', updateOnlineStatus);
    };
  }, []);

  useEffect(() => {
    // Redirect to dashboard when back online
    if (isOnline) {
      router.push('/dashboard');
    }
  }, [isOnline, router]);

  const handleRetry = () => {
    if (navigator.onLine) {
      router.push('/dashboard');
    } else {
      // Force a reload to check connection
      window.location.reload();
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        <div className="mb-8">
          <div className="mx-auto h-24 w-24 bg-muted rounded-full flex items-center justify-center mb-6">
            <WifiIcon className="h-12 w-12 text-muted-foreground" />
          </div>
          
          <h1 className="text-2xl font-bold text-foreground mb-2">
            You&apos;re offline
          </h1>
          
          <p className="text-muted-foreground mb-6">
            It looks like you&apos;ve lost your internet connection. Don&apos;t worry - you can still browse cached content and create new issues offline.
          </p>
        </div>

        <div className="space-y-4">
          <Button 
            onClick={handleRetry}
            className="w-full"
            variant="default"
          >
            Try again
          </Button>
          
          <Button 
            onClick={() => router.push('/repositories')}
            variant="ghost"
            className="w-full"
          >
            Browse cached repositories
          </Button>
        </div>

        <div className="mt-8 p-4 bg-muted rounded-lg">
          <h3 className="font-medium text-foreground mb-2">
            What you can do offline:
          </h3>
          <ul className="text-sm text-muted-foreground space-y-1 text-left">
            <li>• View recently accessed repositories</li>
            <li>• Read cached issues and pull requests</li>
            <li>• Browse cached code files</li>
            <li>• Create new issues (will sync when online)</li>
          </ul>
        </div>

        <div className="mt-6 text-xs text-muted-foreground">
          Connection status: {isOnline ? 'Online' : 'Offline'}
        </div>
      </div>
    </div>
  );
}