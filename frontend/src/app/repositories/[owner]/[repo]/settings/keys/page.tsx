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

interface DeployKey {
  id: string;
  title: string;
  key: string;
  read_only: boolean;
  verified: boolean;
  created_at: string;
  last_used?: string;
}

export default function DeployKeysPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const [deployKeys, setDeployKeys] = useState<DeployKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  
  const [formData, setFormData] = useState({
    title: '',
    key: '',
    read_only: true
  });

  const fetchDeployKeys = async () => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
      const response = await api.get(`/repositories/${owner}/${repo}/keys`);
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setDeployKeys(result);
    }
  };

  useEffect(() => {
    fetchDeployKeys();
  }, [owner, repo]);

  const handleCreateDeployKey = async () => {
    if (!formData.title || !formData.key) return;

    const handleError = createErrorHandler(setError, setCreateLoading);
    
    const operation = async () => {
      const response = await api.post(`/repositories/${owner}/${repo}/keys`, {
        title: formData.title,
        key: formData.key,
        read_only: formData.read_only
      });
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setDeployKeys([...deployKeys, result]);
      setShowCreateModal(false);
      setFormData({
        title: '',
        key: '',
        read_only: true
      });
    }
  };

  const handleDeleteDeployKey = async (keyId: string) => {
    if (!confirm('Are you sure you want to delete this deploy key?')) return;

    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      await api.delete(`/repositories/${owner}/${repo}/keys/${keyId}`);
    };

    const result = await handleError(operation);
    if (result !== null) {
      setDeployKeys(deployKeys.filter(k => k.id !== keyId));
    }
  };

  const getKeyFingerprint = (key: string) => {
    // Simple approximation of key fingerprint display
    if (key.length > 40) {
      return key.substring(0, 20) + '...' + key.substring(key.length - 20);
    }
    return key;
  };

  const validateSSHKey = (key: string) => {
    const sshKeyPattern = /^(ssh-rsa|ssh-dss|ssh-ed25519|ecdsa-sha2-nistp256|ecdsa-sha2-nistp384|ecdsa-sha2-nistp521)\s+[A-Za-z0-9+/]+[=]{0,3}(\s+.*)?$/;
    return sshKeyPattern.test(key.trim());
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
            <Button onClick={fetchDeployKeys} disabled={loading}>
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
          <span className="text-foreground font-medium">Deploy Keys</span>
        </nav>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Deploy Keys</h1>
            <p className="text-muted-foreground mt-2">
              Deploy keys allow read-only or read-write access to your repository
            </p>
          </div>
          <Button onClick={() => setShowCreateModal(true)}>
            Add Deploy Key
          </Button>
        </div>

        {/* Deploy Keys List */}
        <div className="space-y-4">
          {deployKeys.length === 0 ? (
            <Card>
              <div className="p-8 text-center">
                <svg className="w-12 h-12 mx-auto mb-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
                <h3 className="text-lg font-medium text-foreground mb-2">No deploy keys</h3>
                <p className="text-muted-foreground">
                  Deploy keys allow servers to access your repository
                </p>
              </div>
            </Card>
          ) : (
            deployKeys.map((deployKey) => (
              <Card key={deployKey.id}>
                <div className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <h3 className="text-lg font-semibold text-foreground">{deployKey.title}</h3>
                        <Badge variant={deployKey.read_only ? 'secondary' : 'default'}>
                          {deployKey.read_only ? 'Read-only' : 'Read-write'}
                        </Badge>
                        {deployKey.verified && (
                          <Badge variant="default" className="bg-green-100 text-green-800">
                            Verified
                          </Badge>
                        )}
                      </div>
                      <div className="bg-muted p-3 rounded-md font-mono text-sm mb-3">
                        {getKeyFingerprint(deployKey.key)}
                      </div>
                      <div className="text-xs text-muted-foreground space-y-1">
                        <p>Added {new Date(deployKey.created_at).toLocaleDateString()}</p>
                        {deployKey.last_used && (
                          <p>Last used {new Date(deployKey.last_used).toLocaleDateString()}</p>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDeleteDeployKey(deployKey.id)}
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

        {/* Help Section */}
        <Card className="mt-8">
          <div className="p-6">
            <h3 className="text-lg font-semibold text-foreground mb-4">About Deploy Keys</h3>
            <div className="space-y-3 text-sm text-muted-foreground">
              <p>
                Deploy keys are SSH keys that grant access to a single repository. 
                They can be read-only or have read-write permissions.
              </p>
              <p>
                <strong>Read-only keys</strong> can pull (clone/fetch) from the repository but cannot push changes.
              </p>
              <p>
                <strong>Read-write keys</strong> can both pull from and push to the repository.
              </p>
              <p>
                Deploy keys are useful for CI/CD systems, deployment scripts, and other automated services.
              </p>
            </div>
          </div>
        </Card>

        {/* Create Deploy Key Modal */}
        <Modal 
          open={showCreateModal} 
          onClose={() => setShowCreateModal(false)}
          title="Add Deploy Key"
        >
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Title
              </label>
              <Input
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                placeholder="My Deploy Key"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                Key
              </label>
              <textarea
                value={formData.key}
                onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                placeholder="ssh-rsa AAAAB3NzaC1yc2EAAAA..."
                rows={6}
                className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring font-mono text-sm"
              />
              {formData.key && !validateSSHKey(formData.key) && (
                <p className="text-sm text-red-600 mt-1">
                  Please enter a valid SSH public key
                </p>
              )}
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="read_only"
                checked={formData.read_only}
                onChange={(e) => setFormData({ ...formData, read_only: e.target.checked })}
                className="rounded border-border"
              />
              <label htmlFor="read_only" className="text-sm font-medium text-foreground">
                Allow write access
              </label>
              <span className="text-xs text-muted-foreground">
                (uncheck for read-write access)
              </span>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
              <p className="text-sm text-blue-800">
                <strong>Tip:</strong> Generate an SSH key with: <code>ssh-keygen -t ed25519 -C &quot;your_email@example.com&quot;</code>
              </p>
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <Button variant="outline" onClick={() => setShowCreateModal(false)}>
                Cancel
              </Button>
              <Button 
                onClick={handleCreateDeployKey} 
                disabled={createLoading || !formData.title || !formData.key || !validateSSHKey(formData.key)}
              >
                {createLoading ? 'Adding...' : 'Add Key'}
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </AppLayout>
  );
}