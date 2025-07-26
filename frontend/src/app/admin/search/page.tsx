'use client';

import { useState, useEffect } from 'react';
import { Card, Button, Input } from '@/components/ui';
import { 
  MagnifyingGlassIcon, 
  Cog6ToothIcon,
  ChartBarIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  ClockIcon,
  ArrowPathIcon
} from '@heroicons/react/24/outline';
import { apiClient } from '@/lib/api';

interface SearchStats {
  total_documents: number;
  indices: {
    [key: string]: {
      documents: number;
      size_mb: number;
      last_updated: string;
    };
  };
  search_requests_today: number;
  avg_response_time_ms: number;
  index_health: 'green' | 'yellow' | 'red';
}

interface SearchConfig {
  elasticsearch_enabled: boolean;
  auto_index: boolean;
  max_results_per_page: number;
  search_timeout_ms: number;
  highlight_enabled: boolean;
  fuzzy_search_enabled: boolean;
  analytics_enabled: boolean;
}

export default function AdminSearchPage() {
  const [stats, setStats] = useState<SearchStats | null>(null);
  const [config, setConfig] = useState<SearchConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [reindexing, setReindexing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    loadSearchData();
  }, []);

  const loadSearchData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const [statsRes, configRes] = await Promise.all([
        apiClient.get<SearchStats>('/admin/search/stats'),
        apiClient.get<SearchConfig>('/admin/search/config')
      ]);
      if (statsRes.success && statsRes.data) {
        setStats(statsRes.data);
      }
      if (configRes.success && configRes.data) {
        setConfig(configRes.data);
      }
    } catch (err) {
      console.error('Error loading search data:', err);
      setError('Failed to load search configuration');
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async () => {
    if (!config) return;

    setSaving(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const res = await apiClient.put('/admin/search/config', config);
      if (res.success) {
        setSuccessMessage('Search configuration saved successfully');
        await loadSearchData();
        setTimeout(() => setSuccessMessage(null), 3000);
      }
    } catch (err) {
      console.error('Error saving search config:', err);
      setError('Failed to save search configuration');
    } finally {
      setSaving(false);
    }
  };

  const reindexAll = async () => {
    setReindexing(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const res = await apiClient.post('/admin/search/reindex');
      if (res.success) {
        setSuccessMessage('Reindexing completed successfully');
        await loadSearchData();
        setTimeout(() => setSuccessMessage(null), 3000);
      } else {
        setError('Reindexing failed');
      }
    } catch (err) {
      console.error('Error reindexing:', err);
      setError('Reindexing failed');
    } finally {
      setReindexing(false);
    }
  };

  const updateConfig = (key: keyof SearchConfig, value: string | number | boolean) => {
    if (!config) return;
    setConfig({ ...config, [key]: value });
  };

  const getHealthIcon = (health: string) => {
    switch (health) {
      case 'green':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />;
      case 'yellow':
        return <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />;
      case 'red':
        return <ExclamationTriangleIcon className="h-5 w-5 text-red-500" />;
      default:
        return <ClockIcon className="h-5 w-5 text-gray-500" />;
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading search configuration...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Cog6ToothIcon className="h-8 w-8 text-primary" />
            <h1 className="text-3xl font-bold">Search Configuration</h1>
          </div>
          <p className="text-muted-foreground">
            Manage search settings, monitoring, and index operations
          </p>
        </div>

        {/* Status Messages */}
        {error && (
          <Card className="p-4 mb-6 border-destructive bg-red-50">
            <p className="text-destructive">{error}</p>
          </Card>
        )}

        {successMessage && (
          <Card className="p-4 mb-6 border-green-200 bg-green-50">
            <p className="text-green-800">{successMessage}</p>
          </Card>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          {/* Search Statistics */}
          <div className="lg:col-span-2">
            <Card className="p-6">
              <div className="flex items-center gap-3 mb-4">
                <ChartBarIcon className="h-6 w-6 text-primary" />
                <h2 className="text-xl font-semibold">Search Statistics</h2>
                {stats && getHealthIcon(stats.index_health)}
              </div>

              {stats && (
                <div className="space-y-6">
                  {/* Overview Stats */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div className="text-center">
                      <div className="text-2xl font-bold text-primary">{stats.total_documents.toLocaleString()}</div>
                      <div className="text-sm text-muted-foreground">Total Documents</div>
                    </div>
                    <div className="text-center">
                      <div className="text-2xl font-bold text-primary">{stats.search_requests_today.toLocaleString()}</div>
                      <div className="text-sm text-muted-foreground">Searches Today</div>
                    </div>
                    <div className="text-center">
                      <div className="text-2xl font-bold text-primary">{stats.avg_response_time_ms}ms</div>
                      <div className="text-sm text-muted-foreground">Avg Response Time</div>
                    </div>
                    <div className="text-center">
                      <div className="flex items-center justify-center gap-1">
                        <div className="text-2xl font-bold text-primary capitalize">{stats.index_health}</div>
                        {getHealthIcon(stats.index_health)}
                      </div>
                      <div className="text-sm text-muted-foreground">Index Health</div>
                    </div>
                  </div>

                  {/* Index Details */}
                  <div>
                    <h3 className="font-semibold mb-3">Index Details</h3>
                    <div className="space-y-3">
                      {Object.entries(stats.indices).map(([indexName, indexData]) => (
                        <div key={indexName} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                          <div>
                            <div className="font-medium capitalize">{indexName}</div>
                            <div className="text-sm text-muted-foreground">
                              Last updated: {new Date(indexData.last_updated).toLocaleString()}
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="font-medium">{indexData.documents.toLocaleString()} docs</div>
                            <div className="text-sm text-muted-foreground">{indexData.size_mb} MB</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </Card>
          </div>

          {/* Quick Actions */}
          <div>
            <Card className="p-6">
              <div className="flex items-center gap-3 mb-4">
                <ArrowPathIcon className="h-6 w-6 text-primary" />
                <h2 className="text-xl font-semibold">Quick Actions</h2>
              </div>

              <div className="space-y-3">
                <Button
                  onClick={reindexAll}
                  disabled={reindexing}
                  variant="outline"
                  className="w-full justify-start"
                >
                  {reindexing ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary mr-2"></div>
                      Reindexing...
                    </>
                  ) : (
                    <>
                      <ArrowPathIcon className="h-4 w-4 mr-2" />
                      Reindex All Data
                    </>
                  )}
                </Button>

                <Button
                  onClick={() => window.location.href = '/admin/search/analytics'}
                  variant="outline"
                  className="w-full justify-start"
                >
                  <ChartBarIcon className="h-4 w-4 mr-2" />
                  View Analytics
                </Button>

                <Button
                  onClick={() => window.location.href = '/search'}
                  variant="outline"
                  className="w-full justify-start"
                >
                  <MagnifyingGlassIcon className="h-4 w-4 mr-2" />
                  Test Search
                </Button>
              </div>
            </Card>
          </div>
        </div>

        {/* Configuration Settings */}
        <Card className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <Cog6ToothIcon className="h-6 w-6 text-primary" />
              <h2 className="text-xl font-semibold">Search Configuration</h2>
            </div>
            <Button onClick={saveConfig} disabled={saving || !config}>
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>

          {config && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Basic Settings */}
              <div className="space-y-4">
                <h3 className="font-semibold">Basic Settings</h3>
                
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Elasticsearch Enabled</label>
                  <input
                    type="checkbox"
                    checked={config.elasticsearch_enabled}
                    onChange={(e) => updateConfig('elasticsearch_enabled', e.target.checked)}
                    className="rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Auto Index Updates</label>
                  <input
                    type="checkbox"
                    checked={config.auto_index}
                    onChange={(e) => updateConfig('auto_index', e.target.checked)}
                    className="rounded"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Max Results Per Page</label>
                  <Input
                    type="number"
                    value={config.max_results_per_page}
                    onChange={(e) => updateConfig('max_results_per_page', parseInt(e.target.value))}
                    min="10"
                    max="100"
                    className="w-full"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Search Timeout (ms)</label>
                  <Input
                    type="number"
                    value={config.search_timeout_ms}
                    onChange={(e) => updateConfig('search_timeout_ms', parseInt(e.target.value))}
                    min="1000"
                    max="30000"
                    className="w-full"
                  />
                </div>
              </div>

              {/* Advanced Settings */}
              <div className="space-y-4">
                <h3 className="font-semibold">Advanced Settings</h3>

                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Syntax Highlighting</label>
                  <input
                    type="checkbox"
                    checked={config.highlight_enabled}
                    onChange={(e) => updateConfig('highlight_enabled', e.target.checked)}
                    className="rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Fuzzy Search</label>
                  <input
                    type="checkbox"
                    checked={config.fuzzy_search_enabled}
                    onChange={(e) => updateConfig('fuzzy_search_enabled', e.target.checked)}
                    className="rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium">Search Analytics</label>
                  <input
                    type="checkbox"
                    checked={config.analytics_enabled}
                    onChange={(e) => updateConfig('analytics_enabled', e.target.checked)}
                    className="rounded"
                  />
                </div>
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
