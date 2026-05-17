/**
 * pages/logout.js
 *
 * Calls backend POST /api/v1/logout, clears client auth state,
 * re-renders the navbar in guest mode, then navigates home.
 *
 * The BFF's Logout handler did:
 *   1. POST to backend /logout (with cookies so the session is deleted server-side)
 *   2. Clear the access_token and refresh_token cookies (MaxAge: -1)
 *   3. Redirect to /
 *
 * In CSR the cookie clearing happens server-side when the backend logout
 * endpoint responds — it sets Set-Cookie headers with MaxAge:-1, which the
 * browser processes automatically because we send credentials:"include".
 */

import { api } from '../api.js';
import { navigate } from '../router.js';
import { closeNotificationStream } from '../navbar.js';
import { cleanupChat } from '../chat.js';

export async function renderLogoutPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  if (!user) {
    navigate('/login');
    return;
  }

  root.innerHTML = buildHTML('Signing out…');

  try {
    // POST /api/v1/logout — backend deletes the session and clears cookies
    await api.post('/logout');
  } catch (err) {
    console.warn('Backend logout failed, clearing client state anyway:', err.message);
  }

  //the cookies are deleted from the backend response, so we just go to the home page and the middleware will get the user and render the navbar

  closeNotificationStream();
  cleanupChat();

  navigate('/');
}

function buildHTML(message, showRetry = false) {
  return /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <div class="page-loading">
        <div class="page-loading-spinner"></div>
      </div>
      <p style="margin-top:1.5rem;color:var(--white-background-light);font-size:1.1rem">
        ${message}
      </p>
      ${
        showRetry
          ? `<div style="margin-top:1.5rem">
             <button id="logout-retry" class="btn-signup">Retry</button>
           </div>`
          : ''
      }
    </div>
  `;
}
