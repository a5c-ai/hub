'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import api from '@/lib/api';

interface SAMLProvider {
  id?: string;
  name: string;
  metadata_url: string;
  entity_id: string;
  sso_url: string;
  slo_url?: string;
  certificate?: string;
  enabled: boolean;
  attribute_mappings: {
    email?: string;
    first_name?: string;
    last_name?: string;
    groups?: string;
  };
  just_in_time_provisioning: boolean;
}

interface SAMLConfigurationProps {
  onSave?: (config: SAMLProvider) => void;
  initialConfig?: SAMLProvider;
}

export function SAMLConfiguration({ onSave, initialConfig }: SAMLConfigurationProps) {
  const [config, setConfig] = useState<SAMLProvider>({
    name: '',
    metadata_url: '',
    entity_id: '',
    sso_url: '',
    slo_url: '',
    certificate: '',
    enabled: false,
    attribute_mappings: {
      email: 'email',
      first_name: 'firstName',
      last_name: 'lastName',
      groups: 'groups'
    },
    just_in_time_provisioning: false,
    ...initialConfig
  });

  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [metadataGeneratedUrl, setMetadataGeneratedUrl] = useState<string>('');

  useEffect(() => {
    // Generate metadata URL for this instance
    const baseUrl = window.location.origin;
    setMetadataGeneratedUrl(`${baseUrl}/api/v1/auth/saml/metadata`);
  }, []);

  const handleInputChange = (field: keyof SAMLProvider, value: string | boolean) => {
    setConfig(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleAttributeMappingChange = (field: keyof SAMLProvider['attribute_mappings'], value: string) => {
    setConfig(prev => ({
      ...prev,
      attribute_mappings: {
        ...prev.attribute_mappings,
        [field]: value
      }
    }));
  };

  const handleMetadataUrlLoad = async () => {
    if (!config.metadata_url) return;

    setLoading(true);
    try {
      const response = await api.post('/auth/saml/parse-metadata', {
        metadata_url: config.metadata_url
      });

      const metadata = response.data;
      setConfig(prev => ({
        ...prev,
        entity_id: metadata.entity_id || prev.entity_id,
        sso_url: metadata.sso_url || prev.sso_url,
        slo_url: metadata.slo_url || prev.slo_url,
        certificate: metadata.certificate || prev.certificate
      }));
    } catch (error) {
      console.error('Failed to parse metadata:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async () => {
    setTesting(true);
    setTestResult(null);

    try {
      const response = await api.post('/auth/saml/test', config);
      setTestResult({
        success: true,
        message: response.data.message || 'SAML configuration test successful'
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
        : 'Failed to test SAML configuration';
      
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
      const endpoint = config.id ? `/auth/saml/providers/${config.id}` : '/auth/saml/providers';
      const method = config.id ? 'put' : 'post';
      
      const response = await api[method](endpoint, config);
      
      if (onSave) {
        onSave(response.data);
      }
    } catch (error) {
      console.error('Failed to save SAML configuration:', error);
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
              <h3 className="text-lg font-semibold text-foreground">SAML 2.0 Configuration</h3>
              <p className="text-sm text-muted-foreground">Configure SAML Single Sign-On for your organization</p>
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
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Provider Name
              </label>
              <Input
                value={config.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                placeholder="e.g., Company Azure AD"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Metadata URL
              </label>
              <div className="flex space-x-2">
                <Input
                  value={config.metadata_url}
                  onChange={(e) => handleInputChange('metadata_url', e.target.value)}
                  placeholder="https://login.microsoftonline.com/..."
                />
                <Button
                  variant="outline"
                  onClick={handleMetadataUrlLoad}
                  disabled={loading || !config.metadata_url}
                >
                  Load
                </Button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Entity ID (Identifier)
              </label>
              <Input
                value={config.entity_id}
                onChange={(e) => handleInputChange('entity_id', e.target.value)}
                placeholder="urn:yourcompany:hub"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                SSO URL (Login URL)
              </label>
              <Input
                value={config.sso_url}
                onChange={(e) => handleInputChange('sso_url', e.target.value)}
                placeholder="https://login.microsoftonline.com/..."
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                SLO URL (Logout URL) <span className="text-muted-foreground">(Optional)</span>
              </label>
              <Input
                value={config.slo_url || ''}
                onChange={(e) => handleInputChange('slo_url', e.target.value)}
                placeholder="https://login.microsoftonline.com/..."
              />
            </div>
          </div>

          <div className="mt-6">
            <label className="block text-sm font-medium text-foreground mb-2">
              X.509 Certificate
            </label>
            <textarea
              value={config.certificate || ''}
              onChange={(e) => handleInputChange('certificate', e.target.value)}
              placeholder="-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKoK..."
              rows={4}
              className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent font-mono text-sm"
            />
          </div>
        </div>
      </Card>

      <Card>
        <div className="p-6">
          <h4 className="text-md font-semibold text-foreground mb-4">Attribute Mapping</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Email Attribute
              </label>
              <Input
                value={config.attribute_mappings.email || ''}
                onChange={(e) => handleAttributeMappingChange('email', e.target.value)}
                placeholder="email"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                First Name Attribute
              </label>
              <Input
                value={config.attribute_mappings.first_name || ''}
                onChange={(e) => handleAttributeMappingChange('first_name', e.target.value)}
                placeholder="firstName"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Last Name Attribute
              </label>
              <Input
                value={config.attribute_mappings.last_name || ''}
                onChange={(e) => handleAttributeMappingChange('last_name', e.target.value)}
                placeholder="lastName"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Groups Attribute
              </label>
              <Input
                value={config.attribute_mappings.groups || ''}
                onChange={(e) => handleAttributeMappingChange('groups', e.target.value)}
                placeholder="groups"
              />
            </div>
          </div>

          <div className="mt-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={config.just_in_time_provisioning}
                onChange={(e) => handleInputChange('just_in_time_provisioning', e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm font-medium">Enable Just-in-Time User Provisioning</span>
            </label>
            <p className="text-xs text-muted-foreground mt-1">
              Automatically create user accounts when users log in via SAML for the first time
            </p>
          </div>
        </div>
      </Card>

      <Card>
        <div className="p-6">
          <h4 className="text-md font-semibold text-foreground mb-4">Service Provider Information</h4>
          <p className="text-sm text-muted-foreground mb-4">
            Provide these details to your Identity Provider administrator:
          </p>
          
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                SP Entity ID / Audience
              </label>
              <div className="flex">
                <Input
                  value={`${window.location.origin}/api/v1/auth/saml/acs`}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="ml-2"
                  onClick={() => navigator.clipboard.writeText(`${window.location.origin}/api/v1/auth/saml/acs`)}
                >
                  Copy
                </Button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Assertion Consumer Service (ACS) URL
              </label>
              <div className="flex">
                <Input
                  value={`${window.location.origin}/api/v1/auth/saml/acs`}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="ml-2"
                  onClick={() => navigator.clipboard.writeText(`${window.location.origin}/api/v1/auth/saml/acs`)}
                >
                  Copy
                </Button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                SP Metadata URL
              </label>
              <div className="flex">
                <Input
                  value={metadataGeneratedUrl}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="ml-2"
                  onClick={() => navigator.clipboard.writeText(metadataGeneratedUrl)}
                >
                  Copy
                </Button>
              </div>
            </div>
          </div>
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
        <Button
          variant="outline"
          onClick={handleTestConnection}
          disabled={testing || !config.sso_url}
        >
          {testing ? 'Testing...' : 'Test Connection'}
        </Button>

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