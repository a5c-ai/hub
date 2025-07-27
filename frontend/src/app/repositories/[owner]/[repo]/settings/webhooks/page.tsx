'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import api from '@/lib/api';
import { createErrorHandler } from '@/lib/utils';


interface Webhook {
  id: string;
  name: string;
  config: {
    url: string;
    content_type: 'json' | 'form';
    secret?: string;
  };
  events: string[];
  active: boolean;
  created_at: string;
  updated_at: string;
}

export default function WebhooksPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [selectedWebhook, setSelectedWebhook] = useState<Webhook | null>(null);
  const [testingWebhook, setTestingWebhook] = useState<Webhook | null>(null);
  
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    content_type: 'json' as 'json' | 'form',
    secret: '',
    events: ['push'] as string[],
    active: true
  });

  const availableEvents = [
    'push', 'pull_request', 'issues', 'issue_comment', 'create', 'delete',
    'fork', 'star', 'watch', 'release', 'pull_request_review', 'pull_request_review_comment'
  ];

  const fetchWebhooks = async () => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
      const response = await api.get(`/repositories/${owner}/${repo}/hooks`);
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setWebhooks(result);
    }
  };

  useEffect(() => {
    fetchWebhooks();
  }, [owner, repo]);

  const handleCreateWebhook = async () => {
    if (!formData.name || !formData.url) return;

    const handleError = createErrorHandler(setError, setCreateLoading);
    
    const operation = async () => {
      const response = await api.post(`/repositories/${owner}/${repo}/hooks`, {
        name: formData.name,
        config: {
          url: formData.url,
          content_type: formData.content_type,
          secret: formData.secret || undefined
        },
        events: formData.events,
        active: formData.active
      });
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setWebhooks([...webhooks, result]);
      setShowCreateModal(false);
      setFormData({
        name: '',
        url: '',
        content_type: 'json',
        secret: '',
        events: ['push'],
        active: true
      });
    }
  };

  const handleDeleteWebhook = async (webhookId: string) => {
    if (!confirm('Are you sure you want to delete this webhook?')) return;

    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      await api.delete(`/repositories/${owner}/${repo}/hooks/${webhookId}`);
    };

    const result = await handleError(operation);
    if (result !== null) {
      setWebhooks(webhooks.filter(w => w.id !== webhookId));
    }
  };

  const handleToggleWebhook = async (webhookId: string, active: boolean) => {
    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      const response = await api.patch(`/repositories/${owner}/${repo}/hooks/${webhookId}`, {
        active: !active
      });
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setWebhooks(webhooks.map(w => w.id === webhookId ? { ...w, active: !active } : w));
    }
  };

  const handlePingWebhook = async (webhookId: string) => {
    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      await api.post(`/repositories/${owner}/${repo}/hooks/${webhookId}/pings`);
    };

    const result = await handleError(operation);
    if (result !== null) {
      alert('Ping sent successfully!');
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse space-y-4">
            <div className="h-8 bg-muted rounded w-1/3"></div>
            <div className="h-32 bg-muted rounded"></div>
            <div className="h-32 bg-muted rounded"></div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error}</div>
            <Button onClick={fetchWebhooks} disabled={loading}>
              {loading ? 'Retrying...' : 'Try Again'}
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

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
          <Link 
            href={`/repositories/${owner}/${repo}/settings`}
            className="hover:text-foreground transition-colors"
          >
            Settings
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">Webhooks</span>
        </nav>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Webhooks</h1>
            <p className="text-muted-foreground mt-2">
              Webhooks allow external services to be notified when certain events happen
            </p>
          </div>
          <Button onClick={() => setShowCreateModal(true)}>
            Add Webhook
          </Button>
        </div>

        {/* Webhooks List */}
        <div className="space-y-4">
          {webhooks.length === 0 ? (
            <Card>
              <div className="p-8 text-center">
                <svg className="w-12 h-12 mx-auto mb-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                </svg>
                <h3 className="text-lg font-medium text-foreground mb-2">No webhooks</h3>
                <p className="text-muted-foreground">
                  Get started by creating your first webhook
                </p>
              </div>
            </Card>
          ) : (
            webhooks.map((webhook) => (
              <Card key={webhook.id}>
                <div className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <h3 className="text-lg font-semibold text-foreground">{webhook.name}</h3>
                        <Badge variant={webhook.active ? 'default' : 'secondary'}>
                          {webhook.active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mb-3">{webhook.config.url}</p>
                      <div className="flex flex-wrap gap-2 mb-3">
                        {webhook.events.map((event) => (
                          <Badge key={event} variant="outline" className="text-xs">
                            {event}
                          </Badge>
                        ))}
                      </div>
                      <p className="text-xs text-muted-foreground">
                        Created {new Date(webhook.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handlePingWebhook(webhook.id)}
                      >
                        Ping
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setTestingWebhook(webhook)}
                      >
                        Test
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleToggleWebhook(webhook.id, webhook.active)}
                      >
                        {webhook.active ? 'Disable' : 'Enable'}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setSelectedWebhook(webhook)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDeleteWebhook(webhook.id)}
                        className="text-red-600 hover:bg-red-50"
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        {/* Create Webhook Modal */}
        <Modal 
          open={showCreateModal} 
          onClose={() => setShowCreateModal(false)}
          title="Add Webhook"
        >
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Name
              </label>
              <Input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Webhook name"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Payload URL
              </label>
              <Input
                value={formData.url}
                onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                placeholder="https://example.com/webhook"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Content Type
              </label>
              <select
                value={formData.content_type}
                onChange={(e) => setFormData({ ...formData, content_type: e.target.value as 'json' | 'form' })}
                className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground"
              >
                <option value="json">application/json</option>
                <option value="form">application/x-www-form-urlencoded</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Secret (optional)
              </label>
              <Input
                type="password"
                value={formData.secret}
                onChange={(e) => setFormData({ ...formData, secret: e.target.value })}
                placeholder="Secret for webhook validation"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Events
              </label>
              <div className="grid grid-cols-2 gap-2 max-h-40 overflow-y-auto border border-input rounded-md p-3">
                {availableEvents.map((event) => (
                  <label key={event} className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={formData.events.includes(event)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setFormData({
                            ...formData,
                            events: [...formData.events, event]
                          });
                        } else {
                          setFormData({
                            ...formData,
                            events: formData.events.filter(e => e !== event)
                          });
                        }
                      }}
                      className="rounded border-border"
                    />
                    <span className="text-sm">{event}</span>
                  </label>
                ))}
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="active"
                checked={formData.active}
                onChange={(e) => setFormData({ ...formData, active: e.target.checked })}
                className="rounded border-border"
              />
              <label htmlFor="active" className="text-sm font-medium text-foreground">
                Active
              </label>
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <Button variant="outline" onClick={() => setShowCreateModal(false)}>
                Cancel
              </Button>
              <Button 
                onClick={handleCreateWebhook} 
                disabled={createLoading || !formData.name || !formData.url}
              >
                {createLoading ? 'Creating...' : 'Create Webhook'}
              </Button>
            </div>
          </div>
        </Modal>


      </div>
    </AppLayout>
  );
}