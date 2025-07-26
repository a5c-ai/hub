'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface TeamSyncProvider {
  id: string;
  name: string;
  type: 'ldap' | 'active_directory' | 'okta' | 'github' | 'azure_ad';
  connection_string: string;
  is_active: boolean;
  last_sync_at?: string;
  last_sync_status: 'success' | 'failed' | 'pending' | 'never';
  sync_interval: number; // in minutes
  created_at: string;
}

interface SyncJob {
  id: string;
  provider_id: string;
  provider_name: string;
  started_at: string;
  completed_at?: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  teams_created: number;
  teams_updated: number;
  users_synced: number;
  error_message?: string;
  details: Record<string, any>;
}

interface TeamSyncStats {
  total_providers: number;
  active_providers: number;
  successful_syncs_24h: number;
  failed_syncs_24h: number;
  teams_synced: number;
  users_synced: number;
}

interface ExternalTeam {
  id: string;
  external_id: string;
  provider_id: string;
  provider_name: string;
  name: string;
  description?: string;
  member_count: number;
  last_synced_at: string;
  sync_status: 'synced' | 'pending' | 'error';
  organization_id?: string;
  organization_name?: string;
}

export function TeamSyncDashboard() {
  const [providers, setProviders] = useState<TeamSyncProvider[]>([]);
  const [syncJobs, setSyncJobs] = useState<SyncJob[]>([]);
  const [externalTeams, setExternalTeams] = useState<ExternalTeam[]>([]);
  const [stats, setStats] = useState<TeamSyncStats>({
    total_providers: 0,
    active_providers: 0,
    successful_syncs_24h: 0,
    failed_syncs_24h: 0,
    teams_synced: 0,
    users_synced: 0
  });

  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'providers' | 'jobs' | 'teams'>('providers');
  const [searchQuery, setSearchQuery] = useState('');
  const [showAddProviderModal, setShowAddProviderModal] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<string>('');

  // New provider form state
  const [newProvider, setNewProvider] = useState({
    name: '',
    type: 'ldap' as TeamSyncProvider['type'],
    connection_string: '',
    sync_interval: 60
  });

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [providersResponse, jobsResponse, teamsResponse, statsResponse] = await Promise.all([
        api.get('/admin/auth/team-sync/providers'),
        api.get('/admin/auth/team-sync/jobs?limit=20'),
        api.get('/admin/auth/team-sync/teams'),
        api.get('/admin/auth/team-sync/stats')
      ]);

      setProviders(providersResponse.data.providers || []);
      setSyncJobs(jobsResponse.data.jobs || []);
      setExternalTeams(teamsResponse.data.teams || []);
      setStats(statsResponse.data);
    } catch (error) {
      console.error('Failed to fetch team sync data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAddProvider = async () => {
    if (!newProvider.name.trim() || !newProvider.connection_string.trim()) {
      return;
    }

    try {
      const response = await api.post('/admin/auth/team-sync/providers', newProvider);
      setProviders(prev => [...prev, response.data]);
      setNewProvider({
        name: '',
        type: 'ldap',
        connection_string: '',
        sync_interval: 60
      });
      setShowAddProviderModal(false);
      fetchData(); // Refresh stats
    } catch (error) {
      console.error('Failed to add team sync provider:', error);
    }
  };

  const handleToggleProvider = async (providerId: string, isActive: boolean) => {
    try {
      await api.patch(`/admin/auth/team-sync/providers/${providerId}`, { is_active: !isActive });
      setProviders(prev => prev.map(p => 
        p.id === providerId ? { ...p, is_active: !isActive } : p
      ));
    } catch (error) {
      console.error('Failed to toggle provider:', error);
    }
  };

  const handleDeleteProvider = async (providerId: string) => {
    if (!confirm('Are you sure you want to delete this team sync provider? This will stop all synchronization.')) {
      return;
    }

    try {
      await api.delete(`/admin/auth/team-sync/providers/${providerId}`);
      setProviders(prev => prev.filter(p => p.id !== providerId));
      fetchData(); // Refresh stats
    } catch (error) {
      console.error('Failed to delete provider:', error);
    }
  };

  const handleManualSync = async (providerId: string) => {
    try {
      const response = await api.post(`/admin/auth/team-sync/providers/${providerId}/sync`);
      setSyncJobs(prev => [response.data, ...prev]);
    } catch (error) {
      console.error('Failed to start manual sync:', error);
    }
  };

  const handleTestConnection = async (providerId: string) => {
    try {
      const response = await api.post(`/admin/auth/team-sync/providers/${providerId}/test`);
      alert(`Connection test result: ${response.data.message}`);
    } catch (error) {
      console.error('Failed to test connection:', error);
      alert('Connection test failed. Check console for details.');
    }
  };

  const handleCancelJob = async (jobId: string) => {
    try {
      await api.post(`/admin/auth/team-sync/jobs/${jobId}/cancel`);
      setSyncJobs(prev => prev.map(job => 
        job.id === jobId ? { ...job, status: 'cancelled' } : job
      ));
    } catch (error) {
      console.error('Failed to cancel sync job:', error);
    }
  };

  const getProviderTypeIcon = (type: string) => {
    switch (type) {
      case 'ldap': return '🏢';
      case 'active_directory': return '🏛️';
      case 'okta': return '🔐';
      case 'github': return '🐙';
      case 'azure_ad': return '☁️';
      default: return '🔗';
    }
  };

  const getStatusBadgeColor = (status: string) => {
    switch (status) {
      case 'success':
      case 'completed':
      case 'synced':
        return 'bg-green-100 text-green-800';
      case 'failed':
      case 'error':
        return 'bg-red-100 text-red-800';
      case 'pending':
      case 'running':
        return 'bg-blue-100 text-blue-800';
      case 'cancelled':
        return 'bg-orange-100 text-orange-800';
      case 'never':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);
    
    if (diffDays > 0) return `${diffDays}d ago`;
    if (diffHours > 0) return `${diffHours}h ago`;
    return 'Recently';
  };

  const filteredProviders = providers.filter(provider =>
    provider.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    provider.type.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const filteredJobs = syncJobs.filter(job =>
    job.provider_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    job.status.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const filteredTeams = externalTeams.filter(team =>
    team.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    team.provider_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    team.organization_name?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Total Providers</p>
              <p className="text-2xl font-bold text-foreground">{stats.total_providers}</p>
            </div>
            <div className="h-8 w-8 bg-blue-100 rounded-full flex items-center justify-center">
              <span className="text-blue-600 text-sm">🔗</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Active Providers</p>
              <p className="text-2xl font-bold text-green-600">{stats.active_providers}</p>
            </div>
            <div className="h-8 w-8 bg-green-100 rounded-full flex items-center justify-center">
              <span className="text-green-600 text-sm">✅</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Success (24h)</p>
              <p className="text-2xl font-bold text-green-600">{stats.successful_syncs_24h}</p>
            </div>
            <div className="h-8 w-8 bg-green-100 rounded-full flex items-center justify-center">
              <span className="text-green-600 text-sm">📈</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Failed (24h)</p>
              <p className="text-2xl font-bold text-red-600">{stats.failed_syncs_24h}</p>
            </div>
            <div className="h-8 w-8 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-red-600 text-sm">📉</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Teams Synced</p>
              <p className="text-2xl font-bold text-purple-600">{stats.teams_synced}</p>
            </div>
            <div className="h-8 w-8 bg-purple-100 rounded-full flex items-center justify-center">
              <span className="text-purple-600 text-sm">👥</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Users Synced</p>
              <p className="text-2xl font-bold text-orange-600">{stats.users_synced}</p>
            </div>
            <div className="h-8 w-8 bg-orange-100 rounded-full flex items-center justify-center">
              <span className="text-orange-600 text-sm">👤</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Navigation Tabs */}
      <Card>
        <div className="p-4">
          <div className="flex justify-between items-center mb-4">
            <div className="flex space-x-1">
              {[
                { id: 'providers', name: 'Providers', icon: '🔗' },
                { id: 'jobs', name: 'Sync Jobs', icon: '⚙️' },
                { id: 'teams', name: 'External Teams', icon: '👥' }
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as typeof activeTab)}
                  className={`px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    activeTab === tab.id
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  }`}
                >
                  <span className="mr-2">{tab.icon}</span>
                  {tab.name}
                </button>
              ))}
            </div>
            
            <div className="flex items-center space-x-2">
              <Input
                placeholder={`Search ${activeTab}...`}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64"
              />
              {activeTab === 'providers' && (
                <Button onClick={() => setShowAddProviderModal(true)}>
                  Add Provider
                </Button>
              )}
            </div>
          </div>

          {/* Providers Tab */}
          {activeTab === 'providers' && (
            <div className="space-y-3">
              {loading ? (
                <div className="text-center py-8">Loading providers...</div>
              ) : filteredProviders.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No team sync providers found
                </div>
              ) : (
                filteredProviders.map((provider) => (
                  <div key={provider.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <span className="text-2xl">{getProviderTypeIcon(provider.type)}</span>
                        <div>
                          <div className="flex items-center space-x-2">
                            <span className="font-medium text-foreground">{provider.name}</span>
                            <Badge className={getStatusBadgeColor(provider.last_sync_status)}>
                              {provider.last_sync_status}
                            </Badge>
                            <Badge className={provider.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                              {provider.is_active ? 'Active' : 'Inactive'}
                            </Badge>
                          </div>
                          <div className="text-sm text-muted-foreground mt-1">
                            Type: {provider.type} • Interval: {provider.sync_interval}min
                            {provider.last_sync_at && (
                              <span> • Last sync: {formatTimestamp(provider.last_sync_at)}</span>
                            )}
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleTestConnection(provider.id)}
                        >
                          Test
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleManualSync(provider.id)}
                          disabled={!provider.is_active}
                        >
                          Sync Now
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleToggleProvider(provider.id, provider.is_active)}
                          className={provider.is_active ? 'text-orange-600' : 'text-green-600'}
                        >
                          {provider.is_active ? 'Disable' : 'Enable'}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDeleteProvider(provider.id)}
                          className="text-red-600 hover:text-red-700"
                        >
                          Delete
                        </Button>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* Sync Jobs Tab */}
          {activeTab === 'jobs' && (
            <div className="space-y-3">
              {loading ? (
                <div className="text-center py-8">Loading sync jobs...</div>
              ) : filteredJobs.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No sync jobs found
                </div>
              ) : (
                filteredJobs.map((job) => (
                  <div key={job.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-foreground">{job.provider_name}</span>
                          <Badge className={getStatusBadgeColor(job.status)}>
                            {job.status}
                          </Badge>
                        </div>
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>Started: {formatTimestamp(job.started_at)}</span>
                            {job.completed_at && (
                              <span>Completed: {formatTimestamp(job.completed_at)}</span>
                            )}
                            <span>Teams: {job.teams_created + job.teams_updated}</span>
                            <span>Users: {job.users_synced}</span>
                          </div>
                          {job.error_message && (
                            <div className="mt-1 text-red-600 text-xs">
                              Error: {job.error_message}
                            </div>
                          )}
                        </div>
                      </div>

                      <div className="flex items-center space-x-2">
                        {job.status === 'running' && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleCancelJob(job.id)}
                            className="text-orange-600 hover:text-orange-700"
                          >
                            Cancel
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* External Teams Tab */}
          {activeTab === 'teams' && (
            <div className="space-y-3">
              {loading ? (
                <div className="text-center py-8">Loading external teams...</div>
              ) : filteredTeams.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No external teams found
                </div>
              ) : (
                filteredTeams.map((team) => (
                  <div key={team.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-foreground">{team.name}</span>
                          <Badge className={getStatusBadgeColor(team.sync_status)}>
                            {team.sync_status}
                          </Badge>
                          {team.organization_name && (
                            <Badge variant="outline">
                              {team.organization_name}
                            </Badge>
                          )}
                        </div>
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>Provider: {team.provider_name}</span>
                            <span>Members: {team.member_count}</span>
                            <span>Last synced: {formatTimestamp(team.last_synced_at)}</span>
                          </div>
                          {team.description && (
                            <div className="mt-1 text-xs">
                              {team.description}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </Card>

      {/* Add Provider Modal */}
      {showAddProviderModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-md">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Add Team Sync Provider</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Provider Name
                  </label>
                  <Input
                    value={newProvider.name}
                    onChange={(e) => setNewProvider(prev => ({ ...prev, name: e.target.value }))}
                    placeholder="e.g., Company LDAP, GitHub Organization"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Provider Type
                  </label>
                  <select
                    value={newProvider.type}
                    onChange={(e) => setNewProvider(prev => ({ ...prev, type: e.target.value as TeamSyncProvider['type'] }))}
                    className="w-full px-3 py-2 border rounded-md"
                  >
                    <option value="ldap">LDAP</option>
                    <option value="active_directory">Active Directory</option>
                    <option value="okta">Okta</option>
                    <option value="github">GitHub</option>
                    <option value="azure_ad">Azure AD</option>
                  </select>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Connection String
                  </label>
                  <Input
                    value={newProvider.connection_string}
                    onChange={(e) => setNewProvider(prev => ({ ...prev, connection_string: e.target.value }))}
                    placeholder="Provider-specific connection details"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Sync Interval (minutes)
                  </label>
                  <Input
                    type="number"
                    value={newProvider.sync_interval}
                    onChange={(e) => setNewProvider(prev => ({ ...prev, sync_interval: parseInt(e.target.value) }))}
                    min="5"
                    max="1440"
                  />
                </div>
              </div>
              
              <div className="flex justify-end space-x-2 mt-6">
                <Button
                  variant="outline"
                  onClick={() => setShowAddProviderModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleAddProvider}
                  disabled={!newProvider.name.trim() || !newProvider.connection_string.trim()}
                >
                  Add Provider
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}