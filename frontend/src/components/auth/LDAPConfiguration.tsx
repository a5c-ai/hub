'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import api from '@/lib/api';

interface LDAPConfig {
  id?: string;
  name: string;
  host: string;
  port: number;
  use_tls: boolean;
  use_ssl: boolean;
  skip_tls_verify: boolean;
  bind_dn: string;
  bind_password: string;
  user_search_base: string;
  user_search_filter: string;
  group_search_base?: string;
  group_search_filter?: string;
  user_attributes: {
    username: string;
    email: string;
    first_name?: string;
    last_name?: string;
    display_name?: string;
  };
  group_attributes: {
    name?: string;
    members?: string;
  };
  enabled: boolean;
  sync_enabled: boolean;
  sync_interval: number; // in hours
}

interface LDAPConfigurationProps {
  onSave?: (config: LDAPConfig) => void;
  initialConfig?: LDAPConfig;
}

export function LDAPConfiguration({ onSave, initialConfig }: LDAPConfigurationProps) {
  const [config, setConfig] = useState<LDAPConfig>({
    name: '',
    host: '',
    port: 389,
    use_tls: false,
    use_ssl: false,
    skip_tls_verify: false,
    bind_dn: '',
    bind_password: '',
    user_search_base: '',
    user_search_filter: '(&(objectClass=person)(uid={0}))',
    group_search_base: '',
    group_search_filter: '(&(objectClass=group)(member={0}))',
    user_attributes: {
      username: 'uid',
      email: 'mail',
      first_name: 'givenName',
      last_name: 'sn',
      display_name: 'displayName'
    },
    group_attributes: {
      name: 'cn',
      members: 'member'
    },
    enabled: false,
    sync_enabled: false,
    sync_interval: 24,
    ...initialConfig
  });

  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string; details?: unknown } | null>(null);

  const handleInputChange = (field: keyof LDAPConfig, value: string | number | boolean) => {
    setConfig(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleUserAttributeChange = (field: keyof LDAPConfig['user_attributes'], value: string) => {
    setConfig(prev => ({
      ...prev,
      user_attributes: {
        ...prev.user_attributes,
        [field]: value
      }
    }));
  };

  const handleGroupAttributeChange = (field: keyof LDAPConfig['group_attributes'], value: string) => {
    setConfig(prev => ({
      ...prev,
      group_attributes: {
        ...prev.group_attributes,
        [field]: value
      }
    }));
  };

  const handleTestConnection = async () => {
    setTesting(true);
    setTestResult(null);

    try {
      const response = await api.post('/auth/ldap/test', config);
      setTestResult({
        success: true,
        message: 'LDAP connection successful',
        details: response.data
      });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && 
        error.response !== null &&
        'data' in error.response &&
        typeof error.response.data === 'object' &&
        error.response.data !== null &&
        'error' in error.response.data &&
        typeof error.response.data.error === 'string'
        ? error.response.data.error
        : 'Failed to connect to LDAP server';
      
      setTestResult({
        success: false,
        message: errorMessage
      });
    } finally {
      setTesting(false);
    }
  };

  const handleTestUserSearch = async () => {
    const testUsername = prompt('Enter a username to test user search:');
    if (!testUsername) return;

    setTesting(true);
    try {
      const response = await api.post('/auth/ldap/test-user-search', {
        ...config,
        test_username: testUsername
      });
      
      setTestResult({
        success: true,
        message: `User search successful for: ${testUsername}`,
        details: response.data
      });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && 
        error.response !== null &&
        'data' in error.response &&
        typeof error.response.data === 'object' &&
        error.response.data !== null &&
        'error' in error.response.data &&
        typeof error.response.data.error === 'string'
        ? error.response.data.error
        : 'User search failed';
      
      setTestResult({
        success: false,
        message: errorMessage
      });
    } finally {
      setTesting(false);
    }
  };

  const handleSave = async () => {
    setLoading(true);
    try {
      const endpoint = config.id ? `/auth/ldap/config/${config.id}` : '/auth/ldap/config';
      const method = config.id ? 'put' : 'post';
      
      const response = await api[method](endpoint, config);
      
      if (onSave) {
        onSave(response.data);
      }
    } catch (error) {
      console.error('Failed to save LDAP configuration:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h3 className="text-lg font-semibold text-foreground">LDAP / Active Directory Configuration</h3>
              <p className="text-sm text-muted-foreground">Configure LDAP authentication for your organization</p>
            </div>
            <div className="flex items-center space-x-2">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={config.enabled}
                  onChange={(e) => handleInputChange('enabled', e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm font-medium">Enabled</span>
              </label>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-foreground mb-2">
                Configuration Name
              </label>
              <Input
                value={config.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                placeholder="e.g., Company Active Directory"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                LDAP Host
              </label>
              <Input
                value={config.host}
                onChange={(e) => handleInputChange('host', e.target.value)}
                placeholder="ldap.company.com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Port
              </label>
              <Input
                type="number"
                value={config.port}
                onChange={(e) => handleInputChange('port', parseInt(e.target.value))}
                placeholder="389"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-foreground mb-2">
                Security Options
              </label>
              <div className="flex flex-wrap gap-4">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.use_tls}
                    onChange={(e) => handleInputChange('use_tls', e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm">Use StartTLS</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.use_ssl}
                    onChange={(e) => handleInputChange('use_ssl', e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm">Use SSL/LDAPS</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={config.skip_tls_verify}
                    onChange={(e) => handleInputChange('skip_tls_verify', e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm">Skip TLS Verification</span>
                </label>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Bind DN
              </label>
              <Input
                value={config.bind_dn}
                onChange={(e) => handleInputChange('bind_dn', e.target.value)}
                placeholder="cn=admin,dc=company,dc=com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Bind Password
              </label>
              <Input
                type="password"
                value={config.bind_password}
                onChange={(e) => handleInputChange('bind_password', e.target.value)}
                placeholder="••••••••"
              />
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <div className="p-6">
          <h4 className="text-md font-semibold text-foreground mb-4">User Search Configuration</h4>
          
          <div className="grid grid-cols-1 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                User Search Base DN
              </label>
              <Input
                value={config.user_search_base}
                onChange={(e) => handleInputChange('user_search_base', e.target.value)}
                placeholder="ou=users,dc=company,dc=com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                User Search Filter
              </label>
              <Input
                value={config.user_search_filter}
                onChange={(e) => handleInputChange('user_search_filter', e.target.value)}
                placeholder="(&(objectClass=person)(uid={0}))"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Use {0} as placeholder for the username
              </p>
            </div>
          </div>

          <div className="mt-6">
            <h5 className="text-sm font-semibold text-foreground mb-3">User Attribute Mapping</h5>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Username Attribute
                </label>
                <Input
                  value={config.user_attributes.username}
                  onChange={(e) => handleUserAttributeChange('username', e.target.value)}
                  placeholder="uid"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Email Attribute
                </label>
                <Input
                  value={config.user_attributes.email}
                  onChange={(e) => handleUserAttributeChange('email', e.target.value)}
                  placeholder="mail"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  First Name Attribute
                </label>
                <Input
                  value={config.user_attributes.first_name || ''}
                  onChange={(e) => handleUserAttributeChange('first_name', e.target.value)}
                  placeholder="givenName"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Last Name Attribute
                </label>
                <Input
                  value={config.user_attributes.last_name || ''}
                  onChange={(e) => handleUserAttributeChange('last_name', e.target.value)}
                  placeholder="sn"
                />
              </div>
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h4 className="text-md font-semibold text-foreground">Group Synchronization</h4>
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={config.sync_enabled}
                onChange={(e) => handleInputChange('sync_enabled', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Enable Group Sync</span>
            </label>
          </div>

          {config.sync_enabled && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Group Search Base DN
                </label>
                <Input
                  value={config.group_search_base || ''}
                  onChange={(e) => handleInputChange('group_search_base', e.target.value)}
                  placeholder="ou=groups,dc=company,dc=com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Group Search Filter
                </label>
                <Input
                  value={config.group_search_filter || ''}
                  onChange={(e) => handleInputChange('group_search_filter', e.target.value)}
                  placeholder="(&(objectClass=group)(member={0}))"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Group Name Attribute
                  </label>
                  <Input
                    value={config.group_attributes.name || ''}
                    onChange={(e) => handleGroupAttributeChange('name', e.target.value)}
                    placeholder="cn"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Sync Interval (hours)
                  </label>
                  <Input
                    type="number"
                    value={config.sync_interval}
                    onChange={(e) => handleInputChange('sync_interval', parseInt(e.target.value))}
                    placeholder="24"
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      </Card>

      {testResult && (
        <Card className={testResult.success ? 'border-green-200' : 'border-red-200'}>
          <div className="p-4">
            <div className={`flex items-center space-x-2 ${testResult.success ? 'text-green-600' : 'text-red-600'}`}>
              <span className="font-medium">
                Test Result: {testResult.success ? 'Success' : 'Failed'}
              </span>
            </div>
            <p className="text-sm text-muted-foreground mt-1">{testResult.message}</p>
          </div>
        </Card>
      )}

      <div className="flex justify-between">
        <div className="flex space-x-2">
          <Button
            variant="outline"
            onClick={handleTestConnection}
            disabled={testing || !config.host}
          >
            {testing ? 'Testing...' : 'Test Connection'}
          </Button>
          
          <Button
            variant="outline"
            onClick={handleTestUserSearch}
            disabled={testing || !config.host}
          >
            Test User Search
          </Button>
        </div>

        <div className="flex space-x-3">
          <Button variant="outline">
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={loading}>
            {loading ? 'Saving...' : 'Save Configuration'}
          </Button>
        </div>
      </div>
    </div>
  );
}