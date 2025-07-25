'use client';

import { useState } from 'react';
import { Button } from './Button';
import { Card } from './Card';
import { usePWAContext } from '@/components/providers/PWAProvider';
import { XMarkIcon, ArrowDownTrayIcon } from '@heroicons/react/24/outline';

interface PWAInstallPromptProps {
  onDismiss?: () => void;
  className?: string;
}

export function PWAInstallPrompt({ onDismiss, className }: PWAInstallPromptProps) {
  const { canInstall, installApp } = usePWAContext();
  const [isInstalling, setIsInstalling] = useState(false);
  const [isDismissed, setIsDismissed] = useState(false);

  const handleInstall = async () => {
    setIsInstalling(true);
    try {
      const result = await installApp();
      if (result) {
        setIsDismissed(true);
        onDismiss?.();
      }
    } catch (error) {
      console.error('Failed to install PWA:', error);
    } finally {
      setIsInstalling(false);
    }
  };

  const handleDismiss = () => {
    setIsDismissed(true);
    onDismiss?.();
  };

  if (!canInstall || isDismissed) {
    return null;
  }

  return (
    <Card className={`relative border-primary/20 bg-primary/5 ${className}`}>
      <button
        onClick={handleDismiss}
        className="absolute top-2 right-2 p-1 rounded-md hover:bg-background/10 transition-colors"
        aria-label="Dismiss install prompt"
      >
        <XMarkIcon className="h-4 w-4 text-muted-foreground" />
      </button>
      
      <div className="p-4">
        <div className="flex items-start space-x-3">
          <div className="flex-shrink-0">
            <div className="h-10 w-10 rounded-lg bg-primary flex items-center justify-center">
              <ArrowDownTrayIcon className="h-5 w-5 text-primary-foreground" />
            </div>
          </div>
          
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-medium text-foreground">
              Install Hub App
            </h3>
            <p className="text-sm text-muted-foreground mt-1">
              Install Hub as an app for faster access and offline functionality.
            </p>
            
            <div className="mt-3 flex items-center space-x-2">
              <Button
                size="sm"
                onClick={handleInstall}
                disabled={isInstalling}
                className="h-8"
              >
                {isInstalling ? 'Installing...' : 'Install'}
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