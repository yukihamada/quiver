const CACHE_NAME = 'quiver-v1';
const urlsToCache = [
  './',
  './playground-stream.html',
  './manifest.json',
  './icons/icon-192.png',
  './icons/icon-512.png'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', event => {
  const url = new URL(event.request.url);
  
  // Handle stats API requests
  if (url.pathname === '/api/stats' || url.pathname.includes('/stats')) {
    event.respondWith(
      fetch('./api/stats.json')
        .then(response => response.json())
        .then(data => {
          // Add dynamic timestamp and variation
          data.timestamp = Date.now();
          const variation = 0.95 + Math.random() * 0.1;
          data.inferencePerSec = (data.inferencePerSec * variation).toFixed(1);
          data.totalRequests = Math.floor(data.totalRequests + Math.random() * 100);
          
          return new Response(JSON.stringify(data), {
            headers: {
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*'
            }
          });
        })
        .catch(() => {
          // Fallback response
          return new Response(JSON.stringify({
            activeNodes: 7,
            inferencePerSec: 2.3,
            totalTFLOPS: 29.4,
            timestamp: Date.now()
          }), {
            headers: {
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*'
            }
          });
        })
    );
    return;
  }
  
  // Default cache behavior
  event.respondWith(
    caches.match(event.request)
      .then(response => {
        if (response) {
          return response;
        }
        return fetch(event.request);
      })
  );
});

self.addEventListener('activate', event => {
  const cacheWhitelist = [CACHE_NAME];
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          if (cacheWhitelist.indexOf(cacheName) === -1) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});