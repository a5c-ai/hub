'use client';

import React, { useState, useEffect } from 'react';
import AnalyticsDashboard from '@/components/analytics/AnalyticsDashboard';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { 
  ExclamationTriangleIcon, 
  ArrowDownTrayIcon,
  Cog6ToothIcon,
  ServerIcon,
  CpuChipIcon,
  ClockIcon
} from '@heroicons/react/24/outline';

// API response types
interface SystemInsights {
  total_views: number;
  total_users: number;
  avg_response_time: number;
  total_repositories: number;
  trends: {
    views: Array<{ date: string; value: number }>;
    users: Array<{ date: string; value: number }>;
    response_time: Array<{ date: string; value: number }>;
    repositories: Array<{ date: string; value: number }>;
  };
}

interface TransformedAnalyticsData {
  totalViews: number;
  totalUsers: number;
  avgResponseTime: number;
  totalRepositories: number;
  viewsTrend: Array<{ date: string; value: number }>;
  usersTrend: Array<{ date: string; value: number }>;
  responseTrend: Array<{ date: string; value: number }>;
  repositoriesTrend: Array<{ date: string; value: number }>;
}

interface PerformanceMetrics {
  cpu_usage_percent: number;
  memory_usage_percent: number;
  disk_usage_percent: number;
  uptime_percent: number;
  error_rate_percent: number;
  active_connections: number;
}

interface HealthData {
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  uptime: number;
  errorRate: number;
  activeConnections: number;
}

export default function AdminAnalyticsPage() {
  const [data, setData] = useState<TransformedAnalyticsData | null>(null);
  const [healthData, setHealthData] = useState<HealthData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<'daily' | 'weekly' | 'monthly' | 'yearly'>('daily');
  const [isExporting, setIsExporting] = useState(false);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        setIsLoading(true);
        setError(null);

        const [analyticsResponse, performanceResponse] = await Promise.all([
          fetch(`/api/v1/admin/analytics/platform?period=${timeRange}`),
          fetch('/api/v1/admin/analytics/performance')
        ]);
        
        if (!analyticsResponse.ok) {
          throw new Error(`Failed to fetch analytics data: ${analyticsResponse.statusText}`);
        }
        
        const analyticsData: SystemInsights = await analyticsResponse.json();
        
        // Convert backend format to frontend format
        const transformedData = {
          totalViews: analyticsData.total_views,
          totalUsers: analyticsData.total_users,
          avgResponseTime: analyticsData.avg_response_time,
          totalRepositories: analyticsData.total_repositories,
          viewsTrend: analyticsData.trends?.views || [],
          usersTrend: analyticsData.trends?.users || [],
          responseTrend: analyticsData.trends?.response_time || [],
          repositoriesTrend: analyticsData.trends?.repositories || [],
        };
        
        setData(transformedData);
        
        // Handle performance data if available
        if (performanceResponse.ok) {
          const performanceData: PerformanceMetrics = await performanceResponse.json();
          setHealthData({
            cpuUsage: performanceData.cpu_usage_percent,
            memoryUsage: performanceData.memory_usage_percent,
            diskUsage: performanceData.disk_usage_percent,
            uptime: performanceData.uptime_percent,
            errorRate: performanceData.error_rate_percent,
            activeConnections: performanceData.active_connections,
          });
        } else {
          // If performance endpoint fails, set basic health data
          setHealthData({
            cpuUsage: 0,
            memoryUsage: 0,
            diskUsage: 0,
            uptime: 100,
            errorRate: 0,
            activeConnections: 0,
          });
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred while fetching analytics data');
        console.error('Analytics fetch error:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchAnalytics();
  }, [timeRange]);

  const handleTimeRangeChange = (newTimeRange: 'daily' | 'weekly' | 'monthly' | 'yearly') => {
    setTimeRange(newTimeRange);
  };

  const handleExport = async (format: 'json' | 'csv' | 'xlsx') => {
    try {
      setIsExporting(true);
      
      const response = await fetch(`/api/v1/admin/analytics/export?format=${format}&period=${timeRange}`);
      if (!response.ok) {
        throw new Error(`Failed to export data: ${response.statusText}`);
      }
      
      // Download the exported file
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `analytics-${timeRange}.${format}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export data');
      console.error('Export error:', err);
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8 space-y-8">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Admin Analytics</h1>
          <p className="text-muted-foreground mt-2">System-wide analytics and performance monitoring</p>
        </div>
        
        <div className="flex space-x-3">
          <Button
            variant="secondary"
            onClick={() => handleExport('csv')}
            disabled={isExporting}
            className="flex items-center space-x-2"
          >
            <ArrowDownTrayIcon className="h-4 w-4" />
            <span>{isExporting ? 'Exporting...' : 'Export CSV'}</span>
          </Button>
          
          <Button
            variant="secondary"
            onClick={() => handleExport('json')}
            disabled={isExporting}
            className="flex items-center space-x-2"
          >
            <ArrowDownTrayIcon className="h-4 w-4" />
            <span>Export JSON</span>
          </Button>
        </div>
      </div>

      {/* System Health Cards */}
      {healthData && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <SystemHealthCard
            icon={CpuChipIcon}
            label="CPU Usage"
            value={`${healthData.cpuUsage}%`}
            status={healthData.cpuUsage > 80 ? 'critical' : healthData.cpuUsage > 60 ? 'warning' : 'good'}
          />
          <SystemHealthCard
            icon={ServerIcon}
            label="Memory Usage"
            value={`${healthData.memoryUsage}%`}
            status={healthData.memoryUsage > 85 ? 'critical' : healthData.memoryUsage > 70 ? 'warning' : 'good'}
          />
          <SystemHealthCard
            icon={Cog6ToothIcon}
            label="Disk Usage"
            value={`${healthData.diskUsage}%`}
            status={healthData.diskUsage > 90 ? 'critical' : healthData.diskUsage > 75 ? 'warning' : 'good'}
          />
          <SystemHealthCard
            icon={ClockIcon}
            label="Uptime"
            value={`${healthData.uptime}%`}
            status={healthData.uptime < 99 ? 'critical' : healthData.uptime < 99.5 ? 'warning' : 'good'}
          />
        </div>
      )}

      {/* Main Analytics Dashboard */}
      <AnalyticsDashboard
        data={data}
        isLoading={isLoading}
        title="Platform Analytics"
        timeRange={timeRange}
        onTimeRangeChange={handleTimeRangeChange}
      />

      {/* Additional Admin Metrics */}
      {healthData && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="p-6">
            <h3 className="text-lg font-medium text-foreground mb-4">System Performance</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold text-red-600">
                  {healthData ? `${healthData.errorRate.toFixed(1)}%` : '0%'}
                </div>
                <span className="text-sm text-muted-foreground">Error Rate</span>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-green-600">
                  {healthData ? healthData.activeConnections.toLocaleString() : '0'}
                </div>
                <span className="text-sm text-muted-foreground">Active Connections</span>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <h3 className="text-lg font-medium text-foreground mb-4">Quick Actions</h3>
            <div className="space-y-3">
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => handleExport('json')}
                disabled={isExporting}
              >
                Generate Performance Report
              </Button>
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => window.location.reload()}
              >
                Refresh System Health
              </Button>
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => handleExport('csv')}
                disabled={isExporting}
              >
                Export All Analytics Data
              </Button>
            </div>
          </Card>
        </div>
      )}

      {error && (
        <Card className="p-6">
          <div className="text-center">
            <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-foreground mb-2">Error Loading Analytics</h3>
            <p className="text-muted-foreground">{error}</p>
          </div>
        </Card>
      )}
    </div>
  );
}

// System Health Card Component
interface SystemHealthCardProps {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
  status: 'good' | 'warning' | 'critical';
}

const SystemHealthCard: React.FC<SystemHealthCardProps> = ({ icon: Icon, label, value, status }) => {
  const statusColors = {
    good: 'text-green-600',
    warning: 'text-yellow-600',
    critical: 'text-red-600',
  };

  const statusBgColors = {
    good: 'bg-green-50',
    warning: 'bg-yellow-50',
    critical: 'bg-red-50',
  };

  return (
    <Card className={`p-6 ${statusBgColors[status]}`}>
      <div className="flex items-center">
        <div className="flex-shrink-0">
          <Icon className={`h-8 w-8 ${statusColors[status]}`} />
        </div>
        <div className="ml-4 flex-1">
          <p className="text-sm font-medium text-muted-foreground">{label}</p>
          <p className={`text-2xl font-bold ${statusColors[status]}`}>{value}</p>
        </div>
      </div>
    </Card>
  );
};