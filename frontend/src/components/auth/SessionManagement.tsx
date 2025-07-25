'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface UserSession {
  id: string;
  user_id: string;
  username: string;
  device_info: {
    browser?: string;
    os?: string;
    device_type?: string;
    ip_address: string;
    user_agent: string;
  };
  location?: {
    country?: string;
    city?: string;
    region?: string;
  };
  created_at: string;
  last_activity: string;
  is_current: boolean;
  is_suspicious: boolean;
}

interface SessionSettings {
  session_timeout: number; // in minutes
  max_concurrent_sessions: number;
  require_reauth_for_sensitive: boolean;
  track_location: boolean;
  alert_on_new_device: boolean;
  alert_on_suspicious_activity: boolean;
}

interface SessionManagementProps {
  userId?: string;
  isAdminView?: boolean;
}

export function SessionManagement({ userId, isAdminView = false }: SessionManagementProps) {
  const [sessions, setSessions] = useState<UserSession[]>([]);
  const [settings, setSettings] = useState<SessionSettings>({
    session_timeout: 480, // 8 hours
    max_concurrent_sessions: 5,
    require_reauth_for_sensitive: true,
    track_location: true,
    alert_on_new_device: true,
    alert_on_suspicious_activity: true
  });
  
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSessions, setSelectedSessions] = useState<string[]>([]);

  useEffect(() => {
    fetchSessions();
    fetchSettings();
  }, [userId]);

  const fetchSessions = async () => {
    try {
      const endpoint = isAdminView 
        ? `/admin/sessions${userId ? `?user_id=${userId}` : ''}`
        : '/user/sessions';
      
      const response = await api.get(endpoint);
      setSessions(response.data.sessions || []);
    } catch (error) {
      console.error('Failed to fetch sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchSettings = async () => {
    try {
      const endpoint = isAdminView ? '/admin/session-settings' : '/user/session-settings';
      const response = await api.get(endpoint);
      setSettings(response.data);
    } catch (error) {
      console.error('Failed to fetch session settings:', error);
    }
  };

  const handleSettingChange = (field: keyof SessionSettings, value: number | boolean) => {
    setSettings(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSaveSettings = async () => {
    setSaving(true);
    try {
      const endpoint = isAdminView ? '/admin/session-settings' : '/user/session-settings';
      await api.put(endpoint, settings);
    } catch (error) {
      console.error('Failed to save session settings:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleRevokeSession = async (sessionId: string) => {
    try {
      const endpoint = isAdminView 
        ? `/admin/sessions/${sessionId}/revoke`
        : `/user/sessions/${sessionId}/revoke`;
      
      await api.post(endpoint);
      setSessions(prev => prev.filter(s => s.id !== sessionId));
    } catch (error) {
      console.error('Failed to revoke session:', error);
    }
  };

  const handleRevokeSelected = async () => {
    if (selectedSessions.length === 0) return;
    
    try {
      const endpoint = isAdminView ? '/admin/sessions/revoke-multiple' : '/user/sessions/revoke-multiple';
      await api.post(endpoint, { session_ids: selectedSessions });
      
      setSessions(prev => prev.filter(s => !selectedSessions.includes(s.id)));
      setSelectedSessions([]);
    } catch (error) {
      console.error('Failed to revoke sessions:', error);
    }
  };

  const handleRevokeAllOthers = async () => {
    if (!confirm('Are you sure you want to revoke all other sessions? This will log out all other devices.')) {
      return;
    }

    try {
      const endpoint = isAdminView ? '/admin/sessions/revoke-others' : '/user/sessions/revoke-others';
      await api.post(endpoint, userId ? { user_id: userId } : {});
      
      setSessions(prev => prev.filter(s => s.is_current));
    } catch (error) {
      console.error('Failed to revoke other sessions:', error);
    }
  };

  const toggleSessionSelection = (sessionId: string) => {
    setSelectedSessions(prev => 
      prev.includes(sessionId)
        ? prev.filter(id => id !== sessionId)
        : [...prev, sessionId]
    );
  };

  const filteredSessions = sessions.filter(session =>
    session.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    session.device_info.ip_address.includes(searchQuery) ||
    session.device_info.browser?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatLastActivity = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  return (
    <div className="space-y-6">
      {/* Session Settings */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-semibold text-foreground mb-4">Session Settings</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Session Timeout (minutes)
              </label>
              <Input
                type="number"
                value={settings.session_timeout}
                onChange={(e) => handleSettingChange('session_timeout', parseInt(e.target.value))}
                min="5"
                max="10080"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Sessions will expire after this period of inactivity
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Max Concurrent Sessions
              </label>
              <Input
                type="number"
                value={settings.max_concurrent_sessions}
                onChange={(e) => handleSettingChange('max_concurrent_sessions', parseInt(e.target.value))}
                min="1"
                max="20"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Maximum number of active sessions per user
              </p>
            </div>
          </div>

          <div className="mt-6 space-y-3">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.require_reauth_for_sensitive}
                onChange={(e) => handleSettingChange('require_reauth_for_sensitive', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Require re-authentication for sensitive operations</span>
            </label>

            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.track_location}
                onChange={(e) => handleSettingChange('track_location', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Track session location</span>
            </label>

            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.alert_on_new_device}
                onChange={(e) => handleSettingChange('alert_on_new_device', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Alert on new device login</span>
            </label>

            <label className="flex items-center">
              <input
                type="checkbox"
                checked={settings.alert_on_suspicious_activity}
                onChange={(e) => handleSettingChange('alert_on_suspicious_activity', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Alert on suspicious activity</span>
            </label>
          </div>

          <div className="mt-6">
            <Button onClick={handleSaveSettings} disabled={saving}>
              {saving ? 'Saving...' : 'Save Settings'}
            </Button>
          </div>
        </div>
      </Card>

      {/* Active Sessions */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">Active Sessions</h3>
            <div className="flex space-x-2">
              {selectedSessions.length > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRevokeSelected}
                  className="text-red-600 hover:text-red-700"
                >
                  Revoke Selected ({selectedSessions.length})
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={handleRevokeAllOthers}
                className="text-red-600 hover:text-red-700"
              >
                Revoke All Others
              </Button>
            </div>
          </div>

          <div className="mb-4">
            <Input
              placeholder="Search sessions by username, IP, or browser..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-md"
            />
          </div>

          {loading ? (
            <div className="text-center py-8">Loading sessions...</div>
          ) : filteredSessions.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No sessions found
            </div>
          ) : (
            <div className="space-y-3">
              {filteredSessions.map((session) => (
                <div
                  key={session.id}
                  className={`border rounded-lg p-4 ${session.is_current ? 'border-green-200 bg-green-50' : 'border-border'} ${session.is_suspicious ? 'border-red-200 bg-red-50' : ''}`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <input
                        type="checkbox"
                        checked={selectedSessions.includes(session.id)}
                        onChange={() => toggleSessionSelection(session.id)}
                        disabled={session.is_current}
                        className="rounded"
                      />
                      
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          {isAdminView && (
                            <span className="font-medium text-foreground">{session.username}</span>
                          )}
                          
                          <div className="flex items-center space-x-1">
                            {session.is_current && (
                              <Badge variant="default" className="bg-green-100 text-green-800">
                                Current
                              </Badge>
                            )}
                            {session.is_suspicious && (
                              <Badge variant="destructive">
                                Suspicious
                              </Badge>
                            )}
                          </div>
                        </div>
                        
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>{session.device_info.browser} on {session.device_info.os}</span>
                            <span>{session.device_info.ip_address}</span>
                            {session.location && (
                              <span>{session.location.city}, {session.location.country}</span>
                            )}
                          </div>
                          <div className="mt-1">
                            Last activity: {formatLastActivity(session.last_activity)}
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      {!session.is_current && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleRevokeSession(session.id)}
                          className="text-red-600 hover:text-red-700"
                        >
                          Revoke
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
}