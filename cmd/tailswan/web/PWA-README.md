# TailSwan PWA Implementation

This document explains the Progressive Web App (PWA) features added to TailSwan's control panel.

## What Was Added

### 1. Web App Manifest (`manifest.json`)
- Defines app metadata for installation
- Specifies app icons in 8 different sizes (72-512px)
- Theme colors matching TailSwan's dark UI
- Standalone display mode (fullscreen, no browser chrome)

### 2. Service Worker (`sw.js`)
- **Cache-first strategy** for static assets (CSS, JS, images)
- **Network-first strategy** for API calls with cache fallback
- Pre-caches app shell on installation
- Automatic cache versioning and cleanup
- Offline support with cached data

### 3. PWA Meta Tags
Added to `index.html`:
- iOS support (Apple-specific meta tags)
- Android support (mobile-web-app-capable)
- Windows support (msapplication tiles)
- Theme color for status bar

### 4. App Icons
Generated 8 PNG icons from the existing logo:
- 72x72 - Android legacy
- 96x96 - Android legacy
- 128x128 - Chrome Web Store
- 144x144 - Windows tiles
- 152x152 - iOS
- 192x192 - Android home screen
- 384x384 - Android splash
- 512x512 - PWA standard

## Testing the PWA

### Desktop (Chrome/Edge)

1. **Start the server**:
   ```bash
   # Build and run the Docker container, or
   # Run the control server directly
   go run ./cmd/controlserver
   ```

2. **Open in Chrome**:
   - Navigate to `http://localhost:8080/`
   - Or via Tailscale: `https://tailswan/` (or your configured hostname)

3. **Verify manifest**:
   - Open DevTools (F12)
   - Go to Application tab → Manifest
   - Check that all fields load correctly
   - Verify icons are displayed

4. **Test service worker**:
   - DevTools → Application → Service Workers
   - Should show service worker registered for scope "/"
   - Check "Offline" checkbox to test offline mode
   - Reload page - should still work with cached content

5. **Install the app**:
   - Look for install icon in address bar (⊕)
   - Click to install
   - Or: Chrome menu → "Install TailSwan Control Panel"
   - App should open in standalone window

6. **Test installed app**:
   - Launch from Chrome apps or desktop shortcut
   - Should open without browser UI
   - All features should work normally
   - Test offline mode by disabling network

### Mobile (iOS - Safari)

1. **Open in Safari**:
   - Navigate to TailSwan control panel URL
   - Via Tailscale: `https://tailswan/`

2. **Add to Home Screen**:
   - Tap Share button (box with arrow)
   - Scroll down and tap "Add to Home Screen"
   - Customize name if desired
   - Tap "Add"

3. **Test installed app**:
   - Tap the TailSwan icon on home screen
   - Should launch fullscreen without Safari UI
   - Test all features

**Note**: iOS Safari has limited PWA support:
- No install prompt (must use Share menu)
- Limited background capabilities
- Service worker support is basic

### Mobile (Android - Chrome)

1. **Open in Chrome**:
   - Navigate to TailSwan URL

2. **Install prompt**:
   - Should see automatic "Add to Home Screen" prompt
   - Or: Chrome menu → "Add to Home Screen"
   - Or: Chrome menu → "Install app"

3. **Test installed app**:
   - Launch from home screen or app drawer
   - Should show splash screen with icon
   - Full offline support
   - Test features and offline mode

## Lighthouse PWA Audit

Run a Lighthouse audit to verify PWA quality:

1. Open Chrome DevTools (F12)
2. Go to "Lighthouse" tab
3. Select "Progressive Web App" category
4. Click "Analyze page load"

**Target scores**:
- PWA: 100/100
- Installability: ✓ All checks passing
- PWA Optimized: ✓ All checks passing

Common checks:
- ✓ Registers a service worker
- ✓ Responds with 200 when offline
- ✓ Provides valid web app manifest
- ✓ Has a `<meta name="viewport">` tag
- ✓ Content sized correctly for viewport
- ✓ Has Apple touch icon
- ✓ Configured for custom splash screen

## Architecture Details

### Caching Strategy

**App Shell (pre-cached on install)**:
```
/               (index.html)
/static/style.css
/static/app.js
/static/logo.jpeg
/static/favicon.svg
```

**Static Assets**:
- Strategy: Cache first, fallback to network
- Updates cached on new versions

**API Calls**:
- Strategy: Network first, fallback to cache
- Cache TTL: 5 minutes
- Excludes `/api/events` (SSE stream - never cached)

### Service Worker Lifecycle

1. **Install**: Pre-caches app shell resources
2. **Activate**: Cleans up old caches
3. **Fetch**: Intercepts requests and applies caching strategy
4. **Update**: New version installs in background

### Cache Versioning

- Cache name: `tailswan-v1`
- Increment version in `sw.js` when making updates
- Old caches automatically deleted on activation

## Updating the PWA

When you update the web app:

1. **Update version in service worker**:
   ```javascript
   // In sw.js
   const CACHE_VERSION = 'v2';  // Increment version
   ```

2. **Rebuild/restart server**:
   - Embedded files need rebuild for changes to take effect
   - Service worker will auto-update on next visit

3. **Users see update**:
   - New service worker installs in background
   - On next reload, new version activates
   - Console shows: "New service worker available. Refresh to update."

## Troubleshooting

### Service Worker Won't Register

**Check**:
- Server must use HTTPS (Tailscale provides this) or localhost
- `/sw.js` must return `Content-Type: application/javascript`
- Check browser console for errors

**Fix**:
- Verify route in `internal/server/server.go`
- Check service worker syntax for errors

### Icons Not Showing

**Check**:
- Icons exist in `/static/icons/` path
- Manifest references correct paths
- Icon files are valid PNG format

**Fix**:
- Re-run icon generation script
- Verify manifest.json paths
- Clear browser cache

### Offline Mode Not Working

**Check**:
- Service worker is active (DevTools → Application)
- Resources are cached (DevTools → Application → Cache Storage)
- Network strategy is correct for resource type

**Fix**:
- Unregister and re-register service worker
- Clear cache storage
- Check service worker caching logic

### App Not Installable

**Check Lighthouse audit** for specific issues:
- Manifest must be valid JSON
- Must have icons (192x192 minimum)
- Must have service worker
- Must be served over HTTPS
- Must have proper `<meta name="viewport">`

## Regenerating Icons

If you update the logo and need new icons:

```bash
cd cmd/controlserver/web/icons
./generate-icons.sh
```

The script will regenerate all 8 icon sizes from `../logo.jpeg`.

## Browser Support

| Browser | Install | Offline | Notes |
|---------|---------|---------|-------|
| Chrome (Desktop) | ✓ | ✓ | Full support |
| Edge (Desktop) | ✓ | ✓ | Full support |
| Safari (Desktop) | ⚠️ | ✓ | Limited (no install prompt) |
| Firefox (Desktop) | ⚠️ | ✓ | Service worker only |
| Chrome (Android) | ✓ | ✓ | Full support + splash screen |
| Safari (iOS) | ✓ | ⚠️ | Manual install, limited features |

## Files Modified/Created

**New Files**:
- `cmd/controlserver/web/manifest.json` - PWA manifest
- `cmd/controlserver/web/sw.js` - Service worker
- `cmd/controlserver/web/icons/icon-*.png` - 8 app icons
- `cmd/controlserver/web/icons/generate-icons.sh` - Icon generation script

**Modified Files**:
- `cmd/controlserver/web/index.html` - Added PWA meta tags and service worker registration
- `internal/server/server.go` - Added `/sw.js` route

## Resources

- [MDN: Progressive Web Apps](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
- [web.dev: PWA Checklist](https://web.dev/pwa-checklist/)
- [Service Worker API](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API)
- [Web App Manifest](https://developer.mozilla.org/en-US/docs/Web/Manifest)
