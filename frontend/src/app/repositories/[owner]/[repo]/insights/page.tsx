'use client';

import React from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import RepositoryStats from '@/components/repository/RepositoryStats';
import LanguageStats from '@/components/repository/LanguageStats';
import ContributorStats from '@/components/repository/ContributorStats';
import { useRepositoryStats } from '@/hooks/useRepositoryStats';
import { ExclamationTriangleIcon, ChartBarIcon } from '@heroicons/react/24/outline';

export default function RepositoryInsightsPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const { 
    statistics, 
    languages, 
    contributors, 
    loading, 
    error, 
    refetch 
  } = useRepositoryStats(owner, repo);

  if (!owner || !repo) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Card className="p-6">
            <div className="text-center">
              <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-2">Invalid Repository</h3>
              <p className="text-muted-foreground">The repository owner or name is missing from the URL.</p>
            </div>
          </Card>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <Link 
            href="/repositories" 
            className="hover:text-foreground transition-colors"
          >
            Repositories
          </Link>
          <span>/</span>
          <Link 
            href={`/repositories/${owner}/${repo}`}
            className="hover:text-foreground transition-colors"
          >
            {owner}/{repo}
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">Insights</span>
        </nav>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-3">
            <ChartBarIcon className="h-8 w-8 text-primary" />
            <div>
              <h1 className="text-3xl font-bold text-foreground">
                Repository Insights
              </h1>
              <p className="text-muted-foreground">
                Statistics and analytics for {owner}/{repo}
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-3">
            <Button variant="outline" size="sm" onClick={refetch} disabled={loading}>
              {loading ? 'Refreshing...' : 'Refresh'}
            </Button>
            <Link href={`/repositories/${owner}/${repo}`}>
              <Button variant="secondary" size="sm">
                Back to Repository
              </Button>
            </Link>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div className="border-b border-border mb-8">
          <nav className="-mb-px flex space-x-8">
            <div className="border-b-2 border-blue-500 py-2 px-1 text-sm font-medium text-blue-600">
              <ChartBarIcon className="w-4 h-4 mr-2 inline" />
              Overview
            </div>
          </nav>
        </div>

        {/* Loading State */}
        {loading && !statistics && (
          <div className="space-y-6">
            <div className="animate-pulse">
              <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
              <div className="h-4 bg-muted rounded w-2/3 mb-6"></div>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <Card key={i} className="p-6">
                  <div className="animate-pulse">
                    <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
                    <div className="h-8 bg-muted rounded w-1/2"></div>
                  </div>
                </Card>
              ))}
            </div>
          </div>
        )}

        {/* Error State */}
        {error && (
          <Card className="p-6">
            <div className="text-center">
              <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-2">Failed to load insights</h3>
              <p className="text-muted-foreground mb-4">{error}</p>
              <Button onClick={refetch} disabled={loading}>
                {loading ? 'Retrying...' : 'Retry'}
              </Button>
            </div>
          </Card>
        )}

        {/* Content */}
        {!loading && !error && (
          <div className="space-y-8">
            {/* Repository Statistics */}
            {statistics && (
              <RepositoryStats statistics={statistics} />
            )}

            {/* Two Column Layout */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              {/* Language Statistics */}
              {languages && languages.length > 0 && (
                <LanguageStats
                  languages={languages}
                  primaryLanguage={statistics?.primary_language}
                  showPercentages={true}
                  showBytes={true}
                  compact={false}
                />
              )}

              {/* Contributor Statistics */}
              {contributors && contributors.length > 0 && (
                <ContributorStats
                  contributors={contributors}
                  totalContributors={statistics?.contributors}
                  showDetails={true}
                  maxDisplay={10}
                />
              )}
            </div>

            {/* Empty State */}
            {!statistics && !languages?.length && !contributors?.length && (
              <Card className="p-12">
                <div className="text-center">
                  <ChartBarIcon className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-xl font-medium text-foreground mb-2">No insights available</h3>
                  <p className="text-muted-foreground mb-6">
                    Repository statistics will appear here once the repository has been analyzed.
                  </p>
                  <div className="flex items-center justify-center space-x-3">
                    <Button onClick={refetch} disabled={loading}>
                      {loading ? 'Analyzing...' : 'Analyze Repository'}
                    </Button>
                    <Badge variant="outline" className="text-xs">
                      This may take a few moments
                    </Badge>
                  </div>
                </div>
              </Card>
            )}
          </div>
        )}
      </div>
    </AppLayout>
  );
}