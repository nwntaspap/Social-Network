/**
 * router.js
 *
 * Client-side SPA router.
 *
 * Maps URL paths to page modules, handles browser back/forward, and
 * intercepts all <a data-link> clicks to keep navigation inside the SPA.
 *
 * This replaces the Go BFF's SetupRoutes() + http.HandleFunc registrations.
 * The protection that RequireAuth middleware provided is replicated via the
 * `protected: true` flag on route definitions — the router calls requireAuth()
 * before rendering those pages.
 *
 * Adding a new page:
 *   1. Create frontend/static/js/pages/myPage.js with a renderMyPage(user) export
 *   2. Add an entry to ROUTES below
 *   3. Done — no server changes needed
 */

import { getUser, authMiddleware } from './auth.js';
import { renderNavbar } from './navbar.js';
import { renderFooter } from './footer.js';

// ─── Route definitions ────────────────────────────────────────────────────────

/**
 * Each route:
 *   path      — exact path OR a prefix string ending with "*" for dynamic segments
 *   loader    — async function that imports the page module
 *   render    — name of the exported render function in that module
 *   protected — if true, redirects to /login when no user is in auth state
 *   title     — document.title to set on navigation
 */
const ROUTES = [
  {
    path: '/',
    loader: () => import('./pages/home.js'),
    render: 'renderHomePage',
    protected: true,
    title: 'Forum — Home',
  },
  {
    path: '/categories',
    loader: () => import('./pages/categories.js'),
    render: 'renderCategoriesPage',
    protected: true,
    title: 'Forum — Categories',
  },
  {
    path: '/topics',
    loader: () => import('./pages/topics.js'),
    render: 'renderTopicsPage',
    protected: true,
    title: 'Forum — Topics',
  },
  {
    path: '/topic/*',
    loader: () => import('./pages/topic.js'),
    render: 'renderTopicPage',
    protected: true,
    title: 'Forum — Topic',
  },
  {
    path: '/topics/create',
    loader: () => import('./pages/createPost.js'),
    render: 'renderCreatePostPage',
    protected: true,
    title: 'Forum — New Post',
  },
  {
    path: '/login',
    loader: () => import('./pages/login.js'),
    render: 'renderLoginPage',
    protected: false,
    title: 'Forum — Login',
  },
  {
    path: '/auth/google/login',
    loader: () => import('./pages/googlelogin.js'),
    render: 'handleGoogleOAuthLogin',
    protected: false,
    title: 'Forum — Login',
  },
  {
    path: '/auth/github/login',
    loader: () => import('./pages/githublogin.js'),
    render: 'handleGithubOAuthLogin',
    protected: false,
    title: 'Forum — Login',
  },
  {
    path: '/register',
    loader: () => import('./pages/register.js'),
    render: 'renderRegisterPage',
    protected: false,
    title: 'Forum — Register',
  },
  {
    path: '/activity',
    loader: () => import('./pages/activity.js'),
    render: 'renderActivityPage',
    protected: true,
    title: 'Forum — Activity',
  },
  {
    path: '/logout',
    loader: () => import('./pages/logout.js'),
    render: 'renderLogoutPage',
    protected: true,
    title: 'Forum — Logout',
  },
];

// ─── Router state ─────────────────────────────────────────────────────────────

let _isInitialised = false;

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Initialises the router. Call once from app.js after auth is resolved.
 * Handles the initial page load and sets up all link interception.
 */
export function initRouter() {
  if (_isInitialised) return;
  _isInitialised = true;

  // Handle browser back/forward
  window.addEventListener('popstate', () => {
    handleRoute(window.location.pathname + window.location.search);
  });

  // Intercept ALL clicks on [data-link] anchors (event delegation on document)
  document.addEventListener('click', (e) => {
    const anchor = e.target.closest('a[data-link]');
    if (!anchor) return;

    e.preventDefault();
    const href = anchor.getAttribute('href');
    if (href && href !== window.location.pathname) {
      navigate(href);
    }
  });

  // Render the current URL on startup
  handleRoute(window.location.pathname + window.location.search);
}

/**
 * Navigates to a new path, updating the browser URL and rendering the page.
 *
 * @param {string} path
 */
export function navigate(path) {
  if (path !== window.location.pathname + window.location.search) {
    history.pushState(null, '', path);
  }
  handleRoute(path);
}

// ─── Internal routing logic ───────────────────────────────────────────────────

async function handleRoute(fullPath) {
  // Split path from query string
  const [pathname] = fullPath.split('?');

  const route = matchRoute(pathname);

  await authMiddleware()
  // Always re-render the navbar so it reflects the latest auth state
  const user = getUser();
  renderNavbar(user);
  renderFooter();

  // Update document title
  if (route) document.title = route.title ?? 'Forum';

  if (!route) {
    renderNotFound();
    return;
  }

  // Auth guard — mirrors RequireAuth middleware
  if (route.protected && !user) {
    navigate('/login');
    return;
  }

  // Show loading state while the page module loads
  showAppLoading();

  try {
    const module = await route.loader();
    const renderFn = module[route.render];

    if (typeof renderFn !== 'function') {
      throw new Error(`Page module missing export: ${route.render}`);
    }

    await renderFn(user);
  } catch (err) {
    console.error('Route render error:', err);
    renderError(err.message);
  }
}

/**
 * Finds the best matching route for a given pathname.
 * Supports exact matches and wildcard suffix (*).
 */
function matchRoute(pathname) {
  // Exact match first
  const exact = ROUTES.find((r) => r.path === pathname);
  if (exact) return exact;

  // Wildcard match: "/topic/*" matches "/topic/123"
  const wildcard = ROUTES.find((r) => {
    if (!r.path.endsWith('*')) return false;
    const prefix = r.path.slice(0, -1); // strip trailing *
    return pathname.startsWith(prefix);
  });

  return wildcard ?? null;
}

// ─── App-root state helpers ───────────────────────────────────────────────────

function showAppLoading() {
  const root = document.getElementById('app-root');
  if (!root) return;
  root.innerHTML = /* html */ `
    <div class="page-loading">
      <div class="page-loading-spinner"></div>
    </div>
  `;
}

function renderNotFound() {
  const root = document.getElementById('app-root');
  if (!root) return;
  document.title = 'Forum — Not Found';
  root.innerHTML = /* html */ `
    <div class="not-found-container main-container" style="text-align:center;padding:4rem 1rem">
      <h1 style="font-size:4rem;color:var(--dark-background)">404</h1>
      <p style="font-size:1.3rem;color:var(--grey-color);margin-top:1rem">
        Oops! The page you're looking for has vanished into the digital void.
      </p>
      <a href="/" data-link style="
        display:inline-block;
        margin-top:2rem;
        padding:0.8rem 1.5rem;
        background:var(--primary-color);
        color:#fff;
        border-radius:8px;
        text-decoration:none;
        transition:background 0.3s
      ">Go Home</a>
    </div>
  `;
}

function renderError(message) {
  const root = document.getElementById('app-root');
  if (!root) return;
  root.innerHTML = /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <h2 style="color:#e53e3e">Something went wrong</h2>
      <p style="color:var(--grey-color);margin-top:1rem">${message ?? 'An unexpected error occurred.'}</p>
      <a href="/" data-link style="
        display:inline-block;
        margin-top:2rem;
        padding:0.8rem 1.5rem;
        background:var(--secondary-color);
        color:#fff;
        border-radius:8px;
        text-decoration:none
      ">Go Home</a>
    </div>
  `;
}
