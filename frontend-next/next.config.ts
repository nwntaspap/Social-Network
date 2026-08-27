import type { NextConfig } from 'next';

// Origin of the Go backend. Defaults to the local dev setup; docker-compose
// sets BACKEND_ORIGIN=https://forum:8080 for the containerised frontend.
const backendOrigin = process.env.BACKEND_ORIGIN || 'https://localhost:8080';

// Origin of the notifications microservice. Direct browser calls work in local
// dev; in Docker the request is proxied through Next so session cookies stay
// first-party (SameSite) and no extra CORS/TLS setup is needed.
const notificationsOrigin = process.env.NOTIFICATIONS_ORIGIN || 'http://localhost:8081';

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Emits .next/standalone so the production image needs no node_modules.
  output: 'standalone',

  // DiceBear avatars used by dev seed data (db/seeds/dev_data.sql);
  // Google (lh3.googleusercontent.com) and GitHub (avatars.githubusercontent.com)
  // avatars come from OAuth sign-in.
  images: {
    dangerouslyAllowSVG: true,
    remotePatterns: [
      { protocol: 'https', hostname: 'api.dicebear.com', pathname: '/7.x/**' },
      { protocol: 'https', hostname: 'lh3.googleusercontent.com' },
      { protocol: 'https', hostname: 'avatars.githubusercontent.com' },
    ],
  },

  // Proxy API requests and uploaded static files to the Go backend
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${backendOrigin}/api/:path*`,
      },
      {
        source: '/uploads/:path*',
        destination: `${backendOrigin}/uploads/:path*`,
      },
      {
        // Client paths already carry /api/v1 (see NOTIFICATIONS_BASE), so the
        // capture is appended verbatim — adding /api here would double it.
        source: '/notifications-api/:path*',
        destination: `${notificationsOrigin}/:path*`,
      },
    ];
  },
};

export default nextConfig;
