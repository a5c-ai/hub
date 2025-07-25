'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Modal } from '@/components/ui/Modal';
import { sshKeyApi } from '@/lib/api';
import { SSHKey, CreateSSHKeyRequest } from '@/types';

export function SSHKeyManagement() {
  const [sshKeys, setSSHKeys] = useState<SSHKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [addingKey, setAddingKey] = useState(false);
  const [deletingKey, setDeletingKey] = useState<string | null>(null);

  const [newKey, setNewKey] = useState<CreateSSHKeyRequest>({
    title: '',
    key_data: '',
  });

  useEffect(() => {
    loadSSHKeys();
  }, []);

  const loadSSHKeys = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await sshKeyApi.getSSHKeys();
      const keysData = Array.isArray(response) ? response : (response.data || []);
      setSSHKeys(keysData as SSHKey[]);
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to load SSH keys');
    } finally {
      setLoading(false);
    }
  };

  const handleAddKey = async () => {
    if (!newKey.title.trim() || !newKey.key_data.trim()) {
      setError('Please fill in all fields');
      return;
    }

    try {
      setAddingKey(true);
      setError(null);
      const response = await sshKeyApi.createSSHKey(newKey);
      const newSSHKey = (response.data || response) as SSHKey;
      setSSHKeys([...sshKeys, newSSHKey]);
      setShowAddModal(false);
      setNewKey({ title: '', key_data: '' });
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to add SSH key');
    } finally {
      setAddingKey(false);
    }
  };

  const handleDeleteKey = async (keyId: string) => {
    try {
      setDeletingKey(keyId);
      setError(null);
      await sshKeyApi.deleteSSHKey(keyId);
      setSSHKeys(sshKeys.filter(key => key.id !== keyId));
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Failed to delete SSH key');
    } finally {
      setDeletingKey(null);
    }
  };

  const formatFingerprint = (fingerprint: string) => {
    if (fingerprint.startsWith('SHA256:')) {
      return fingerprint;
    }
    // Format MD5 fingerprint if needed
    return fingerprint;
  };

  const formatLastUsed = (lastUsed?: string) => {
    if (!lastUsed) return 'Never';
    return new Date(lastUsed).toLocaleDateString();
  };

  const getKeyTypeIcon = (keyType: string) => {
    switch (keyType) {
      case 'ssh-rsa':
        return '🔑';
      case 'ssh-ed25519':
        return '🛡️';
      case 'ecdsa-sha2-nistp256':
      case 'ecdsa-sha2-nistp384':
      case 'ecdsa-sha2-nistp521':
        return '🔐';
      default:
        return '🔑';
    }
  };

  const validateSSHKey = (keyData: string) => {
    const trimmed = keyData.trim();
    if (!trimmed) return false;
    
    // Basic SSH key format validation
    const sshKeyRegex = /^ssh-(rsa|ed25519|ecdsa-sha2-nistp(256|384|521)) [A-Za-z0-9+/=]+ .*$/;
    return sshKeyRegex.test(trimmed) || trimmed.startsWith('ecdsa-sha2-');
  };

  if (loading) {
    return (
      <Card>
        <div className="p-6">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p className="text-center text-muted-foreground mt-2">Loading SSH keys...</p>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h3 className="text-lg font-semibold text-foreground">SSH Keys</h3>
              <p className="text-sm text-muted-foreground">
                SSH keys allow you to securely connect to Git repositories without passwords
              </p>
            </div>
            <Button onClick={() => setShowAddModal(true)}>
              Add SSH Key
            </Button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          {sshKeys.length === 0 ? (
            <div className="text-center py-8">
              <div className="text-4xl mb-4">🔑</div>
              <h4 className="text-lg font-medium text-foreground mb-2">No SSH keys</h4>
              <p className="text-muted-foreground mb-4">
                You haven&apos;t added any SSH keys yet. Add one to securely access your repositories.
              </p>
              <Button onClick={() => setShowAddModal(true)}>
                Add your first SSH key
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {sshKeys.map((key) => (
                <div
                  key={key.id}
                  className="border border-border rounded-lg p-4 bg-card hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-lg">{getKeyTypeIcon(key.key_type)}</span>
                        <h4 className="font-medium text-foreground">{key.title}</h4>
                        <span className="text-xs bg-muted px-2 py-1 rounded text-muted-foreground">
                          {key.key_type}
                        </span>
                      </div>
                      <div className="space-y-1 text-sm text-muted-foreground">
                        <div>
                          <span className="font-medium">Fingerprint:</span>{' '}
                          <code className="bg-muted px-1 py-0.5 rounded text-xs">
                            {formatFingerprint(key.fingerprint)}
                          </code>
                        </div>
                        <div>
                          <span className="font-medium">Last used:</span> {formatLastUsed(key.last_used_at)}
                        </div>
                        <div>
                          <span className="font-medium">Added:</span> {new Date(key.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDeleteKey(key.id)}
                      disabled={deletingKey === key.id}
                      className="text-destructive hover:text-destructive"
                    >
                      {deletingKey === key.id ? 'Deleting...' : 'Delete'}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Add SSH Key Modal */}
      <Modal
        open={showAddModal}
        onClose={() => {
          setShowAddModal(false);
          setNewKey({ title: '', key_data: '' });
          setError(null);
        }}
        title="Add SSH Key"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              Title
            </label>
            <Input
              value={newKey.title}
              onChange={(e) => setNewKey({ ...newKey, title: e.target.value })}
              placeholder="My laptop key"
              disabled={addingKey}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Give your key a descriptive name
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              SSH Public Key
            </label>
            <textarea
              value={newKey.key_data}
              onChange={(e) => setNewKey({ ...newKey, key_data: e.target.value })}
              placeholder="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQ... user@example.com"
              rows={6}
              disabled={addingKey}
              className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 resize-none font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Paste your public key here. It should start with ssh-rsa, ssh-ed25519, or ecdsa-sha2-
            </p>
          </div>

          {error && (
            <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          <div className="bg-muted/30 border border-border rounded-md p-4">
            <h4 className="font-medium text-foreground mb-2">💡 How to generate an SSH key</h4>
            <div className="text-sm text-muted-foreground space-y-2">
              <p>1. Open your terminal</p>
              <p>2. Run: <code className="bg-muted px-1 py-0.5 rounded">ssh-keygen -t ed25519 -C &quot;your-email@example.com&quot;</code></p>
              <p>3. Copy the public key: <code className="bg-muted px-1 py-0.5 rounded">cat ~/.ssh/id_ed25519.pub</code></p>
              <p>4. Paste the output above</p>
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <Button
              variant="outline"
              onClick={() => {
                setShowAddModal(false);
                setNewKey({ title: '', key_data: '' });
                setError(null);
              }}
              disabled={addingKey}
            >
              Cancel
            </Button>
            <Button
              onClick={handleAddKey}
              disabled={addingKey || !newKey.title.trim() || !validateSSHKey(newKey.key_data)}
            >
              {addingKey ? 'Adding...' : 'Add SSH Key'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}