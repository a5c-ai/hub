'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api, { repoApi } from '@/lib/api';
import { Repository } from '@/types';

export default function RepositorySettingsPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const [, setRepository] = useState<Repository | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState('general');
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    private: false,
    default_branch: 'main',
  });

  useEffect(() => {
    const fetchRepository = async () => {
      try {
        setLoading(true);
        const response = await repoApi.getRepositorySettings(owner, repo);
        const settings = response.data;
        setFormData({
          name: settings.name,
          description: settings.description || '',
          private: settings.private,
          default_branch: settings.default_branch,
        });
      } catch (err) {
        console.error('Failed to fetch repository', err);
      } finally {
        setLoading(false);
      }
    };

    fetchRepository();
  }, [owner, repo]);

  const handleSave = async () => {
    try {
      setSaving(true);
      await repoApi.updateRepositorySettings(owner, repo, formData);
      // Show success message
    } catch (err) {
      console.error('Failed to save repository settings', err);
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteRepository = async () => {
    if (window.confirm(`Are you sure you want to delete ${owner}/${repo}? This action cannot be undone.`)) {
      try {
        await api.delete(`/repositories/${owner}/${repo}`);
        window.location.href = '/repositories';
      } catch (err) {
        console.error('Failed to delete repository', err);
      }
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-8"></div>
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
              <div>
                <div className="h-64 bg-muted rounded"></div>
              </div>
              <div className="lg:col-span-3">
                <div className="h-96 bg-muted rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  const tabs = [
    { id: 'general', name: 'General', icon: '⚙️' },
    { id: 'access', name: 'Access', icon: '👥' },
    { id: 'branches', name: 'Branches', icon: '🌿' },
    { id: 'webhooks', name: 'Webhooks', icon: '🔗' },
    { id: 'security', name: 'Security', icon: '🔒' },
    { id: 'danger', name: 'Danger Zone', icon: '⚠️' }
  ];

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <Link href="/repositories" className="hover:text-foreground transition-colors">
            Repositories
          </Link>
          <span>/</span>
          <Link 
            href={`/repositories/${owner}/${repo}`}
            className="hover:text-foreground transition-colors"
          >
            {owner}/{repo}
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">Settings</span>
        </nav>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Repository Settings</h1>
            <p className="text-muted-foreground mt-2">Manage repository configuration and access</p>
          </div>
        </div>

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
                      ? 'bg-blue-100 text-blue-700 border-r-2 border-blue-500'
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
            {activeTab === 'general' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Repository Details</h3>
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Repository Name
                        </label>
                        <Input
                          value={formData.name}
                          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                          placeholder="Repository name"
                        />
                      </div>
                      
                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Description
                        </label>
                        <textarea
                          value={formData.description}
                          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                          placeholder="Short description of your repository"
                          rows={3}
                          className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring disabled:cursor-not-allowed disabled:opacity-50"
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-foreground mb-2">
                          Default Branch
                        </label>
                        <Input
                          value={formData.default_branch}
                          onChange={(e) => setFormData({ ...formData, default_branch: e.target.value })}
                          placeholder="main"
                        />
                      </div>

                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          id="private"
                          checked={formData.private}
                          onChange={(e) => setFormData({ ...formData, private: e.target.checked })}
                          className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
                        />
                        <label htmlFor="private" className="ml-2 block text-sm text-foreground">
                          Make this repository private
                        </label>
                      </div>
                    </div>
                    
                    <div className="mt-6">
                      <Button onClick={handleSave} disabled={saving}>
                        {saving ? 'Saving...' : 'Save Changes'}
                      </Button>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'access' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Repository Access</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Public Access</h4>
                          <p className="text-sm text-muted-foreground">Anyone can view this repository</p>
                        </div>
                        <Badge variant={!formData.private ? 'default' : 'secondary'}>
                          {!formData.private ? 'Enabled' : 'Disabled'}
                        </Badge>
                      </div>
                      
                      <div className="border-t pt-4">
                        <h4 className="font-medium text-foreground mb-3">Collaborators</h4>
                        <div className="flex items-center justify-between">
                          <p className="text-sm text-muted-foreground">Manage who can access this repository</p>
                          <Button size="sm">Add Collaborator</Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'branches' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Branch Protection</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Default Branch: {formData.default_branch}</h4>
                          <p className="text-sm text-muted-foreground">Configure protection rules for your default branch</p>
                        </div>
                        <Link href={`/repositories/${owner}/${repo}/settings/branches`}>
                          <Button size="sm" variant="outline">Configure</Button>
                        </Link>
                      </div>
                      
                      <div className="border-t pt-4">
                        <h4 className="font-medium text-foreground mb-3">Branch Protection Rules</h4>
                        <p className="text-sm text-muted-foreground mb-4">Protect branches by requiring status checks, reviews, or restrictions</p>
                        <Link href={`/repositories/${owner}/${repo}/settings/branches`}>
                          <Button size="sm">Manage Branch Protection</Button>
                        </Link>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'webhooks' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-lg font-semibold text-foreground">Webhooks</h3>
                      <Link href={`/repositories/${owner}/${repo}/settings/webhooks`}>
                        <Button size="sm">Manage Webhooks</Button>
                      </Link>
                    </div>
                    <p className="text-muted-foreground mb-4">
                      Webhooks allow external services to be notified when certain events happen
                    </p>
                    <div className="text-center py-8 text-muted-foreground">
                      <svg className="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                      </svg>
                      <p>Manage webhooks for this repository</p>
                      <Link href={`/repositories/${owner}/${repo}/settings/webhooks`}>
                        <Button variant="outline" size="sm" className="mt-3">
                          Go to Webhooks
                        </Button>
                      </Link>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'security' && (
              <div className="space-y-6">
                <Card>
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-foreground mb-4">Security Settings</h3>
                    <div className="space-y-4">
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Vulnerability Alerts</h4>
                          <p className="text-sm text-muted-foreground">Get notified about security vulnerabilities</p>
                        </div>
                        <input type="checkbox" defaultChecked className="h-4 w-4 text-primary focus:ring-primary border-border rounded" />
                      </div>
                      
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Dependency Graph</h4>
                          <p className="text-sm text-muted-foreground">Understand your dependencies</p>
                        </div>
                        <input type="checkbox" defaultChecked className="h-4 w-4 text-primary focus:ring-primary border-border rounded" />
                      </div>
                      
                      <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
                        <div>
                          <h4 className="font-medium text-foreground">Deploy Keys</h4>
                          <p className="text-sm text-muted-foreground">Manage SSH keys for repository access</p>
                        </div>
                        <Link href={`/repositories/${owner}/${repo}/settings/keys`}>
                          <Button size="sm" variant="outline">Manage Keys</Button>
                        </Link>
                      </div>
                    </div>
                  </div>
                </Card>
              </div>
            )}

            {activeTab === 'danger' && (
              <div className="space-y-6">
                <Card className="border-red-200">
                  <div className="p-6">
                    <h3 className="text-lg font-semibold text-red-900 mb-4">Danger Zone</h3>
                    <div className="space-y-4">
                      <div className="border border-red-200 rounded-lg p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-medium text-red-900">Transfer Repository</h4>
                            <p className="text-sm text-red-600">Transfer this repository to another user or organization</p>
                          </div>
                          <Button variant="outline" size="sm" className="border-red-300 text-red-700 hover:bg-red-50">
                            Transfer
                          </Button>
                        </div>
                      </div>
                      
                      <div className="border border-red-200 rounded-lg p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-medium text-red-900">Archive Repository</h4>
                            <p className="text-sm text-red-600">Make this repository read-only</p>
                          </div>
                          <Button variant="outline" size="sm" className="border-red-300 text-red-700 hover:bg-red-50">
                            Archive
                          </Button>
                        </div>
                      </div>
                      
                      <div className="border border-red-200 rounded-lg p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-medium text-red-900">Delete Repository</h4>
                            <p className="text-sm text-red-600">Permanently delete this repository and all of its contents</p>
                          </div>
                          <Button 
                            variant="outline" 
                            size="sm" 
                            className="border-red-300 text-red-700 hover:bg-red-50"
                            onClick={handleDeleteRepository}
                          >
                            Delete
                          </Button>
                        </div>
                      </div>
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
