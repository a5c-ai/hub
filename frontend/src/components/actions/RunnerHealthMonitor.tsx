'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';

interface RunnerHealthStats {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'busy';
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  current_job?: {
    id: string;
    workflow_name: string;
    started_at: string;
  };
  queue_length: number;
  last_heartbeat: string;
}

interface RunnerHealthMonitorProps {
  runnerId: string;
  onError?: (error: string) => void;
}

export default function RunnerHealthMonitor({ 
  runnerId, 
  onError 
}: RunnerHealthMonitorProps) {
  const [stats, setStats] = useState<RunnerHealthStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchStats = async () => {
    try {
      const response = await fetch(`/api/v1/runners/${runnerId}/health`);
      if (response.ok) {
        const data = await response.json();
        setStats(data);
        setError(null);
      } else {
        throw new Error('Failed to fetch runner stats');
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMsg);
      onError?.(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    
    let interval: NodeJS.Timeout | null = null;
    if (autoRefresh) {
      interval = setInterval(fetchStats, 10000); // Refresh every 10 seconds
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [runnerId, autoRefresh]);

  const getUsageColor = (usage: number) => {
    if (usage < 50) return 'text-green-600';
    if (usage < 80) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getUsageBarColor = (usage: number) => {
    if (usage < 50) return 'bg-green-500';
    if (usage < 80) return 'bg-yellow-500';
    return 'bg-red-500';
  };

  const formatUptime = (lastHeartbeat: string) => {
    const diff = Date.now() - new Date(lastHeartbeat).getTime();
    const minutes = Math.floor(diff / (1000 * 60));
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ${hours % 24}h`;
    if (hours > 0) return `${hours}h ${minutes % 60}m`;
    return `${minutes}m`;
  };

  if (loading) {
    return (
      <Card className="p-4">
        <div className="animate-pulse space-y-4">
          <div className="h-4 bg-muted rounded w-1/3"></div>
          <div className="space-y-2">
            <div className="h-2 bg-muted rounded"></div>
            <div className="h-2 bg-muted rounded"></div>
            <div className="h-2 bg-muted rounded"></div>
          </div>
        </div>
      </Card>
    );
  }

  if (error || !stats) {
    return (
      <Card className="p-4">
        <div className="text-center">
          <p className="text-red-600 mb-2">{error || 'Failed to load runner stats'}</p>
          <Button variant="outline" size="sm" onClick={fetchStats}>
            Retry
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <Card className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h4 className="font-medium">Health Monitor</h4>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            {autoRefresh ? '⏸️' : '▶️'} {autoRefresh ? 'Auto' : 'Manual'}
          </Button>
          <Button variant="outline" size="sm" onClick={fetchStats}>
            🔄 Refresh
          </Button>
        </div>
      </div>

      {/* Status */}
      <div className="grid grid-cols-2 gap-4 mb-4">
        <div>
          <span className="text-sm text-muted-foreground">Status</span>
          <div className="flex items-center gap-2">
            <Badge variant={
              stats.status === 'online' ? 'default' : 
              stats.status === 'busy' ? 'outline' : 'secondary'
            }>
              {stats.status}
            </Badge>
            {stats.status === 'busy' && stats.current_job && (
              <span className="text-xs text-muted-foreground">
                Running: {stats.current_job.workflow_name}
              </span>
            )}
          </div>
        </div>
        <div>
          <span className="text-sm text-muted-foreground">Queue</span>
          <div className="text-lg font-semibold">
            {stats.queue_length} jobs
          </div>
        </div>
      </div>

      {/* Resource Usage */}
      <div className="space-y-3">
        <div>
          <div className="flex justify-between text-sm mb-1">
            <span>CPU Usage</span>
            <span className={getUsageColor(stats.cpu_usage)}>
              {stats.cpu_usage.toFixed(1)}%
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full ${getUsageBarColor(stats.cpu_usage)}`}
              style={{ width: `${Math.min(stats.cpu_usage, 100)}%` }}
            ></div>
          </div>
        </div>

        <div>
          <div className="flex justify-between text-sm mb-1">
            <span>Memory Usage</span>
            <span className={getUsageColor(stats.memory_usage)}>
              {stats.memory_usage.toFixed(1)}%
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full ${getUsageBarColor(stats.memory_usage)}`}
              style={{ width: `${Math.min(stats.memory_usage, 100)}%` }}
            ></div>
          </div>
        </div>

        <div>
          <div className="flex justify-between text-sm mb-1">
            <span>Disk Usage</span>
            <span className={getUsageColor(stats.disk_usage)}>
              {stats.disk_usage.toFixed(1)}%
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`h-2 rounded-full ${getUsageBarColor(stats.disk_usage)}`}
              style={{ width: `${Math.min(stats.disk_usage, 100)}%` }}
            ></div>
          </div>
        </div>
      </div>

      {/* Current Job */}
      {stats.current_job && (
        <div className="mt-4 p-3 bg-blue-50 rounded-lg">
          <h5 className="font-medium text-sm mb-1">Current Job</h5>
          <p className="text-sm text-muted-foreground">
            {stats.current_job.workflow_name}
          </p>
          <p className="text-xs text-muted-foreground">
            Started {formatUptime(stats.current_job.started_at)} ago
          </p>
        </div>
      )}

      {/* Last Heartbeat */}
      <div className="mt-4 text-xs text-muted-foreground">
        Last heartbeat: {formatUptime(stats.last_heartbeat)} ago
      </div>
    </Card>
  );
}