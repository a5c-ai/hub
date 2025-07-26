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
import LanguageStats from '@/components/repository/LanguageStats';
import api, { repoApi } from '@/lib/api';
import { Repository } from '@/types';
import { createErrorHandler } from '@/lib/utils';
import { useRepositoryStats } from '@/hooks/useRepositoryStats';

export default function RepositoryDetailsPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const [repository, setRepository] = useState<Repository | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isStarred, setIsStarred] = useState(false);
  const [starCount, setStarCount] = useState(0);
  const [forkCount, setForkCount] = useState(0);
  const [starLoading, setStarLoading] = useState(false);
  const [forkLoading, setForkLoading] = useState(false);

  // Fetch repository statistics
  const { 
    statistics, 
    languages, 
    contributors, 
    loading: statsLoading, 
    error: statsError 
  } = useRepositoryStats(owner, repo);

  const fetchRepository = async () => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
      const response = await api.get(`/repositories/${owner}/${repo}`);
      
      // Check if user has starred this repository
      let starred = false;
      try {
        const starResponse = await repoApi.checkStarred(owner, repo);
        starred = (starResponse.data as { starred: boolean }).starred;
      } catch (starErr) {
        // If user is not authenticated, star check will fail - that's okay
        console.debug('Could not check star status:', starErr);
      }
      
      return { repository: response.data, starred };
    };

    const result = await handleError(operation);
    if (result) {
      setRepository(result.repository);
      setStarCount(result.repository.stargazers_count || 0);
      setForkCount(result.repository.forks_count || 0);
      setIsStarred(result.starred);
    }
  };

  useEffect(() => {
    fetchRepository();
  }, [owner, repo]);

  const handleStarToggle = async () => {
    if (starLoading) return;
    
    try {
      setStarLoading(true);
      if (isStarred) {
        await repoApi.unstarRepository(owner, repo);
        setIsStarred(false);
        setStarCount(prev => Math.max(0, prev - 1));
      } else {
        await repoApi.starRepository(owner, repo);
        setIsStarred(true);
        setStarCount(prev => prev + 1);
      }
    } catch (err) {
      console.error('Failed to toggle star:', err);
      // You might want to show a toast notification here
    } finally {
      setStarLoading(false);
    }
  };

  const handleFork = async () => {
    if (forkLoading) return;
    
    try {
      setForkLoading(true);
      await repoApi.forkRepository(owner, repo);
      setForkCount(prev => prev + 1);
      // You might want to show a success message or redirect to the forked repository
    } catch (err) {
      console.error('Failed to fork repository:', err);
      // You might want to show an error message here
    } finally {
      setForkLoading(false);
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
            <div className="h-4 bg-muted rounded w-2/3 mb-8"></div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2">
                <div className="h-64 bg-muted rounded"></div>
              </div>
              <div>
                <div className="h-48 bg-muted rounded"></div>
              </div>
            </div>
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
            <Button onClick={fetchRepository} disabled={loading}>
              {loading ? 'Retrying...' : 'Try Again'}
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

  // Generate fallback URLs if they're not provided by the backend
  const cloneUrl = repository.clone_url || `https://github.com/${owner}/${repo}.git`;
  const sshUrl = repository.ssh_url || `git@github.com:${owner}/${repo}.git`;

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).catch(err => {
      console.error('Failed to copy text: ', err);
    });
  };

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
          <span className="text-foreground font-medium">{owner}/{repo}</span>
        </nav>

        {/* Repository Header */}
        <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between mb-8">
          <div className="flex items-start space-x-4 mb-4 lg:mb-0">
            <Avatar
              src={repository.owner?.avatar_url}
              alt={repository.owner?.username || 'Repository owner'}
              size="lg"
            />
            <div>
              <div className="flex items-center space-x-2 mb-2">
                <h1 className="text-3xl font-bold text-foreground">
                  {repository.owner?.username || owner}/{repository.name}
                </h1>
                <Badge variant={repository.private ? 'secondary' : 'default'}>
                  {repository.private ? 'Private' : 'Public'}
                </Badge>
                {repository.fork && (
                  <Badge variant="outline">Fork</Badge>
                )}
              </div>
              {repository.description && (
                <p className="text-muted-foreground text-lg">{repository.description}</p>
              )}
            </div>
          </div>
          
          <div className="flex items-center space-x-3">
            <Button 
              variant={isStarred ? "default" : "outline"} 
              size="sm" 
              onClick={handleStarToggle} 
              disabled={starLoading}
            >
              <svg className="w-4 h-4 mr-2" fill={isStarred ? "currentColor" : "none"} stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
              </svg>
              {starLoading ? 'Loading...' : isStarred ? 'Starred' : 'Star'}
              <Badge variant="outline" className="ml-2">
                {starCount}
              </Badge>
            </Button>
            
            <Button variant="outline" size="sm" onClick={handleFork} disabled={forkLoading}>
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
              </svg>
              {forkLoading ? 'Forking...' : 'Fork'}
              <Badge variant="outline" className="ml-2">
                {forkCount}
              </Badge>
            </Button>

            <Link href={`/repositories/${owner}/${repo}/settings`}>
              <Button variant="outline" size="sm">
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                Settings
              </Button>
            </Link>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            {/* Navigation Tabs */}
            <div className="border-b border-border mb-6">
              <nav className="-mb-px flex space-x-8">
                <Link
                  href={`/repositories/${owner}/${repo}`}
                  className="border-b-2 border-blue-500 py-2 px-1 text-sm font-medium text-blue-600"
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Code
                </Link>
                <Link
                  href={`/repositories/${owner}/${repo}/issues`}
                  className="border-transparent text-muted-foreground hover:text-foreground hover:border-border border-b-2 py-2 px-1 text-sm font-medium"
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.732L13.732 4.268c-.77-1.064-2.694-1.064-3.464 0L3.34 16.268C2.57 17.333 3.532 19 5.072 19z" />
                  </svg>
                  Issues
                  <Badge variant="secondary" className="ml-2">
                    {repository.open_issues_count}
                  </Badge>
                </Link>
                <Link
                  href={`/repositories/${owner}/${repo}/pulls`}
                  className="border-transparent text-muted-foreground hover:text-foreground hover:border-border border-b-2 py-2 px-1 text-sm font-medium"
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                  Pull requests
                </Link>
                <Link
                  href={`/repositories/${owner}/${repo}/actions`}
                  className="border-transparent text-muted-foreground hover:text-foreground hover:border-border border-b-2 py-2 px-1 text-sm font-medium"
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  Actions
                </Link>
                <Link
                  href={`/repositories/${owner}/${repo}/insights`}
                  className="border-transparent text-muted-foreground hover:text-foreground hover:border-border border-b-2 py-2 px-1 text-sm font-medium"
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                  Insights
                </Link>
              </nav>
            </div>

            {/* Repository Content */}
            <Card>
              <div className="p-6">
                <RepositoryBrowser
                  owner={owner}
                  repo={repo}
                  repository={repository}
                />
              </div>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* About */}
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-foreground mb-4">About</h3>
                {repository.description && (
                  <p className="text-muted-foreground mb-4">{repository.description}</p>
                )}
                
                <div className="space-y-3">
                  {repository.language && (
                    <div className="flex items-center">
                      <div className="w-3 h-3 rounded-full bg-blue-500 mr-2"></div>
                      <span className="text-sm text-muted-foreground">{repository.language}</span>
                    </div>
                  )}
                  
                  <div className="flex items-center text-sm text-muted-foreground">
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                    </svg>
                    {repository.stargazers_count} stars
                  </div>
                  
                  <div className="flex items-center text-sm text-muted-foreground">
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    {repository.watchers_count} watching
                  </div>
                  
                  <div className="flex items-center text-sm text-muted-foreground">
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
                    </svg>
                    {repository.forks_count} forks
                  </div>
                </div>
              </div>
            </Card>

            {/* Clone URLs */}
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-foreground mb-4">Clone</h3>
                <div className="space-y-3">
                  <div>
                    <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">HTTPS</label>
                    <div className="flex mt-1">
                      <input 
                        type="text" 
                        value={cloneUrl}
                        readOnly
                        title="HTTPS clone URL"
                        aria-label="HTTPS clone URL"
                        className="flex-1 text-sm bg-background border border-input rounded-l-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring text-foreground"
                      />
                      <Button 
                        size="sm" 
                        variant="outline" 
                        className="rounded-l-none border-l-0"
                        onClick={() => copyToClipboard(cloneUrl)}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                  <div>
                    <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">SSH</label>
                    <div className="flex mt-1">
                      <input 
                        type="text" 
                        value={sshUrl}
                        readOnly
                        title="SSH clone URL"
                        aria-label="SSH clone URL"
                        className="flex-1 text-sm bg-background border border-input rounded-l-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-ring focus:border-ring text-foreground"
                      />
                      <Button 
                        size="sm" 
                        variant="outline" 
                        className="rounded-l-none border-l-0"
                        onClick={() => copyToClipboard(sshUrl)}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </Card>

            {/* Language Statistics */}
            {languages && languages.length > 0 && (
              <LanguageStats
                languages={languages}
                primaryLanguage={statistics?.primary_language}
                showPercentages={true}
                showBytes={true}
                compact={false}
              />
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  );
}