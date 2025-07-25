'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { apiClient } from '@/lib/api';
import { Dropdown } from '@/components/ui/Dropdown';
import { formatDistanceToNow } from 'date-fns';

interface Runner {
  id: string;
  name: string;
  labels: string[];
  status: 'online' | 'offline' | 'busy';
  type: 'kubernetes' | 'self-hosted';
  version?: string;
  os?: string;
  architecture?: string;
  last_seen_at?: string;
  created_at: string;
}

export default function RunnersPage() {
  const params = useParams();
  const { owner, repo } = params;
  
  const [runners, setRunners] = useState<Runner[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState<string | null>(null);
  const [registrationToken, setRegistrationToken] = useState<string>('');

  const fetchRunners = useCallback(async () => {
    try {
      const response = await apiClient.get<{ runners: Runner[] }>(`/repositories/${owner}/${repo}/actions/runners`);
      setRunners(response.data.runners || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  }, [owner, repo]);

  useEffect(() => {
    fetchRunners();
  }, [fetchRunners]);

  const generateRegistrationToken = async () => {
    try {
      const response = await apiClient.post<{ token: string }>(`/repositories/${owner}/${repo}/actions/runners/registration-token`);
      setRegistrationToken(response.data.token);
      setShowAddModal(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };

  const handleDeleteRunner = async (runnerId: string) => {
    try {
      await apiClient.delete(`/repositories/${owner}/${repo}/actions/runners/${runnerId}`);
      await fetchRunners();
      setShowDeleteModal(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'text-green-600';
      case 'busy': return 'text-yellow-600';
      case 'offline': return 'text-muted-foreground';
      default: return 'text-muted-foreground';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online': return '🟢';
      case 'busy': return '🟡';
      case 'offline': return '🔴';
      default: return '❓';
    }
  };

  const formatLastSeen = (lastSeenAt?: string) => {
    if (!lastSeenAt) return 'Never';
    const date = new Date(lastSeenAt);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffMins > 0) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/4 mb-4"></div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-20 bg-muted rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold">Runners</h1>
          <p className="text-muted-foreground mt-2">
            Self-hosted runners for this repository. Kubernetes runners are managed automatically.
          </p>
        </div>
        <Button onClick={generateRegistrationToken}>
          New self-hosted runner
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
          <h3 className="text-lg font-medium text-red-800">Error</h3>
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {/* Runner Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card className="p-4">
          <div className="text-2xl font-bold text-green-600">
            {runners.filter(r => r.status === 'online').length}
          </div>
                          <div className="text-sm text-muted-foreground">Online</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold text-yellow-600">
            {runners.filter(r => r.status === 'busy').length}
          </div>
                          <div className="text-sm text-muted-foreground">Busy</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold text-muted-foreground">
            {runners.filter(r => r.status === 'offline').length}
          </div>
          <div className="text-sm text-muted-foreground">Offline</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold text-primary">
            {runners.filter(r => r.type === 'kubernetes').length}
          </div>
          <div className="text-sm text-muted-foreground">Kubernetes</div>
        </Card>
      </div>

      {/* Runners List */}
      {runners.length === 0 ? (
        <Card className="p-8 text-center">
          <h3 className="text-lg font-medium mb-2">No runners yet</h3>
          <p className="text-gray-600 mb-4">
            Set up a self-hosted runner to run workflows on your own infrastructure.
          </p>
          <Button onClick={generateRegistrationToken}>
            Add your first runner
          </Button>
        </Card>
      ) : (
        <Card>
          <div className="divide-y">
            {runners.map((runner) => (
              <div key={runner.id} className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <span className="text-2xl">
                      {getStatusIcon(runner.status)}
                    </span>
                    
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-medium">{runner.name}</h3>
                        <Badge variant={getStatusColor(runner.status) as 'default' | 'secondary' | 'outline' | 'destructive'}>
                          {runner.status}
                        </Badge>
                        <Badge variant="outline">
                          {runner.type}
                        </Badge>
                      </div>
                      
                      <div className="flex items-center gap-4 text-sm text-gray-500 mt-1">
                        {runner.os && (
                          <span>{runner.os} {runner.architecture}</span>
                        )}
                        {runner.version && (
                          <span>v{runner.version}</span>
                        )}
                        <span>Last seen: {formatLastSeen(runner.last_seen_at)}</span>
                      </div>
                      
                      {runner.labels.length > 0 && (
                        <div className="flex gap-1 mt-2">
                          {runner.labels.map((label, index) => (
                            <Badge key={index} variant="outline" className="text-xs">
                              {label}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                  
                  <div className="flex gap-2">
                    {runner.type === 'self-hosted' && (
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => setShowDeleteModal(runner.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        Remove
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Add Runner Modal */}
      <Modal
        open={showAddModal}
        onClose={() => {
          setShowAddModal(false);
          setRegistrationToken('');
        }}
        title="Add self-hosted runner"
      >
        <div className="space-y-4">
          <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
            <h4 className="font-medium text-blue-800 mb-2">Runner registration token</h4>
            <div className="bg-white border rounded p-2 font-mono text-sm break-all">
              {registrationToken}
            </div>
            <p className="text-sm text-blue-700 mt-2">
              This token expires in 1 hour. Use it to configure your self-hosted runner.
            </p>
          </div>
          
          <div>
            <h4 className="font-medium mb-2">Setup instructions</h4>
            <div className="bg-gray-50 rounded p-4 text-sm">
              <p className="mb-2">1. Download and extract the runner package:</p>
              <code className="block bg-gray-800 text-white p-2 rounded text-xs mb-2">
                curl -o actions-runner-linux-x64-2.311.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz
              </code>
              
              <p className="mb-2">2. Configure the runner:</p>
              <code className="block bg-gray-800 text-white p-2 rounded text-xs mb-2">
                ./config.sh --url https://github.com/{owner}/{repo} --token {registrationToken}
              </code>
              
              <p className="mb-2">3. Start the runner:</p>
              <code className="block bg-gray-800 text-white p-2 rounded text-xs">
                ./run.sh
              </code>
            </div>
          </div>
          
          <div className="flex justify-end">
            <Button
              onClick={() => {
                setShowAddModal(false);
                setRegistrationToken('');
              }}
            >
              Done
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        open={!!showDeleteModal}
        onClose={() => setShowDeleteModal(null)}
        title="Remove runner"
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to remove this runner? This action cannot be undone.
          </p>
          
          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowDeleteModal(null)}
            >
              Cancel
            </Button>
            <Button
              variant="outline"
              className="text-red-600 hover:text-red-700"
              onClick={() => showDeleteModal && handleDeleteRunner(showDeleteModal)}
            >
              Remove runner
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}