"use client";

import Link from 'next/link';
import { Card, CardContent, Button } from '@/components/ui';
import type { PluginManifest } from '@/types';

interface PluginCardProps {
  manifest: PluginManifest;
}

/**
 * Card component to display plugin summary in the marketplace.
 */
export function PluginCard({ manifest }: PluginCardProps) {
  const { metadata } = manifest;
  return (
    <Card className="h-full">
      <CardContent className="flex flex-col">
        <Link href={`/plugins/${metadata.name}`}> 
          <h2 className="text-lg font-semibold hover:underline">{metadata.name}</h2>
        </Link>
        {metadata.description && (
          <p className="text-muted-foreground flex-1 mt-2">{metadata.description}</p>
        )}
        <div className="mt-4 flex items-center justify-between">
          <span className="text-sm text-muted-foreground">v{metadata.version}</span>
          <Button asChild size="sm">
            <Link href={`/plugins/${metadata.name}`}>View Details</Link>
          </Button>
        </div>
        {metadata.website && (
          <a
            href={metadata.website}
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-primary mt-2"
          >
            Visit Website
          </a>
        )}
      </CardContent>
    </Card>
  );
}
