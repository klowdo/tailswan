const CACHE_VERSION = 'v1';
const CACHE_NAME = `tailswan-${CACHE_VERSION}`;

const APP_SHELL = [
  '/',
  '/static/style.css',
  '/static/app.js',
  '/static/logo.jpeg',
  '/static/favicon.svg'
];

const API_CACHE_DURATION = 5 * 60 * 1000;

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(APP_SHELL);
      })
      .then(() => {
        return self.skipWaiting();
      })
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames
            .filter((name) => name.startsWith('tailswan-') && name !== CACHE_NAME)
            .map((name) => {
              return caches.delete(name);
            })
        );
      })
      .then(() => {
        return self.clients.claim();
      })
  );
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  if (url.pathname === '/api/events') {
    return;
  }

  if (url.pathname.startsWith('/api/')) {
    event.respondWith(networkFirstStrategy(request));
  } else if (url.pathname.startsWith('/static/') || url.pathname === '/') {
    event.respondWith(cacheFirstStrategy(request));
  } else {
    event.respondWith(fetch(request));
  }
});

async function cacheFirstStrategy(request) {
  try {
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }

    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }

    return networkResponse;
  } catch (error) {
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }

    throw error;
  }
}

async function networkFirstStrategy(request) {
  try {
    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      const responseToCache = networkResponse.clone();

      const cacheEntry = {
        response: responseToCache,
        timestamp: Date.now()
      };

      cache.put(request, new Response(
        JSON.stringify({
          data: await responseToCache.clone().text(),
          timestamp: cacheEntry.timestamp
        })
      ));
    }

    return networkResponse;
  } catch (error) {
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      const cached = await cachedResponse.json();

      if (Date.now() - cached.timestamp < API_CACHE_DURATION) {
        return new Response(cached.data, {
          headers: { 'Content-Type': 'application/json' }
        });
      }
    }

    throw error;
  }
}

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }

  if (event.data && event.data.type === 'CLEAR_CACHE') {
    event.waitUntil(
      caches.delete(CACHE_NAME).then(() => {
        return caches.open(CACHE_NAME);
      }).then((cache) => {
        return cache.addAll(APP_SHELL);
      })
    );
  }
});
