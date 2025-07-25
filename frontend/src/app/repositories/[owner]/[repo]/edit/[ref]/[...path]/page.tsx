'use client';

import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import { Input } from '@/components/ui/Input';
import { repoApi } from '@/lib/api';
import api from '@/lib/api';
import { Repository, File } from '@/types';
import { useAuthStore } from '@/store/auth';

export default function EditPage() {
  const params = useParams();
  const router = useRouter();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const ref = params.ref as string;
  const pathArray = params.path as string[];
  const filePath = pathArray ? pathArray.join('/') : '';
  
  const { user, isAuthenticated } = useAuthStore();
  const [repository, setRepository] = useState<Repository | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [content, setContent] = useState('');
  const [commitMessage, setCommitMessage] = useState('');
  const [commitDescription, setCommitDescription] = useState('');
  const [isCommitting, setIsCommitting] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch repository info and file content in parallel
        const [repoResponse, fileResponse] = await Promise.all([
          api.get(`/repositories/${owner}/${repo}`),
          repoApi.getFile(owner, repo, filePath, ref)
        ]);
        
        setRepository(repoResponse.data);
        setFile(fileResponse.data);
        setContent(fileResponse.data.content);
        setCommitMessage(`Update ${fileResponse.data.name}`);
      } catch (err: unknown) {
        const errorMessage = err instanceof Error && 'response' in err && 
          typeof err.response === 'object' && err.response && 
          'data' in err.response && 
          typeof err.response.data === 'object' && err.response.data &&
          'message' in err.response.data && 
          typeof err.response.data.message === 'string'
          ? err.response.data.message 
          : 'Failed to fetch file';
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [owner, repo, filePath, ref]);

  const handleCommit = async () => {
    if (!file || !commitMessage.trim() || !user || !isAuthenticated) return;
    
    setIsCommitting(true);
    try {
      const now = new Date();
      const commitData = {
        content: content,
        message: commitMessage,
        branch: ref,
        sha: file.sha, // Include the file SHA for conflict detection
        author: {
          name: user.name || user.username,
          email: user.email,
          date: now.toISOString()
        },
        committer: {
          name: user.name || user.username,
          email: user.email,
          date: now.toISOString()
        }
      };

      await api.put(`/repositories/${owner}/${repo}/contents/${filePath}`, commitData);
      
      // Redirect back to the file view
      router.push(`/repositories/${owner}/${repo}/blob/${ref}/${filePath}`);
    } catch (err: unknown) {
      const errorMessage = err instanceof Error && 'response' in err && 
        typeof err.response === 'object' && err.response && 
        'data' in err.response && 
        typeof err.response.data === 'object' && err.response.data &&
        'message' in err.response.data && 
        typeof err.response.data.message === 'string'
        ? err.response.data.message 
        : 'Failed to commit changes';
      setError(errorMessage);
    } finally {
      setIsCommitting(false);
    }
  };

  const cancelEdit = () => {
    router.push(`/repositories/${owner}/${repo}/blob/${ref}/${filePath}`);
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
            <div className="h-64 bg-muted rounded"></div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (!isAuthenticated || !user) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: You must be authenticated to edit files</div>
            <Button onClick={() => router.push('/login')}>
              Go to Login
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error || !repository || !file) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error || 'File not found'}</div>
            <Button onClick={() => window.location.reload()}>
              Try Again
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
          <Link 
            href="/repositories" 
            className="hover:text-foreground transition-colors"
          >
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
          <span className="text-foreground font-medium">edit</span>
          <span>/</span>
          <span className="text-foreground font-medium">{ref}</span>
          <span>/</span>
          <span className="text-foreground font-medium">{filePath}</span>
        </nav>

        {/* Repository Header */}
        <div className="flex items-center space-x-4 mb-6">
          <Avatar
            src={repository.owner?.avatar_url}
            alt={repository.owner?.username || 'Repository owner'}
            size="md"
          />
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              Edit {filePath}
            </h1>
            <div className="flex items-center space-x-2">
              <Badge variant={repository.private ? 'secondary' : 'default'}>
                {repository.private ? 'Private' : 'Public'}
              </Badge>
            </div>
          </div>
        </div>

        {/* Edit Form */}
        <div className="space-y-6">
          {/* File Editor */}
          <Card>
            <div className="border-b border-border px-6 py-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4">
                  <h2 className="text-lg font-semibold text-foreground">
                    Editing {file.name}
                  </h2>
                  <span className="text-sm text-muted-foreground">
                    on branch {ref}
                  </span>
                </div>
                <div className="flex items-center space-x-2">
                  <Button size="sm" variant="outline" onClick={cancelEdit}>
                    Cancel
                  </Button>
                  <Link 
                    href={`/repositories/${owner}/${repo}/blob/${ref}/${filePath}`}
                    className="inline-flex"
                  >
                    <Button size="sm" variant="outline">
                      Preview
                    </Button>
                  </Link>
                </div>
              </div>
            </div>
            <div className="p-6">
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                className="w-full h-96 p-4 border border-input rounded-md font-mono text-sm resize-y bg-background text-foreground focus:ring-2 focus:ring-ring focus:border-ring placeholder:text-muted-foreground"
                placeholder="Enter file content..."
                style={{ minHeight: '400px' }}
              />
            </div>
          </Card>

          {/* Commit Form */}
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">
                Commit changes
              </h3>
              <div className="space-y-4">
                <div>
                  <Input
                    type="text"
                    value={commitMessage}
                    onChange={(e) => setCommitMessage(e.target.value)}
                    placeholder="Commit message (required)"
                    className="w-full"
                    required
                  />
                </div>
                <div>
                  <textarea
                    value={commitDescription}
                    onChange={(e) => setCommitDescription(e.target.value)}
                    placeholder="Extended description (optional)"
                    className="w-full p-3 border border-gray-300 dark:border-gray-600 rounded-md resize-y bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    rows={3}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <div className="text-sm text-gray-500 dark:text-gray-400">
                    Commit directly to the <code className="bg-gray-100 dark:bg-gray-700 px-1 rounded">{ref}</code> branch.
                  </div>
                  <div className="flex items-center space-x-3">
                    <Button 
                      variant="outline" 
                      onClick={cancelEdit}
                      disabled={isCommitting}
                    >
                      Cancel
                    </Button>
                    <Button 
                      onClick={handleCommit}
                      disabled={!commitMessage.trim() || isCommitting || !user || !isAuthenticated}
                      className="min-w-[120px]"
                    >
                      {isCommitting ? (
                        <div className="flex items-center space-x-2">
                          <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                          <span>Committing...</span>
                        </div>
                      ) : (
                        'Commit changes'
                      )}
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
} 