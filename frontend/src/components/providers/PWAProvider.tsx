'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { PWAManager } from '@/lib/pwa';

interface PWAContextType {
  canInstall: boolean;
  isInstalled: boolean;
  updateAvailable: boolean;
  isOnline: boolean;
  installApp: () => Promise<boolean>;
  updateApp: () => Promise<void>;
  checkForUpdates: () => Promise<void>;
  subscribeToPushNotifications: () => Promise<PushSubscription | null>;
  unsubscribeFromPushNotifications: () => Promise<boolean>;
}

const PWAContext = createContext<PWAContextType | undefined>(undefined);

interface PWAProviderProps {
  children: ReactNode;
}

export function PWAProvider({ children }: PWAProviderProps) {
  const [canInstall, setCanInstall] = useState(false);
  const [isInstalled, setIsInstalled] = useState(false);
  const [updateAvailable, setUpdateAvailable] = useState(false);
  const [isOnline, setIsOnline] = useState(true);

  useEffect(() => {
    const pwa = PWAManager.getInstance();

    // Initial status check
    const updateStatus = () => {
      const status = pwa.getInstallationStatus();
      setCanInstall(status.canInstall);
      setIsInstalled(status.isInstalled);
      setUpdateAvailable(status.updateAvailable);
      setIsOnline(pwa.isOnline());
    };

    updateStatus();

    // Listen for PWA update events
    const handleUpdateAvailable = () => {
      setUpdateAvailable(true);
    };

    window.addEventListener('pwa-update-available', handleUpdateAvailable);

    // Listen for online/offline status
    const cleanupOnlineStatus = pwa.onlineStatusChanged(setIsOnline);

    // Listen for install prompt changes
    const handleBeforeInstallPrompt = () => {
      setTimeout(updateStatus, 100); // Small delay to let PWA manager update
    };

    const handleAppInstalled = () => {
      setIsInstalled(true);
      setCanInstall(false);
    };

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    window.addEventListener('appinstalled', handleAppInstalled);

    // Periodic status updates
    const interval = setInterval(updateStatus, 30000); // Check every 30 seconds

    return () => {
      window.removeEventListener('pwa-update-available', handleUpdateAvailable);
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
      window.removeEventListener('appinstalled', handleAppInstalled);
      cleanupOnlineStatus();
      clearInterval(interval);
    };
  }, []);

  const installApp = async (): Promise<boolean> => {
    const pwa = PWAManager.getInstance();
    const result = await pwa.installApp();
    
    if (result) {
      setIsInstalled(true);
      setCanInstall(false);
    }
    
    return result;
  };

  const updateApp = async (): Promise<void> => {
    const pwa = PWAManager.getInstance();
    await pwa.updateApp();
    setUpdateAvailable(false);
  };

  const checkForUpdates = async (): Promise<void> => {
    const pwa = PWAManager.getInstance();
    await pwa.checkForUpdates();
  };

  const subscribeToPushNotifications = async (): Promise<PushSubscription | null> => {
    const pwa = PWAManager.getInstance();
    return await pwa.subscribeToPushNotifications();
  };

  const unsubscribeFromPushNotifications = async (): Promise<boolean> => {
    const pwa = PWAManager.getInstance();
    return await pwa.unsubscribeFromPushNotifications();
  };

  const contextValue: PWAContextType = {
    canInstall,
    isInstalled,
    updateAvailable,
    isOnline,
    installApp,
    updateApp,
    checkForUpdates,
    subscribeToPushNotifications,
    unsubscribeFromPushNotifications,
  };

  return (
    <PWAContext.Provider value={contextValue}>
      {children}
    </PWAContext.Provider>
  );
}

export function usePWAContext(): PWAContextType {
  const context = useContext(PWAContext);
  if (context === undefined) {
    throw new Error('usePWAContext must be used within a PWAProvider');
  }
  return context;
}

// Hook for checking if running as PWA
export function useIsPWA(): boolean {
  const [isPWA, setIsPWA] = useState(false);

  useEffect(() => {
    const checkPWA = () => {
      // Check if running as standalone app
      const isStandalone = window.matchMedia('(display-mode: standalone)').matches;
      const isInWebAppiOS = (window.navigator as { standalone?: boolean }).standalone === true;
      const isInWebAppChrome = window.matchMedia('(display-mode: standalone)').matches;
      
      setIsPWA(isStandalone || isInWebAppiOS || isInWebAppChrome);
    };

    checkPWA();

    // Listen for display mode changes
    const mediaQuery = window.matchMedia('(display-mode: standalone)');
    mediaQuery.addListener(checkPWA);

    return () => {
      mediaQuery.removeListener(checkPWA);
    };
  }, []);

  return isPWA;
}