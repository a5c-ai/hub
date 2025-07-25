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
  // Optimize for CI builds - skip type checking and linting during build
  // since they are done separately in the workflow
  ...(process.env.CI === 'true' && {
    typescript: {
      // Skip type checking during build to prevent hanging
      ignoreBuildErrors: true,
    },
    eslint: {
      // Skip ESLint during build to prevent hanging
      ignoreDuringBuilds: true,
    },
  }),
};

export default nextConfig;
