'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useAuthStore } from '@/store/auth';
import { apiClient } from '@/lib/api';

interface EmailPreferences {
  issues_and_prs: boolean;
  repository_updates: boolean;
  security_alerts: boolean;
  mfa_notifications: boolean;
  password_reset: boolean;
  email_verification: boolean;
  weekly_digest: boolean;
  marketing_emails: boolean;
}

interface EmailVerificationStatus {
  verified: boolean;
  pending_token?: {
    created_at: string;
    expires_at: string;
  };
}

export default function EmailSettingsPage() {
  const { user } = useAuthStore();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [sending, setSending] = useState(false);
  const [preferences, setPreferences] = useState<EmailPreferences>({
    issues_and_prs: true,
    repository_updates: true,
    security_alerts: true,
    mfa_notifications: true,
    password_reset: true,
    email_verification: true,
    weekly_digest: false,
    marketing_emails: false,
  });
  const [verificationStatus, setVerificationStatus] = useState<EmailVerificationStatus | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    loadEmailData();
  }, []);

  const loadEmailData = async () => {
    try {
      setLoading(true);
      
      // Load preferences and verification status in parallel
      const [preferencesResponse, statusResponse] = await Promise.all([
        apiClient.get<EmailPreferences>('/user/email/preferences'),
        apiClient.get<EmailVerificationStatus>('/user/email/verification-status')
      ]);

      if (preferencesResponse.success && preferencesResponse.data) {
        setPreferences(preferencesResponse.data);
      }

      if (statusResponse.success && statusResponse.data) {
        setVerificationStatus(statusResponse.data);
      }
    } catch (error) {
      console.error('Failed to load email data:', error);
      setMessage({ type: 'error', text: 'Failed to load email settings' });
    } finally {
      setLoading(false);
    }
  };

  const savePreferences = async () => {
    try {
      setSaving(true);
      const response = await apiClient.put('/user/email/preferences', preferences);
      
      if (response.success) {
        setMessage({ type: 'success', text: 'Email preferences saved successfully' });
        setTimeout(() => setMessage(null), 3000);
      }
    } catch (error) {
      console.error('Failed to save preferences:', error);
      setMessage({ type: 'error', text: 'Failed to save email preferences' });
    } finally {
      setSaving(false);
    }
  };

  const resendVerificationEmail = async () => {
    try {
      setSending(true);
      const response = await apiClient.post('/user/email/resend-verification');
      
      if (response.success) {
        setMessage({ type: 'success', text: 'Verification email sent successfully' });
        // Refresh verification status
        await loadEmailData();
        setTimeout(() => setMessage(null), 3000);
      }
    } catch (error) {
      console.error('Failed to send verification email:', error);
      setMessage({ type: 'error', text: 'Failed to send verification email' });
    } finally {
      setSending(false);
    }
  };

  const handlePreferenceChange = (key: keyof EmailPreferences, value: boolean) => {
    setPreferences(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
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
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Email Settings</h1>
          <p className="text-muted-foreground mt-2">
            Manage your email preferences and verification status
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

        <div className="space-y-6">
          {/* Email Verification Status */}
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Email Verification</h3>
              
              <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                <div>
                  <h4 className="font-medium text-foreground">Email Address</h4>
                  <p className="text-sm text-muted-foreground">{user?.email}</p>
                  <div className="flex items-center mt-2 space-x-2">
                    {verificationStatus?.verified ? (
                      <Badge variant="default" className="bg-success text-success-foreground">
                        Verified
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="border-warning text-warning">
                        Unverified
                      </Badge>
                    )}
                  </div>
                </div>
                
                {!verificationStatus?.verified && (
                  <div className="text-right">
                    <Button 
                      onClick={resendVerificationEmail} 
                      disabled={sending}
                      size="sm"
                    >
                      {sending ? 'Sending...' : 'Resend Verification'}
                    </Button>
                    {verificationStatus?.pending_token && (
                      <p className="text-xs text-muted-foreground mt-2">
                        Expires: {formatDate(verificationStatus.pending_token.expires_at)}
                      </p>
                    )}
                  </div>
                )}
              </div>

              {!verificationStatus?.verified && (
                <div className="mt-4 p-4 bg-warning/10 border border-warning/20 rounded-lg">
                  <p className="text-sm text-warning-foreground">
                    <strong>Important:</strong> Please verify your email address to receive important 
                    security notifications and enable all email features.
                  </p>
                </div>
              )}
            </div>
          </Card>

          {/* Email Preferences */}
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Notification Preferences</h3>
              
              <div className="space-y-4">
                {/* Essential Notifications */}
                <div>
                  <h4 className="font-medium text-foreground mb-3">Essential Notifications</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Security Alerts</h5>
                        <p className="text-sm text-muted-foreground">
                          Critical security events and vulnerability notifications
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.security_alerts}
                        onChange={(e) => handlePreferenceChange('security_alerts', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Password Reset</h5>
                        <p className="text-sm text-muted-foreground">
                          Password reset confirmation emails
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.password_reset}
                        onChange={(e) => handlePreferenceChange('password_reset', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Email Verification</h5>
                        <p className="text-sm text-muted-foreground">
                          Email address verification messages
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.email_verification}
                        onChange={(e) => handlePreferenceChange('email_verification', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Two-Factor Authentication</h5>
                        <p className="text-sm text-muted-foreground">
                          MFA setup confirmations and backup codes
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.mfa_notifications}
                        onChange={(e) => handlePreferenceChange('mfa_notifications', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>
                  </div>
                </div>

                {/* Activity Notifications */}
                <div className="pt-4 border-t border-border">
                  <h4 className="font-medium text-foreground mb-3">Activity Notifications</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Issues and Pull Requests</h5>
                        <p className="text-sm text-muted-foreground">
                          Notifications about issues and pull requests you&apos;re involved in
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.issues_and_prs}
                        onChange={(e) => handlePreferenceChange('issues_and_prs', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Repository Updates</h5>
                        <p className="text-sm text-muted-foreground">
                          Updates from repositories you&apos;re watching
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.repository_updates}
                        onChange={(e) => handlePreferenceChange('repository_updates', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>
                  </div>
                </div>

                {/* Optional Notifications */}
                <div className="pt-4 border-t border-border">
                  <h4 className="font-medium text-foreground mb-3">Optional Notifications</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Weekly Digest</h5>
                        <p className="text-sm text-muted-foreground">
                          Weekly summary of your activity and updates
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.weekly_digest}
                        onChange={(e) => handlePreferenceChange('weekly_digest', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <div>
                        <h5 className="font-medium text-foreground">Product Updates</h5>
                        <p className="text-sm text-muted-foreground">
                          New features, tips, and product announcements
                        </p>
                      </div>
                      <input
                        type="checkbox"
                        checked={preferences.marketing_emails}
                        onChange={(e) => handlePreferenceChange('marketing_emails', e.target.checked)}
                        className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div className="mt-6 pt-6 border-t border-border">
                <Button onClick={savePreferences} disabled={saving}>
                  {saving ? 'Saving...' : 'Save Preferences'}
                </Button>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
}