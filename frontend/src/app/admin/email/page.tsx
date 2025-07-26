'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { apiClient } from '@/lib/api';

interface EmailConfig {
  smtp: {
    host: string;
    port: string;
    username: string;
    from: string;
    use_tls: boolean;
    configured: boolean;
  };
  application: {
    base_url: string;
    name: string;
  };
}

interface EmailHealth {
  status: string;
  configured: boolean;
  smtp: {
    host_configured: boolean;
    auth_configured: boolean;
    from_configured: boolean;
  };
  stats: {
    emails_sent_today: number;
    emails_sent_this_week: number;
    failed_emails_today: number;
    success_rate: number;
  };
  last_check: string;
}

interface EmailLog {
  id: string;
  to: string;
  subject: string;
  type: string;
  status: string;
  sent_at: string;
  delivered: boolean;
  error?: string;
}

export default function AdminEmailPage() {
  const [activeTab, setActiveTab] = useState('config');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [, setConfig] = useState<EmailConfig | null>(null);
  const [health, setHealth] = useState<EmailHealth | null>(null);
  const [logs, setLogs] = useState<EmailLog[]>([]);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  
  // Form state
  const [formData, setFormData] = useState({
    smtp: {
      host: '',
      port: '587',
      username: '',
      password: '',
      from: '',
      use_tls: true,
    },
    application: {
      base_url: '',
      name: 'A5C Hub',
    },
  });

  // Test email form
  const [testEmail, setTestEmail] = useState({
    to: '',
    subject: 'Test Email from A5C Hub',
    body: 'This is a test email to verify your SMTP configuration is working correctly.',
  });

  useEffect(() => {
    loadEmailData();
  }, []);

  const loadEmailData = async () => {
    try {
      setLoading(true);
      
      const [configResponse, healthResponse, logsResponse] = await Promise.all([
        apiClient.get<EmailConfig>('/admin/email/config'),
        apiClient.get<EmailHealth>('/admin/email/health'),
        apiClient.get<{ logs: EmailLog[] }>('/admin/email/logs')
      ]);

      if (configResponse.success && configResponse.data) {
        setConfig(configResponse.data);
        setFormData({
          smtp: {
            host: configResponse.data.smtp.host,
            port: configResponse.data.smtp.port,
            username: configResponse.data.smtp.username,
            password: '', // Don't populate password for security
            from: configResponse.data.smtp.from,
            use_tls: configResponse.data.smtp.use_tls,
          },
          application: {
            base_url: configResponse.data.application.base_url,
            name: configResponse.data.application.name,
          },
        });
      }

      if (healthResponse.success && healthResponse.data) {
        setHealth(healthResponse.data);
      }

      if (logsResponse.success && logsResponse.data) {
        setLogs(logsResponse.data.logs);
      }
    } catch (error) {
      console.error('Failed to load email data:', error);
      setMessage({ type: 'error', text: 'Failed to load email configuration' });
    } finally {
      setLoading(false);
    }
  };

  const saveConfig = async () => {
    try {
      setSaving(true);
      const response = await apiClient.put('/admin/email/config', formData);
      
      if (response.success) {
        setMessage({ type: 'success', text: 'Email configuration saved successfully' });
        await loadEmailData(); // Reload data
        setTimeout(() => setMessage(null), 3000);
      }
    } catch (error) {
      console.error('Failed to save config:', error);
      setMessage({ type: 'error', text: 'Failed to save email configuration' });
    } finally {
      setSaving(false);
    }
  };

  const sendTestEmail = async () => {
    try {
      setTesting(true);
      const response = await apiClient.post('/admin/email/test', testEmail);
      
      if (response.success) {
        setMessage({ type: 'success', text: 'Test email sent successfully' });
        await loadEmailData(); // Reload logs
        setTimeout(() => setMessage(null), 3000);
      }
    } catch (error) {
      console.error('Failed to send test email:', error);
      setMessage({ type: 'error', text: 'Failed to send test email' });
    } finally {
      setTesting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'healthy':
        return <Badge variant="default" className="bg-success text-success-foreground">Healthy</Badge>;
      case 'not_configured':
        return <Badge variant="outline" className="border-warning text-warning">Not Configured</Badge>;
      case 'error':
        return <Badge variant="destructive">Error</Badge>;
      default:
        return <Badge variant="outline">Unknown</Badge>;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const tabs = [
    { id: 'config', name: 'Configuration', icon: '⚙️' },
    { id: 'health', name: 'Health & Stats', icon: '📊' },
    { id: 'logs', name: 'Email Logs', icon: '📋' },
    { id: 'test', name: 'Test Email', icon: '🧪' },
  ];

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
            <div className="space-y-4">
              <div className="h-32 bg-muted rounded"></div>
              <div className="h-64 bg-muted rounded"></div>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Email Configuration</h1>
          <p className="text-muted-foreground mt-2">
            Manage SMTP settings, monitor email health, and view delivery logs
          </p>
        </div>

        {/* Message Display */}
        {message && (
          <div className={`mb-6 p-4 rounded-md ${
            message.type === 'success' 
              ? 'bg-success/10 text-success border border-success/20' 
              : 'bg-destructive/10 text-destructive border border-destructive/20'
          }`}>
            {message.text}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar */}
          <div className="lg:col-span-1">
            <nav className="space-y-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    activeTab === tab.id
                      ? 'bg-primary/10 text-primary border-r-2 border-primary'
                      : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                  }`}
                >
                  <span className="mr-3">{tab.icon}</span>
                  {tab.name}
                </button>
              ))}
            </nav>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-3">
            {activeTab === 'config' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">SMTP Configuration</h3>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          SMTP Host
                        </label>
                        <Input
                          value={formData.smtp.host}
                          onChange={(e) => setFormData({
                            ...formData,
                            smtp: { ...formData.smtp, host: e.target.value }
                          })}
                          placeholder="smtp.gmail.com"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Port
                        </label>
                        <Input
                          value={formData.smtp.port}
                          onChange={(e) => setFormData({
                            ...formData,
                            smtp: { ...formData.smtp, port: e.target.value }
                          })}
                          placeholder="587"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Username
                        </label>
                        <Input
                          value={formData.smtp.username}
                          onChange={(e) => setFormData({
                            ...formData,
                            smtp: { ...formData.smtp, username: e.target.value }
                          })}
                          placeholder="your-email@gmail.com"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Password
                        </label>
                        <Input
                          type="password"
                          value={formData.smtp.password}
                          onChange={(e) => setFormData({
                            ...formData,
                            smtp: { ...formData.smtp, password: e.target.value }
                          })}
                          placeholder="Enter password to update"
                        />
                      </div>
                      
                      <div className="md:col-span-2">
                        <label className="block text-sm font-medium text-foreground mb-2">
                          From Address
                        </label>
                        <Input
                          type="email"
                          value={formData.smtp.from}
                          onChange={(e) => setFormData({
                            ...formData,
                            smtp: { ...formData.smtp, from: e.target.value }
                          })}
                          placeholder="noreply@yourcompany.com"
                        />
                      </div>
                      
                      <div className="md:col-span-2">
                        <div className="flex items-center">
                          <input
                            type="checkbox"
                            id="use_tls"
                            checked={formData.smtp.use_tls}
                            onChange={(e) => setFormData({
                              ...formData,
                              smtp: { ...formData.smtp, use_tls: e.target.checked }
                            })}
                            className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                          />
                          <label htmlFor="use_tls" className="ml-2 text-sm font-medium text-foreground">
                            Use TLS/SSL encryption
                          </label>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>

                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Application Settings</h3>
                    
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Application Base URL
                        </label>
                        <Input
                          value={formData.application.base_url}
                          onChange={(e) => setFormData({
                            ...formData,
                            application: { ...formData.application, base_url: e.target.value }
                          })}
                          placeholder="https://your-domain.com"
                        />
                        <p className="text-sm text-muted-foreground mt-1">
                          Used for email links and verification URLs
                        </p>
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Application Name
                        </label>
                        <Input
                          value={formData.application.name}
                          onChange={(e) => setFormData({
                            ...formData,
                            application: { ...formData.application, name: e.target.value }
                          })}
                          placeholder="A5C Hub"
                        />
                        <p className="text-sm text-muted-foreground mt-1">
                          Used in email headers and templates
                        </p>
                      </div>
                    </div>
                  </div>
                </Card>

                <div className="flex justify-end">
                  <Button onClick={saveConfig} disabled={saving}>
                    {saving ? 'Saving...' : 'Save Configuration'}
                  </Button>
                </div>
              </div>
            )}

            {activeTab === 'health' && health && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Email Service Health</h3>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-4">
                        <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                          <div>
                            <h4 className="font-medium text-foreground">Service Status</h4>
                            <p className="text-sm text-muted-foreground">Overall email service health</p>
                          </div>
                          {getStatusBadge(health.status)}
                        </div>
                        
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">SMTP Host</span>
                            <Badge variant={health.smtp.host_configured ? 'default' : 'outline'}>
                              {health.smtp.host_configured ? 'Configured' : 'Not Set'}
                            </Badge>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">Authentication</span>
                            <Badge variant={health.smtp.auth_configured ? 'default' : 'outline'}>
                              {health.smtp.auth_configured ? 'Configured' : 'Not Set'}
                            </Badge>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">From Address</span>
                            <Badge variant={health.smtp.from_configured ? 'default' : 'outline'}>
                              {health.smtp.from_configured ? 'Configured' : 'Not Set'}
                            </Badge>
                          </div>
                        </div>
                      </div>
                      
                      <div className="space-y-4">
                        <h4 className="font-medium text-foreground">Email Statistics</h4>
                        <div className="space-y-3">
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">Emails sent today</span>
                            <span className="font-medium">{health.stats.emails_sent_today}</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">Emails sent this week</span>
                            <span className="font-medium">{health.stats.emails_sent_this_week}</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">Failed emails today</span>
                            <span className="font-medium text-destructive">{health.stats.failed_emails_today}</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-muted-foreground">Success rate</span>
                            <span className={`font-medium ${health.stats.success_rate >= 95 ? 'text-success' : 'text-warning'}`}>
                              {health.stats.success_rate}%
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <div className="mt-6 pt-4 border-t border-border">
                      <p className="text-sm text-muted-foreground">
                        Last health check: {formatDate(health.last_check)}
                      </p>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'logs' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Email Delivery Logs</h3>
                    
                    <div className="overflow-x-auto">
                      <table className="w-full">
                        <thead>
                          <tr className="border-b border-border">
                            <th className="text-left py-2 text-sm font-medium text-muted-foreground">Recipient</th>
                            <th className="text-left py-2 text-sm font-medium text-muted-foreground">Subject</th>
                            <th className="text-left py-2 text-sm font-medium text-muted-foreground">Type</th>
                            <th className="text-left py-2 text-sm font-medium text-muted-foreground">Status</th>
                            <th className="text-left py-2 text-sm font-medium text-muted-foreground">Sent At</th>
                          </tr>
                        </thead>
                        <tbody>
                          {logs.map((log) => (
                            <tr key={log.id} className="border-b border-border">
                              <td className="py-3 text-sm">{log.to}</td>
                              <td className="py-3 text-sm font-medium">{log.subject}</td>
                              <td className="py-3">
                                <Badge variant="outline" className="text-xs">
                                  {log.type}
                                </Badge>
                              </td>
                              <td className="py-3">
                                <Badge variant={log.delivered ? 'default' : 'destructive'} className="text-xs">
                                  {log.status}
                                </Badge>
                              </td>
                              <td className="py-3 text-sm text-muted-foreground">
                                {formatDate(log.sent_at)}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                      
                      {logs.length === 0 && (
                        <div className="text-center py-8 text-muted-foreground">
                          No email logs found
                        </div>
                      )}
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'test' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Send Test Email</h3>
                    <p className="text-muted-foreground mb-6">
                      Send a test email to verify your SMTP configuration is working correctly.
                    </p>
                    
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Recipient Email
                        </label>
                        <Input
                          type="email"
                          value={testEmail.to}
                          onChange={(e) => setTestEmail({ ...testEmail, to: e.target.value })}
                          placeholder="test@example.com"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Subject
                        </label>
                        <Input
                          value={testEmail.subject}
                          onChange={(e) => setTestEmail({ ...testEmail, subject: e.target.value })}
                          placeholder="Test Email from A5C Hub"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Message Body
                        </label>
                        <textarea
                          value={testEmail.body}
                          onChange={(e) => setTestEmail({ ...testEmail, body: e.target.value })}
                          rows={4}
                          className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
                          placeholder="Enter your test message here..."
                        />
                      </div>
                    </div>
                    
                    <div className="mt-6">
                      <Button 
                        onClick={sendTestEmail} 
                        disabled={testing || !testEmail.to}
                      >
                        {testing ? 'Sending...' : 'Send Test Email'}
                      </Button>
                    </div>
                  </div>
                </Card>
              </div>
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  );
}