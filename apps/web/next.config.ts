import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Self-contained server bundle for the Docker image.
  output: "standalone",
  poweredByHeader: false,
  async rewrites() {
    // Local `npm run dev` without nginx: forward API calls to the Go service.
    // In Docker/production nginx routes /api before it ever reaches Next.js.
    const api = process.env.API_URL ?? "http://localhost:8081";
    return [{ source: "/api/:path*", destination: `${api}/api/:path*` }];
  },
};

export default nextConfig;
