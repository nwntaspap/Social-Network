import { fetchCurrentUser } from '../api.js';
import { setUser } from '../auth.js';
import { navigate } from '../router.js';

// Opens the OAuth popup and handles postMessage from the bridge page.
export async function handleGoogleOAuthLogin(user) {
  const BACKEND_LOGIN_URL = '/api/v1/auth/google/login';
  const FRONTEND_ORIGIN = window.location.origin;

  const width = 500;
  const height = 700;
  const left = Math.round((screen.width - width) / 2);
  const top = Math.round((screen.height - height) / 2);

  const popup = window.open(
    BACKEND_LOGIN_URL,
    'oauth_popup',
    `width=${width},height=${height},left=${left},top=${top}`
  );

  if (!popup) {
    // Popup blocked — fallback to full redirect
    window.location.href = BACKEND_LOGIN_URL;
    return;
  }

  function onMessage(event) {
    // Validate origin
    if (event.origin !== FRONTEND_ORIGIN) return;

    const payload = event.data;
    if (!payload || payload.type !== 'oauth') return;

    window.removeEventListener('message', onMessage);

  }

  window.addEventListener('message', onMessage);

  // Detect popup closed without message
  const popupCheck = setInterval(() => {
    if (popup.closed) {
      clearInterval(popupCheck);
      window.removeEventListener('message', onMessage);
      navigate('/');
      // Optionally, navigate back to login or show cancellation
    }
  }, 500);
}
