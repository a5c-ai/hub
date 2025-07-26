'use client';

import { useEffect } from 'react';
import { Button } from '@/components/ui/Button';

interface ErrorProps {
  error: Error;
  reset: () => void;
}

export default function GlobalError({ error, reset }: ErrorProps) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        <h1 className="text-2xl font-bold text-foreground mb-4">Something went wrong</h1>
        <p className="text-muted-foreground mb-6">{error.message}</p>
        <Button onClick={() => reset()} variant="default">
          Retry
        </Button>
      </div>
    </div>
  );
}
