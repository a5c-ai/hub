# Mobile and Progressive Web App Documentation

## Overview

The A5C Hub includes comprehensive mobile support and Progressive Web App (PWA) capabilities, providing a native app-like experience with offline functionality, touch optimization, and app installation support.

## Features

### Progressive Web App (PWA)
- **App Installation**: Native app installation prompts with proper lifecycle management
- **Offline Support**: Graceful offline functionality when network is unavailable
- **Service Worker**: Advanced caching strategies with multiple cache policies
- **App Manifest**: Complete manifest with shortcuts, icons, and app metadata
- **Background Sync**: Foundation for offline action synchronization

### Mobile Optimization
- **Responsive Design**: Optimized layouts for mobile devices and tablets
- **Touch Interface**: Touch-friendly navigation and controls
- **Mobile Components**: Specialized mobile UI components
- **Safe Area Support**: Proper handling of device notches and safe areas
- **Performance**: Optimized for mobile device constraints

## PWA Installation

### Automatic Installation Prompts

The PWA installation prompt appears automatically when criteria are met:
- User has visited the site multiple times
- User has engaged with the content
- Site meets PWA installability requirements
- User is on a supported browser/platform

### Manual Installation

**Chrome/Edge (Desktop)**
1. Click the install icon in the address bar
2. Or use the "Install App" option in the menu

**Chrome (Android)**
1. Tap the menu (three dots)
2. Select "Add to Home screen" or "Install app"

**Safari (iOS)**
1. Tap the share button
2. Select "Add to Home Screen"
3. Confirm installation

## Offline Capabilities

### Caching Strategy

The PWA implements multiple caching strategies for optimal performance:

```javascript
// Service Worker caching strategies
const CACHE_STRATEGIES = {
  pages: 'NetworkFirst',        // HTML pages
  api: 'NetworkFirst',         // API responses
  assets: 'CacheFirst',        // CSS, JS, images
  fonts: 'CacheFirst',         // Web fonts
  static: 'CacheFirst'         // Static assets
};
```

### Offline Functionality

**Available Offline:**
- Previously viewed repositories and files
- Cached issue and pull request data
- User profile and organization information
- Static pages and documentation
- Basic search within cached content

**Requires Online Connection:**
- Git operations (push, pull, clone)
- Creating new issues or pull requests
- Real-time notifications
- Live collaboration features
- Administrative functions

### Network Status Detection

```javascript
// Automatic network status detection
window.addEventListener('online', () => {
  // Resume online functionality
  showNetworkStatusBanner('Connected');
  syncOfflineActions();
});

window.addEventListener('offline', () => {
  // Switch to offline mode
  showNetworkStatusBanner('Offline - Limited functionality');
});
```

## Mobile Interface Components

### Mobile Navigation

**Bottom Navigation Bar**
- Quick access to main sections
- Touch-optimized button sizes (44px minimum)
- Visual indicators for current section
- Swipe gestures for navigation

**Mobile Repository Browser**
- Touch-friendly file tree navigation
- Swipe actions for common operations
- Search and filtering capabilities
- Breadcrumb navigation

### Mobile Code Viewer

**Features:**
- Syntax highlighting optimized for mobile
- Zoom and pan support
- Search within files
- Line number toggle
- Mobile-friendly selection

**Controls:**
- Pinch to zoom
- Double-tap to fit width
- Swipe between files
- Touch-friendly scrolling

### Mobile Forms

**Issue Creation Form**
- Touch-optimized input fields
- Auto-complete and suggestions
- File attachment support
- Markdown preview toggle
- Voice input support (where available)

## Configuration

### PWA Manifest

```json
{
  "name": "A5C Hub",
  "short_name": "Hub",
  "description": "Self-hosted Git platform",
  "start_url": "/",
  "display": "standalone",
  "theme_color": "#0066cc",
  "background_color": "#ffffff",
  "orientation": "portrait",
  "categories": ["productivity", "developer"],
  "shortcuts": [
    {
      "name": "Repositories",
      "short_name": "Repos",
      "description": "View repositories",
      "url": "/repositories",
      "icons": [{"src": "/icons/repos-96x96.png", "sizes": "96x96"}]
    },
    {
      "name": "Issues",
      "short_name": "Issues",
      "description": "View issues",
      "url": "/issues",
      "icons": [{"src": "/icons/issues-96x96.png", "sizes": "96x96"}]
    }
  ],
  "icons": [
    {
      "src": "/icons/icon-72x72.png",
      "sizes": "72x72",
      "type": "image/png",
      "purpose": "maskable any"
    },
    {
      "src": "/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "maskable any"
    },
    {
      "src": "/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable any"
    }
  ]
}
```

### Service Worker Configuration

```javascript
// Service worker registration
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('/sw.js')
    .then(registration => {
      console.log('SW registered:', registration);
      
      // Check for updates
      registration.addEventListener('updatefound', () => {
        const newWorker = registration.installing;
        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            // New version available
            showUpdateAvailableNotification();
          }
        });
      });
    })
    .catch(error => {
      console.error('SW registration failed:', error);
    });
}
```

## Mobile Performance Optimization

### Image Optimization

```javascript
// Responsive image loading
const imageObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      const img = entry.target;
      img.src = img.dataset.src;
      img.classList.remove('lazy');
      imageObserver.unobserve(img);
    }
  });
});

// Lazy load images
document.querySelectorAll('img[data-src]').forEach(img => {
  imageObserver.observe(img);
});
```

### Touch Optimization

```css
/* Touch-friendly button sizing */
.touch-target {
  min-height: 44px;
  min-width: 44px;
  padding: 12px;
  touch-action: manipulation;
}

/* Improve touch scrolling */
.scrollable {
  -webkit-overflow-scrolling: touch;
  overflow-scrolling: touch;
}

/* Prevent zoom on input focus (iOS) */
input, select, textarea {
  font-size: 16px;
}
```

## Development

### PWA Testing

**Chrome DevTools**
1. Open DevTools → Application tab
2. Check "Manifest" section for manifest validation
3. Use "Service Workers" section to test SW functionality
4. Simulate offline mode in Network tab

**Lighthouse Audit**
1. Open DevTools → Lighthouse tab
2. Select "Progressive Web App" category
3. Run audit to check PWA compliance
4. Review recommendations for improvements

### Mobile Testing

**Device Testing**
- Test on real devices when possible
- Use browser device emulation
- Test various screen sizes and orientations
- Verify touch interactions work correctly

**Performance Testing**
- Measure load times on slow networks
- Test offline functionality
- Monitor memory usage
- Verify battery impact

## Browser Support

### PWA Features
- **Chrome (Android)**: Full PWA support including installation
- **Edge (Windows)**: Full PWA support with app store integration
- **Safari (iOS)**: Limited PWA support, Add to Home Screen available
- **Firefox**: Service worker support, limited installation features

### Mobile Optimization
- **iOS Safari**: 12.0+
- **Chrome Android**: 70+
- **Samsung Internet**: 8.0+
- **Edge Mobile**: All versions

## Analytics and Monitoring

### PWA Metrics

```javascript
// Track PWA installation
window.addEventListener('beforeinstallprompt', (e) => {
  analytics.track('pwa_install_prompt_shown');
  
  e.userChoice.then(choiceResult => {
    analytics.track('pwa_install_choice', {
      choice: choiceResult.outcome
    });
  });
});

// Track offline usage
window.addEventListener('offline', () => {
  analytics.track('app_went_offline');
});

window.addEventListener('online', () => {
  analytics.track('app_came_online');
});
```

### Performance Monitoring

```javascript
// Core Web Vitals tracking
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

getCLS(metric => analytics.track('core_web_vital', metric));
getFID(metric => analytics.track('core_web_vital', metric));
getFCP(metric => analytics.track('core_web_vital', metric));
getLCP(metric => analytics.track('core_web_vital', metric));
getTTFB(metric => analytics.track('core_web_vital', metric));
```

## Troubleshooting

### Common Issues

**PWA not installing**
- Check manifest.json is valid and accessible
- Verify HTTPS is enabled
- Ensure service worker is registered successfully
- Check browser console for errors

**Offline mode not working**
- Verify service worker is active
- Check cache storage in DevTools
- Review network requests during offline testing
- Validate cache strategies

**Touch interactions not working**
- Check touch event handlers
- Verify touch-action CSS properties
- Test on actual devices
- Review viewport meta tag

**Performance issues on mobile**
- Optimize images and assets
- Reduce JavaScript bundle size
- Implement lazy loading
- Use service worker caching effectively

### Debug Tools

```javascript
// Service worker debug logging
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.addEventListener('message', event => {
    console.log('SW message:', event.data);
  });
}

// Cache inspection
caches.keys().then(cacheNames => {
  console.log('Available caches:', cacheNames);
  
  cacheNames.forEach(cacheName => {
    caches.open(cacheName).then(cache => {
      cache.keys().then(requests => {
        console.log(`Cache ${cacheName}:`, requests.map(r => r.url));
      });
    });
  });
});
```

## Best Practices

### PWA Development
- Implement progressive enhancement
- Design for offline-first experience
- Use app shell architecture
- Provide clear offline indicators
- Handle background sync appropriately

### Mobile UX
- Follow platform-specific design guidelines
- Optimize for one-handed use
- Provide haptic feedback where appropriate
- Ensure accessibility compliance
- Test with real users on actual devices

### Performance
- Minimize initial bundle size
- Implement code splitting
- Use efficient caching strategies
- Optimize for Core Web Vitals
- Monitor real user metrics

## Future Enhancements

### Planned Features
- Push notification support
- Background sync for offline actions
- Native file system access APIs
- Biometric authentication support
- Camera integration for file uploads

### Advanced PWA Features
- App shortcuts customization
- Share target API integration
- Payment request API support
- Contact picker API integration
- Web locks for data synchronization

## Support

For mobile and PWA issues:
- Test on multiple devices and browsers
- Check service worker registration and caching
- Review console logs for errors
- Validate PWA compliance with Lighthouse
- Contact system administrators for assistance

## References

- [PWA Documentation](https://web.dev/progressive-web-apps/)
- [Service Worker API](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API)
- [Web App Manifest](https://developer.mozilla.org/en-US/docs/Web/Manifest)
- [Mobile Web Best Practices](https://developers.google.com/web/fundamentals/design-and-ux/principles)