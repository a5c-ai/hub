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
    allowlist_ips?: string[];
  };
  events: string[];
  active: boolean;
  created_at: string;
  updated_at: string;
}

export default function OrgWebhooksPage() {
  const params = useParams();
  const org = params.org as string;

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
    allowlist_ips: [] as string[],
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
      const response = await api.get(`/organizations/${org}/hooks`);
      return response.data;
    };
    const result = await handleError(operation);
    if (result) {
      setWebhooks(result);
    }
  };

  useEffect(() => {
    fetchWebhooks();
  }, [org]);

  const handleCreateWebhook = async () => {
    if (!formData.name || !formData.url) return;
    const handleError = createErrorHandler(setError, setCreateLoading);
    const operation = async () => {
      const response = await api.post(`/organizations/${org}/hooks`, {
        name: formData.name,
        config: {
          url: formData.url,
          content_type: formData.content_type,
          secret: formData.secret || undefined,
          allowlist_ips: formData.allowlist_ips.length ? formData.allowlist_ips : undefined
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
        name: '', url: '', content_type: 'json', secret: '', allowlist_ips: [], events: ['push'], active: true
      });
    }
  };

  const handleDeleteWebhook = async (webhookId: string) => {
    if (!confirm('Are you sure you want to delete this webhook?')) return;
    const handleError = createErrorHandler(setError);
    const operation = async () => {
      await api.delete(`/organizations/${org}/hooks/${webhookId}`);
    };
    const result = await handleError(operation);
    if (result !== null) {
      setWebhooks(webhooks.filter(w => w.id !== webhookId));
    }
  };

  const handleToggleWebhook = async (webhookId: string, active: boolean) => {
    const handleError = createErrorHandler(setError);
    const operation = async () => {
      const response = await api.patch(`/organizations/${org}/hooks/${webhookId}`, { active: !active });
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
      await api.post(`/organizations/${org}/hooks/${webhookId}/pings`);
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
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <Link href="/organizations" className="hover:text-foreground transition-colors">
            Organizations
          </Link>
          <span>/</span>
          <Link href={`/organizations/${org}`} className="hover:text-foreground transition-colors">
            {org}
          </Link>
          <span>/</span>
          <Link href={`/organizations/${org}/settings`} className="hover:text-foreground transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">Webhooks</span>
        </nav>

        <div className="space-y-6">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-3xl font-bold text-foreground">Webhooks</h1>
            <Button onClick={() => setShowCreateModal(true)}>New Webhook</Button>
          </div>
          {webhooks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <h3 className="text-lg font-medium mb-2">No webhooks</h3>
              <p className="text-sm">Get started by creating your first webhook</p>
            </div>
          ) : (
            webhooks.map((webhook) => (
              <Card key={webhook.id} className="mb-4">
                <div className="p-6">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-semibold text-foreground">{webhook.name}</h3>
                    <Badge variant={webhook.active ? 'default' : 'secondary'}>
                      {webhook.active ? 'Active' : 'Inactive'}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mb-3">{webhook.config.url}</p>
                  <div className="mb-3">
                    {webhook.events.map((event) => (
                      <Badge key={event} className="mr-2 mb-2">{event}</Badge>
                    ))}
                  </div>
                  <p className="text-sm mb-3">Created {new Date(webhook.created_at).toLocaleDateString()}</p>
                  <div className="space-x-2">
                    <Button size="sm" onClick={() => handlePingWebhook(webhook.id)}>Ping</Button>
                    <Button size="sm" variant="outline" onClick={() => setSelectedWebhook(webhook)}>Edit</Button>
                    <Button size="sm" variant="outline" onClick={() => handleToggleWebhook(webhook.id, webhook.active)}>
                      {webhook.active ? 'Disable' : 'Enable'}
                    </Button>
                    <Button size="sm" variant="destructive" onClick={() => handleDeleteWebhook(webhook.id)}>Delete</Button>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        <Modal open={showCreateModal} onClose={() => setShowCreateModal(false)} title="Add Webhook">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Name</label>
                <Input value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} placeholder="Webhook name" />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Payload URL</label>
                <Input value={formData.url} onChange={(e) => setFormData({ ...formData, url: e.target.value })} placeholder="https://example.com/webhook" />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Secret</label>
                <Input value={formData.secret} onChange={(e) => setFormData({ ...formData, secret: e.target.value })} placeholder="Secret for webhook validation" />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">IP Allow List</label>
                <Input value={formData.allowlist_ips.join(',')} onChange={(e) => setFormData({ ...formData, allowlist_ips: e.target.value.split(',').map(ip => ip.trim()) })} placeholder="Comma-separated IP addresses" />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Events</label>
                <div className="grid grid-cols-2 gap-2 max-h-32 overflow-y-auto">
                  {availableEvents.map((event) => (
                    <label key={event} className="flex items-center space-x-2">
                      <input type="checkbox" checked={formData.events.includes(event)} onChange={() => {
                        const events = formData.events.includes(event)
                          ? formData.events.filter(e => e !== event)
                          : [...formData.events, event];
                        setFormData({ ...formData, events });
                      }} />
                      <span className="text-sm text-foreground">{event}</span>
                    </label>
                  ))}
                </div>
              </div>
              <div className="flex items-center space-x-4 mt-2">
                <label className="flex items-center space-x-2">
                  <input type="checkbox" checked={formData.active} onChange={() => setFormData({ ...formData, active: !formData.active })} />
                  <span className="text-sm text-foreground">Active</span>
                </label>
              </div>
            </div>
            <div className="mt-6 flex justify-end">
              <Button onClick={handleCreateWebhook} disabled={createLoading}>
                {createLoading ? 'Creating…' : 'Create Webhook'}
              </Button>
            </div>
        </Modal>
      </div>
    </AppLayout>
  );
}
