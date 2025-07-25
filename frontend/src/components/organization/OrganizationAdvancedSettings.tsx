'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';
import { 
  ShieldCheckIcon, 
  UsersIcon, 
  DocumentTextIcon, 
  ChartBarIcon,
  CogIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  PlusIcon,
  TrashIcon
} from '@heroicons/react/24/outline';

interface OrganizationSettingsProps {
  orgName: string;
}

interface CustomRole {
  id: string;
  name: string;
  description: string;
  permissions: Record<string, boolean | string | number>;
  color: string;
  is_default: boolean;
  created_at: string;
}

interface Policy {
  id: string;
  policy_type: string;
  name: string;
  description: string;
  enabled: boolean;
  enforcement: string;
  configuration: Record<string, boolean | string | number>;
  created_at: string;
}

interface OrganizationSettings {
  // Security Settings
  require_two_factor: boolean;
  allowed_ip_ranges: string[];
  sso_provider: string;
  session_timeout: number;
  
  // Repository Settings
  default_visibility: string;
  allow_private_repos: boolean;
  allow_internal_repos: boolean;
  allow_forking: boolean;
  allow_outside_collaborators: boolean;
  
  // Branding
  primary_color: string;
  secondary_color: string;
  logo_url: string;
  
  // Billing
  billing_plan: string;
  seat_count: number;
  storage_limit_gb: number;
  bandwidth_limit_gb: number;
}

export function OrganizationAdvancedSettings({ orgName }: OrganizationSettingsProps) {
  const [activeTab, setActiveTab] = useState<'roles' | 'policies' | 'settings' | 'analytics'>('roles');
  const [customRoles, setCustomRoles] = useState<CustomRole[]>([]);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [settings, setSettings] = useState<OrganizationSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Role management state
  const [showCreateRole, setShowCreateRole] = useState(false);
  const [newRole, setNewRole] = useState({
    name: '',
    description: '',
    permissions: {},
    color: '#6b7280'
  });

  // Policy management state
  const [showCreatePolicy, setShowCreatePolicy] = useState(false);
  const [newPolicy, setNewPolicy] = useState({
    policy_type: 'repository_creation',
    name: '',
    description: '',
    enabled: true,
    enforcement: 'warn',
    configuration: {}
  });

  useEffect(() => {
    fetchAdvancedSettings();
  }, [orgName]);

  const fetchAdvancedSettings = async () => {
    try {
      setLoading(true);
      const [rolesResponse, policiesResponse, settingsResponse] = await Promise.all([
        api.get(`/organizations/${orgName}/roles`),
        api.get(`/organizations/${orgName}/policies`),
        api.get(`/organizations/${orgName}/settings`)
      ]);
      
      setCustomRoles(rolesResponse.data.roles || []);
      setPolicies(policiesResponse.data.policies || []);
      setSettings(settingsResponse.data);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      setError(error.response?.data?.message || 'Failed to fetch advanced settings');
    } finally {
      setLoading(false);
    }
  };

  const createCustomRole = async () => {
    try {
      const response = await api.post(`/organizations/${orgName}/roles`, newRole);
      setCustomRoles([...customRoles, response.data]);
      setShowCreateRole(false);
      setNewRole({ name: '', description: '', permissions: {}, color: '#6b7280' });
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      setError(error.response?.data?.message || 'Failed to create custom role');
    }
  };

  const deleteCustomRole = async (roleId: string) => {
    try {
      await api.delete(`/organizations/${orgName}/roles/${roleId}`);
      setCustomRoles(customRoles.filter(role => role.id !== roleId));
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      setError(error.response?.data?.message || 'Failed to delete custom role');
    }
  };

  const createPolicy = async () => {
    try {
      const response = await api.post(`/organizations/${orgName}/policies`, newPolicy);
      setPolicies([...policies, response.data]);
      setShowCreatePolicy(false);
      setNewPolicy({
        policy_type: 'repository_creation',
        name: '',
        description: '',
        enabled: true,
        enforcement: 'warn',
        configuration: {}
      });
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      setError(error.response?.data?.message || 'Failed to create policy');
    }
  };

  const togglePolicy = async (policyId: string, enabled: boolean) => {
    try {
      await api.put(`/organizations/${orgName}/policies/${policyId}`, { enabled });
      setPolicies(policies.map(policy => 
        policy.id === policyId ? { ...policy, enabled } : policy
      ));
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      setError(error.response?.data?.message || 'Failed to update policy');
    }
  };

  const updateSettings = async (updatedSettings: Partial<OrganizationSettings>) => {
    try {
      const response = await api.put(`/organizations/${orgName}/settings`, updatedSettings);
      setSettings(response.data);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      setError(error.response?.data?.message || 'Failed to update settings');
    }
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/3 mb-6"></div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">Advanced Organization Settings</h1>
        <p className="text-muted-foreground">Manage roles, policies, security, and analytics for {orgName}</p>
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
            { id: 'roles', name: 'Custom Roles', icon: UsersIcon },
            { id: 'policies', name: 'Policies', icon: ShieldCheckIcon },
            { id: 'settings', name: 'Settings', icon: CogIcon },
            { id: 'analytics', name: 'Analytics', icon: ChartBarIcon }
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as 'roles' | 'policies' | 'settings' | 'analytics')}
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

      {/* Custom Roles Tab */}
      {activeTab === 'roles' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold text-foreground">Custom Roles</h2>
              <p className="text-muted-foreground">Create and manage custom roles with granular permissions</p>
            </div>
            <Button onClick={() => setShowCreateRole(true)}>
              <PlusIcon className="w-4 h-4 mr-2" />
              Create Role
            </Button>
          </div>

          {showCreateRole && (
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-medium mb-4">Create Custom Role</h3>
                <div className="grid grid-cols-1 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Name</label>
                    <Input
                      value={newRole.name}
                      onChange={(e) => setNewRole({ ...newRole, name: e.target.value })}
                      placeholder="Role name"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Description</label>
                    <Input
                      value={newRole.description}
                      onChange={(e) => setNewRole({ ...newRole, description: e.target.value })}
                      placeholder="Role description"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Color</label>
                    <input
                      type="color"
                      value={newRole.color}
                      onChange={(e) => setNewRole({ ...newRole, color: e.target.value })}
                      className="w-16 h-8 rounded border border-input"
                    />
                  </div>
                  <div className="flex space-x-2">
                    <Button onClick={createCustomRole}>Create Role</Button>
                    <Button variant="outline" onClick={() => setShowCreateRole(false)}>Cancel</Button>
                  </div>
                </div>
              </div>
            </Card>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {customRoles.map((role) => (
              <Card key={role.id}>
                <div className="p-6">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center">
                      <div 
                        className="w-4 h-4 rounded-full mr-3"
                        style={{ backgroundColor: role.color }}
                      ></div>
                      <div>
                        <h3 className="text-lg font-medium text-foreground">{role.name}</h3>
                        {role.is_default && (
                          <Badge variant="secondary" className="mt-1">Default</Badge>
                        )}
                      </div>
                    </div>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => deleteCustomRole(role.id)}
                    >
                      <TrashIcon className="w-4 h-4" />
                    </Button>
                  </div>
                  <p className="text-muted-foreground mb-4">{role.description}</p>
                  <div className="text-sm text-muted-foreground">
                    Created {new Date(role.created_at).toLocaleDateString()}
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Policies Tab */}
      {activeTab === 'policies' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold text-foreground">Organization Policies</h2>
              <p className="text-muted-foreground">Enforce rules and compliance across your organization</p>
            </div>
            <Button onClick={() => setShowCreatePolicy(true)}>
              <PlusIcon className="w-4 h-4 mr-2" />
              Create Policy
            </Button>
          </div>

          {showCreatePolicy && (
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-medium mb-4">Create Policy</h3>
                <div className="grid grid-cols-1 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Policy Type</label>
                    <select
                      value={newPolicy.policy_type}
                      onChange={(e) => setNewPolicy({ ...newPolicy, policy_type: e.target.value })}
                      className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground"
                    >
                      <option value="repository_creation">Repository Creation</option>
                      <option value="member_invitation">Member Invitation</option>
                      <option value="branch_protection">Branch Protection</option>
                      <option value="secret_management">Secret Management</option>
                      <option value="2fa_enforcement">2FA Enforcement</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Name</label>
                    <Input
                      value={newPolicy.name}
                      onChange={(e) => setNewPolicy({ ...newPolicy, name: e.target.value })}
                      placeholder="Policy name"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Description</label>
                    <Input
                      value={newPolicy.description}
                      onChange={(e) => setNewPolicy({ ...newPolicy, description: e.target.value })}
                      placeholder="Policy description"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">Enforcement</label>
                    <select
                      value={newPolicy.enforcement}
                      onChange={(e) => setNewPolicy({ ...newPolicy, enforcement: e.target.value })}
                      className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground"
                    >
                      <option value="warn">Warn</option>
                      <option value="block">Block</option>
                    </select>
                  </div>
                  <div className="flex space-x-2">
                    <Button onClick={createPolicy}>Create Policy</Button>
                    <Button variant="outline" onClick={() => setShowCreatePolicy(false)}>Cancel</Button>
                  </div>
                </div>
              </div>
            </Card>
          )}

          <div className="space-y-4">
            {policies.map((policy) => (
              <Card key={policy.id}>
                <div className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center mb-2">
                        <h3 className="text-lg font-medium text-foreground mr-3">{policy.name}</h3>
                        <Badge variant={policy.enabled ? 'default' : 'secondary'}>
                          {policy.enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                        <Badge variant="outline" className="ml-2">
                          {policy.enforcement}
                        </Badge>
                      </div>
                      <p className="text-muted-foreground mb-2">{policy.description}</p>
                      <div className="text-sm text-muted-foreground">
                        Type: {policy.policy_type.replace('_', ' ')} • 
                        Created {new Date(policy.created_at).toLocaleDateString()}
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        size="sm"
                        variant={policy.enabled ? "outline" : "default"}
                        onClick={() => togglePolicy(policy.id, !policy.enabled)}
                      >
                        {policy.enabled ? 'Disable' : 'Enable'}
                      </Button>
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Settings Tab */}
      {activeTab === 'settings' && settings && (
        <div className="space-y-8">
          <div>
            <h2 className="text-xl font-semibold text-foreground mb-4">Organization Settings</h2>
            
            {/* Security Settings */}
            <Card className="mb-6">
              <div className="p-6">
                <h3 className="text-lg font-medium text-foreground mb-4 flex items-center">
                  <ShieldCheckIcon className="w-5 h-5 mr-2" />
                  Security Settings
                </h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.require_two_factor}
                        onChange={(e) => updateSettings({ require_two_factor: e.target.checked })}
                        className="mr-2"
                      />
                      <span className="text-sm font-medium text-foreground">
                        Require Two-Factor Authentication
                      </span>
                    </label>
                    <p className="text-xs text-muted-foreground mt-1">
                      All organization members must enable 2FA
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                      Session Timeout (seconds)
                    </label>
                    <Input
                      type="number"
                      value={settings.session_timeout}
                      onChange={(e) => updateSettings({ session_timeout: parseInt(e.target.value) })}
                    />
                  </div>
                </div>
              </div>
            </Card>

            {/* Repository Settings */}
            <Card className="mb-6">
              <div className="p-6">
                <h3 className="text-lg font-medium text-foreground mb-4 flex items-center">
                  <DocumentTextIcon className="w-5 h-5 mr-2" />
                  Repository Settings
                </h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                      Default Visibility
                    </label>
                    <select
                      value={settings.default_visibility}
                      onChange={(e) => updateSettings({ default_visibility: e.target.value })}
                      className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground"
                    >
                      <option value="private">Private</option>
                      <option value="internal">Internal</option>
                      <option value="public">Public</option>
                    </select>
                  </div>
                  <div className="space-y-3">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.allow_private_repos}
                        onChange={(e) => updateSettings({ allow_private_repos: e.target.checked })}
                        className="mr-2"
                      />
                      <span className="text-sm font-medium text-foreground">Allow Private Repositories</span>
                    </label>
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.allow_forking}
                        onChange={(e) => updateSettings({ allow_forking: e.target.checked })}
                        className="mr-2"
                      />
                      <span className="text-sm font-medium text-foreground">Allow Forking</span>
                    </label>
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.allow_outside_collaborators}
                        onChange={(e) => updateSettings({ allow_outside_collaborators: e.target.checked })}
                        className="mr-2"
                      />
                      <span className="text-sm font-medium text-foreground">Allow Outside Collaborators</span>
                    </label>
                  </div>
                </div>
              </div>
            </Card>

            {/* Billing Information */}
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-medium text-foreground mb-4">Billing & Usage</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div>
                    <div className="text-2xl font-bold text-primary">{settings.billing_plan}</div>
                    <div className="text-sm text-muted-foreground">Current Plan</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-primary">{settings.seat_count}</div>
                    <div className="text-sm text-muted-foreground">Seats Used</div>
                  </div>
                  <div>
                    <div className="text-2xl font-bold text-primary">{settings.storage_limit_gb}GB</div>
                    <div className="text-sm text-muted-foreground">Storage Limit</div>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>
      )}

      {/* Analytics Tab */}
      {activeTab === 'analytics' && (
        <div className="space-y-6">
          <div>
            <h2 className="text-xl font-semibold text-foreground mb-4">Analytics & Insights</h2>
            <p className="text-muted-foreground">Monitor organization performance and compliance</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <Card>
              <div className="p-6 text-center">
                <div className="text-2xl font-bold text-primary">85%</div>
                <div className="text-sm text-muted-foreground">Security Score</div>
                <CheckCircleIcon className="w-6 h-6 mx-auto mt-2 text-green-500" />
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <div className="text-2xl font-bold text-primary">12</div>
                <div className="text-sm text-muted-foreground">Active Policies</div>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <div className="text-2xl font-bold text-primary">3</div>
                <div className="text-sm text-muted-foreground">Custom Roles</div>
              </div>
            </Card>
            <Card>
              <div className="p-6 text-center">
                <div className="text-2xl font-bold text-primary">0</div>
                <div className="text-sm text-muted-foreground">Policy Violations</div>
                <CheckCircleIcon className="w-6 h-6 mx-auto mt-2 text-green-500" />
              </div>
            </Card>
          </div>

          <Card>
            <div className="p-6">
              <h3 className="text-lg font-medium text-foreground mb-4">Compliance Status</h3>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">GDPR Compliance</span>
                  <Badge variant="default">Compliant</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">SOC 2 Type II</span>
                  <Badge variant="secondary">In Progress</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">ISO 27001</span>
                  <Badge variant="outline">Not Started</Badge>
                </div>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}