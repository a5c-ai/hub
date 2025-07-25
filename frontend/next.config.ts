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
  // Reduce memory usage in CI
  ...(process.env.CI === 'true' && {
    typescript: {
      // Skip type checking during build (already done in linting phase)
      ignoreBuildErrors: false,
    },
    eslint: {
      // Skip ESLint during build (already done in linting phase)
      ignoreDuringBuilds: false,
    },
  }),
};

export default nextConfig;
