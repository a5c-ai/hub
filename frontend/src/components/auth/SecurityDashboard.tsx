'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { RefreshTokenManager } from './RefreshTokenManager';
import { TokenBlacklistManager } from './TokenBlacklistManager';
import { OIDCOrganizationMapper } from './OIDCOrganizationMapper';
import { SAMLOrganizationMapper } from './SAMLOrganizationMapper';
import { TeamSyncDashboard } from './TeamSyncDashboard';
import { OAuthDebugger } from './OAuthDebugger';
import { AuthAuditLogViewer } from './AuthAuditLogViewer';
import api from '@/lib/api';

interface SecurityEvent {
  id: string;
  user_id?: string;
  username?: string;
  event_type: 'login_attempt' | 'login_success' | 'login_failure' | 'account_locked' | 'password_reset' | 'suspicious_activity' | 'rate_limit_exceeded';
  severity: 'low' | 'medium' | 'high' | 'critical';
  ip_address: string;
  user_agent: string;
  location?: {
    country?: string;
    city?: string;
  };
  details: Record<string, unknown>;
  timestamp: string;
}

interface SecurityMetrics {
  total_login_attempts: number;
  successful_logins: number;
  failed_logins: number;
  locked_accounts: number;
  suspicious_activities: number;
  rate_limit_violations: number;
  unique_ips: number;
  new_device_logins: number;
}

interface RateLimitConfig {
  login_attempts_per_minute: number;
  login_attempts_per_hour: number;
  account_lockout_threshold: number;
  lockout_duration: number; // in minutes
  enabled: boolean;
}

type SecurityTab = 'overview' | 'tokens' | 'blacklist' | 'oidc' | 'saml' | 'team-sync' | 'oauth-debug' | 'audit-logs';

export function SecurityDashboard() {
  const [activeTab, setActiveTab] = useState<SecurityTab>('overview');
  const [events, setEvents] = useState<SecurityEvent[]>([]);
  const [metrics, setMetrics] = useState<SecurityMetrics>({
    total_login_attempts: 0,
    successful_logins: 0,
    failed_logins: 0,
    locked_accounts: 0,
    suspicious_activities: 0,
    rate_limit_violations: 0,
    unique_ips: 0,
    new_device_logins: 0
  });
  
  const [rateLimitConfig, setRateLimitConfig] = useState<RateLimitConfig>({
    login_attempts_per_minute: 5,
    login_attempts_per_hour: 30,
    account_lockout_threshold: 5,
    lockout_duration: 30,
    enabled: true
  });

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [timeRange, setTimeRange] = useState<'1h' | '24h' | '7d' | '30d'>('24h');
  const [eventFilter, setEventFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    fetchSecurityData();
  }, [timeRange]);

  const fetchSecurityData = async () => {
    setLoading(true);
    try {
      const [eventsResponse, metricsResponse, configResponse] = await Promise.all([
        api.get(`/admin/security/events?range=${timeRange}&limit=100`),
        api.get(`/admin/security/metrics?range=${timeRange}`),
        api.get('/admin/security/rate-limit-config')
      ]);

      setEvents(eventsResponse.data.events || []);
      setMetrics(metricsResponse.data);
      setRateLimitConfig(configResponse.data);
    } catch (error) {
      console.error('Failed to fetch security data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRateLimitConfigChange = (field: keyof RateLimitConfig, value: number | boolean) => {
    setRateLimitConfig(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSaveRateLimitConfig = async () => {
    setSaving(true);
    try {
      await api.put('/admin/security/rate-limit-config', rateLimitConfig);
    } catch (error) {
      console.error('Failed to save rate limit configuration:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleUnlockAccount = async (userId: string) => {
    try {
      await api.post(`/admin/security/unlock-account/${userId}`);
      // Refresh data
      fetchSecurityData();
    } catch (error) {
      console.error('Failed to unlock account:', error);
    }
  };

  const exportSecurityReport = async (format: 'csv' | 'json') => {
    try {
      const response = await api.get(`/admin/security/export?format=${format}&range=${timeRange}`, {
        responseType: 'blob'
      });
      
      const blob = new Blob([response.data]);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `security-report-${timeRange}.${format}`;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export security report:', error);
    }
  };

  const getEventSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-100 text-red-800';
      case 'high': return 'bg-orange-100 text-orange-800';
      case 'medium': return 'bg-yellow-100 text-yellow-800';
      case 'low': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getEventTypeLabel = (eventType: string) => {
    const labels: { [key: string]: string } = {
      'login_attempt': 'Login Attempt',
      'login_success': 'Login Success',
      'login_failure': 'Login Failed',
      'account_locked': 'Account Locked',
      'password_reset': 'Password Reset',
      'suspicious_activity': 'Suspicious Activity',
      'rate_limit_exceeded': 'Rate Limit Exceeded'
    };
    return labels[eventType] || eventType;
  };

  const filteredEvents = events.filter(event => {
    const matchesFilter = eventFilter === 'all' || event.event_type === eventFilter;
    const matchesSearch = !searchQuery || 
      event.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      event.ip_address.includes(searchQuery) ||
      event.event_type.toLowerCase().includes(searchQuery.toLowerCase());
    
    return matchesFilter && matchesSearch;
  });

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const tabs = [
    { id: 'overview' as SecurityTab, name: 'Overview', icon: '📊' },
    { id: 'tokens' as SecurityTab, name: 'Refresh Tokens', icon: '🔄' },
    { id: 'blacklist' as SecurityTab, name: 'Token Blacklist', icon: '🚫' },
    { id: 'oidc' as SecurityTab, name: 'OIDC Mapping', icon: '🔗' },
    { id: 'saml' as SecurityTab, name: 'SAML Mapping', icon: '🏢' },
    { id: 'team-sync' as SecurityTab, name: 'Team Sync', icon: '👥' },
    { id: 'oauth-debug' as SecurityTab, name: 'OAuth Debug', icon: '🧪' },
    { id: 'audit-logs' as SecurityTab, name: 'Audit Logs', icon: '📋' },
  ];

  const renderTabContent = () => {
    switch (activeTab) {
      case 'tokens':
        return <RefreshTokenManager isAdminView={true} />;
      case 'blacklist':
        return <TokenBlacklistManager isAdminView={true} />;
      case 'oidc':
        return <OIDCOrganizationMapper />;
      case 'saml':
        return <SAMLOrganizationMapper />;
      case 'team-sync':
        return <TeamSyncDashboard />;
      case 'oauth-debug':
        return <OAuthDebugger />;
      case 'audit-logs':
        return <AuthAuditLogViewer />;
      case 'overview':
      default:
        return renderOverviewContent();
    }
  };

  const renderOverviewContent = () => (
    <div className="space-y-6">{/* Time Range Selector */}
      <Card>
        <div className="p-4">
          <div className="flex space-x-2">
            {['1h', '24h', '7d', '30d'].map((range) => (
              <Button
                key={range}
                variant={timeRange === range ? 'default' : 'outline'}
                size="sm"
                onClick={() => setTimeRange(range as typeof timeRange)}
              >
                {range === '1h' ? '1 Hour' : range === '24h' ? '24 Hours' : range === '7d' ? '7 Days' : '30 Days'}
              </Button>
            ))}
          </div>
        </div>
      </Card>

      {/* Security Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Total Login Attempts</p>
              <p className="text-2xl font-bold text-foreground">{metrics.total_login_attempts}</p>
            </div>
            <div className="h-8 w-8 bg-blue-100 rounded-full flex items-center justify-center">
              <span className="text-blue-600 text-sm">🔐</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Failed Logins</p>
              <p className="text-2xl font-bold text-red-600">{metrics.failed_logins}</p>
            </div>
            <div className="h-8 w-8 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-red-600 text-sm">❌</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Locked Accounts</p>
              <p className="text-2xl font-bold text-orange-600">{metrics.locked_accounts}</p>
            </div>
            <div className="h-8 w-8 bg-orange-100 rounded-full flex items-center justify-center">
              <span className="text-orange-600 text-sm">🔒</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Suspicious Activities</p>
              <p className="text-2xl font-bold text-purple-600">{metrics.suspicious_activities}</p>
            </div>
            <div className="h-8 w-8 bg-purple-100 rounded-full flex items-center justify-center">
              <span className="text-purple-600 text-sm">⚠️</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Rate Limiting Configuration */}
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-foreground">Rate Limiting Configuration</h3>
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={rateLimitConfig.enabled}
                onChange={(e) => handleRateLimitConfigChange('enabled', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Enabled</span>
            </label>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Login attempts per minute
              </label>
              <Input
                type="number"
                value={rateLimitConfig.login_attempts_per_minute}
                onChange={(e) => handleRateLimitConfigChange('login_attempts_per_minute', parseInt(e.target.value))}
                min="1"
                max="100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Login attempts per hour
              </label>
              <Input
                type="number"
                value={rateLimitConfig.login_attempts_per_hour}
                onChange={(e) => handleRateLimitConfigChange('login_attempts_per_hour', parseInt(e.target.value))}
                min="1"
                max="1000"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Account lockout threshold
              </label>
              <Input
                type="number"
                value={rateLimitConfig.account_lockout_threshold}
                onChange={(e) => handleRateLimitConfigChange('account_lockout_threshold', parseInt(e.target.value))}
                min="1"
                max="20"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Lockout duration (minutes)
              </label>
              <Input
                type="number"
                value={rateLimitConfig.lockout_duration}
                onChange={(e) => handleRateLimitConfigChange('lockout_duration', parseInt(e.target.value))}
                min="1"
                max="1440"
              />
            </div>
          </div>

          <div className="mt-4">
            <Button onClick={handleSaveRateLimitConfig} disabled={saving}>
              {saving ? 'Saving...' : 'Save Configuration'}
            </Button>
          </div>
        </div>
      </Card>

      {/* Security Events */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">Recent Security Events</h3>
            <div className="flex space-x-2">
              <select
                value={eventFilter}
                onChange={(e) => setEventFilter(e.target.value)}
                className="px-3 py-1 border rounded-md text-sm"
              >
                <option value="all">All Events</option>
                <option value="login_failure">Failed Logins</option>
                <option value="account_locked">Account Lockouts</option>
                <option value="suspicious_activity">Suspicious Activity</option>
                <option value="rate_limit_exceeded">Rate Limit Exceeded</option>
              </select>
            </div>
          </div>

          <div className="mb-4">
            <Input
              placeholder="Search by username, IP address, or event type..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-md"
            />
          </div>

          {loading ? (
            <div className="text-center py-8">Loading security events...</div>
          ) : filteredEvents.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No security events found
            </div>
          ) : (
            <div className="space-y-2">
              {filteredEvents.map((event) => (
                <div
                  key={event.id}
                  className="border rounded-lg p-4 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <Badge className={getEventSeverityColor(event.severity)}>
                        {event.severity.toUpperCase()}
                      </Badge>
                      
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-foreground">
                            {getEventTypeLabel(event.event_type)}
                          </span>
                          {event.username && (
                            <span className="text-muted-foreground">({event.username})</span>
                          )}
                        </div>
                        
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>IP: {event.ip_address}</span>
                            {event.location && (
                              <span>{event.location.city}, {event.location.country}</span>
                            )}
                            <span>{formatTimestamp(event.timestamp)}</span>
                          </div>
                          {event.details && Object.keys(event.details).length > 0 && (
                            <div className="mt-1 text-xs">
                              Details: {JSON.stringify(event.details)}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      {event.event_type === 'account_locked' && event.user_id && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleUnlockAccount(event.user_id!)}
                          className="text-green-600 hover:text-green-700"
                        >
                          Unlock Account
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>
    </div>
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-foreground">Security Dashboard</h2>
          <p className="text-muted-foreground">Monitor authentication security and advanced authentication features</p>
        </div>
        
        {activeTab === 'overview' && (
          <div className="flex space-x-2">
            <Button
              variant="outline"
              onClick={() => exportSecurityReport('csv')}
            >
              Export CSV
            </Button>
            <Button
              variant="outline"
              onClick={() => exportSecurityReport('json')}
            >
              Export JSON
            </Button>
          </div>
        )}
      </div>

      {/* Navigation Tabs */}
      <Card>
        <div className="p-4">
          <nav className="flex space-x-1 overflow-x-auto">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors whitespace-nowrap ${
                  activeTab === tab.id
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.name}
              </button>
            ))}
          </nav>
        </div>
      </Card>

      {/* Tab Content */}
      {renderTabContent()}
    </div>
  );
}