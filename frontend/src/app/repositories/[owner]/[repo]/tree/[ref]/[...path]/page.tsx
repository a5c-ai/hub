'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import RepositoryBrowser from '@/components/repository/RepositoryBrowser';
import { MobileRepositoryBrowser } from '@/components/mobile/MobileRepositoryBrowser';
import { useMobile } from '@/hooks/useDevice';
import api from '@/lib/api';
import { Repository, TreeEntry, Tree } from '@/types';
import { repoApi } from '@/lib/api';

export default function TreePage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const ref = params.ref as string;
  const pathArray = params.path as string[];
  const currentPath = pathArray ? pathArray.join('/') : '';
  
  const [repository, setRepository] = useState<Repository | null>(null);
  const [tree, setTree] = useState<Tree | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const isMobile = useMobile();

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch repository and tree data in parallel
        const [repoResponse, treeResponse] = await Promise.all([
          api.get(`/repositories/${owner}/${repo}`),
          repoApi.getTree(owner, repo, currentPath, ref)
        ]);
        
        setRepository(repoResponse.data);
        setTree(treeResponse.data);
      } catch (err: unknown) {
        const errorMessage = err instanceof Error && 'response' in err && 
          typeof err.response === 'object' && err.response && 
          'data' in err.response && 
          typeof err.response.data === 'object' && err.response.data &&
          'message' in err.response.data && 
          typeof err.response.data.message === 'string'
          ? err.response.data.message 
          : 'Failed to fetch data';
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [owner, repo, currentPath, ref]);

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
            <div className="h-4 bg-muted rounded w-2/3 mb-8"></div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    // Check if this might be an empty repository (404 error)
    const isEmptyRepo = error.includes('404') || error.includes('Not Found') || error.includes('repository is empty');
    
    if (isEmptyRepo) {
      // Redirect to main repository page to show empty repo instructions
      window.location.href = `/repositories/${owner}/${repo}`;
      return null;
    }

    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-destructive text-lg mb-4">Error: {error}</div>
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (!repository) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-muted-foreground text-lg">Repository not found</div>
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
          <span className="text-foreground font-medium">tree</span>
          <span>/</span>
          <span className="text-foreground font-medium">{ref}</span>
          {currentPath && (
            <>
              <span>/</span>
              <span className="text-foreground font-medium">{currentPath}</span>
            </>
          )}
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
              {repository.owner?.username || owner}/{repository.name}
            </h1>
            <div className="flex items-center space-x-2">
              <Badge variant={repository.private ? 'secondary' : 'default'}>
                {repository.private ? 'Private' : 'Public'}
              </Badge>
            </div>
          </div>
        </div>

        {/* Repository Browser */}
        {isMobile ? (
          <MobileRepositoryBrowser
            files={tree?.entries.map((entry: TreeEntry) => ({
              name: entry.name,
              type: entry.type === 'tree' ? 'directory' : 'file',
              size: entry.size?.toString(),
              lastModified: undefined, // TreeEntry doesn't include commit info
              url: entry.type === 'tree' 
                ? `/repositories/${owner}/${repo}/tree/${ref}/${currentPath ? currentPath + '/' : ''}${entry.name}`
                : `/repositories/${owner}/${repo}/blob/${ref}/${currentPath ? currentPath + '/' : ''}${entry.name}`
            })) || []}
            currentPath={currentPath}
            repositoryUrl={`/repositories/${owner}/${repo}`}
            onRefresh={() => window.location.reload()}
            loading={loading}
          />
        ) : (
          <Card>
            <div className="p-6">
              <RepositoryBrowser
                owner={owner}
                repo={repo}
                repository={repository}
                currentPath={currentPath}
                currentRef={ref}
              />
            </div>
          </Card>
        )}
      </div>
    </AppLayout>
  );
} 