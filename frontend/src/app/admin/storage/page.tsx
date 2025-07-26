'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface StorageConfig {
  backend: string;
  max_size_mb: number;
  retention_days: number;
  health: string;
}

interface StorageUsage {
  total_artifacts: number;
  total_size_bytes: number;
  total_size_mb: number;
  expired_count: number;
  retention_days: number;
  max_size_mb: number;
}

interface StorageHealth {
  status: string;
  checks: {
    upload?: { status: string; error?: string };
    download?: { status: string; error?: string };
  };
}

export default function AdminStoragePage() {
  const [config, setConfig] = useState<StorageConfig | null>(null);
  const [usage, setUsage] = useState<StorageUsage | null>(null);
  const [health, setHealth] = useState<StorageHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Configuration modal state
  const [showConfigModal, setShowConfigModal] = useState(false);
  const [configForm, setConfigForm] = useState({
    backend: 'filesystem',
    max_size_mb: 100,
    retention_days: 90,
  });
  const [saving, setSaving] = useState(false);

  // Retention policy modal state
  const [showRetentionModal, setShowRetentionModal] = useState(false);
  const [retentionDays, setRetentionDays] = useState(90);

  // Cleanup state
  const [cleaning, setCleaning] = useState(false);

  const fetchStorageData = async () => {
    try {
      setLoading(true);
      
      // Fetch all storage data in parallel
      const [configRes, usageRes, healthRes] = await Promise.all([
        fetch('/api/v1/admin/storage/config'),
        fetch('/api/v1/admin/storage/usage'),
        fetch('/api/v1/admin/storage/health'),
      ]);

      if (configRes.ok) {
        const configData = await configRes.json();
        setConfig(configData);
        setConfigForm({
          backend: configData.backend,
          max_size_mb: configData.max_size_mb,
          retention_days: configData.retention_days,
        });
        setRetentionDays(configData.retention_days);
      }

      if (usageRes.ok) {
        const usageData = await usageRes.json();
        setUsage(usageData);
      }

      if (healthRes.ok) {
        const healthData = await healthRes.json();
        setHealth(healthData);
      }

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch storage data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStorageData();
  }, []);

  const handleConfigSave = async () => {
    setSaving(true);
    try {
      const response = await fetch('/api/v1/admin/storage/config', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(configForm),
      });

      if (response.ok) {
        setShowConfigModal(false);
        await fetchStorageData();
      } else {
        const error = await response.json();
        setError(error.error || 'Failed to update configuration');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update configuration');
    } finally {
      setSaving(false);
    }
  };

  const handleRetentionSave = async () => {
    setSaving(true);
    try {
      const response = await fetch('/api/v1/admin/storage/retention', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ retention_days: retentionDays }),
      });

      if (response.ok) {
        setShowRetentionModal(false);
        await fetchStorageData();
      } else {
        const error = await response.json();
        setError(error.error || 'Failed to update retention policy');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update retention policy');
    } finally {
      setSaving(false);
    }
  };

  const handleManualCleanup = async () => {
    if (!confirm('Are you sure you want to run a manual cleanup? This will remove all expired artifacts and build logs.')) {
      return;
    }

    setCleaning(true);
    try {
      const response = await fetch('/api/v1/admin/storage/cleanup', {
        method: 'DELETE',
      });

      if (response.ok) {
        await fetchStorageData();
        alert('Storage cleanup completed successfully');
      } else {
        const error = await response.json();
        setError(error.error || 'Failed to run cleanup');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to run cleanup');
    } finally {
      setCleaning(false);
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
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-6">Storage Management</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="p-4">
              <div className="animate-pulse space-y-4">
                <div className="h-4 bg-muted rounded w-1/2"></div>
                <div className="h-8 bg-muted rounded w-3/4"></div>
              </div>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Storage Management</h1>
        <div className="flex gap-2">
          <Button 
            onClick={() => setShowConfigModal(true)}
            variant="outline"
          >
            Configure Storage
          </Button>
          <Button 
            onClick={handleManualCleanup}
            disabled={cleaning}
            variant="outline"
          >
            {cleaning ? 'Cleaning...' : 'Manual Cleanup'}
          </Button>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-700">{error}</p>
          <Button 
            size="sm" 
            variant="outline" 
            onClick={() => setError(null)}
            className="mt-2"
          >
            Dismiss
          </Button>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-6">
        {/* Storage Configuration */}
        <Card>
          <div className="p-4 border-b">
            <div className="flex items-center justify-between">
              <h3 className="font-medium">Storage Configuration</h3>
              {health && (
                <Badge 
                  variant={health.status === 'healthy' ? 'default' : 'destructive'}
                >
                  {health.status}
                </Badge>
              )}
            </div>
          </div>
          <div className="p-4 space-y-3">
            {config ? (
              <>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Backend:</span>
                  <Badge variant="outline">{config.backend}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Max Size:</span>
                  <span className="text-sm font-medium">{config.max_size_mb} MB</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Retention:</span>
                  <span className="text-sm font-medium">{config.retention_days} days</span>
                </div>
              </>
            ) : (
              <p className="text-sm text-muted-foreground">Loading configuration...</p>
            )}
          </div>
        </Card>

        {/* Storage Usage */}
        <Card>
          <div className="p-4 border-b">
            <h3 className="font-medium">Storage Usage</h3>
          </div>
          <div className="p-4 space-y-3">
            {usage ? (
              <>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Total Artifacts:</span>
                  <span className="text-sm font-medium">{usage.total_artifacts.toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Total Size:</span>
                  <span className="text-sm font-medium">{formatBytes(usage.total_size_bytes)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-muted-foreground">Expired Count:</span>
                  <span className="text-sm font-medium">{usage.expired_count.toLocaleString()}</span>
                </div>
                {usage.max_size_mb > 0 && (
                  <div className="mt-4">
                    <div className="flex justify-between mb-1">
                      <span className="text-sm text-muted-foreground">Quota Usage:</span>
                      <span className="text-sm font-medium">
                        {((usage.total_size_mb / usage.max_size_mb) * 100).toFixed(1)}%
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                        style={{ 
                          width: `${Math.min((usage.total_size_mb / usage.max_size_mb) * 100, 100)}%` 
                        }}
                      ></div>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <p className="text-sm text-muted-foreground">Loading usage data...</p>
            )}
          </div>
        </Card>

        {/* Health Status */}
        <Card>
          <div className="p-4 border-b">
            <h3 className="font-medium">Backend Health</h3>
          </div>
          <div className="p-4 space-y-3">
            {health ? (
              <>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">Overall Status:</span>
                  <Badge 
                    variant={health.status === 'healthy' ? 'default' : 'destructive'}
                  >
                    {health.status}
                  </Badge>
                </div>
                {health.checks.upload && (
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">Upload:</span>
                    <Badge 
                      variant={health.checks.upload.status === 'pass' ? 'default' : 'destructive'}
                      className="text-xs"
                    >
                      {health.checks.upload.status}
                    </Badge>
                  </div>
                )}
                {health.checks.download && (
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">Download:</span>
                    <Badge 
                      variant={health.checks.download.status === 'pass' ? 'default' : 'destructive'}
                      className="text-xs"
                    >
                      {health.checks.download.status}
                    </Badge>
                  </div>
                )}
              </>
            ) : (
              <p className="text-sm text-muted-foreground">Loading health data...</p>
            )}
          </div>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <div className="p-4 border-b">
          <h3 className="font-medium">Quick Actions</h3>
        </div>
        <div className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Button 
              onClick={() => setShowRetentionModal(true)}
              variant="outline"
              className="h-20 flex flex-col items-center justify-center"
            >
              <svg className="w-6 h-6 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm">Retention Policy</span>
            </Button>
            
            <Button 
              onClick={handleManualCleanup}
              disabled={cleaning}
              variant="outline"
              className="h-20 flex flex-col items-center justify-center"
            >
              <svg className="w-6 h-6 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              <span className="text-sm">{cleaning ? 'Cleaning...' : 'Manual Cleanup'}</span>
            </Button>

            <Button 
              onClick={fetchStorageData}
              variant="outline"
              className="h-20 flex flex-col items-center justify-center"
            >
              <svg className="w-6 h-6 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <span className="text-sm">Refresh Data</span>
            </Button>
          </div>
        </div>
      </Card>

      {/* Configuration Modal */}
      <Modal
        open={showConfigModal}
        onClose={() => setShowConfigModal(false)}
        title="Storage Configuration"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Storage Backend
            </label>
            <select
              value={configForm.backend}
              onChange={(e) => setConfigForm({ ...configForm, backend: e.target.value })}
              className="w-full px-3 py-2 border border-input rounded-md bg-background"
            >
              <option value="filesystem">Filesystem</option>
              <option value="azure">Azure Blob Storage</option>
              <option value="s3">Amazon S3</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Maximum File Size (MB)
            </label>
            <Input
              type="number"
              min="1"
              max="10240"
              value={configForm.max_size_mb}
              onChange={(e) => setConfigForm({ ...configForm, max_size_mb: parseInt(e.target.value) })}
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              Retention Period (Days)
            </label>
            <Input
              type="number"
              min="1"
              max="365"
              value={configForm.retention_days}
              onChange={(e) => setConfigForm({ ...configForm, retention_days: parseInt(e.target.value) })}
            />
          </div>

          <div className="bg-yellow-50 border border-yellow-200 rounded-md p-3">
            <h6 className="font-medium text-yellow-800 mb-1">Important</h6>
            <p className="text-sm text-yellow-700">
              Changing storage backend requires server restart and may affect existing artifacts.
            </p>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowConfigModal(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleConfigSave}
              disabled={saving}
            >
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Retention Policy Modal */}
      <Modal
        open={showRetentionModal}
        onClose={() => setShowRetentionModal(false)}
        title="Retention Policy"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Retention Period (Days)
            </label>
            <Input
              type="number"
              min="1"
              max="365"
              value={retentionDays}
              onChange={(e) => setRetentionDays(parseInt(e.target.value))}
            />
            <p className="text-sm text-muted-foreground mt-1">
              Artifacts older than this period will be automatically deleted during cleanup.
            </p>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
            <h6 className="font-medium text-blue-800 mb-1">Current Usage</h6>
            <p className="text-sm text-blue-700">
              {usage?.expired_count || 0} artifacts are currently expired and can be cleaned up.
            </p>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowRetentionModal(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleRetentionSave}
              disabled={saving}
            >
              {saving ? 'Saving...' : 'Update Policy'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}