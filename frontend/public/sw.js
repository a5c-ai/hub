const CACHE_NAME = 'hub-v2';
const STATIC_CACHE_NAME = 'hub-static-v2';
const DYNAMIC_CACHE_NAME = 'hub-dynamic-v2';

// Cache configuration
const CACHE_DURATION = {
  static: 60 * 60 * 24 * 30, // 30 days for static assets
  api: 60 * 5, // 5 minutes for API responses
  pages: 60 * 60 * 24, // 24 hours for pages
};

// URLs to cache on install
const STATIC_ASSETS = [
  '/',
  '/dashboard',
  '/login',
  '/register',
  '/offline',
  '/manifest.json',
  '/_next/static/css/',
  '/_next/static/js/',
];

// API endpoints to cache
const API_CACHE_PATTERNS = [
  /^\/api\/v1\/repositories/,
  /^\/api\/v1\/user/,
  /^\/api\/v1\/organizations/,
];

// Network-first endpoints (always try network first)
const NETWORK_FIRST_PATTERNS = [
  /^\/api\/v1\/auth/,
  /^\/api\/v1\/.*\/commits/,
  /^\/api\/v1\/.*\/pulls/,
  /^\/api\/v1\/.*\/issues/,
];

// Cache-first patterns (static assets)
const CACHE_FIRST_PATTERNS = [
  /\/_next\/static\//,
  /\.(?:js|css|woff2?|png|jpg|jpeg|gif|svg|ico)$/,
];

self.addEventListener('install', (event) => {
  console.log('[Service Worker] Installing...');
  
  event.waitUntil(
    (async () => {
      try {
        const cache = await caches.open(STATIC_CACHE_NAME);
        console.log('[Service Worker] Caching static assets');
        await cache.addAll(STATIC_ASSETS);
        console.log('[Service Worker] Static assets cached successfully');
        
        // Skip waiting to activate immediately
        self.skipWaiting();
      } catch (error) {
        console.error('[Service Worker] Failed to cache static assets:', error);
      }
    })()
  );
});

self.addEventListener('activate', (event) => {
  console.log('[Service Worker] Activating...');
  
  event.waitUntil(
    (async () => {
      try {
        // Clean up old caches
        const cacheNames = await caches.keys();
        const oldCaches = cacheNames.filter(name => 
          name.startsWith('hub-') && 
          name !== STATIC_CACHE_NAME && 
          name !== DYNAMIC_CACHE_NAME
        );
        
        await Promise.all(
          oldCaches.map(cacheName => {
            console.log('[Service Worker] Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          })
        );
        
        // Claim clients
        await self.clients.claim();
        console.log('[Service Worker] Activated successfully');
      } catch (error) {
        console.error('[Service Worker] Activation failed:', error);
      }
    })()
  );
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);
  
  // Skip non-GET requests
  if (request.method !== 'GET') {
    return;
  }
  
  // Skip cross-origin requests
  if (url.origin !== location.origin) {
    return;
  }
  
  event.respondWith(handleFetch(request));
});

async function handleFetch(request) {
  const url = new URL(request.url);
  const pathname = url.pathname;
  
  try {
    // Network-first strategy for critical API endpoints
    if (NETWORK_FIRST_PATTERNS.some(pattern => pattern.test(pathname))) {
      return await networkFirstStrategy(request);
    }
    
    // Cache-first strategy for static assets
    if (CACHE_FIRST_PATTERNS.some(pattern => pattern.test(pathname))) {
      return await cacheFirstStrategy(request);
    }
    
    // Stale-while-revalidate for API endpoints
    if (API_CACHE_PATTERNS.some(pattern => pattern.test(pathname))) {
      return await staleWhileRevalidateStrategy(request);
    }
    
    // Default: network-first with cache fallback
    return await networkFirstStrategy(request);
    
  } catch (error) {
    console.error('[Service Worker] Fetch failed:', error);
    
    // Return offline page for navigation requests
    if (request.mode === 'navigate') {
      const cache = await caches.open(STATIC_CACHE_NAME);
      const offlinePage = await cache.match('/offline');
      return offlinePage || new Response('Offline', { status: 503 });
    }
    
    return new Response('Network error', { status: 503 });
  }
}

async function networkFirstStrategy(request) {
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(DYNAMIC_CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('[Service Worker] Network failed, trying cache:', request.url);
    const cache = await caches.open(DYNAMIC_CACHE_NAME);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }
    
    throw error;
  }
}

async function cacheFirstStrategy(request) {
  const cache = await caches.open(STATIC_CACHE_NAME);
  const cachedResponse = await cache.match(request);
  
  if (cachedResponse) {
    return cachedResponse;
  }
  
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.error('[Service Worker] Cache-first strategy failed:', error);
    throw error;
  }
}

async function staleWhileRevalidateStrategy(request) {
  const cache = await caches.open(DYNAMIC_CACHE_NAME);
  const cachedResponse = await cache.match(request);
  
  // Always try to update the cache in the background
  const networkResponsePromise = fetch(request).then(response => {
    if (response.ok) {
      cache.put(request, response.clone());
    }
    return response;
  }).catch(error => {
    console.log('[Service Worker] Background fetch failed:', error);
  });
  
  // Return cached response immediately if available
  if (cachedResponse) {
    return cachedResponse;
  }
  
  // If no cache, wait for network
  try {
    return await networkResponsePromise;
  } catch (error) {
    console.error('[Service Worker] Stale-while-revalidate failed:', error);
    throw error;
  }
}

// Background sync for offline actions
self.addEventListener('sync', (event) => {
  console.log('[Service Worker] Background sync triggered:', event.tag);
  
  if (event.tag === 'background-sync') {
    event.waitUntil(handleBackgroundSync());
  }
});

async function handleBackgroundSync() {
  try {
    // Get stored offline actions from IndexedDB
    const offlineActions = await getOfflineActions();
    
    for (const action of offlineActions) {
      try {
        await processOfflineAction(action);
        await removeOfflineAction(action.id);
      } catch (error) {
        console.error('[Service Worker] Failed to process offline action:', error);
      }
    }
  } catch (error) {
    console.error('[Service Worker] Background sync failed:', error);
  }
}

// Push notification handling
self.addEventListener('push', (event) => {
  console.log('[Service Worker] Push notification received:', event);
  
  const options = {
    body: 'You have new activity in Hub',
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge-72x72.png',
    vibrate: [100, 50, 100],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1,
    },
    actions: [
      {
        action: 'explore',
        title: 'View',
        icon: '/icons/checkmark.png',
      },
      {
        action: 'close',
        title: 'Close',
        icon: '/icons/xmark.png',
      },
    ],
  };
  
  if (event.data) {
    try {
      const payload = event.data.json();
      options.body = payload.body || options.body;
      options.title = payload.title || 'Hub Notification';
      options.data = { ...options.data, ...payload.data };
    } catch (error) {
      console.error('[Service Worker] Failed to parse push payload:', error);
    }
  }
  
  event.waitUntil(
    self.registration.showNotification('Hub', options)
  );
});

// Notification click handling
self.addEventListener('notificationclick', (event) => {
  console.log('[Service Worker] Notification clicked:', event);
  
  event.notification.close();
  
  if (event.action === 'close') {
    return;
  }
  
  // Default action or 'explore' action
  event.waitUntil(
    clients.matchAll().then((clientList) => {
      if (clientList.length > 0) {
        return clientList[0].focus();
      }
      return clients.openWindow('/dashboard');
    })
  );
});

// Helper functions for IndexedDB operations (offline actions)
async function getOfflineActions() {
  // TODO: Implement IndexedDB operations for offline actions
  return [];
}

async function processOfflineAction(action) {
  // TODO: Implement processing of offline actions
  console.log('[Service Worker] Processing offline action:', action);
}

async function removeOfflineAction(id) {
  // TODO: Implement removal of processed offline actions
  console.log('[Service Worker] Removing offline action:', id);
}

// Message handling from main thread
self.addEventListener('message', (event) => {
  console.log('[Service Worker] Message received:', event.data);
  
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});

console.log('[Service Worker] Service worker script loaded');