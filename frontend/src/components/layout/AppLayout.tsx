'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { useAppStore } from '@/store/app';
import { useAuthStore } from '@/store/auth';
import { cn } from '@/lib/utils';
import { PWAInstallPrompt } from '@/components/ui/PWAInstallPrompt';
import { PWAUpdatePrompt } from '@/components/ui/PWAUpdatePrompt';
import { OfflineBanner } from '@/components/ui/OfflineIndicator';
import { MobileBottomNav } from './MobileBottomNav';

interface AppLayoutProps {
  children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const router = useRouter();
  const { sidebarOpen, setSidebarOpen } = useAppStore();
  const { checkAuth, isAuthenticated, isLoading, token } = useAuthStore();

  useEffect(() => {
    // Check auth status on initial load
    console.log('AppLayout: Checking auth on mount...');
    checkAuth();
  }, [checkAuth]);

  useEffect(() => {
    console.log('AppLayout: Auth state changed:', { isLoading, isAuthenticated, token });
    // Only redirect after auth check is complete and we're definitely not authenticated
    // Give a small delay to allow for rehydration to complete
    if (!isLoading && !isAuthenticated && !token) {
      console.log('AppLayout: Not authenticated, redirecting to login...');
      const timeoutId = setTimeout(() => {
        router.push('/login');
      }, 100); // Small delay to allow rehydration
      
      return () => clearTimeout(timeoutId);
    }
  }, [isAuthenticated, isLoading, token, router]);

  useEffect(() => {
    // Close sidebar on mobile when clicking outside
    const handleResize = () => {
      if (window.innerWidth >= 1024) {
        setSidebarOpen(true);
      } else {
        setSidebarOpen(false);
      }
    };

    handleResize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [setSidebarOpen]);

  // Show loading screen while checking authentication
  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <div className="text-muted-foreground">Loading...</div>
        </div>
      </div>
    );
  }

  // Don't render layout if not authenticated
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="h-screen flex overflow-hidden bg-background">
      {/* Sidebar */}
      <Sidebar />

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black bg-opacity-25 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Main content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <OfflineBanner />
        
        <main
          data-testid="main-content"
          className={cn(
            'flex-1 overflow-y-auto transition-all duration-200 pb-16 lg:pb-0',
            sidebarOpen ? 'lg:ml-0' : 'lg:ml-0'
          )}
        >
          {/* PWA Prompts */}
          <div className="sticky top-0 z-30 space-y-2 p-4">
            <PWAUpdatePrompt />
            <PWAInstallPrompt />
          </div>
          
          {children}
        </main>
      </div>
      
      {/* Mobile bottom navigation */}
      <MobileBottomNav />
    </div>
  );
}