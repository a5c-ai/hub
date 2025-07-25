'use client';

import React, { useState, useEffect } from 'react';
import { Card } from '../ui/Card';
import { Button } from '../ui/Button';
import { ChartBarIcon, ClockIcon, EyeIcon, UserGroupIcon } from '@heroicons/react/24/outline';

interface AnalyticsData {
  totalViews: number;
  totalUsers: number;
  avgResponseTime: number;
  totalRepositories: number;
  viewsTrend: Array<{ date: string; value: number }>;
  usersTrend: Array<{ date: string; value: number }>;
  responseTrend: Array<{ date: string; value: number }>;
  repositoriesTrend: Array<{ date: string; value: number }>;
}

interface AnalyticsDashboardProps {
  data?: AnalyticsData | null;
  isLoading?: boolean;
  title?: string;
  timeRange?: 'daily' | 'weekly' | 'monthly' | 'yearly';
  onTimeRangeChange?: (range: 'daily' | 'weekly' | 'monthly' | 'yearly') => void;
}

export const AnalyticsDashboard: React.FC<AnalyticsDashboardProps> = ({
  data,
  isLoading = false,
  title = 'Analytics Dashboard',
  timeRange = 'daily',
  onTimeRangeChange,
}) => {
  const [selectedTimeRange, setSelectedTimeRange] = useState(timeRange);

  const handleTimeRangeChange = (range: 'daily' | 'weekly' | 'monthly' | 'yearly') => {
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

  const formatResponseTime = (ms: number): string => {
    if (ms >= 1000) {
      return (ms / 1000).toFixed(1) + 's';
    }
    return ms.toFixed(0) + 'ms';
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
          <div className="flex space-x-2">
            {(['daily', 'weekly', 'monthly', 'yearly'] as const).map((range) => (
              <div key={range} className="h-8 w-16 bg-muted rounded animate-pulse"></div>
            ))}
          </div>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="space-y-4">
              <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
              <div className="h-8 bg-muted rounded w-1/2 mb-4"></div>
              <div className="h-16 bg-muted rounded"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
        <div className="flex space-x-2">
          {(['daily', 'weekly', 'monthly', 'yearly'] as const).map((range) => (
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

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Total Views */}
        <Card className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <EyeIcon className="h-8 w-8 text-blue-600" />
            </div>
            <div className="ml-4 flex-1">
              <p className="text-sm font-medium text-gray-500">Total Views</p>
              <p className="text-2xl font-bold text-gray-900">
                {data ? formatNumber(data.totalViews) : '0'}
              </p>
            </div>
          </div>
          <div className="mt-4">
            <SimpleChart 
              data={data?.viewsTrend || []} 
              color="blue" 
              height={40}
            />
          </div>
        </Card>

        {/* Total Users */}
        <Card className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <UserGroupIcon className="h-8 w-8 text-green-600" />
            </div>
            <div className="ml-4 flex-1">
              <p className="text-sm font-medium text-gray-500">Active Users</p>
              <p className="text-2xl font-bold text-gray-900">
                {data ? formatNumber(data.totalUsers) : '0'}
              </p>
            </div>
          </div>
          <div className="mt-4">
            <SimpleChart 
              data={data?.usersTrend || []} 
              color="green" 
              height={40}
            />
          </div>
        </Card>

        {/* Avg Response Time */}
        <Card className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <ClockIcon className="h-8 w-8 text-yellow-600" />
            </div>
            <div className="ml-4 flex-1">
              <p className="text-sm font-medium text-gray-500">Avg Response</p>
              <p className="text-2xl font-bold text-gray-900">
                {data ? formatResponseTime(data.avgResponseTime) : '0ms'}
              </p>
            </div>
          </div>
          <div className="mt-4">
            <SimpleChart 
              data={data?.responseTrend || []} 
              color="yellow" 
              height={40}
            />
          </div>
        </Card>

        {/* Total Repositories */}
        <Card className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <ChartBarIcon className="h-8 w-8 text-purple-600" />
            </div>
            <div className="ml-4 flex-1">
              <p className="text-sm font-medium text-gray-500">Repositories</p>
              <p className="text-2xl font-bold text-gray-900">
                {data ? formatNumber(data.totalRepositories) : '0'}
              </p>
            </div>
          </div>
          <div className="mt-4">
            <SimpleChart 
              data={data?.repositoriesTrend || []} 
              color="purple" 
              height={40}
            />
          </div>
        </Card>
      </div>

      {/* Detailed Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Activity Overview</h3>
          <div className="h-64 flex items-center justify-center text-gray-500">
            <p>Detailed chart component would go here</p>
          </div>
        </Card>

        <Card className="p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Performance Metrics</h3>
          <div className="h-64 flex items-center justify-center text-gray-500">
            <p>Performance chart component would go here</p>
          </div>
        </Card>
      </div>
    </div>
  );
};

// Simple chart component for trend visualization
interface SimpleChartProps {
  data: Array<{ date: string; value: number }>;
  color: 'blue' | 'green' | 'yellow' | 'purple';
  height: number;
}

const SimpleChart: React.FC<SimpleChartProps> = ({ data, color, height }) => {
  if (!data || data.length === 0) {
    return <div style={{ height }} className="bg-muted rounded"></div>;
  }

  const maxValue = Math.max(...data.map(d => d.value));
  const minValue = Math.min(...data.map(d => d.value));
  const range = maxValue - minValue || 1;

  const colorClasses = {
    blue: 'stroke-blue-500',
    green: 'stroke-green-500',
    yellow: 'stroke-yellow-500',
    purple: 'stroke-purple-500',
  };

  const points = data.map((point, index) => {
    const x = (index / (data.length - 1)) * 100;
    const y = ((maxValue - point.value) / range) * 100;
    return `${x},${y}`;
  }).join(' ');

  return (
    <div style={{ height }} className="w-full">
      <svg
        width="100%"
        height="100%"
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
        className="overflow-visible"
      >
        <polyline
          fill="none"
          strokeWidth="2"
          className={`${colorClasses[color]} opacity-80`}
          points={points}
        />
      </svg>
    </div>
  );
};

export default AnalyticsDashboard;