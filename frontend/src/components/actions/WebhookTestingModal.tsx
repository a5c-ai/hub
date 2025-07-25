'use client';

import { useState } from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Card } from '@/components/ui/Card';

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

interface DeliveryResult {
  id: string;
  webhook_id: string;
  event_type: string;
  response_status: number;
  response_time_ms: number;
  delivered_at: string;
  error?: string;
}

interface WebhookTestingModalProps {
  webhook: Webhook;
  open: boolean;
  onClose: () => void;
  owner: string;
  repo: string;
}

export default function WebhookTestingModal({
  webhook,
  open,
  onClose,
  owner,
  repo
}: WebhookTestingModalProps) {
  const [testing, setTesting] = useState(false);
  const [testEventType, setTestEventType] = useState('push');
  const [deliveries, setDeliveries] = useState<DeliveryResult[]>([]);
  const [showPayload, setShowPayload] = useState<string | null>(null);

  const testWebhook = async () => {
    setTesting(true);
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/hooks/${webhook.id}/test`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          event_type: testEventType
        })
      });

      if (response.ok) {
        const result = await response.json();
        setDeliveries(prev => [result, ...prev]);
      } else {
        console.error('Failed to test webhook');
      }
    } catch (error) {
      console.error('Error testing webhook:', error);
    } finally {
      setTesting(false);
    }
  };

  const getDeliveryStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'text-green-600';
    if (status >= 400 && status < 500) return 'text-yellow-600';
    if (status >= 500) return 'text-red-600';
    return 'text-gray-600';
  };

  const getDeliveryStatusBadge = (status: number) => {
    if (status >= 200 && status < 300) return 'default';
    if (status >= 400 && status < 500) return 'outline';
    if (status >= 500) return 'destructive';
    return 'secondary';
  };

  const samplePayloads = {
    push: {
      ref: "refs/heads/main",
      before: "0000000000000000000000000000000000000000",
      after: "1234567890abcdef1234567890abcdef12345678",
      repository: {
        id: "12345",
        name: repo,
        full_name: `${owner}/${repo}`,
        owner: {
          login: owner,
          id: "67890"
        }
      },
      pusher: {
        name: "test-user",
        email: "test@example.com"
      },
      sender: {
        login: "test-user",
        id: "54321"
      },
      commits: [
        {
          id: "1234567890abcdef1234567890abcdef12345678",
          message: "Test commit",
          timestamp: new Date().toISOString(),
          added: ["file1.txt"],
          removed: [],
          modified: []
        }
      ]
    },
    pull_request: {
      action: "opened",
      number: 123,
      pull_request: {
        id: "pr-123",
        number: 123,
        title: "Test PR",
        body: "This is a test pull request",
        head: {
          ref: "feature-branch",
          sha: "abcdef1234567890abcdef1234567890abcdef12"
        },
        base: {
          ref: "main",
          sha: "1234567890abcdef1234567890abcdef12345678"
        }
      },
      repository: {
        id: "12345",
        name: repo,
        full_name: `${owner}/${repo}`
      },
      sender: {
        login: "test-user",
        id: "54321"
      }
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Test Webhook: ${webhook.name}`}
      size="large"
    >
      <div className="space-y-6">
        {/* Webhook Info */}
        <Card className="p-4">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium mb-1">Webhook URL</label>
              <code className="text-xs bg-gray-100 p-2 rounded block break-all">
                {webhook.config.url}
              </code>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Content Type</label>
              <Badge variant="outline">{webhook.config.content_type}</Badge>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Events</label>
            <div className="flex flex-wrap gap-1">
              {webhook.events.map(event => (
                <Badge key={event} variant="outline" className="text-xs">
                  {event}
                </Badge>
              ))}
            </div>
          </div>
        </Card>

        {/* Test Section */}
        <Card className="p-4">
          <h4 className="font-medium mb-4">Send Test Event</h4>
          <div className="flex gap-4 items-end">
            <div className="flex-1">
              <label className="block text-sm font-medium mb-2">Event Type</label>
              <select
                value={testEventType}
                onChange={(e) => setTestEventType(e.target.value)}
                className="w-full px-3 py-2 border border-input rounded-md bg-background"
              >
                {webhook.events.map(event => (
                  <option key={event} value={event}>{event}</option>
                ))}
              </select>
            </div>
            <Button
              onClick={testWebhook}
              disabled={testing}
              className="px-6"
            >
              {testing ? 'Sending...' : 'Send Test'}
            </Button>
          </div>

          {/* Sample Payload Preview */}
          {samplePayloads[testEventType as keyof typeof samplePayloads] && (
            <div className="mt-4">
              <label className="block text-sm font-medium mb-2">Sample Payload</label>
              <pre className="bg-gray-900 text-gray-100 p-4 rounded text-xs overflow-x-auto max-h-40">
                {JSON.stringify(samplePayloads[testEventType as keyof typeof samplePayloads], null, 2)}
              </pre>
            </div>
          )}
        </Card>

        {/* Recent Deliveries */}
        <Card className="p-4">
          <h4 className="font-medium mb-4">Recent Deliveries</h4>
          
          {deliveries.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <p>No recent deliveries</p>
              <p className="text-sm">Send a test event to see delivery results</p>
            </div>
          ) : (
            <div className="space-y-3">
              {deliveries.map((delivery) => (
                <div key={delivery.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {delivery.event_type}
                      </Badge>
                      <Badge variant={getDeliveryStatusBadge(delivery.response_status)}>
                        {delivery.response_status}
                      </Badge>
                      <span className="text-sm text-gray-500">
                        {delivery.response_time_ms}ms
                      </span>
                    </div>
                    <span className="text-xs text-gray-500">
                      {new Date(delivery.delivered_at).toLocaleString()}
                    </span>
                  </div>
                  
                  {delivery.error && (
                    <div className="bg-red-50 border border-red-200 rounded p-2 text-sm text-red-700">
                      {delivery.error}
                    </div>
                  )}
                  
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowPayload(delivery.id)}
                    className="mt-2"
                  >
                    View Details
                  </Button>
                </div>
              ))}
            </div>
          )}
        </Card>

        {/* Action Buttons */}
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </Modal>
  );
}