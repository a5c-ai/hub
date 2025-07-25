// PWA Service Worker registration and management

interface InstallPromptEvent extends Event {
  readonly platforms: string[];
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed';
    platform: string;
  }>;
  prompt(): Promise<void>;
}

export class PWAManager {
  private static instance: PWAManager | null = null;
  private swRegistration: ServiceWorkerRegistration | null = null;
  private installPrompt: InstallPromptEvent | null = null;
  private updateAvailable = false;
  private isInstalled = false;

  private constructor() {
    this.init();
  }

  static getInstance(): PWAManager {
    if (!PWAManager.instance) {
      PWAManager.instance = new PWAManager();
    }
    return PWAManager.instance;
  }

  private async init() {
    if (typeof window === 'undefined' || !('serviceWorker' in navigator)) {
      console.log('[PWA] Service Worker not supported');
      return;
    }

    try {
      // Register service worker
      this.swRegistration = await navigator.serviceWorker.register('/sw.js', {
        scope: '/',
      });

      console.log('[PWA] Service Worker registered successfully');

      // Check for updates
      this.swRegistration.addEventListener('updatefound', () => {
        console.log('[PWA] Update found');
        const newWorker = this.swRegistration?.installing;
        
        if (newWorker) {
          newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              console.log('[PWA] Update available');
              this.updateAvailable = true;
              this.notifyUpdateAvailable();
            }
          });
        }
      });

      // Listen for install prompt
      window.addEventListener('beforeinstallprompt', (e) => {
        console.log('[PWA] Install prompt available');
        e.preventDefault();
        this.installPrompt = e as InstallPromptEvent;
      });

      // Check if app is already installed
      window.addEventListener('appinstalled', () => {
        console.log('[PWA] App installed');
        this.isInstalled = true;
        this.installPrompt = null;
      });

      // Check install status on load
      this.checkInstallStatus();

    } catch (error) {
      console.error('[PWA] Service Worker registration failed:', error);
    }
  }

  async installApp(): Promise<boolean> {
    if (!this.installPrompt) {
      console.log('[PWA] Install prompt not available');
      return false;
    }

    try {
      await this.installPrompt.prompt();
      const { outcome } = await this.installPrompt.userChoice;
      
      console.log('[PWA] Install prompt result:', outcome);
      
      if (outcome === 'accepted') {
        this.isInstalled = true;
        this.installPrompt = null;
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('[PWA] Install failed:', error);
      return false;
    }
  }

  async updateApp(): Promise<void> {
    if (!this.swRegistration) {
      return;
    }

    const waitingWorker = this.swRegistration.waiting;
    if (waitingWorker) {
      waitingWorker.postMessage({ type: 'SKIP_WAITING' });
      
      // Reload the page after the new service worker takes control
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        window.location.reload();
      });
    }
  }

  async checkForUpdates(): Promise<void> {
    if (!this.swRegistration) {
      return;
    }

    try {
      await this.swRegistration.update();
    } catch (error) {
      console.error('[PWA] Update check failed:', error);
    }
  }

  getInstallationStatus(): {
    canInstall: boolean;
    isInstalled: boolean;
    updateAvailable: boolean;
  } {
    return {
      canInstall: !!this.installPrompt && !this.isInstalled,
      isInstalled: this.isInstalled,
      updateAvailable: this.updateAvailable,
    };
  }

  private checkInstallStatus() {
    // Check if running as standalone app
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches;
    const isInWebAppiOS = (window.navigator as { standalone?: boolean }).standalone === true;
    const isInWebAppChrome = window.matchMedia('(display-mode: standalone)').matches;
    
    this.isInstalled = isStandalone || isInWebAppiOS || isInWebAppChrome;
    
    console.log('[PWA] Install status:', {
      isStandalone,
      isInWebAppiOS,
      isInWebAppChrome,
      isInstalled: this.isInstalled,
    });
  }

  private notifyUpdateAvailable() {
    // Create a custom event for update notification
    const updateEvent = new CustomEvent('pwa-update-available', {
      detail: { updateAvailable: true },
    });
    window.dispatchEvent(updateEvent);
  }

  // Push notifications management
  async subscribeToPushNotifications(): Promise<PushSubscription | null> {
    if (!this.swRegistration || !('PushManager' in window)) {
      console.log('[PWA] Push notifications not supported');
      return null;
    }

    try {
      const permission = await Notification.requestPermission();
      if (permission !== 'granted') {
        console.log('[PWA] Notification permission denied');
        return null;
      }

      const subscription = await this.swRegistration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: this.urlBase64ToUint8Array(
          process.env.NEXT_PUBLIC_VAPID_PUBLIC_KEY || ''
        ),
      });

      console.log('[PWA] Push subscription created:', subscription);
      return subscription;
    } catch (error) {
      console.error('[PWA] Push subscription failed:', error);
      return null;
    }
  }

  async unsubscribeFromPushNotifications(): Promise<boolean> {
    if (!this.swRegistration) {
      return false;
    }

    try {
      const subscription = await this.swRegistration.pushManager.getSubscription();
      if (subscription) {
        await subscription.unsubscribe();
        console.log('[PWA] Push subscription removed');
        return true;
      }
      return false;
    } catch (error) {
      console.error('[PWA] Push unsubscription failed:', error);
      return false;
    }
  }

  private urlBase64ToUint8Array(base64String: string): Uint8Array {
    const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);

    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i);
    }
    return outputArray;
  }

  // Offline support
  async cacheImportantData(): Promise<void> {
    if (!this.swRegistration) {
      return;
    }

    try {
      // Send message to service worker to cache important data
      const channel = new MessageChannel();
      this.swRegistration.active?.postMessage(
        { type: 'CACHE_IMPORTANT_DATA' },
        [channel.port2]
      );
    } catch (error) {
      console.error('[PWA] Failed to cache important data:', error);
    }
  }

  // Network status monitoring
  onlineStatusChanged(callback: (isOnline: boolean) => void): () => void {
    const handleOnline = () => callback(true);
    const handleOffline = () => callback(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Return cleanup function
    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }

  isOnline(): boolean {
    return navigator.onLine;
  }
}

// React hook for PWA functionality
export function usePWA() {
  const pwa = PWAManager.getInstance();
  
  return {
    installApp: () => pwa.installApp(),
    updateApp: () => pwa.updateApp(),
    checkForUpdates: () => pwa.checkForUpdates(),
    getInstallationStatus: () => pwa.getInstallationStatus(),
    subscribeToPushNotifications: () => pwa.subscribeToPushNotifications(),
    unsubscribeFromPushNotifications: () => pwa.unsubscribeFromPushNotifications(),
    cacheImportantData: () => pwa.cacheImportantData(),
    onlineStatusChanged: (callback: (isOnline: boolean) => void) => 
      pwa.onlineStatusChanged(callback),
    isOnline: () => pwa.isOnline(),
  };
}