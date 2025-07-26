'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface AuditLogEntry {
  id: string;
  user_id?: string;
  username?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  details: Record<string, any>;
  ip_address: string;
  user_agent: string;
  location?: {
    country?: string;
    city?: string;
    region?: string;
  };
  outcome: 'success' | 'failure' | 'warning';
  risk_level: 'low' | 'medium' | 'high' | 'critical';
  timestamp: string;
  session_id?: string;
}

interface AuditStats {
  total_entries: number;
  success_rate: number;
  failed_attempts_24h: number;
  high_risk_events_24h: number;
  unique_users_24h: number;
  unique_ips_24h: number;
}

interface ExportRequest {
  format: 'csv' | 'json' | 'pdf';
  date_range: string;
  filters: Record<string, any>;
}

export function AuthAuditLogViewer() {
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [stats, setStats] = useState<AuditStats>({
    total_entries: 0,
    success_rate: 0,
    failed_attempts_24h: 0,
    high_risk_events_24h: 0,
    unique_users_24h: 0,
    unique_ips_24h: 0
  });
  
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [dateRange, setDateRange] = useState<'1h' | '24h' | '7d' | '30d'>('24h');
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [outcomeFilter, setOutcomeFilter] = useState<string>('all');
  const [riskFilter, setRiskFilter] = useState<string>('all');
  const [userFilter, setUserFilter] = useState('');
  const [ipFilter, setIpFilter] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(20);
  const [totalPages, setTotalPages] = useState(1);
  const [showExportModal, setShowExportModal] = useState(false);
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    fetchAuditLogs();
    fetchStats();
  }, [dateRange, actionFilter, outcomeFilter, riskFilter, userFilter, ipFilter, currentPage]);

  const fetchAuditLogs = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        range: dateRange,
        page: currentPage.toString(),
        limit: pageSize.toString(),
        ...(actionFilter !== 'all' && { action: actionFilter }),
        ...(outcomeFilter !== 'all' && { outcome: outcomeFilter }),
        ...(riskFilter !== 'all' && { risk_level: riskFilter }),
        ...(userFilter && { username: userFilter }),
        ...(ipFilter && { ip_address: ipFilter }),
        ...(searchQuery && { search: searchQuery })
      });

      const response = await api.get(`/admin/auth/audit-logs?${params}`);
      setAuditLogs(response.data.logs || []);
      setTotalPages(Math.ceil((response.data.total || 0) / pageSize));
    } catch (error) {
      console.error('Failed to fetch audit logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await api.get(`/admin/auth/audit-logs/stats?range=${dateRange}`);
      setStats(response.data);
    } catch (error) {
      console.error('Failed to fetch audit stats:', error);
    }
  };

  const handleExport = async (format: 'csv' | 'json' | 'pdf') => {
    setExporting(true);
    try {
      const params = new URLSearchParams({
        format,
        range: dateRange,
        ...(actionFilter !== 'all' && { action: actionFilter }),
        ...(outcomeFilter !== 'all' && { outcome: outcomeFilter }),
        ...(riskFilter !== 'all' && { risk_level: riskFilter }),
        ...(userFilter && { username: userFilter }),
        ...(ipFilter && { ip_address: ipFilter }),
        ...(searchQuery && { search: searchQuery })
      });

      const response = await api.get(`/admin/auth/audit-logs/export?${params}`, {
        responseType: 'blob'
      });
      
      const blob = new Blob([response.data]);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `auth-audit-${dateRange}-${Date.now()}.${format}`;
      link.click();
      window.URL.revokeObjectURL(url);
      
      setShowExportModal(false);
    } catch (error) {
      console.error('Failed to export audit logs:', error);
    } finally {
      setExporting(false);
    }
  };

  const filteredLogs = auditLogs.filter(log => {
    const matchesSearch = !searchQuery || 
      log.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.action.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.resource_type.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.ip_address.includes(searchQuery) ||
      JSON.stringify(log.details).toLowerCase().includes(searchQuery.toLowerCase());
    
    return matchesSearch;
  });

  const getOutcomeBadgeColor = (outcome: string) => {
    switch (outcome) {
      case 'success': return 'bg-green-100 text-green-800';
      case 'failure': return 'bg-red-100 text-red-800';
      case 'warning': return 'bg-yellow-100 text-yellow-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getRiskBadgeColor = (riskLevel: string) => {
    switch (riskLevel) {
      case 'critical': return 'bg-red-100 text-red-800';
      case 'high': return 'bg-orange-100 text-orange-800';
      case 'medium': return 'bg-yellow-100 text-yellow-800';
      case 'low': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const formatDetails = (details: Record<string, any>) => {
    return JSON.stringify(details, null, 2);
  };

  const getActionIcon = (action: string) => {
    switch (action.toLowerCase()) {
      case 'login': return '🔑';
      case 'logout': return '🚪';
      case 'password_reset': return '🔒';
      case 'mfa_setup': return '📱';
      case 'oauth_auth': return '🔗';
      case 'saml_auth': return '🏢';
      case 'token_refresh': return '🔄';
      case 'account_locked': return '🔐';
      case 'password_changed': return '🔑';
      case 'email_verified': return '✉️';
      default: return '📝';
    }
  };

  const commonActions = [
    'login', 'logout', 'password_reset', 'mfa_setup', 'oauth_auth', 
    'saml_auth', 'token_refresh', 'account_locked', 'password_changed', 'email_verified'
  ];

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Total Entries</p>
              <p className="text-2xl font-bold text-foreground">{stats.total_entries}</p>
            </div>
            <div className="h-8 w-8 bg-blue-100 rounded-full flex items-center justify-center">
              <span className="text-blue-600 text-sm">📊</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Success Rate</p>
              <p className="text-2xl font-bold text-green-600">{stats.success_rate.toFixed(1)}%</p>
            </div>
            <div className="h-8 w-8 bg-green-100 rounded-full flex items-center justify-center">
              <span className="text-green-600 text-sm">✅</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Failed (24h)</p>
              <p className="text-2xl font-bold text-red-600">{stats.failed_attempts_24h}</p>
            </div>
            <div className="h-8 w-8 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-red-600 text-sm">❌</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">High Risk (24h)</p>
              <p className="text-2xl font-bold text-orange-600">{stats.high_risk_events_24h}</p>
            </div>
            <div className="h-8 w-8 bg-orange-100 rounded-full flex items-center justify-center">
              <span className="text-orange-600 text-sm">⚠️</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Users (24h)</p>
              <p className="text-2xl font-bold text-purple-600">{stats.unique_users_24h}</p>
            </div>
            <div className="h-8 w-8 bg-purple-100 rounded-full flex items-center justify-center">
              <span className="text-purple-600 text-sm">👥</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">IPs (24h)</p>
              <p className="text-2xl font-bold text-indigo-600">{stats.unique_ips_24h}</p>
            </div>
            <div className="h-8 w-8 bg-indigo-100 rounded-full flex items-center justify-center">
              <span className="text-indigo-600 text-sm">🌐</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Filters and Controls */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">Authentication Audit Logs</h3>
            <Button
              variant="outline"
              onClick={() => setShowExportModal(true)}
            >
              Export Logs
            </Button>
          </div>

          {/* Time Range Selector */}
          <div className="flex space-x-2 mb-4">
            {['1h', '24h', '7d', '30d'].map((range) => (
              <Button
                key={range}
                variant={dateRange === range ? 'default' : 'outline'}
                size="sm"
                onClick={() => setDateRange(range as typeof dateRange)}
              >
                {range === '1h' ? '1 Hour' : range === '24h' ? '24 Hours' : range === '7d' ? '7 Days' : '30 Days'}
              </Button>
            ))}
          </div>

          {/* Filters */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4 mb-4">
            <Input
              placeholder="Search logs..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            
            <select
              value={actionFilter}
              onChange={(e) => setActionFilter(e.target.value)}
              className="px-3 py-2 border rounded-md"
            >
              <option value="all">All Actions</option>
              {commonActions.map((action) => (
                <option key={action} value={action}>
                  {action.replace('_', ' ').toUpperCase()}
                </option>
              ))}
            </select>

            <select
              value={outcomeFilter}
              onChange={(e) => setOutcomeFilter(e.target.value)}
              className="px-3 py-2 border rounded-md"
            >
              <option value="all">All Outcomes</option>
              <option value="success">Success</option>
              <option value="failure">Failure</option>
              <option value="warning">Warning</option>
            </select>

            <select
              value={riskFilter}
              onChange={(e) => setRiskFilter(e.target.value)}
              className="px-3 py-2 border rounded-md"
            >
              <option value="all">All Risk Levels</option>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>

            <Input
              placeholder="Filter by username..."
              value={userFilter}
              onChange={(e) => setUserFilter(e.target.value)}
            />

            <Input
              placeholder="Filter by IP address..."
              value={ipFilter}
              onChange={(e) => setIpFilter(e.target.value)}
            />
          </div>

          {/* Audit Logs */}
          {loading ? (
            <div className="text-center py-8">Loading audit logs...</div>
          ) : filteredLogs.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No audit logs found
            </div>
          ) : (
            <div className="space-y-3">
              {filteredLogs.map((log) => (
                <div
                  key={log.id}
                  className={`border rounded-lg p-4 ${
                    log.risk_level === 'critical' || log.risk_level === 'high'
                      ? 'border-red-200 bg-red-50'
                      : log.outcome === 'failure'
                        ? 'border-orange-200 bg-orange-50'
                        : ''
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3 flex-1">
                      <span className="text-2xl">{getActionIcon(log.action)}</span>
                      
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 flex-wrap">
                          <span className="font-medium text-foreground">{log.action}</span>
                          <Badge className={getOutcomeBadgeColor(log.outcome)}>
                            {log.outcome}
                          </Badge>
                          <Badge className={getRiskBadgeColor(log.risk_level)}>
                            {log.risk_level} risk
                          </Badge>
                          {log.username && (
                            <span className="text-sm text-muted-foreground">by {log.username}</span>
                          )}
                        </div>
                        
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4 flex-wrap">
                            <span>Resource: {log.resource_type}</span>
                            <span>IP: {log.ip_address}</span>
                            {log.location && (
                              <span>Location: {log.location.city}, {log.location.country}</span>
                            )}
                            <span>Time: {formatTimestamp(log.timestamp)}</span>
                            {log.session_id && (
                              <span>Session: {log.session_id.substring(0, 8)}...</span>
                            )}
                          </div>
                        </div>

                        {Object.keys(log.details).length > 0 && (
                          <details className="mt-2">
                            <summary className="text-sm cursor-pointer text-blue-600">View Details</summary>
                            <pre className="text-xs bg-gray-100 p-2 mt-1 rounded overflow-x-auto max-h-40">
                              {formatDetails(log.details)}
                            </pre>
                          </details>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-6">
              <div className="text-sm text-muted-foreground">
                Page {currentPage} of {totalPages}
              </div>
              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                  disabled={currentPage === 1}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                  disabled={currentPage === totalPages}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* Export Modal */}
      {showExportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-md">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Export Audit Logs</h3>
              
              <div className="space-y-4">
                <div>
                  <p className="text-sm text-muted-foreground mb-3">
                    Export will include all logs matching current filters and date range ({dateRange}).
                  </p>
                </div>
                
                <div className="grid grid-cols-3 gap-2">
                  <Button
                    variant="outline"
                    onClick={() => handleExport('csv')}
                    disabled={exporting}
                    className="flex flex-col items-center py-4"
                  >
                    <span className="text-lg mb-1">📊</span>
                    <span className="text-xs">CSV</span>
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => handleExport('json')}
                    disabled={exporting}
                    className="flex flex-col items-center py-4"
                  >
                    <span className="text-lg mb-1">📄</span>
                    <span className="text-xs">JSON</span>
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => handleExport('pdf')}
                    disabled={exporting}
                    className="flex flex-col items-center py-4"
                  >
                    <span className="text-lg mb-1">📑</span>
                    <span className="text-xs">PDF</span>
                  </Button>
                </div>
              </div>
              
              <div className="flex justify-end space-x-2 mt-6">
                <Button
                  variant="outline"
                  onClick={() => setShowExportModal(false)}
                  disabled={exporting}
                >
                  Cancel
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}