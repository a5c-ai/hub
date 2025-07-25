'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';
import {
  ChartBarIcon,
  UsersIcon,
  ServerIcon,
  ShieldCheckIcon,
  DocumentArrowDownIcon,
  ExclamationTriangleIcon,
  CalendarIcon,
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon
} from '@heroicons/react/24/outline';

interface AnalyticsDashboardProps {
  orgName: string;
}

interface DashboardMetrics {
  overview: {
    total_members: number;
    total_repositories: number;
    total_teams: number;
    active_members_30d: number;
    commits_this_month: number;
    issues_open: number;
    pull_requests_open: number;
    security_score: number;
  };
  recent_activity: Array<{
    date: string;
    action: string;
    actor_name: string;
    target_type: string;
    target_name: string;
    description: string;
  }>;
  top_repositories: Array<{
    name: string;
    language: string;
    stars: number;
    forks: number;
    contributors: number;
    commits_30d: number;
    last_activity_at: string;
  }>;
  active_members: Array<{
    username: string;
    name: string;
    role: string;
    commits_30d: number;
    issues_30d: number;
    pull_requests_30d: number;
    last_active_at: string;
  }>;
  security_alerts: Array<{
    type: string;
    severity: string;
    title: string;
    description: string;
    repository: string;
    created_at: string;
  }>;
  storage_usage: {
    total_used_gb: number;
    total_limit_gb: number;
    usage_percent: number;
  };
  bandwidth_usage: {
    total_used_gb: number;
    total_limit_gb: number;
    usage_percent: number;
  };
}

interface MemberActivityMetrics {
  period: string;
  total_members: number;
  active_members: number;
  member_growth: Array<{
    date: string;
    added: number;
    removed: number;
    total: number;
  }>;
  activity_trends: Array<{
    date: string;
    commits: number;
    issues: number;
    pull_requests: number;
    comments: number;
  }>;
  top_contributors: Array<{
    username: string;
    name: string;
    commits_30d: number;
    issues_30d: number;
    pull_requests_30d: number;
  }>;
}

interface SecurityMetrics {
  period: string;
  security_score: number;
  vulnerabilities_found: number;
  vulnerabilities_fixed: number;
  security_alerts: Array<{
    type: string;
    severity: string;
    title: string;
    created_at: string;
  }>;
  compliance_status: Record<string, boolean>;
  policy_violations: Array<{
    policy_name: string;
    count: number;
    last_occurred: string;
  }>;
}

export function OrganizationAnalyticsDashboard({ orgName }: AnalyticsDashboardProps) {
  const [activeTab, setActiveTab] = useState<'overview' | 'members' | 'repositories' | 'security' | 'usage'>('overview');
  const [period, setPeriod] = useState('30d');
  const [dashboardMetrics, setDashboardMetrics] = useState<DashboardMetrics | null>(null);
  const [memberMetrics, setMemberMetrics] = useState<MemberActivityMetrics | null>(null);
  const [securityMetrics, setSecurityMetrics] = useState<SecurityMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchAnalyticsData();
  }, [orgName, period]);

  const fetchAnalyticsData = async () => {
    try {
      setLoading(true);
      const [dashboardResponse, memberResponse, securityResponse] = await Promise.all([
        api.get(`/organizations/${orgName}/analytics/dashboard`),
        api.get(`/organizations/${orgName}/analytics/members?period=${period}`),
        api.get(`/organizations/${orgName}/analytics/security?period=${period}`)
      ]);

      setDashboardMetrics(dashboardResponse.data);
      setMemberMetrics(memberResponse.data);
      setSecurityMetrics(securityResponse.data);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to fetch analytics data');
    } finally {
      setLoading(false);
    }
  };

  const exportAnalytics = async (format: 'csv' | 'json') => {
    try {
      const response = await api.get(`/organizations/${orgName}/analytics/export?format=${format}&period=${period}`, {
        responseType: 'blob'
      });
      
      const blob = new Blob([response.data]);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${orgName}-analytics-${period}.${format}`;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to export analytics');
    }
  };

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  };

  const formatPercentage = (percent: number) => {
    return `${percent.toFixed(1)}%`;
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/3 mb-6"></div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-32 bg-muted rounded"></div>
            ))}
          </div>
          <div className="h-96 bg-muted rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Analytics Dashboard</h1>
          <p className="text-muted-foreground">Insights and metrics for {orgName}</p>
        </div>
        <div className="flex items-center space-x-4">
          <select
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
            className="px-3 py-2 border border-input rounded-md bg-background text-foreground"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
            <option value="1y">Last year</option>
          </select>
          <Button variant="outline" onClick={() => exportAnalytics('csv')}>
            <DocumentArrowDownIcon className="w-4 h-4 mr-2" />
            Export CSV
          </Button>
          <Button variant="outline" onClick={() => exportAnalytics('json')}>
            <DocumentArrowDownIcon className="w-4 h-4 mr-2" />
            Export JSON
          </Button>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
          <div className="flex">
            <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />
            <div className="ml-3">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* Navigation Tabs */}
      <div className="border-b border-border mb-8">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'overview', name: 'Overview', icon: ChartBarIcon },
            { id: 'members', name: 'Members', icon: UsersIcon },
            { id: 'repositories', name: 'Repositories', icon: ServerIcon },
            { id: 'security', name: 'Security', icon: ShieldCheckIcon },
            { id: 'usage', name: 'Usage & Costs', icon: CalendarIcon }
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors flex items-center ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <tab.icon className="w-5 h-5 mr-2" />
              {tab.name}
            </button>
          ))}
        </nav>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && dashboardMetrics && (
        <div className="space-y-8">
          {/* Key Metrics */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <div className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Members</p>
                    <p className="text-2xl font-bold text-foreground">{dashboardMetrics.overview.total_members}</p>
                  </div>
                  <UsersIcon className="h-8 w-8 text-blue-500" />
                </div>
                <div className="mt-2">
                  <p className="text-sm text-muted-foreground">
                    {dashboardMetrics.overview.active_members_30d} active this month
                  </p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Repositories</p>
                    <p className="text-2xl font-bold text-foreground">{dashboardMetrics.overview.total_repositories}</p>
                  </div>
                  <ServerIcon className="h-8 w-8 text-green-500" />
                </div>
                <div className="mt-2">
                  <p className="text-sm text-muted-foreground">
                    {dashboardMetrics.overview.commits_this_month} commits this month
                  </p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Open Issues</p>
                    <p className="text-2xl font-bold text-foreground">{dashboardMetrics.overview.issues_open}</p>
                  </div>
                  <ExclamationTriangleIcon className="h-8 w-8 text-yellow-500" />
                </div>
                <div className="mt-2">
                  <p className="text-sm text-muted-foreground">
                    {dashboardMetrics.overview.pull_requests_open} open PRs
                  </p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Security Score</p>
                    <p className="text-2xl font-bold text-foreground">{formatPercentage(dashboardMetrics.overview.security_score)}</p>
                  </div>
                  <ShieldCheckIcon className="h-8 w-8 text-purple-500" />
                </div>
                <div className="mt-2">
                  <Badge variant={dashboardMetrics.overview.security_score >= 80 ? 'default' : 'secondary'}>
                    {dashboardMetrics.overview.security_score >= 80 ? 'Good' : 'Needs Attention'}
                  </Badge>
                </div>
              </div>
            </Card>
          </div>

          {/* Top Repositories */}
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Top Repositories</h3>
              <div className="space-y-4">
                {dashboardMetrics.top_repositories.slice(0, 5).map((repo, index) => (
                  <div key={index} className="flex items-center justify-between p-4 border border-border rounded-lg">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3">
                        <span className="text-lg font-medium text-foreground">{repo.name}</span>
                        <Badge variant="outline">{repo.language}</Badge>
                      </div>
                      <div className="flex items-center space-x-4 mt-2 text-sm text-muted-foreground">
                        <span>★ {repo.stars}</span>
                        <span>⑂ {repo.forks}</span>
                        <span>{repo.contributors} contributors</span>
                        <span>{repo.commits_30d} commits (30d)</span>
                      </div>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {new Date(repo.last_activity_at).toLocaleDateString()}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>

          {/* Storage and Bandwidth Usage */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-medium text-foreground mb-4">Storage Usage</h3>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">Used</span>
                    <span className="text-sm font-medium">{dashboardMetrics.storage_usage.total_used_gb}GB</span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div 
                      className="bg-blue-500 h-2 rounded-full" 
                      style={{ width: `${dashboardMetrics.storage_usage.usage_percent}%` }}
                    ></div>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">
                      {formatPercentage(dashboardMetrics.storage_usage.usage_percent)} used
                    </span>
                    <span className="text-muted-foreground">
                      {dashboardMetrics.storage_usage.total_limit_gb}GB total
                    </span>
                  </div>
                </div>
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <h3 className="text-lg font-medium text-foreground mb-4">Bandwidth Usage</h3>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">Used</span>
                    <span className="text-sm font-medium">{dashboardMetrics.bandwidth_usage.total_used_gb}GB</span>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div 
                      className="bg-green-500 h-2 rounded-full" 
                      style={{ width: `${dashboardMetrics.bandwidth_usage.usage_percent}%` }}
                    ></div>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">
                      {formatPercentage(dashboardMetrics.bandwidth_usage.usage_percent)} used
                    </span>
                    <span className="text-muted-foreground">
                      {dashboardMetrics.bandwidth_usage.total_limit_gb}GB total
                    </span>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>
      )}

      {/* Members Tab */}
      {activeTab === 'members' && memberMetrics && (
        <div className="space-y-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">{memberMetrics.total_members}</p>
                <p className="text-sm text-muted-foreground">Total Members</p>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">{memberMetrics.active_members}</p>
                <p className="text-sm text-muted-foreground">Active Members</p>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">
                  {formatPercentage((memberMetrics.active_members / memberMetrics.total_members) * 100)}
                </p>
                <p className="text-sm text-muted-foreground">Activity Rate</p>
              </div>
            </Card>
          </div>

          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Top Contributors</h3>
              <div className="space-y-4">
                {memberMetrics.top_contributors.slice(0, 10).map((contributor, index) => (
                  <div key={index} className="flex items-center justify-between p-4 border border-border rounded-lg">
                    <div>
                      <p className="font-medium text-foreground">{contributor.name}</p>
                      <p className="text-sm text-muted-foreground">@{contributor.username}</p>
                    </div>
                    <div className="flex space-x-6 text-sm text-muted-foreground">
                      <div className="text-center">
                        <p className="font-medium text-foreground">{contributor.commits_30d}</p>
                        <p>Commits</p>
                      </div>
                      <div className="text-center">
                        <p className="font-medium text-foreground">{contributor.issues_30d}</p>
                        <p>Issues</p>
                      </div>
                      <div className="text-center">
                        <p className="font-medium text-foreground">{contributor.pull_requests_30d}</p>
                        <p>PRs</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Security Tab */}
      {activeTab === 'security' && securityMetrics && (
        <div className="space-y-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">{formatPercentage(securityMetrics.security_score)}</p>
                <p className="text-sm text-muted-foreground">Security Score</p>
                <Badge variant={securityMetrics.security_score >= 80 ? 'default' : 'secondary'} className="mt-2">
                  {securityMetrics.security_score >= 80 ? 'Good' : 'Needs Attention'}
                </Badge>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">{securityMetrics.vulnerabilities_found}</p>
                <p className="text-sm text-muted-foreground">Vulnerabilities Found</p>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">{securityMetrics.vulnerabilities_fixed}</p>
                <p className="text-sm text-muted-foreground">Vulnerabilities Fixed</p>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <p className="text-2xl font-bold text-foreground">{securityMetrics.policy_violations.length}</p>
                <p className="text-sm text-muted-foreground">Policy Violations</p>
              </div>
            </Card>
          </div>

          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Compliance Status</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {Object.entries(securityMetrics.compliance_status).map(([standard, compliant]) => (
                  <div key={standard} className="flex items-center justify-between p-4 border border-border rounded-lg">
                    <span className="font-medium text-foreground">{standard}</span>
                    <Badge variant={compliant ? 'default' : 'secondary'}>
                      {compliant ? 'Compliant' : 'Non-Compliant'}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          </Card>

          {securityMetrics.security_alerts.length > 0 && (
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-medium text-foreground mb-4">Recent Security Alerts</h3>
                <div className="space-y-4">
                  {securityMetrics.security_alerts.slice(0, 5).map((alert, index) => (
                    <div key={index} className="flex items-start space-x-4 p-4 border border-border rounded-lg">
                      <div className="flex-shrink-0">
                        <Badge variant={alert.severity === 'high' ? 'destructive' : alert.severity === 'medium' ? 'default' : 'secondary'}>
                          {alert.severity}
                        </Badge>
                      </div>
                      <div className="flex-1">
                        <p className="font-medium text-foreground">{alert.title}</p>
                        <p className="text-sm text-muted-foreground mt-1">{alert.type}</p>
                        <p className="text-xs text-muted-foreground mt-2">
                          {new Date(alert.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}