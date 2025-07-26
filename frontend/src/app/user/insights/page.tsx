'use client';

import React, { useState, useEffect } from 'react';
import UserAnalyticsDashboard from '@/components/analytics/UserAnalyticsDashboard';
import { apiClient } from '@/lib/api';
import { useAuthStore } from '@/store/auth';


export default function UserInsightsPage() {
  const [data, setData] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);
  const [timeRange, setTimeRange] = useState<'weekly' | 'monthly' | 'yearly'>('monthly');

  const { user } = useAuthStore();

  useEffect(() => {
    const fetchUserAnalytics = async () => {
      try {
        setIsLoading(true);
        setError(undefined);

        const [activityRes, contributionsRes, reposRes] = await Promise.all([
          apiClient.get(
            `/user/analytics/activity?period=${timeRange}`
          ),
          apiClient.get(
            `/user/analytics/contributions?period=${timeRange}`
          ),
          apiClient.get(
            `/user/analytics/repositories?period=${timeRange}`
          ),
        ]);

        setData({
          user: {
            id: user?.id ?? '',
            username: user?.username ?? '',
            fullName: user?.name || user?.username,
            avatarUrl: user?.avatar_url,
            joinedAt: user?.created_at || '',
          },
          activityStats: {
            totalLogins: activityRes.data.total_logins,
            totalSessions: activityRes.data.total_sessions,
            avgSessionTime: activityRes.data.avg_session_time ?? 0,
            totalPageViews: activityRes.data.total_page_views,
            activityTrend: activityRes.data.activity_trend,
          },
          contributionStats: {
            totalCommits: contributionsRes.data.total_commits,
            totalPullRequests:
              contributionsRes.data.total_pull_requests,
            totalIssues: contributionsRes.data.total_issues,
            totalComments: contributionsRes.data.total_comments,
            contributionTrend:
              contributionsRes.data.contribution_trend,
          },
          repositoryStats: {
            totalRepositories:
              reposRes.data.total_repositories,
            totalStars: reposRes.data.total_stars,
            totalForks: reposRes.data.total_forks,
            repositoryTrend: reposRes.data.repository_trend,
          },
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setIsLoading(false);
      }
    };

    fetchUserAnalytics();
  }, [timeRange, user]);

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
