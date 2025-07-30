"use client";

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { pluginApi } from '@/lib/api';
import { Button } from '@/components/ui';
import type { PluginManifest } from '@/types';

/**
 * Plugin detail page showing manifest info and install action.
 */
export default function PluginDetailPage() {
  const { name } = useParams() as { name: string };
  const [plugin, setPlugin] = useState<PluginManifest>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const [installing, setInstalling] = useState(false);

  useEffect(() => {
    pluginApi
      .listMarketplace()
      .then((resp) => {
        const found = resp.data.find((p) => p.metadata.name === name);
        if (!found) throw new Error('Plugin not found');
        setPlugin(found);
      })
      .catch((err) => {
        console.error(err);
        setError('Failed to load plugin details.');
      })
      .finally(() => setLoading(false));
  }, [name]);

  const handleInstall = () => {
    if (!plugin) return;
    setInstalling(true);
    // Example: install for organization; adapt context as needed
    pluginApi
      .installOrgPlugin(name, name, {})
      .then(() => alert('Plugin installed'))
      .catch(() => alert('Installation failed'))
      .finally(() => setInstalling(false));
  };

  if (loading) return <AppLayout><div className="p-6">Loading...</div></AppLayout>;
  if (error || !plugin)
    return <AppLayout><div className="p-6 text-destructive">{error || 'Plugin not found'}</div></AppLayout>;

  const { metadata } = plugin;

  return (
    <AppLayout>
      <div className="p-6">
        <h1 className="text-3xl font-bold">{metadata.name}</h1>
        {metadata.description && (
          <p className="text-muted-foreground mt-2">{metadata.description}</p>
        )}
        <div className="mt-4 space-y-1 text-sm">
          <div><strong>Version:</strong> {metadata.version}</div>
          {metadata.author && (<div><strong>Author:</strong> {metadata.author}</div>)}
          {metadata.website && (
            <div>
              <strong>Website:</strong>{' '}
              <a
                href={metadata.website}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline"
              >
                {metadata.website}
              </a>
            </div>
          )}
        </div>
        <div className="mt-6">
          <Button onClick={handleInstall} disabled={installing}>
            {installing ? 'Installing...' : 'Install Plugin'}
          </Button>
        </div>
      </div>
    </AppLayout>
  );
}
