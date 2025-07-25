'use client';

import React, { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import RepositoryInsights from '@/components/analytics/RepositoryInsights';
import { Card } from '@/components/ui/Card';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';

// Mock data for development - replace with actual API calls
const mockRepositoryInsights = {
  repository: {
    id: '1',
    name: 'example-repo',
    description: 'An example repository for testing analytics',
    language: 'TypeScript',
    starsCount: 245,
    forksCount: 67,
    watchersCount: 32,
  },
  codeStats: {
    totalLinesOfCode: 15420,
    totalFiles: 156,
    totalCommits: 342,
    totalBranches: 8,
    languageBreakdown: {
      'TypeScript': 65.4,
      'JavaScript': 20.1,
      'CSS': 8.2,
      'HTML': 4.1,
      'JSON': 2.2,
    },
  },
  activityStats: {
    totalViews: 1240,
    totalClones: 89,
    activityTrend: [
      { date: '2024-01-01', commits: 12, views: 45 },
      { date: '2024-01-02', commits: 8, views: 52 },
      { date: '2024-01-03', commits: 15, views: 38 },
    ],
  },
  contributorStats: {
    totalContributors: 8,
    activeContributors: 3,
    topContributors: [
      { username: 'alice', commitCount: 156, linesAdded: 4521, linesDeleted: 1204 },
      { username: 'bob', commitCount: 89, linesAdded: 2103, linesDeleted: 567 },
      { username: 'charlie', commitCount: 67, linesAdded: 1890, linesDeleted: 345 },
      { username: 'diana', commitCount: 30, linesAdded: 890, linesDeleted: 123 },
    ],
  },
  issueStats: {
    totalIssues: 45,
    openIssues: 8,
    closedIssues: 37,
    avgTimeToClose: 48.5,
  },
  pullRequestStats: {
    totalPullRequests: 78,
    openPullRequests: 5,
    mergedPullRequests: 68,
    avgTimeToMerge: 24.2,
  },
};

export default function RepositoryInsightsPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const [data, setData] = useState<typeof mockRepositoryInsights | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);

  useEffect(() => {
    const fetchInsights = async () => {
      try {
        setIsLoading(true);
        setError(undefined);

        // Replace with actual API call
        // const response = await fetch(`/api/v1/repositories/${owner}/${repo}/analytics`);
        // if (!response.ok) {
        //   throw new Error('Failed to fetch repository insights');
        // }
        // const result = await response.json();
        
        // For now, use mock data with a delay to simulate loading
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        setData(mockRepositoryInsights);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setIsLoading(false);
      }
    };

    if (owner && repo) {
      fetchInsights();
    }
  }, [owner, repo]);

  if (!owner || !repo) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-foreground mb-2">Invalid Repository</h3>
          <p className="text-muted-foreground">The repository owner or name is missing from the URL.</p>
        </div>
      </Card>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <RepositoryInsights
        owner={owner}
        repo={repo}
        data={data}
        isLoading={isLoading}
        error={error}
      />
    </div>
  );
}