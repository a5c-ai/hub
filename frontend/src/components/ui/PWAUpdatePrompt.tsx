'use client';

import { useState } from 'react';
import { Button } from './Button';
import { Card } from './Card';
import { usePWAContext } from '@/components/providers/PWAProvider';
import { XMarkIcon, ArrowPathIcon } from '@heroicons/react/24/outline';

interface PWAUpdatePromptProps {
  onDismiss?: () => void;
  className?: string;
}

export function PWAUpdatePrompt({ onDismiss, className }: PWAUpdatePromptProps) {
  const { updateAvailable, updateApp } = usePWAContext();
  const [isUpdating, setIsUpdating] = useState(false);
  const [isDismissed, setIsDismissed] = useState(false);

  const handleUpdate = async () => {
    setIsUpdating(true);
    try {
      await updateApp();
      setIsDismissed(true);
      onDismiss?.();
    } catch (error) {
      console.error('Failed to update PWA:', error);
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDismiss = () => {
    setIsDismissed(true);
    onDismiss?.();
  };

  if (!updateAvailable || isDismissed) {
    return null;
  }

  return (
    <Card className={`relative border-warning/20 bg-warning/5 ${className}`}>
      <button
        onClick={handleDismiss}
        className="absolute top-2 right-2 p-1 rounded-md hover:bg-background/10 transition-colors"
        aria-label="Dismiss update prompt"
      >
        <XMarkIcon className="h-4 w-4 text-muted-foreground" />
      </button>
      
      <div className="p-4">
        <div className="flex items-start space-x-3">
          <div className="flex-shrink-0">
            <div className="h-10 w-10 rounded-lg bg-warning flex items-center justify-center">
              <ArrowPathIcon className="h-5 w-5 text-warning-foreground" />
            </div>
          </div>
          
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-medium text-foreground">
              Update Available
            </h3>
            <p className="text-sm text-muted-foreground mt-1">
              A new version of Hub is available with improvements and bug fixes.
            </p>
            
            <div className="mt-3 flex items-center space-x-2">
              <Button
                size="sm"
                onClick={handleUpdate}
                disabled={isUpdating}
                className="h-8"
                variant="default"
              >
                {isUpdating ? 'Updating...' : 'Update Now'}
              </Button>
              
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDismiss}
                className="h-8 text-muted-foreground"
              >
                Later
              </Button>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}