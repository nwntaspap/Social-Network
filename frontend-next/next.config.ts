import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // That's it! Next.js handles CSS automatically

  // DiceBear avatars used by dev seed data (db/seeds/dev_data.sql)
  images: {
    dangerouslyAllowSVG: true,
    remotePatterns: [{ protocol: 'https', hostname: 'api.dicebear.com', pathname: '/7.x/**' }],
  },

  // Proxy API requests to Go backend
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
};

export default nextConfig;
