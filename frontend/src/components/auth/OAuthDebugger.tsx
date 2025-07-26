'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface OAuthProvider {
  id: string;
  name: string;
  client_id: string;
  auth_url: string;
  token_url: string;
  scope: string;
  is_active: boolean;
}

interface StateValidation {
  state_parameter: string;
  is_valid: boolean;
  created_at: string;
  expires_at: string;
  user_ip?: string;
  user_agent?: string;
}

interface OAuthFlow {
  id: string;
  provider: string;
  state: string;
  step: 'initiated' | 'auth_code_received' | 'token_exchange' | 'user_info_fetched' | 'completed' | 'failed';
  started_at: string;
  completed_at?: string;
  error_message?: string;
  debug_info: Record<string, any>;
}

export function OAuthDebugger() {
  const [providers, setProviders] = useState<OAuthProvider[]>([]);
  const [activeFlows, setActiveFlows] = useState<OAuthFlow[]>([]);
  const [stateValidations, setStateValidations] = useState<StateValidation[]>([]);
  
  const [loading, setLoading] = useState(true);
  const [selectedProvider, setSelectedProvider] = useState('');
  const [testState, setTestState] = useState('');
  const [testAuthCode, setTestAuthCode] = useState('');
  const [testResults, setTestResults] = useState<Record<string, any>>({});
  const [showTestModal, setShowTestModal] = useState(false);
  const [activeTab, setActiveTab] = useState<'flows' | 'states' | 'test'>('flows');

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [providersResponse, flowsResponse, statesResponse] = await Promise.all([
        api.get('/admin/auth/oauth/providers'),
        api.get('/admin/auth/oauth/debug/flows?limit=20'),
        api.get('/admin/auth/oauth/debug/states?limit=20')
      ]);

      setProviders(providersResponse.data.providers || []);
      setActiveFlows(flowsResponse.data.flows || []);
      setStateValidations(statesResponse.data.states || []);
    } catch (error) {
      console.error('Failed to fetch OAuth debug data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTestAuthFlow = async () => {
    if (!selectedProvider) return;

    try {
      const response = await api.post('/admin/auth/oauth/debug/test-flow', {
        provider: selectedProvider
      });
      
      setTestResults(prev => ({
        ...prev,
        authFlow: response.data
      }));
    } catch (error) {
      console.error('Failed to test auth flow:', error);
      setTestResults(prev => ({
        ...prev,
        authFlow: { error: 'Test failed' }
      }));
    }
  };

  const handleValidateState = async () => {
    if (!testState.trim()) return;

    try {
      const response = await api.post('/admin/auth/oauth/debug/validate-state', {
        state: testState
      });
      
      setTestResults(prev => ({
        ...prev,
        stateValidation: response.data
      }));
    } catch (error) {
      console.error('Failed to validate state:', error);
      setTestResults(prev => ({
        ...prev,
        stateValidation: { error: 'Validation failed' }
      }));
    }
  };

  const handleTestTokenExchange = async () => {
    if (!selectedProvider || !testAuthCode.trim()) return;

    try {
      const response = await api.post('/admin/auth/oauth/debug/test-token-exchange', {
        provider: selectedProvider,
        auth_code: testAuthCode,
        state: testState
      });
      
      setTestResults(prev => ({
        ...prev,
        tokenExchange: response.data
      }));
    } catch (error) {
      console.error('Failed to test token exchange:', error);
      setTestResults(prev => ({
        ...prev,
        tokenExchange: { error: 'Token exchange failed' }
      }));
    }
  };

  const handleClearFlow = async (flowId: string) => {
    try {
      await api.delete(`/admin/auth/oauth/debug/flows/${flowId}`);
      setActiveFlows(prev => prev.filter(flow => flow.id !== flowId));
    } catch (error) {
      console.error('Failed to clear flow:', error);
    }
  };

  const handleClearAllFlows = async () => {
    if (!confirm('Are you sure you want to clear all OAuth debug flows?')) {
      return;
    }

    try {
      await api.delete('/admin/auth/oauth/debug/flows');
      setActiveFlows([]);
    } catch (error) {
      console.error('Failed to clear flows:', error);
    }
  };

  const getStepBadgeColor = (step: string) => {
    switch (step) {
      case 'initiated': return 'bg-blue-100 text-blue-800';
      case 'auth_code_received': return 'bg-yellow-100 text-yellow-800';
      case 'token_exchange': return 'bg-purple-100 text-purple-800';
      case 'user_info_fetched': return 'bg-orange-100 text-orange-800';
      case 'completed': return 'bg-green-100 text-green-800';
      case 'failed': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const formatDebugInfo = (info: Record<string, any>) => {
    return JSON.stringify(info, null, 2);
  };

  return (
    <div className="space-y-6">
      {/* OAuth Providers Status */}
      <Card>
        <div className="p-6">
          <h3 className="text-lg font-semibold text-foreground mb-4">OAuth Providers Status</h3>
          
          {providers.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No OAuth providers configured
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {providers.map((provider) => (
                <div key={provider.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-foreground">{provider.name}</h4>
                      <p className="text-sm text-muted-foreground">Client ID: {provider.client_id}</p>
                      <p className="text-sm text-muted-foreground">Scope: {provider.scope}</p>
                    </div>
                    <Badge className={provider.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                      {provider.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Debug Tools */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <div className="flex space-x-1">
              {[
                { id: 'flows', name: 'OAuth Flows', icon: '🔄' },
                { id: 'states', name: 'State Validation', icon: '🔐' },
                { id: 'test', name: 'Test Tools', icon: '🧪' }
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
            
            {activeTab === 'flows' && (
              <Button
                variant="outline"
                onClick={handleClearAllFlows}
                className="text-red-600 hover:text-red-700"
              >
                Clear All Flows
              </Button>
            )}
          </div>

          {/* OAuth Flows Tab */}
          {activeTab === 'flows' && (
            <div className="space-y-3">
              {loading ? (
                <div className="text-center py-8">Loading OAuth flows...</div>
              ) : activeFlows.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No active OAuth flows
                </div>
              ) : (
                activeFlows.map((flow) => (
                  <div key={flow.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          <span className="font-medium text-foreground">{flow.provider}</span>
                          <Badge className={getStepBadgeColor(flow.step)}>
                            {flow.step}
                          </Badge>
                          <span className="text-sm text-muted-foreground">State: {flow.state.substring(0, 8)}...</span>
                        </div>
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>Started: {formatTimestamp(flow.started_at)}</span>
                            {flow.completed_at && (
                              <span>Completed: {formatTimestamp(flow.completed_at)}</span>
                            )}
                          </div>
                          {flow.error_message && (
                            <div className="mt-1 text-red-600 text-xs">
                              Error: {flow.error_message}
                            </div>
                          )}
                        </div>
                        {Object.keys(flow.debug_info).length > 0 && (
                          <details className="mt-2">
                            <summary className="text-sm cursor-pointer text-blue-600">Debug Info</summary>
                            <pre className="text-xs bg-gray-100 p-2 mt-1 rounded overflow-x-auto">
                              {formatDebugInfo(flow.debug_info)}
                            </pre>
                          </details>
                        )}
                      </div>

                      <div className="flex items-center space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleClearFlow(flow.id)}
                          className="text-red-600 hover:text-red-700"
                        >
                          Clear
                        </Button>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* State Validation Tab */}
          {activeTab === 'states' && (
            <div className="space-y-3">
              {loading ? (
                <div className="text-center py-8">Loading state validations...</div>
              ) : stateValidations.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No state validations found
                </div>
              ) : (
                stateValidations.map((validation, index) => (
                  <div key={index} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          <span className="font-mono text-sm">{validation.state_parameter.substring(0, 16)}...</span>
                          <Badge className={validation.is_valid ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                            {validation.is_valid ? 'Valid' : 'Invalid'}
                          </Badge>
                        </div>
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>Created: {formatTimestamp(validation.created_at)}</span>
                            <span>Expires: {formatTimestamp(validation.expires_at)}</span>
                            {validation.user_ip && (
                              <span>IP: {validation.user_ip}</span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}

          {/* Test Tools Tab */}
          {activeTab === 'test' && (
            <div className="space-y-6">
              {/* Provider Selection */}
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Select Provider for Testing
                </label>
                <select
                  value={selectedProvider}
                  onChange={(e) => setSelectedProvider(e.target.value)}
                  className="w-full px-3 py-2 border rounded-md"
                >
                  <option value="">Select a provider</option>
                  {providers.map((provider) => (
                    <option key={provider.id} value={provider.name}>
                      {provider.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* Test Auth Flow */}
              <Card className="p-4">
                <h4 className="font-medium text-foreground mb-3">Test Authorization Flow</h4>
                <div className="flex items-center space-x-2 mb-3">
                  <Button
                    onClick={handleTestAuthFlow}
                    disabled={!selectedProvider}
                  >
                    Test Auth Flow
                  </Button>
                </div>
                {testResults.authFlow && (
                  <div className="bg-gray-100 p-3 rounded">
                    <pre className="text-sm overflow-x-auto">
                      {JSON.stringify(testResults.authFlow, null, 2)}
                    </pre>
                  </div>
                )}
              </Card>

              {/* Test State Validation */}
              <Card className="p-4">
                <h4 className="font-medium text-foreground mb-3">Test State Parameter Validation</h4>
                <div className="flex items-center space-x-2 mb-3">
                  <Input
                    placeholder="Enter state parameter to validate"
                    value={testState}
                    onChange={(e) => setTestState(e.target.value)}
                    className="flex-1"
                  />
                  <Button
                    onClick={handleValidateState}
                    disabled={!testState.trim()}
                  >
                    Validate State
                  </Button>
                </div>
                {testResults.stateValidation && (
                  <div className="bg-gray-100 p-3 rounded">
                    <pre className="text-sm overflow-x-auto">
                      {JSON.stringify(testResults.stateValidation, null, 2)}
                    </pre>
                  </div>
                )}
              </Card>

              {/* Test Token Exchange */}
              <Card className="p-4">
                <h4 className="font-medium text-foreground mb-3">Test Token Exchange</h4>
                <div className="space-y-3">
                  <Input
                    placeholder="Enter authorization code"
                    value={testAuthCode}
                    onChange={(e) => setTestAuthCode(e.target.value)}
                  />
                  <div className="flex items-center space-x-2">
                    <Button
                      onClick={handleTestTokenExchange}
                      disabled={!selectedProvider || !testAuthCode.trim()}
                    >
                      Test Token Exchange
                    </Button>
                  </div>
                </div>
                {testResults.tokenExchange && (
                  <div className="bg-gray-100 p-3 rounded mt-3">
                    <pre className="text-sm overflow-x-auto">
                      {JSON.stringify(testResults.tokenExchange, null, 2)}
                    </pre>
                  </div>
                )}
              </Card>

              {/* Clear Test Results */}
              <div className="flex justify-end">
                <Button
                  variant="outline"
                  onClick={() => setTestResults({})}
                  disabled={Object.keys(testResults).length === 0}
                >
                  Clear Test Results
                </Button>
              </div>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}