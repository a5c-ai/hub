'use client';

import React, { useState, useEffect } from 'react';
import UserAnalyticsDashboard from '@/components/analytics/UserAnalyticsDashboard';
// import { Card } from '@/components/ui/Card'; // Currently unused
// import { ExclamationTriangleIcon } from '@heroicons/react/24/outline'; // Currently unused

// Mock data for development - replace with actual API calls
const mockUserAnalytics = {
  user: {
    id: '1',
    username: 'johndoe',
    fullName: 'John Doe',
    avatarUrl: '/api/placeholder/64/64',
    joinedAt: '2023-06-15T10:30:00Z',
  },
  activityStats: {
    totalLogins: 156,
    totalSessions: 89,
    avgSessionTime: 45.2,
    totalPageViews: 1240,
    activityTrend: [
      { date: '2024-01-01', value: 12 },
      { date: '2024-01-02', value: 18 },
      { date: '2024-01-03', value: 15 },
      { date: '2024-01-04', value: 22 },
      { date: '2024-01-05', value: 19 },
    ],
  },
  contributionStats: {
    totalCommits: 342,
    totalPullRequests: 78,
    totalIssues: 23,
    totalComments: 156,
    contributionTrend: [
      { date: '2024-01-01', commits: 5, prs: 2, issues: 1 },
      { date: '2024-01-02', commits: 8, prs: 1, issues: 0 },
      { date: '2024-01-03', commits: 3, prs: 3, issues: 2 },
      { date: '2024-01-04', commits: 12, prs: 0, issues: 1 },
      { date: '2024-01-05', commits: 6, prs: 2, issues: 0 },
    ],
  },
  repositoryStats: {
    totalRepositories: 24,
    totalStars: 456,
    totalForks: 89,
    repositoryTrend: [
      { date: '2024-01-01', value: 20 },
      { date: '2024-01-02', value: 21 },
      { date: '2024-01-03', value: 22 },
      { date: '2024-01-04', value: 23 },
      { date: '2024-01-05', value: 24 },
    ],
  },
};

export default function UserInsightsPage() {
  const [data, setData] = useState<typeof mockUserAnalytics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);
  const [timeRange, setTimeRange] = useState<'weekly' | 'monthly' | 'yearly'>('monthly');

  useEffect(() => {
    const fetchUserAnalytics = async () => {
      try {
        setIsLoading(true);
        setError(undefined);

        // Replace with actual API call
        // const response = await fetch(`/api/v1/user/analytics/activity?period=${timeRange}`);
        // if (!response.ok) {
        //   throw new Error('Failed to fetch user analytics');
        // }
        // const result = await response.json();
        
        // For now, use mock data with a delay to simulate loading
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        setData(mockUserAnalytics);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setIsLoading(false);
      }
    };

    fetchUserAnalytics();
  }, [timeRange]);

  const handleTimeRangeChange = (newTimeRange: 'weekly' | 'monthly' | 'yearly') => {
    setTimeRange(newTimeRange);
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <UserAnalyticsDashboard
        data={data}
        isLoading={isLoading}
        error={error}
        timeRange={timeRange}
        onTimeRangeChange={handleTimeRangeChange}
      />
    </div>
  );
}