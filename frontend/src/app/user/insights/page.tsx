'use client';

import React, { useState, useEffect } from 'react';
import UserAnalyticsDashboard from '@/components/analytics/UserAnalyticsDashboard';


export default function UserInsightsPage() {
  const [data, setData] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);
  const [timeRange, setTimeRange] = useState<'weekly' | 'monthly' | 'yearly'>('monthly');

  useEffect(() => {
    const fetchUserAnalytics = async () => {
      try {
        setIsLoading(true);
        setError(undefined);

        // Fetch real analytics data from API
        const response = await fetch(
          `/api/v1/user/analytics/activity?period=${timeRange}`
        );
        if (!response.ok) {
          throw new Error(
            `Failed to fetch user analytics: ${response.statusText}`
          );
        }
        const result = await response.json();
        // Transform backend data to frontend format
        setData({
          user: {
            id: result.user.id,
            username: result.user.username,
            fullName: result.user.full_name,
            avatarUrl: result.user.avatar_url,
            joinedAt: result.user.created_at,
          },
          activityStats: {
            totalLogins: result.activity_stats.total_logins,
            totalSessions: result.activity_stats.total_sessions,
            avgSessionTime: result.activity_stats.avg_session_time ?? 0,
            totalPageViews: result.activity_stats.total_page_views,
            activityTrend: result.activity_stats.activity_trend,
          },
          contributionStats: {
            totalCommits: result.contribution_stats.total_commits,
            totalPullRequests:
              result.contribution_stats.total_pull_requests,
            totalIssues: result.contribution_stats.total_issues,
            totalComments: result.contribution_stats.total_comments,
            contributionTrend:
              result.contribution_stats.contribution_trend,
          },
          repositoryStats: {
            totalRepositories:
              result.repository_stats.total_repositories,
            totalStars: result.repository_stats.total_stars,
            totalForks: result.repository_stats.total_forks,
            repositoryTrend: result.repository_stats.repository_trend,
          },
        });
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
