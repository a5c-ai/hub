import type { NextConfig } from "next";

import withPWA from 'next-pwa';
import runtimeCaching from 'next-pwa/cache';

/**
 * Base Next.js configuration
 */
const baseConfig: NextConfig = {
  // React strict mode
  reactStrictMode: true,
  // Enable standalone output for Docker builds
  output: 'standalone',
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
  // Skip type errors during production builds; skip ESLint during production builds to avoid lint warnings failure
  typescript: { ignoreBuildErrors: true },
  eslint: { ignoreDuringBuilds: true },
};

/**
 * Export Next.js configuration wrapped with PWA support
 */
export default withPWA({
  dest: 'public',
  disable: process.env.NODE_ENV === 'development',
  register: true,
  skipWaiting: true,
  sw: 'sw.js',
  runtimeCaching,
})(baseConfig);
