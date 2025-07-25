'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Modal } from '@/components/ui/Modal';

interface Secret {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
  environment?: string;
}

export default function SecretsPage() {
  const params = useParams();
  const { owner, repo } = params;
  
  const [secrets, setSecrets] = useState<Secret[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState<string | null>(null);
  
  // Form state
  const [secretName, setSecretName] = useState('');
  const [secretValue, setSecretValue] = useState('');
  const [secretEnvironment, setSecretEnvironment] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchSecrets();
  }, [owner, repo]);

  const fetchSecrets = async () => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/secrets`);
      if (response.ok) {
        const data = await response.json();
        setSecrets(data.secrets || []);
      } else {
        throw new Error('Failed to fetch secrets');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const handleAddSecret = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/secrets`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: secretName,
          value: secretValue,
          environment: secretEnvironment || undefined,
        }),
      });

      if (response.ok) {
        await fetchSecrets();
        setShowAddModal(false);
        setSecretName('');
        setSecretValue('');
        setSecretEnvironment('');
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create secret');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteSecret = async (secretId: string) => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/secrets/${secretId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        await fetchSecrets();
        setShowDeleteModal(null);
      } else {
        throw new Error('Failed to delete secret');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };

  const validateSecretName = (name: string) => {
    // Secret names must be valid environment variable names
    return /^[A-Z][A-Z0-9_]*$/.test(name);
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-16 bg-gray-200 rounded"></div>
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
          <h1 className="text-3xl font-bold">Secrets</h1>
          <p className="text-gray-600 mt-2">
            Secrets are encrypted environment variables that can be used in your workflows.
          </p>
        </div>
        <Button onClick={() => setShowAddModal(true)}>
          New repository secret
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-6">
          <h3 className="text-lg font-medium text-red-800">Error</h3>
          <p className="text-red-700">{error}</p>
        </div>
      )}

      {/* Secrets List */}
      {secrets.length === 0 ? (
        <Card className="p-8 text-center">
          <h3 className="text-lg font-medium mb-2">No secrets yet</h3>
          <p className="text-gray-600 mb-4">
            Secrets are environment variables that are encrypted and only exposed to selected actions.
          </p>
          <Button onClick={() => setShowAddModal(true)}>
            Add your first secret
          </Button>
        </Card>
      ) : (
        <Card>
          <div className="divide-y">
            {secrets.map((secret) => (
              <div key={secret.id} className="p-4 flex items-center justify-between">
                <div>
                  <h3 className="font-medium">{secret.name}</h3>
                  <p className="text-sm text-gray-500">
                    Updated {new Date(secret.updated_at).toLocaleDateString()}
                    {secret.environment && (
                      <span className="ml-2 px-2 py-0.5 bg-blue-100 text-blue-800 rounded text-xs">
                        {secret.environment}
                      </span>
                    )}
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => {
                      setSecretName(secret.name);
                      setSecretEnvironment(secret.environment || '');
                      setShowAddModal(true);
                    }}
                  >
                    Update
                  </Button>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => setShowDeleteModal(secret.id)}
                    className="text-red-600 hover:text-red-700"
                  >
                    Delete
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Add/Update Secret Modal */}
      <Modal
        isOpen={showAddModal}
        onClose={() => {
          setShowAddModal(false);
          setSecretName('');
          setSecretValue('');
          setSecretEnvironment('');
        }}
        title={secretName ? 'Update secret' : 'New repository secret'}
      >
        <form onSubmit={handleAddSecret}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name
              </label>
              <Input
                type="text"
                value={secretName}
                onChange={(e) => setSecretName(e.target.value.toUpperCase())}
                placeholder="SECRET_NAME"
                required
                disabled={!!secretName} // Disable when updating
              />
              <p className="text-xs text-gray-500 mt-1">
                Must be uppercase letters, numbers, and underscores only
              </p>
              {secretName && !validateSecretName(secretName) && (
                <p className="text-xs text-red-600 mt-1">
                  Invalid secret name format
                </p>
              )}
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Value
              </label>
              <textarea
                value={secretValue}
                onChange={(e) => setSecretValue(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                rows={4}
                placeholder="Enter secret value..."
                required
              />
              <p className="text-xs text-gray-500 mt-1">
                The secret value will be encrypted and stored securely
              </p>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Environment (optional)
              </label>
              <Input
                type="text"
                value={secretEnvironment}
                onChange={(e) => setSecretEnvironment(e.target.value)}
                placeholder="production"
              />
              <p className="text-xs text-gray-500 mt-1">
                Restrict this secret to a specific environment
              </p>
            </div>
          </div>
          
          <div className="flex justify-end gap-2 mt-6">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setShowAddModal(false);
                setSecretName('');
                setSecretValue('');
                setSecretEnvironment('');
              }}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={submitting || !validateSecretName(secretName)}
            >
              {submitting ? 'Saving...' : (secretName ? 'Update secret' : 'Add secret')}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!showDeleteModal}
        onClose={() => setShowDeleteModal(null)}
        title="Delete secret"
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to delete this secret? This action cannot be undone.
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
              onClick={() => showDeleteModal && handleDeleteSecret(showDeleteModal)}
            >
              Delete secret
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}