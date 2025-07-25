'use client';

import React, { useState } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { Avatar } from '../ui/Avatar';
import {
  UserIcon,
  CodeBracketIcon,
  DocumentTextIcon,
  ExclamationTriangleIcon,
  ChatBubbleLeftIcon,
  ClockIcon,
  CalendarIcon,
  ChartBarIcon
} from '@heroicons/react/24/outline';

interface UserAnalyticsData {
  user: {
    id: string;
    username: string;
    fullName?: string;
    avatarUrl?: string;
    joinedAt: string;
  };
  activityStats: {
    totalLogins: number;
    totalSessions: number;
    avgSessionTime?: number;
    totalPageViews: number;
    activityTrend: Array<{ date: string; value: number }>;
  };
  contributionStats: {
    totalCommits: number;
    totalPullRequests: number;
    totalIssues: number;
    totalComments: number;
    contributionTrend: Array<{ date: string; commits: number; prs: number; issues: number }>;
  };
  repositoryStats: {
    totalRepositories: number;
    totalStars: number;
    totalForks: number;
    repositoryTrend: Array<{ date: string; value: number }>;
  };
}

interface UserAnalyticsDashboardProps {
  data?: UserAnalyticsData | null;
  isLoading?: boolean;
  error?: string;
  timeRange?: 'weekly' | 'monthly' | 'yearly';
  onTimeRangeChange?: (range: 'weekly' | 'monthly' | 'yearly') => void;
}

export const UserAnalyticsDashboard: React.FC<UserAnalyticsDashboardProps> = ({
  data,
  isLoading = false,
  error,
  timeRange = 'monthly',
  onTimeRangeChange,
}) => {
  const [selectedTimeRange, setSelectedTimeRange] = useState(timeRange);
  const [activeTab, setActiveTab] = useState<'overview' | 'activity' | 'contributions' | 'repositories'>('overview');

  const handleTimeRangeChange = (range: 'weekly' | 'monthly' | 'yearly') => {
    setSelectedTimeRange(range);
    onTimeRangeChange?.(range);
  };

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatDuration = (minutes: number): string => {
    if (minutes >= 60) {
      const hours = Math.floor(minutes / 60);
      const remainingMinutes = minutes % 60;
      return `${hours}h ${remainingMinutes}m`;
    }
    return `${minutes}m`;
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse">
          <div className="flex items-center space-x-4 mb-6">
            <div className="h-16 w-16 bg-muted rounded-full"></div>
            <div className="flex-1 space-y-2">
              <div className="h-6 bg-muted rounded w-48 mb-2"></div>
              <div className="h-4 bg-muted rounded w-32"></div>
            </div>
          </div>
        </div>
        
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="space-y-2">
              <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
              <div className="h-8 bg-muted rounded w-1/2"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-foreground mb-2">Failed to load analytics</h3>
          <p className="text-muted-foreground">{error}</p>
        </div>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <ChartBarIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-medium text-foreground mb-2">No analytics available</h3>
          <p className="text-muted-foreground">User analytics will appear here once there's activity.</p>
        </div>
      </Card>
    );
  }

  const tabs = [
    { id: 'overview', label: 'Overview', icon: ChartBarIcon },
    { id: 'activity', label: 'Activity', icon: UserIcon },
    { id: 'contributions', label: 'Contributions', icon: CodeBracketIcon },
    { id: 'repositories', label: 'Repositories', icon: DocumentTextIcon },
  ] as const;

  return (
    <div className="space-y-6">
      {/* User Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Avatar
            src={data.user.avatarUrl}
            alt={data.user.username}
            size="lg"
          />
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              {data.user.fullName || data.user.username}
            </h1>
            <p className="text-muted-foreground">@{data.user.username}</p>
            <p className="text-sm text-muted-foreground flex items-center mt-1">
              <CalendarIcon className="h-4 w-4 mr-1" />
              Joined {formatDate(data.user.joinedAt)}
            </p>
          </div>
        </div>
        
        <div className="flex space-x-2">
          {(['weekly', 'monthly', 'yearly'] as const).map((range) => (
            <Button
              key={range}
              variant={selectedTimeRange === range ? 'default' : 'secondary'}
              size="sm"
              onClick={() => handleTimeRangeChange(range)}
            >
              {range.charAt(0).toUpperCase() + range.slice(1)}
            </Button>
          ))}
        </div>
      </div>

      {/* Tabs */}
              <div className="border-b border-border">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                } whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm flex items-center space-x-2`}
              >
                <Icon className="h-4 w-4" />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard
              icon={UserIcon}
              label="Total Logins"
              value={formatNumber(data.activityStats.totalLogins)}
              color="blue"
            />
            <MetricCard
              icon={CodeBracketIcon}
              label="Total Commits"
              value={formatNumber(data.contributionStats.totalCommits)}
              color="green"
            />
            <MetricCard
              icon={DocumentTextIcon}
              label="Repositories"
              value={formatNumber(data.repositoryStats.totalRepositories)}
              color="purple"
            />
            <MetricCard
              icon={ChatBubbleLeftIcon}
              label="Comments"
              value={formatNumber(data.contributionStats.totalComments)}
              color="yellow"
            />
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Activity Summary</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-primary">{data.activityStats.totalSessions}</div>
                  <span className="text-sm text-muted-foreground">Total Sessions</span>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-primary">{data.activityStats.avgSessionTime}</div>
                  <span className="text-sm text-muted-foreground">Avg Session Time</span>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-primary">{data.activityStats.totalPageViews}</div>
                  <span className="text-sm text-muted-foreground">Page Views</span>
                </div>
              </div>
            </Card>

            <Card className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Contribution Summary</h3>
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Pull Requests</span>
                  <span className="font-medium">{formatNumber(data.contributionStats.totalPullRequests)}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Issues Created</span>
                  <span className="font-medium">{formatNumber(data.contributionStats.totalIssues)}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Stars Received</span>
                  <span className="font-medium">{formatNumber(data.repositoryStats.totalStars)}</span>
                </div>
              </div>
            </Card>
          </div>
        </div>
      )}

      {activeTab === 'activity' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <MetricCard
              icon={UserIcon}
              label="Total Logins"
              value={formatNumber(data.activityStats.totalLogins)}
              color="blue"
            />
            <MetricCard
              icon={ClockIcon}
              label="Total Sessions"
              value={formatNumber(data.activityStats.totalSessions)}
              color="green"
            />
            <MetricCard
              icon={ChartBarIcon}
              label="Page Views"
              value={formatNumber(data.activityStats.totalPageViews)}
              color="purple"
            />
          </div>

          <Card className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Activity Trend</h3>
            <div className="h-64 flex items-center justify-center text-gray-500">
              <p>Activity trend chart would go here</p>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'contributions' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard
              icon={CodeBracketIcon}
              label="Commits"
              value={formatNumber(data.contributionStats.totalCommits)}
              color="green"
            />
            <MetricCard
              icon={DocumentTextIcon}
              label="Pull Requests"
              value={formatNumber(data.contributionStats.totalPullRequests)}
              color="blue"
            />
            <MetricCard
              icon={ExclamationTriangleIcon}
              label="Issues"
              value={formatNumber(data.contributionStats.totalIssues)}
              color="yellow"
            />
            <MetricCard
              icon={ChatBubbleLeftIcon}
              label="Comments"
              value={formatNumber(data.contributionStats.totalComments)}
              color="purple"
            />
          </div>

          <Card className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Contribution Activity</h3>
            <div className="h-64 flex items-center justify-center text-gray-500">
              <p>Contribution activity chart would go here</p>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'repositories' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <MetricCard
              icon={DocumentTextIcon}
              label="Total Repositories"
              value={formatNumber(data.repositoryStats.totalRepositories)}
              color="blue"
            />
            <MetricCard
              icon={UserIcon}
              label="Total Stars"
              value={formatNumber(data.repositoryStats.totalStars)}
              color="yellow"
            />
            <MetricCard
              icon={CodeBracketIcon}
              label="Total Forks"
              value={formatNumber(data.repositoryStats.totalForks)}
              color="green"
            />
          </div>

          <Card className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Repository Growth</h3>
            <div className="h-64 flex items-center justify-center text-gray-500">
              <p>Repository growth chart would go here</p>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
};

// Reusable metric card component
interface MetricCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  color: 'blue' | 'green' | 'yellow' | 'purple' | 'red';
}

const MetricCard: React.FC<MetricCardProps> = ({ icon: Icon, label, value, color }) => {
  const colorClasses = {
    blue: 'text-blue-600',
    green: 'text-green-600',
    yellow: 'text-yellow-600',
    purple: 'text-purple-600',
    red: 'text-red-600',
  };

  return (
    <Card className="p-6">
      <div className="flex items-center">
        <div className="flex-shrink-0">
          <Icon className={`h-8 w-8 ${colorClasses[color]}`} />
        </div>
        <div className="ml-4 flex-1">
          <p className="text-sm font-medium text-gray-500">{label}</p>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
        </div>
      </div>
    </Card>
  );
};

export default UserAnalyticsDashboard;