import { fetchCurrentUser } from '../api.js';
import { setUser } from '../auth.js';
import { navigate } from '../router.js';

export async function handleGithubOAuthLogin(user) {
  const BACKEND_LOGIN_URL = 'api/v1/auth/github/login';
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
    window.location.href = BACKEND_LOGIN_URL;
    return;
  }

  function onMessage(event) {
    if (event.origin !== FRONTEND_ORIGIN) return;
    const payload = event.data;
    if (!payload || payload.type !== 'oauth') return;

    window.removeEventListener('message', onMessage);

    if (payload.status === 'success') {
      fetchCurrentUser()
        .then((u) => {
          if (u) setUser(u);
          navigate('/');
        })
        .catch((err) => {
          console.error('Failed to fetch user after oauth:', err);
          navigate('/');
        });
    } else {
      console.error('OAuth error:', payload.message);
      navigate('/login');
    }
  }

  window.addEventListener('message', onMessage);

  const popupCheck = setInterval(() => {
    if (popup.closed) {
      clearInterval(popupCheck);
      window.removeEventListener('message', onMessage);
    }
  }, 500);
}
