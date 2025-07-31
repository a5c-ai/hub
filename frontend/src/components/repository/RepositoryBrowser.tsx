'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { repoApi } from '@/lib/api';
import { TreeEntry, Tree, Repository } from '@/types';
import { Button } from '@/components/ui/Button';
import { createErrorHandler } from '@/lib/utils';

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
        const branchNames = branches.map((branch: { name: string }) => branch.name);
        setAvailableBranches(branchNames);
      } catch (branchError) {
        console.warn('Failed to fetch branches:', branchError);
      }
    };

    fetchBranches();
  }, [owner, repo]);

  // Fetch tree content
  useEffect(() => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const fetchTree = async () => {
      const operation = async () => {
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
            } catch {
              throw firstError;
            }
          } else {
            throw firstError;
          }
        }
        
        return response.data;
      };

      const result = await handleError(operation);
      if (result) {
        setTree(result);
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

  const retryFetch = async () => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
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
          } catch {
            throw firstError;
          }
        } else {
          throw firstError;
        }
      }
      
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setTree(result);
    }
  };

  if (error) {
    // Check if this might be an empty repository (404 error)
    const isEmptyRepo = error.includes('404') || error.includes('Not Found') || error.includes('repository is empty') || availableBranches.length === 0;
    
    if (isEmptyRepo) {
      return (
        <div className="text-center py-16">
          <div className="max-w-4xl mx-auto px-4">
            {/* Header Section */}
            <div className="mb-12">
              <svg className="w-20 h-20 mx-auto text-muted-foreground mb-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <h3 className="text-2xl font-bold text-foreground mb-3">This repository is empty</h3>
              <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
                Get started by creating a new file or uploading existing files.
              </p>
            </div>

            {/* Quick Setup Section */}
            <div className="bg-muted/50 rounded-xl p-8 mb-12 text-left max-w-3xl mx-auto">
              <h4 className="text-lg font-semibold text-foreground mb-6 text-center">
                Quick setup — if you've done this kind of thing before
              </h4>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-medium text-foreground uppercase tracking-wide">HTTPS</span>
                    <Button 
                      size="sm" 
                      variant="outline"
                      onClick={() => navigator.clipboard.writeText(repository.clone_url || `http://localhost:8080/tmuskal/${repo}.git`)}
                    >
                      Copy
                    </Button>
                  </div>
                  <div className="bg-background border rounded-md p-4">
                    <code className="text-sm font-mono text-foreground break-all">
                      {repository.clone_url || `http://localhost:8080/tmuskal/${repo}.git`}
                    </code>
                  </div>
                </div>

                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-medium text-foreground uppercase tracking-wide">SSH</span>
                    <Button 
                      size="sm" 
                      variant="outline"
                      onClick={() => navigator.clipboard.writeText(repository.ssh_url || `git@localhost:tmuskal/${repo}.git`)}
                    >
                      Copy
                    </Button>
                  </div>
                  <div className="bg-background border rounded-md p-4">
                    <code className="text-sm font-mono text-foreground break-all">
                      {repository.ssh_url || `git@localhost:tmuskal/${repo}.git`}
                    </code>
                  </div>
                </div>
              </div>
            </div>

            {/* Command Instructions */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 text-left max-w-6xl mx-auto">
              <div className="bg-background border border-border rounded-xl p-8">
                <h4 className="text-lg font-semibold text-foreground mb-6">
                  …or create a new repository on the command line
                </h4>
                <div className="bg-muted rounded-lg p-6 overflow-x-auto">
                  <pre className="text-sm font-mono text-foreground leading-relaxed">
{`echo "# ${repo}" >> README.md
git init
git add README.md
git commit -m "first commit"
git branch -M main
git remote add origin ${repository.clone_url || `http://localhost:8080/tmuskal/${repo}.git`}
git push -u origin main`}
                  </pre>
                </div>
              </div>

              <div className="bg-background border border-border rounded-xl p-8">
                <h4 className="text-lg font-semibold text-foreground mb-6">
                  …or push an existing repository from the command line
                </h4>
                <div className="bg-muted rounded-lg p-6 overflow-x-auto">
                  <pre className="text-sm font-mono text-foreground leading-relaxed">
{`git remote add origin ${repository.clone_url || `http://localhost:8080/tmuskal/${repo}.git`}
git branch -M main
git push -u origin main`}
                  </pre>
                </div>
              </div>
            </div>

            {/* Refresh Button */}
            <div className="mt-12 text-center">
              <Button 
                variant="outline" 
                onClick={retryFetch}
                disabled={loading}
                className="px-6 py-2"
              >
                {loading ? 'Checking...' : 'Refresh'}
              </Button>
            </div>
          </div>
        </div>
      );
    }

    // For other errors, show the generic error message
    return (
      <div className="text-center py-8">
        <div className="text-destructive mb-4">{error}</div>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={retryFetch}
          disabled={loading}
        >
          {loading ? 'Retrying...' : 'Try Again'}
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