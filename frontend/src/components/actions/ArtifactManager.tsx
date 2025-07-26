'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';

interface Artifact {
  id: string;
  name: string;
  size_bytes: number;
  created_at: string;
  expires_at?: string;
  expired: boolean;
  download_url?: string;
  path?: string;
  workflow_run?: {
    id: string;
    workflow_name: string;
  };
}

interface ArtifactManagerProps {
  workflowRunId: string;
  owner: string;
  repo: string;
}

export default function ArtifactManager({ 
  workflowRunId, 
  owner, 
  repo 
}: ArtifactManagerProps) {
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadName, setUploadName] = useState('');
  const [uploading, setUploading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState<Record<string, number>>({});
  
  // Batch operations state
  const [selectedArtifacts, setSelectedArtifacts] = useState<Set<string>>(new Set());
  const [showBatchActions, setShowBatchActions] = useState(false);
  const [batchOperationInProgress, setBatchOperationInProgress] = useState(false);
  
  // Metadata display state
  const [showMetadata, setShowMetadata] = useState<Record<string, boolean>>({});

  const fetchArtifacts = async () => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/runs/${workflowRunId}/artifacts`);
      if (response.ok) {
        const data = await response.json();
        setArtifacts(data.artifacts || []);
      } else {
        throw new Error('Failed to fetch artifacts');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchArtifacts();
  }, [workflowRunId, owner, repo]);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setUploadFile(file);
      setUploadName(file.name);
    }
  };

  const handleUpload = async () => {
    if (!uploadFile || !uploadName) return;

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append('artifact', uploadFile);
      formData.append('name', uploadName);

      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/runs/${workflowRunId}/artifacts`, {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        setShowUploadModal(false);
        setUploadFile(null);
        setUploadName('');
        await fetchArtifacts();
      } else {
        throw new Error('Failed to upload artifact');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploading(false);
    }
  };

  const handleDownload = async (artifact: Artifact) => {
    try {
      setDownloadProgress({ ...downloadProgress, [artifact.id]: 0 });

      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/artifacts/${artifact.id}/download`);
      
      if (!response.ok) {
        throw new Error('Failed to download artifact');
      }

      // Handle download with progress tracking
      const reader = response.body?.getReader();
      const contentLength = response.headers.get('Content-Length');
      const total = contentLength ? parseInt(contentLength) : 0;
      let received = 0;
      const chunks: Uint8Array[] = [];

      if (reader) {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          chunks.push(value);
          received += value.length;

          if (total > 0) {
            const progress = Math.round((received / total) * 100);
            setDownloadProgress({ ...downloadProgress, [artifact.id]: progress });
          }
        }
      }

      // Create blob and download
      const blob = new Blob(chunks);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = artifact.name;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      // Clear progress
      setDownloadProgress({ ...downloadProgress, [artifact.id]: 100 });
      setTimeout(() => {
        setDownloadProgress(prev => {
          const updated = { ...prev };
          delete updated[artifact.id];
          return updated;
        });
      }, 2000);

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Download failed');
      setDownloadProgress(prev => {
        const updated = { ...prev };
        delete updated[artifact.id];
        return updated;
      });
    }
  };

  const handleDelete = async (artifactId: string) => {
    if (!confirm('Are you sure you want to delete this artifact?')) return;

    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/artifacts/${artifactId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        await fetchArtifacts();
      } else {
        throw new Error('Failed to delete artifact');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Delete failed');
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatExpirationDate = (expiresAt?: string) => {
    if (!expiresAt) return 'Never';
    const date = new Date(expiresAt);
    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));
    
    if (diffDays < 0) return 'Expired';
    if (diffDays === 0) return 'Expires today';
    if (diffDays === 1) return 'Expires tomorrow';
    return `Expires in ${diffDays} days`;
  };

  // Batch operations functions
  const handleSelectArtifact = (artifactId: string, checked: boolean) => {
    const newSelected = new Set(selectedArtifacts);
    if (checked) {
      newSelected.add(artifactId);
    } else {
      newSelected.delete(artifactId);
    }
    setSelectedArtifacts(newSelected);
    setShowBatchActions(newSelected.size > 0);
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      const allIds = new Set(artifacts.filter(a => !a.expired).map(a => a.id));
      setSelectedArtifacts(allIds);
      setShowBatchActions(allIds.size > 0);
    } else {
      setSelectedArtifacts(new Set());
      setShowBatchActions(false);
    }
  };

  const handleBatchDownload = async () => {
    if (selectedArtifacts.size === 0) return;

    setBatchOperationInProgress(true);
    let successCount = 0;
    let errorCount = 0;

    for (const artifactId of selectedArtifacts) {
      const artifact = artifacts.find(a => a.id === artifactId);
      if (artifact && !artifact.expired) {
        try {
          await handleDownload(artifact);
          successCount++;
        } catch (err) {
          errorCount++;
        }
      }
    }

    setBatchOperationInProgress(false);
    setSelectedArtifacts(new Set());
    setShowBatchActions(false);
    
    if (errorCount > 0) {
      setError(`Batch download completed: ${successCount} successful, ${errorCount} failed`);
    }
  };

  const handleBatchDelete = async () => {
    if (selectedArtifacts.size === 0) return;
    
    if (!confirm(`Are you sure you want to delete ${selectedArtifacts.size} selected artifacts?`)) {
      return;
    }

    setBatchOperationInProgress(true);
    
    try {
      const response = await fetch('/api/v1/admin/storage/artifacts/batch-delete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          artifact_ids: Array.from(selectedArtifacts),
        }),
      });

      if (response.ok) {
        const result = await response.json();
        await fetchArtifacts();
        setSelectedArtifacts(new Set());
        setShowBatchActions(false);
        
        if (result.error_count > 0) {
          setError(`Batch delete completed: ${result.success_count} successful, ${result.error_count} failed`);
        }
      } else {
        throw new Error('Failed to perform batch delete');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Batch delete failed');
    } finally {
      setBatchOperationInProgress(false);
    }
  };

  const toggleMetadata = (artifactId: string) => {
    setShowMetadata(prev => ({
      ...prev,
      [artifactId]: !prev[artifactId]
    }));
  };

  if (loading) {
    return (
      <Card className="p-4">
        <div className="animate-pulse space-y-4">
          <div className="h-4 bg-muted rounded w-1/4"></div>
          <div className="space-y-2">
            <div className="h-8 bg-muted rounded"></div>
            <div className="h-8 bg-muted rounded"></div>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <Card>
      <div className="p-4 border-b space-y-3">
        <div className="flex items-center justify-between">
          <h4 className="font-medium">Artifacts</h4>
          <Button onClick={() => setShowUploadModal(true)} size="sm">
            Upload Artifact
          </Button>
        </div>
        
        {artifacts.length > 0 && (
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={selectedArtifacts.size === artifacts.filter(a => !a.expired).length && artifacts.filter(a => !a.expired).length > 0}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  className="rounded"
                />
                Select All ({artifacts.filter(a => !a.expired).length})
              </label>
              {selectedArtifacts.size > 0 && (
                <span className="text-sm text-muted-foreground">
                  {selectedArtifacts.size} selected
                </span>
              )}
            </div>
            
            {showBatchActions && (
              <div className="flex items-center gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleBatchDownload}
                  disabled={batchOperationInProgress}
                >
                  Download Selected
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleBatchDelete}
                  disabled={batchOperationInProgress}
                  className="text-red-600 hover:text-red-700"
                >
                  {batchOperationInProgress ? 'Deleting...' : 'Delete Selected'}
                </Button>
              </div>
            )}
          </div>
        )}
      </div>

      {error && (
        <div className="p-4 bg-red-50 border-b border-red-200">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      <div className="divide-y">
        {artifacts.length === 0 ? (
          <div className="p-8 text-center">
            <div className="text-muted-foreground mb-4">
              <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p className="text-sm">No artifacts found</p>
            </div>
            <Button onClick={() => setShowUploadModal(true)} variant="outline">
              Upload your first artifact
            </Button>
          </div>
        ) : (
          artifacts.map((artifact) => (
            <div key={artifact.id} className="p-4">
              <div className="flex items-start gap-3">
                {/* Checkbox for batch selection */}
                {!artifact.expired && (
                  <input
                    type="checkbox"
                    checked={selectedArtifacts.has(artifact.id)}
                    onChange={(e) => handleSelectArtifact(artifact.id, e.target.checked)}
                    className="mt-1 rounded"
                  />
                )}
                
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <h5 className="font-medium truncate">{artifact.name}</h5>
                        {artifact.expired && (
                          <Badge variant="destructive" className="text-xs">
                            Expired
                          </Badge>
                        )}
                      </div>
                      <div className="text-sm text-muted-foreground flex items-center gap-4 flex-wrap">
                        <span>{formatFileSize(artifact.size_bytes)}</span>
                        <span>Created {new Date(artifact.created_at).toLocaleDateString()}</span>
                        <span>{formatExpirationDate(artifact.expires_at)}</span>
                      </div>
                      
                      {/* Enhanced metadata display */}
                      {showMetadata[artifact.id] && (
                        <div className="mt-3 p-3 bg-gray-50 rounded-md space-y-2">
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                            <div>
                              <span className="font-medium text-muted-foreground">Artifact ID:</span>
                              <p className="font-mono text-xs break-all">{artifact.id}</p>
                            </div>
                            {artifact.path && (
                              <div>
                                <span className="font-medium text-muted-foreground">Storage Path:</span>
                                <p className="font-mono text-xs break-all">{artifact.path}</p>
                              </div>
                            )}
                            <div>
                              <span className="font-medium text-muted-foreground">Created:</span>
                              <p className="text-xs">{new Date(artifact.created_at).toLocaleString()}</p>
                            </div>
                            {artifact.expires_at && (
                              <div>
                                <span className="font-medium text-muted-foreground">Expires:</span>
                                <p className="text-xs">{new Date(artifact.expires_at).toLocaleString()}</p>
                              </div>
                            )}
                            {artifact.workflow_run && (
                              <div className="md:col-span-2">
                                <span className="font-medium text-muted-foreground">Workflow:</span>
                                <p className="text-xs">{artifact.workflow_run.workflow_name}</p>
                              </div>
                            )}
                          </div>
                        </div>
                      )}
                      
                      {downloadProgress[artifact.id] !== undefined && (
                        <div className="mt-2">
                          <div className="text-xs text-muted-foreground mb-1">
                            Downloading... {downloadProgress[artifact.id]}%
                          </div>
                          <div className="w-full bg-gray-200 rounded-full h-1">
                            <div
                              className="bg-blue-500 h-1 rounded-full transition-all duration-300"
                              style={{ width: `${downloadProgress[artifact.id]}%` }}
                            ></div>
                          </div>
                        </div>
                      )}
                    </div>
                    
                    <div className="flex items-center gap-2 ml-4">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => toggleMetadata(artifact.id)}
                        className="text-xs"
                      >
                        {showMetadata[artifact.id] ? 'Hide' : 'Info'}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDownload(artifact)}
                        disabled={artifact.expired || downloadProgress[artifact.id] !== undefined}
                      >
                        {downloadProgress[artifact.id] !== undefined ? 'Downloading...' : 'Download'}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDelete(artifact.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Upload Modal */}
      <Modal
        open={showUploadModal}
        onClose={() => setShowUploadModal(false)}
        title="Upload Artifact"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">
              Artifact Name
            </label>
            <Input
              value={uploadName}
              onChange={(e) => setUploadName(e.target.value)}
              placeholder="my-artifact"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">
              File
            </label>
            <input
              type="file"
              onChange={handleFileSelect}
              className="w-full px-3 py-2 border border-input rounded-md bg-background"
            />
            {uploadFile && (
              <p className="text-sm text-muted-foreground mt-1">
                Selected: {uploadFile.name} ({formatFileSize(uploadFile.size)})
              </p>
            )}
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
            <h6 className="font-medium text-blue-800 mb-1">Retention Policy</h6>
            <p className="text-sm text-blue-700">
              Artifacts are automatically deleted after 90 days. Large artifacts may consume storage quota.
            </p>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowUploadModal(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpload}
              disabled={!uploadFile || !uploadName || uploading}
            >
              {uploading ? 'Uploading...' : 'Upload'}
            </Button>
          </div>
        </div>
      </Modal>
    </Card>
  );
}