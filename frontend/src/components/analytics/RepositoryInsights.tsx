'use client';

import React, { useState } from 'react';
import { Card } from '../ui/Card';
// import { Button } from '../ui/Button';
import { Badge } from '../ui/Badge';
import { 
  CodeBracketIcon, 
  UserGroupIcon, 
  ExclamationTriangleIcon,
  DocumentTextIcon,
  // ClockIcon,
  StarIcon,
  EyeIcon,
  DocumentDuplicateIcon
} from '@heroicons/react/24/outline';

interface RepositoryInsightsData {
  repository: {
    id: string;
    name: string;
    description?: string;
    language: string;
    starsCount: number;
    forksCount: number;
    watchersCount: number;
  };
  codeStats: {
    totalLinesOfCode: number;
    totalFiles: number;
    totalCommits: number;
    totalBranches: number;
    languageBreakdown: Record<string, number>;
  };
  activityStats: {
    totalViews: number;
    totalClones: number;
    activityTrend: Array<{ date: string; commits: number; views: number }>;
  };
  contributorStats: {
    totalContributors: number;
    activeContributors: number;
    topContributors: Array<{
      username: string;
      commitCount: number;
      linesAdded: number;
      linesDeleted: number;
    }>;
  };
  issueStats: {
    totalIssues: number;
    openIssues: number;
    closedIssues: number;
    avgTimeToClose?: number;
  };
  pullRequestStats: {
    totalPullRequests: number;
    openPullRequests: number;
    mergedPullRequests: number;
    avgTimeToMerge?: number;
  };
}

interface RepositoryInsightsProps {
  owner: string;
  repo: string;
  data?: RepositoryInsightsData | null;
  isLoading?: boolean;
  error?: string;
}

export const RepositoryInsights: React.FC<RepositoryInsightsProps> = ({
  owner,
  repo,
  data,
  isLoading = false,
  error,
}) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'code' | 'activity' | 'contributors' | 'issues'>('overview');

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatDuration = (hours: number): string => {
    if (hours < 24) {
      return `${hours.toFixed(1)}h`;
    }
    const days = Math.floor(hours / 24);
    return `${days}d`;
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
          <div className="h-4 bg-muted rounded w-2/3 mb-6"></div>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <Card key={i} className="p-6">
              <div className="animate-pulse">
                <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
                <div className="h-8 bg-muted rounded w-1/2"></div>
              </div>
            </Card>
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
          <h3 className="text-lg font-medium text-foreground mb-2">Failed to load insights</h3>
          <p className="text-muted-foreground">{error}</p>
        </div>
      </Card>
    );
  }

  if (!data) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <DocumentTextIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-medium text-foreground mb-2">No insights available</h3>
          <p className="text-muted-foreground">Insights will appear here once there&apos;s activity in this repository.</p>
        </div>
      </Card>
    );
  }

  const tabs = [
    { id: 'overview', label: 'Overview', icon: DocumentTextIcon },
    { id: 'code', label: 'Code', icon: CodeBracketIcon },
    { id: 'activity', label: 'Activity', icon: EyeIcon },
    { id: 'contributors', label: 'Contributors', icon: UserGroupIcon },
    { id: 'issues', label: 'Issues & PRs', icon: ExclamationTriangleIcon },
  ] as const;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-foreground">
          {owner}/{repo} Insights
        </h1>
        {data.repository.description && (
          <p className="text-muted-foreground mt-2">{data.repository.description}</p>
        )}
        <div className="flex items-center space-x-4 mt-4">
          <Badge variant="secondary">{data.repository.language}</Badge>
          <div className="flex items-center space-x-1 text-sm text-muted-foreground">
            <StarIcon className="h-4 w-4" />
            <span>{formatNumber(data.repository.starsCount)}</span>
          </div>
          <div className="flex items-center space-x-1 text-sm text-muted-foreground">
            <DocumentDuplicateIcon className="h-4 w-4" />
            <span>{formatNumber(data.repository.forksCount)}</span>
          </div>
          <div className="flex items-center space-x-1 text-sm text-muted-foreground">
            <EyeIcon className="h-4 w-4" />
            <span>{formatNumber(data.repository.watchersCount)}</span>
          </div>
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <MetricCard
            icon={CodeBracketIcon}
            label="Lines of Code"
            value={formatNumber(data.codeStats.totalLinesOfCode)}
            color="blue"
          />
          <MetricCard
            icon={DocumentTextIcon}
            label="Total Files"
            value={formatNumber(data.codeStats.totalFiles)}
            color="green"
          />
          <MetricCard
            icon={DocumentDuplicateIcon}
            label="Commits"
            value={formatNumber(data.codeStats.totalCommits)}
            color="purple"
          />
          <MetricCard
            icon={UserGroupIcon}
            label="Contributors"
            value={formatNumber(data.contributorStats.totalContributors)}
            color="yellow"
          />
        </div>
      )}

      {activeTab === 'code' && (
        <div className="space-y-6">
          <Card className="p-6">
            <h3 className="text-lg font-medium text-foreground mb-4">Language Breakdown</h3>
            <div className="space-y-3">
              {Object.entries(data.codeStats.languageBreakdown).map(([language, percentage]) => (
                <div key={language} className="flex items-center">
                  <div className="w-20 text-sm text-muted-foreground">{language}</div>
                  <div className="flex-1 mx-4">
                    <div className="bg-muted rounded-full h-2">
                      <div
                        className="bg-blue-500 h-2 rounded-full"
                        style={{ width: `${percentage}%` }}
                      ></div>
                    </div>
                  </div>
                  <div className="w-12 text-sm text-muted-foreground text-right">{percentage.toFixed(1)}%</div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'activity' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="p-6">
            <h3 className="text-lg font-medium text-foreground mb-4">Repository Activity</h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-primary">{data.activityStats.totalViews}</div>
                <span className="text-sm text-muted-foreground">Total Views</span>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-primary">{data.activityStats.totalClones}</div>
                <span className="text-sm text-muted-foreground">Total Clones</span>
              </div>
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'contributors' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <MetricCard
              icon={UserGroupIcon}
              label="Total Contributors"
              value={formatNumber(data.contributorStats.totalContributors)}
              color="blue"
            />
            <MetricCard
              icon={UserGroupIcon}
              label="Active Contributors"
              value={formatNumber(data.contributorStats.activeContributors)}
              color="green"
            />
          </div>
          
          <Card className="p-6">
            <h3 className="text-lg font-medium text-foreground mb-4">Top Contributors</h3>
            <div className="space-y-3">
              {data.contributorStats.topContributors.slice(0, 10).map((contributor, index) => (
                <div key={contributor.username} className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <span className="text-sm text-muted-foreground w-6">#{index + 1}</span>
                    <span className="font-medium text-foreground">{contributor.username}</span>
                  </div>
                  <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                    <span>{formatNumber(contributor.commitCount)} commits</span>
                    <span className="text-green-600">+{formatNumber(contributor.linesAdded)}</span>
                    <span className="text-red-600">-{formatNumber(contributor.linesDeleted)}</span>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {activeTab === 'issues' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <MetricCard
              icon={ExclamationTriangleIcon}
              label="Total Issues"
              value={formatNumber(data.issueStats.totalIssues)}
              color="red"
            />
            <MetricCard
              icon={ExclamationTriangleIcon}
              label="Open Issues"
              value={formatNumber(data.issueStats.openIssues)}
              color="yellow"
            />
            <MetricCard
              icon={DocumentTextIcon}
              label="Total PRs"
              value={formatNumber(data.pullRequestStats.totalPullRequests)}
              color="blue"
            />
            <MetricCard
              icon={DocumentTextIcon}
              label="Merged PRs"
              value={formatNumber(data.pullRequestStats.mergedPullRequests)}
              color="green"
            />
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Issue Stats</h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {((data.issueStats.closedIssues / data.issueStats.totalIssues) * 100).toFixed(1)}%
                  </div>
                  <span className="text-sm text-muted-foreground">Resolution Rate</span>
                </div>
                {data.issueStats.avgTimeToClose && (
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600">{formatDuration(data.issueStats.avgTimeToClose)}</div>
                    <span className="text-sm text-muted-foreground">Avg Time to Close</span>
                  </div>
                )}
              </div>
            </Card>
            
            <Card className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">PR Stats</h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {((data.pullRequestStats.mergedPullRequests / data.pullRequestStats.totalPullRequests) * 100).toFixed(1)}%
                  </div>
                  <span className="text-sm text-muted-foreground">Merge Rate</span>
                </div>
                {data.pullRequestStats.avgTimeToMerge && (
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600">{formatDuration(data.pullRequestStats.avgTimeToMerge)}</div>
                    <span className="text-sm text-muted-foreground">Avg Time to Merge</span>
                  </div>
                )}
              </div>
            </Card>
          </div>
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
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <p className="text-2xl font-bold text-foreground">{value}</p>
        </div>
      </div>
    </Card>
  );
};

export default RepositoryInsights;