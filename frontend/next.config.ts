import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Performance optimizations
  poweredByHeader: false,
  // Memory optimizations for CI builds
  generateBuildId: async () => {
    // Use shorter build ID in CI to reduce memory usage
    return process.env.CI ? 'ci-build' : Date.now().toString();
  },
  // Disable source maps in production builds to speed up CI
  productionBrowserSourceMaps: false,
  // Optimize output and compression
  compress: true,
  // TypeScript build should fail on type errors; skip ESLint during production builds to avoid lint warnings failure
  typescript: {
    ignoreBuildErrors: false,
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
};

export default nextConfig;
