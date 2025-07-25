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

// Mock data for development - replace with actual API calls
const mockSystemAnalytics = {
  totalViews: 45230,
  totalUsers: 1240,
  avgResponseTime: 185.5,
  totalRepositories: 567,
  viewsTrend: [
    { date: '2024-01-01', value: 1200 },
    { date: '2024-01-02', value: 1450 },
    { date: '2024-01-03', value: 1380 },
    { date: '2024-01-04', value: 1620 },
    { date: '2024-01-05', value: 1580 },
  ],
  usersTrend: [
    { date: '2024-01-01', value: 1180 },
    { date: '2024-01-02', value: 1195 },
    { date: '2024-01-03', value: 1210 },
    { date: '2024-01-04', value: 1225 },
    { date: '2024-01-05', value: 1240 },
  ],
  responseTrend: [
    { date: '2024-01-01', value: 195.2 },
    { date: '2024-01-02', value: 178.8 },
    { date: '2024-01-03', value: 189.4 },
    { date: '2024-01-04', value: 172.1 },
    { date: '2024-01-05', value: 185.5 },
  ],
  repositoriesTrend: [
    { date: '2024-01-01', value: 540 },
    { date: '2024-01-02', value: 548 },
    { date: '2024-01-03', value: 555 },
    { date: '2024-01-04', value: 561 },
    { date: '2024-01-05', value: 567 },
  ],
};

const mockSystemHealth = {
  cpuUsage: 45.2,
  memoryUsage: 67.8,
  diskUsage: 34.1,
  uptime: 99.95,
  errorRate: 0.12,
  activeConnections: 156,
};

export default function AdminAnalyticsPage() {
  const [data, setData] = useState<typeof mockSystemAnalytics | null>(null);
  const [healthData, setHealthData] = useState<typeof mockSystemHealth | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<'daily' | 'weekly' | 'monthly' | 'yearly'>('daily');
  const [isExporting, setIsExporting] = useState(false);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Replace with actual API calls
        // const [analyticsResponse, healthResponse] = await Promise.all([
        //   fetch(`/api/v1/admin/analytics/platform?period=${timeRange}`),
        //   fetch('/api/v1/admin/analytics/system-health')
        // ]);
        
        // if (!analyticsResponse.ok || !healthResponse.ok) {
        //   throw new Error('Failed to fetch analytics data');
        // }
        
        // const analyticsData = await analyticsResponse.json();
        // const healthData = await healthResponse.json();
        
        // For now, use mock data with a delay to simulate loading
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        setData(mockSystemAnalytics);
        setHealthData(mockSystemHealth);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
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
      
      // Replace with actual API call
      // const response = await fetch(`/api/v1/admin/analytics/export?format=${format}&period=${timeRange}`);
      // if (!response.ok) {
      //   throw new Error('Failed to export data');
      // }
      
      // For demo purposes, just show a success message
      await new Promise(resolve => setTimeout(resolve, 2000));
      alert(`Analytics data exported as ${format.toUpperCase()}`);
    } catch {
      alert('Failed to export data');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8 space-y-8">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Admin Analytics</h1>
          <p className="text-gray-600 mt-2">System-wide analytics and performance monitoring</p>
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
            <h3 className="text-lg font-medium text-gray-900 mb-4">System Performance</h3>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Error Rate</span>
                <span className={`font-medium ${healthData.errorRate > 1 ? 'text-red-600' : 'text-green-600'}`}>
                  {healthData.errorRate}%
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Active Connections</span>
                <span className="font-medium">{healthData.activeConnections}</span>
              </div>
            </div>
          </Card>

          <Card className="p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h3>
            <div className="space-y-3">
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => alert('Performance report generated')}
              >
                Generate Performance Report
              </Button>
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => alert('System health check initiated')}
              >
                Run System Health Check
              </Button>
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={() => alert('Cleanup process started')}
              >
                Cleanup Old Analytics Data
              </Button>
            </div>
          </Card>
        </div>
      )}

      {error && (
        <Card className="p-6">
          <div className="text-center">
            <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">Error Loading Analytics</h3>
            <p className="text-gray-500">{error}</p>
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
          <p className="text-sm font-medium text-gray-500">{label}</p>
          <p className={`text-2xl font-bold ${statusColors[status]}`}>{value}</p>
        </div>
      </div>
    </Card>
  );
};