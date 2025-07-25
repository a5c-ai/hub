'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import {
  PlusIcon,
  FolderIcon,
  StarIcon,
  ClockIcon,
  CodeBracketIcon as GitBranchIcon,
} from '@heroicons/react/24/outline';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  Avatar,
  Badge,
} from '@/components/ui';
import { AppLayout } from '@/components/layout/AppLayout';
import { useAuthStore } from '@/store/auth';
import { useRepositoryStore } from '@/store/repository';
import { formatRelativeTime } from '@/lib/utils';

export default function DashboardPage() {
  const { user, isAuthenticated } = useAuthStore();
  const { 
    repositories, 
    isLoading: repoLoading, 
    error: repoError, 
    fetchRepositories,
    clearError 
  } = useRepositoryStore();
  const mockActivity = [
    {
      id: '1',
      type: 'push',
      repository: repositories[0]?.full_name || 'user/repository',
      message: 'Added new authentication middleware',
      timestamp: repositories[0]?.updated_at || new Date(Date.now() - 3600000).toISOString(),
    },
    {
      id: '2',
      type: 'pull_request',
      repository: repositories[1]?.full_name || 'user/another-repo',
      message: 'Opened pull request: Implement user management endpoints',
      timestamp: repositories[1]?.updated_at || new Date(Date.now() - 86400000).toISOString(),
    },
    {
      id: '3',
      type: 'issue',
      repository: repositories[2]?.full_name || 'user/third-repo',
      message: 'Created issue: Fix login screen layout on tablet',
      timestamp: repositories[2]?.updated_at || new Date(Date.now() - 172800000).toISOString(),
    },
  ];

  const [] = useState(mockActivity);

  useEffect(() => {
    if (isAuthenticated) {
      // Fetch recent repositories (limit to 6 for dashboard)
      fetchRepositories({ per_page: 6, sort: 'updated' });
    }
  }, [isAuthenticated, fetchRepositories]);

  useEffect(() => {
    if (repoError) {
      console.error('Repository error:', repoError);
      // Auto-clear error after 5 seconds
      const timer = setTimeout(() => {
        clearError();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [repoError, clearError]);

  // Calculate stats from real repository data
  const totalRepos = repositories.length;
  const totalStars = repositories.reduce((sum, repo) => sum + (repo.stargazers_count || 0), 0);
  const totalForks = repositories.reduce((sum, repo) => sum + (repo.forks_count || 0), 0);

  return (
    <AppLayout>
      <div className="p-6">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">
            Welcome back, {user?.name || user?.username}!
          </h1>
          <p className="text-muted-foreground mt-2">
            Here&apos;s what&apos;s happening with your repositories and projects.
          </p>
        </div>

        {repoError && (
          <div className="mb-6 rounded-md bg-destructive/10 p-4">
            <div className="text-sm text-destructive">
              Failed to load repositories: {repoError}
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          {/* Quick stats */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Repositories</CardTitle>
              <FolderIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {repoLoading ? '...' : totalRepos}
              </div>
              <p className="text-xs text-muted-foreground">
                {repoError ? 'Error loading data' : totalRepos === 0 ? 'No repositories yet' : 'Your repositories'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Stars</CardTitle>
              <StarIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {repoLoading ? '...' : totalStars}
              </div>
              <p className="text-xs text-muted-foreground">
                {repoError ? 'Error loading data' : totalStars === 0 ? 'No stars yet' : 'Across all repositories'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Forks</CardTitle>
              <GitBranchIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {repoLoading ? '...' : totalForks}
              </div>
              <p className="text-xs text-muted-foreground">
                {repoError ? 'Error loading data' : totalForks === 0 ? 'No forks yet' : 'Repository forks'}
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Recent repositories */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Recent Repositories</CardTitle>
                  <CardDescription>Your most recently updated repositories</CardDescription>
                </div>
                <Button asChild>
                  <Link href="/repositories/new">
                    <PlusIcon className="h-4 w-4 mr-2" />
                    New
                  </Link>
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {repoLoading ? (
                <div className="space-y-4">
                  {[1, 2, 3].map((i) => (
                    <div key={i} className="flex items-center justify-between p-3 rounded-lg border animate-pulse">
                      <div className="flex-1">
                        <div className="h-4 bg-muted rounded w-1/3 mb-2"></div>
                        <div className="h-3 bg-muted rounded w-2/3 mb-2"></div>
                        <div className="flex space-x-4">
                          <div className="h-3 bg-muted rounded w-16"></div>
                          <div className="h-3 bg-muted rounded w-12"></div>
                          <div className="h-3 bg-muted rounded w-20"></div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : repoError ? (
                <div className="text-center py-8">
                  <div className="text-destructive">{repoError}</div>
                </div>
              ) : repositories.length === 0 ? (
                <div className="text-center py-8">
                  <FolderIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-foreground mb-2">
                    No repositories yet
                  </h3>
                  <p className="text-muted-foreground mb-4">
                    Create your first repository to get started
                  </p>
                  <Button asChild>
                    <Link href="/repositories/new">
                      <PlusIcon className="h-4 w-4 mr-2" />
                      Create repository
                    </Link>
                  </Button>
                </div>
              ) : (
                repositories.map((repo) => (
                  <div key={repo.id} className="flex items-center justify-between p-3 rounded-lg border">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <Link
                          href={`/repositories/${repo.full_name}`}
                          className="font-medium text-foreground hover:text-primary"
                        >
                          {repo.name}
                        </Link>
                        {repo.private && <Badge variant="secondary" size="sm">Private</Badge>}
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        {repo.description || 'No description provided'}
                      </p>
                      <div className="flex items-center space-x-4 mt-2 text-xs text-muted-foreground">
                        {repo.language && (
                          <span className="flex items-center">
                            <span className="w-2 h-2 rounded-full bg-primary mr-1"></span>
                            {repo.language}
                          </span>
                        )}
                        <span className="flex items-center">
                          <StarIcon className="h-3 w-3 mr-1" />
                          {repo.stargazers_count || 0}
                        </span>
                        <span className="flex items-center">
                          <GitBranchIcon className="h-3 w-3 mr-1" />
                          {repo.forks_count || 0}
                        </span>
                        <span>Updated {formatRelativeTime(repo.updated_at)}</span>
                      </div>
                    </div>
                  </div>
                ))
              )}
              <div className="text-center">
                <Button variant="ghost" asChild>
                  <Link href="/repositories">View all repositories</Link>
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Recent activity */}
          <Card>
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Your recent actions across all repositories</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {repoLoading ? (
                <div className="text-center py-8">
                  <div className="text-muted-foreground">Loading activity...</div>
                </div>
              ) : repositories.length === 0 ? (
                <div className="text-center py-8">
                  <ClockIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-foreground mb-2">
                    No activity yet
                  </h3>
                  <p className="text-muted-foreground">
                    Activity will appear here as you work on repositories
                  </p>
                </div>
              ) : (
                <>
                  {mockActivity.slice(0, repositories.length).map((item) => (
                    <div key={item.id} className="flex items-start space-x-3">
                      <div className="flex-shrink-0">
                        <Avatar size="sm" name={user?.name || user?.username} />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm">
                          <span className="font-medium text-foreground">You</span>
                          <span className="text-muted-foreground"> {item.message}</span>
                        </div>
                        <div className="flex items-center space-x-2 mt-1 text-xs text-muted-foreground">
                          <Link
                            href={`/repositories/${item.repository}`}
                            className="hover:text-primary"
                          >
                            {item.repository}
                          </Link>
                          <span>•</span>
                          <span className="flex items-center">
                            <ClockIcon className="h-3 w-3 mr-1" />
                            {formatRelativeTime(item.timestamp)}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))}
                  <div className="text-center">
                    <Button variant="ghost" asChild>
                      <Link href="/activity">View all activity</Link>
                    </Button>
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
}