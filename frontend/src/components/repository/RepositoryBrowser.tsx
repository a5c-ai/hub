'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { repoApi } from '@/lib/api';
import { TreeEntry, Tree, Repository } from '@/types';
import { Button } from '@/components/ui/Button';

interface RepositoryBrowserProps {
  owner: string;
  repo: string;
  repository: Repository;
  currentPath?: string;
  currentRef?: string;
}

export default function RepositoryBrowser({ 
  owner, 
  repo, 
  repository, 
  currentPath = '', 
  currentRef 
}: RepositoryBrowserProps) {
  const [tree, setTree] = useState<Tree | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actualRef, setActualRef] = useState<string | null>(null);
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  
  const ref = currentRef || actualRef || repository.default_branch;

  // Fetch available branches
  useEffect(() => {
    const fetchBranches = async () => {
      try {
        const branchesResponse = await repoApi.getBranches(owner, repo);
        const branches = branchesResponse.data;
        const branchNames = branches.map((branch: any) => branch.name);
        setAvailableBranches(branchNames);
      } catch (branchError) {
        console.warn('Failed to fetch branches:', branchError);
      }
    };

    fetchBranches();
  }, [owner, repo]);

  // Fetch tree content
  useEffect(() => {
    const fetchTree = async () => {
      try {
        setLoading(true);
        setError(null);

        let response;
        let usedRef = ref;
        try {
          // Try with the specified ref first
          response = await repoApi.getTree(owner, repo, currentPath, ref);
        } catch (firstError) {
          // If that fails and we're using the default branch, try available branches
          if (ref === repository.default_branch && availableBranches.length > 0) {
            try {
              usedRef = availableBranches[0];
              response = await repoApi.getTree(owner, repo, currentPath, usedRef);
              setActualRef(usedRef);
            } catch (branchError) {
              throw firstError;
            }
          } else {
            throw firstError;
          }
        }
        
        setTree(response.data);
      } catch (err: unknown) {
        const errorMessage = err instanceof Error && 'response' in err && 
          typeof err.response === 'object' && err.response && 
          'data' in err.response && 
          typeof err.response.data === 'object' && err.response.data &&
          'message' in err.response.data && 
          typeof err.response.data.message === 'string'
          ? err.response.data.message 
          : 'Failed to load repository content';
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchTree();
  }, [owner, repo, currentPath, ref, repository.default_branch, availableBranches]);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getFileIcon = (entry: TreeEntry) => {
    if (entry.type === 'tree') {
      return (
        <svg className="w-4 h-4 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
        </svg>
      );
    }
    
    // File icon
    return (
      <svg className="w-4 h-4 text-muted-foreground" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
      </svg>
    );
  };

  const getBreadcrumbs = () => {
    if (!currentPath) return [];
    
    const parts = currentPath.split('/').filter(Boolean);
    const breadcrumbs = [];
    
    for (let i = 0; i < parts.length; i++) {
      const path = parts.slice(0, i + 1).join('/');
      breadcrumbs.push({
        name: parts[i],
        path: path
      });
    }
    
    return breadcrumbs;
  };

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="flex items-center space-x-3">
              <div className="w-4 h-4 bg-muted rounded"></div>
              <div className="h-4 bg-muted rounded flex-1"></div>
              <div className="w-16 h-4 bg-muted rounded"></div>
              <div className="w-24 h-4 bg-muted rounded"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <div className="text-red-600 mb-4">{error}</div>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={() => window.location.reload()}
        >
          Try Again
        </Button>
      </div>
    );
  }



  if (!tree || !tree.entries || tree.entries.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <svg className="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
        </svg>
        <p>This directory is empty</p>
      </div>
    );
  }

  const breadcrumbs = getBreadcrumbs();

  return (
    <div>
      {/* Breadcrumb navigation */}
      <div className="flex items-center space-x-2 text-sm text-muted-foreground mb-4 p-3 bg-muted rounded-md">
        <Link 
          href={`/repositories/${owner}/${repo}`}
          className="hover:text-blue-600 dark:hover:text-blue-400 transition-colors font-medium"
        >
          {owner}/{repo}
        </Link>
        {currentPath && (
          <>
            <span>/</span>
            <span className="text-foreground">tree</span>
            <span>/</span>
            <Link 
              href={`/repositories/${owner}/${repo}`}
              className="hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
            >
              {ref}
            </Link>
          </>
        )}
        {breadcrumbs.map((crumb, index) => (
          <div key={index} className="flex items-center space-x-2">
            <span>/</span>
            <Link 
              href={`/repositories/${owner}/${repo}/tree/${ref}/${crumb.path}`}
              className="hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
            >
              {crumb.name}
            </Link>
          </div>
        ))}
      </div>

      {/* File listing */}
              <div className="border border-border rounded-md overflow-hidden">
          <div className="bg-muted px-4 py-2 border-b border-border">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
                                            <select 
                 className="text-sm border border-input rounded-md px-3 py-1 bg-background text-foreground"
                 value={ref}
                 aria-label="Select branch or ref"
                 onChange={(e) => {
                   const newRef = e.target.value;
                   if (currentPath) {
                     window.location.href = `/repositories/${owner}/${repo}/tree/${newRef}/${currentPath}`;
                   } else {
                     window.location.href = `/repositories/${owner}/${repo}?ref=${newRef}`;
                   }
                 }}
               >
                 {availableBranches.length > 0 ? 
                   availableBranches.map((branch) => (
                     <option key={branch} value={branch}>{branch}</option>
                   )) :
                   <option value={repository.default_branch}>{repository.default_branch}</option>
                 }
               </select>
              <span className="text-sm text-muted-foreground">
                {tree.entries.length} {tree.entries.length === 1 ? 'item' : 'items'}
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <Button size="sm" variant="outline">
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Code
              </Button>
            </div>
          </div>
        </div>

        <div className="divide-y divide-gray-200">
          {/* Go up directory link */}
          {currentPath && (
            <div className="px-4 py-3 hover:bg-muted transition-colors">
              <Link 
                href={
                  currentPath.split('/').slice(0, -1).length > 0
                    ? `/repositories/${owner}/${repo}/tree/${ref}/${currentPath.split('/').slice(0, -1).join('/')}`
                    : `/repositories/${owner}/${repo}`
                }
                className="flex items-center space-x-3 text-sm"
              >
                <svg className="w-4 h-4 text-muted-foreground" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M9.707 16.707a1 1 0 01-1.414 0l-6-6a1 1 0 010-1.414l6-6a1 1 0 011.414 1.414L5.414 9H17a1 1 0 110 2H5.414l4.293 4.293a1 1 0 010 1.414z" clipRule="evenodd" />
                </svg>
                <span className="text-blue-600 hover:underline">..</span>
              </Link>
            </div>
          )}

          {/* File/directory entries */}
          {tree.entries.map((entry) => (
            <div key={entry.path} className="px-4 py-3 hover:bg-muted transition-colors">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3 flex-1 min-w-0">
                  {getFileIcon(entry)}
                  <Link 
                    href={
                      entry.type === 'tree' 
                        ? `/repositories/${owner}/${repo}/tree/${ref}/${entry.path}`
                        : `/repositories/${owner}/${repo}/blob/${ref}/${entry.path}`
                    }
                    className="text-sm text-blue-600 hover:underline truncate"
                  >
                    {entry.name}
                  </Link>
                </div>
                <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                  {entry.type === 'blob' && (
                    <span>{formatFileSize(entry.size)}</span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
} 