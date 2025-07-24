import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Optimize for CI environments
  experimental: {
    // Reduce build worker count in CI to prevent memory issues
    cpus: process.env.CI ? 2 : undefined,
  },
  
  // Enable standalone output for better deployment
  output: process.env.NODE_ENV === 'production' ? 'standalone' : undefined,
  
  // Optimize images for better performance
  images: {
    domains: ['localhost'],
    // Minimize image optimization in CI to reduce build time
    unoptimized: process.env.CI === 'true',
  },
  
  // Webpack optimizations
  webpack: (config, { isServer }) => {
    // Optimize memory usage in CI
    if (process.env.CI) {
      config.optimization.splitChunks = {
        chunks: 'all',
        cacheGroups: {
          default: {
            minChunks: 2,
            priority: -20,
            reuseExistingChunk: true,
          },
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            priority: -10,
            chunks: 'all',
          },
        },
      };
    }
    
    return config;
  },
};

export default nextConfig;
