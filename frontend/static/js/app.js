/**
 * app.js
 *
 * SPA entry point. Loaded as <script type="module"> from index.html.
 *
 * Boot sequence:
 *   1. initAuth()   — call /me to find out if the browser session is valid
 *                     (mirrors the Go BFF AuthMiddleware running on every request)
 *                     this step was skipped, now the /me is called inside the router on every navigation
 *                     so doing it on start up as well was redundant
 *   2. initRouter() — render the current URL's page and set up link interception
 *
 * Everything else (navbar, footer, page modules) is lazy-loaded by the router.
 */

import { initRouter } from './router.js';
import { initChatWidget } from './chatWidget.js';

async function boot() {
  // Start the router — this renders the first page.
  initRouter();

  // Initialize the floating chat widget separate from the navbar.
  initChatWidget();
}

boot();
