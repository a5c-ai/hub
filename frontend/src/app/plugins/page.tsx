"use client";

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { PluginCard } from '@/components/plugins/PluginCard';
import { pluginApi } from '@/lib/api';
import type { PluginManifest } from '@/types';

/**
 * Plugin Marketplace listing page.
 */
export default function PluginsPage() {
  const [plugins, setPlugins] = useState<PluginManifest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();

  useEffect(() => {
    pluginApi
      .listMarketplace()
      .then((resp) => setPlugins(Array.isArray(resp) ? resp : []))
      .catch((err) => {
        console.error(err);
        setError('Failed to load plugins.');
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <AppLayout>
      <div className="p-6">
        <h1 className="text-3xl font-bold text-foreground">Plugin Marketplace</h1>
        {error && <div className="text-destructive mt-4">{error}</div>}
        {loading ? (
          <div className="mt-6">Loading plugins...</div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
            {plugins.map((plugin) => (
              <PluginCard key={plugin.metadata.name} manifest={plugin} />
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  );
}
