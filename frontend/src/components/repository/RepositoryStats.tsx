'use client';

import React from 'react';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { RepositoryStatistics } from '@/types';
import {
  CodeBracketIcon,
  DocumentTextIcon,
  TagIcon,
  ClockIcon,
  UserGroupIcon,
  ScaleIcon,
} from '@heroicons/react/24/outline';

interface RepositoryStatsProps {
  statistics: RepositoryStatistics;
  compact?: boolean;
}

const RepositoryStats: React.FC<RepositoryStatsProps> = ({
  statistics,
  compact = false,
}) => {
  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString?: string): string => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const stats = [
    {
      icon: CodeBracketIcon,
      label: 'Commits',
      value: formatNumber(statistics.commit_count),
      color: 'blue',
      description: 'Total commits',
    },
    {
      icon: DocumentTextIcon,
      label: 'Branches',
      value: formatNumber(statistics.branch_count),
      color: 'green',
      description: 'Active branches',
    },
    {
      icon: TagIcon,
      label: 'Tags',
      value: formatNumber(statistics.tag_count),
      color: 'purple',
      description: 'Release tags',
    },
    {
      icon: UserGroupIcon,
      label: 'Contributors',
      value: formatNumber(statistics.contributors),
      color: 'yellow',
      description: 'Unique contributors',
    },
    {
      icon: ScaleIcon,
      label: 'Size',
      value: formatBytes(statistics.size_bytes),
      color: 'red',
      description: 'Repository size',
    },
    {
      icon: CodeBracketIcon,
      label: 'Languages',
      value: formatNumber(statistics.language_count),
      color: 'indigo',
      description: 'Programming languages',
    },
  ];

  if (compact) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <div key={stat.label} className="text-center p-3 bg-background border border-border rounded-lg">
              <Icon className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
              <div className="text-lg font-bold text-foreground">{stat.value}</div>
              <div className="text-xs text-muted-foreground">{stat.label}</div>
            </div>
          );
        })}
      </div>
    );
  }

  return (
    <Card>
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-foreground">Repository Statistics</h3>
          {statistics.last_activity && (
            <Badge variant="outline">
              Last activity: {formatDate(statistics.last_activity)}
            </Badge>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            const colorClasses = {
              blue: 'text-blue-600 bg-blue-50 border-blue-200',
              green: 'text-green-600 bg-green-50 border-green-200',
              purple: 'text-purple-600 bg-purple-50 border-purple-200',
              yellow: 'text-yellow-600 bg-yellow-50 border-yellow-200',
              red: 'text-red-600 bg-red-50 border-red-200',
              indigo: 'text-indigo-600 bg-indigo-50 border-indigo-200',
            };

            return (
              <div key={stat.label} className="p-4 rounded-lg border border-border hover:bg-muted/50 transition-colors">
                <div className="flex items-center space-x-3">
                  <div className={`p-2 rounded-lg border ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
                    <Icon className="h-5 w-5" />
                  </div>
                  <div className="flex-1">
                    <div className="text-2xl font-bold text-foreground">{stat.value}</div>
                    <div className="text-sm font-medium text-foreground">{stat.label}</div>
                    <div className="text-xs text-muted-foreground">{stat.description}</div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Additional Details */}
        <div className="mt-6 pt-6 border-t border-border">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Primary Language:</span>
                <span className="font-medium text-foreground">
                  {statistics.primary_language || 'Not detected'}
                </span>
              </div>
            </div>
            <div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Last Updated:</span>
                <span className="font-medium text-foreground">
                  {formatDate(statistics.updated_at)}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
};

export default RepositoryStats;