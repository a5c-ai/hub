'use client';

import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card } from '@/components/ui/Card';
import { Avatar } from '@/components/ui/Avatar';
import { Button } from '@/components/ui/Button';
import { apiClient } from '@/lib/api';
import Link from 'next/link';

import { createErrorHandler } from '@/lib/utils';

interface ActivityItem {
  id: string;
  type: 'push' | 'pull_request' | 'issue' | 'fork' | 'star' | 'follow' | 'create_repository';
  action: string;
  actor: {
    id: string;
    username: string;
    avatar_url?: string;
  };
  repository?: {
    id: string;
    name: string;
    full_name: string;
    owner: {
      username: string;
    };
  };
  payload: {
    commits?: Array<{ sha?: string; message: string }>;
    number?: number;
    title?: string;
    target?: { username: string };
    action?: string;
    ref?: string;
  };
  created_at: string;
}

interface UserActivityResponse {
  activities: ActivityItem[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
    has_more: boolean;
  };
}

export default function ActivityPage() {
  const [activities, setActivities] = useState<ActivityItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'own' | 'following'>('all');

  const fetchActivity = async () => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
      const response = await apiClient.get(`/user/activity`) as unknown as UserActivityResponse;
      return response?.activities || [];
    };

    const result = await handleError(operation);
    if (result) {
      setActivities(result);
    }
  };

  useEffect(() => {
    fetchActivity();
  }, []);

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'push':
        return (
          <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        );
      case 'pull_request':
        return (
          <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
        );
      case 'issue':
        return (
          <svg className="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.732L13.732 4.268c-.77-1.064-2.694-1.064-3.464 0L3.34 16.268C2.57 17.333 3.532 19 5.072 19z" />
          </svg>
        );
      case 'fork':
        return (
          <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
          </svg>
        );
      case 'star':
        return (
          <svg className="w-4 h-4 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69L11.049 2.927z" />
          </svg>
        );
      case 'follow':
        return (
          <svg className="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
          </svg>
        );
      case 'create_repository':
        return (
          <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 7a2 2 0 012-2h10a2 2 0 012 2v2M7 7h10" />
          </svg>
        );
      default:
        return (
          <svg className="w-4 h-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        );
    }
  };

  const formatActivityMessage = (activity: ActivityItem) => {
    const { type, action, actor, repository, payload } = activity;
    
    switch (type) {
      case 'push':
        return (
          <span>
            <strong>{actor.username}</strong> pushed {payload.commits?.length || 1} commit{payload.commits?.length !== 1 ? 's' : ''} to{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}`} className="text-blue-600 hover:text-blue-800">
                {repository.full_name}
              </Link>
            )}
          </span>
        );
      case 'pull_request':
        return (
          <span>
            <strong>{actor.username}</strong> {action} pull request{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}/pulls/${payload.number}`} className="text-blue-600 hover:text-blue-800">
                #{payload.number}
              </Link>
            )}{' '}
            in{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}`} className="text-blue-600 hover:text-blue-800">
                {repository.full_name}
              </Link>
            )}
          </span>
        );
      case 'issue':
        return (
          <span>
            <strong>{actor.username}</strong> {action} issue{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}/issues/${payload.number}`} className="text-blue-600 hover:text-blue-800">
                #{payload.number}
              </Link>
            )}{' '}
            in{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}`} className="text-blue-600 hover:text-blue-800">
                {repository.full_name}
              </Link>
            )}
          </span>
        );
      case 'fork':
        return (
          <span>
            <strong>{actor.username}</strong> forked{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}`} className="text-blue-600 hover:text-blue-800">
                {repository.full_name}
              </Link>
            )}
          </span>
        );
      case 'star':
        return (
          <span>
            <strong>{actor.username}</strong> starred{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}`} className="text-blue-600 hover:text-blue-800">
                {repository.full_name}
              </Link>
            )}
          </span>
        );
      case 'follow':
        return (
          <span>
            <strong>{actor.username}</strong> followed{' '}
            <Link href={`/users/${payload.target?.username}`} className="text-blue-600 hover:text-blue-800">
              {payload.target?.username}
            </Link>
          </span>
        );
      case 'create_repository':
        return (
          <span>
            <strong>{actor.username}</strong> created repository{' '}
            {repository && (
              <Link href={`/repositories/${repository.full_name}`} className="text-blue-600 hover:text-blue-800">
                {repository.full_name}
              </Link>
            )}
          </span>
        );
      default:
        return (
          <span>
            <strong>{actor.username}</strong> {action}
          </span>
        );
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-8"></div>
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-20 bg-muted rounded"></div>
              ))}
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error}</div>
            <Button onClick={() => fetchActivity()} disabled={loading}>
              {loading ? 'Retrying...' : 'Try Again'}
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Activity Feed</h1>
          <p className="text-muted-foreground mt-2">Stay up to date with what&apos;s happening across your repositories and network</p>
        </div>

        {/* Filter Tabs */}
        <div className="border-b border-border mb-8">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setFilter('all')}
              data-testid="filter-all"
              className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                filter === 'all'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              All activity
            </button>
            
            <button
              onClick={() => setFilter('own')}
              data-testid="filter-own"
              className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                filter === 'own'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              Your activity
            </button>
            
            <button
              onClick={() => setFilter('following')}
              data-testid="filter-following"
              className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                filter === 'following'
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
              }`}
            >
              <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
              </svg>
              Following
            </button>
          </nav>
        </div>

        {/* Activity List */}
        {activities.length > 0 ? (
          <div className="space-y-4">
            {activities.map((activity) => (
              <Card key={activity.id} data-testid="activity-item">
                <div className="p-6">
                  <div className="flex items-start space-x-4">
                    <Avatar
                      src={activity.actor.avatar_url}
                      alt={activity.actor.username}
                      size="sm"
                    />
                    
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        {getActivityIcon(activity.type)}
                        <div className="text-sm text-foreground">
                          {formatActivityMessage(activity)}
                        </div>
                      </div>
                      
                      <div className="text-xs text-muted-foreground">
                        {new Date(activity.created_at).toLocaleString()}
                      </div>
                      
                      {/* Additional payload information */}
                      {activity.payload.title && (
                        <div className="mt-2 text-sm text-foreground">
                          &quot;{activity.payload.title}&quot;
                        </div>
                      )}
                      
                      {activity.type === 'push' && activity.payload.commits && (
                        <div className="mt-2">
                          <div className="text-xs text-muted-foreground mb-1">
                            {activity.payload.commits.length} commit{activity.payload.commits.length !== 1 ? 's' : ''}
                          </div>
                          <div className="space-y-1">
                            {activity.payload.commits.slice(0, 3).map((commit: { sha?: string; message: string }, index: number) => (
                              <div key={index} className="text-xs bg-muted p-2 rounded">
                                <span className="font-mono text-blue-600">{commit.sha?.substring(0, 7)}</span>
                                <span className="ml-2">{commit.message}</span>
                              </div>
                            ))}
                            {activity.payload.commits.length > 3 && (
                              <div className="text-xs text-muted-foreground">
                                ...and {activity.payload.commits.length - 3} more commits
                              </div>
                            )}
                          </div>
                        </div>
                      )}
                    </div>
                    
                    <div className="text-xs text-muted-foreground">
                      {new Date(activity.created_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <Card>
            <div className="p-12 text-center">
              <svg className="w-16 h-16 mx-auto mb-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <h3 className="text-lg font-medium text-foreground mb-2">No activity yet</h3>
              <p className="text-muted-foreground mb-4">
                {filter === 'all' && "There&apos;s no activity to show yet."}
                {filter === 'own' && "You haven&apos;t performed any activities yet."}
                {filter === 'following' && "Follow users to see their activity here."}
              </p>
              {filter === 'following' && (
                <Link href="/search">
                  <Button>Find users to follow</Button>
                </Link>
              )}
            </div>
          </Card>
        )}
      </div>
    </AppLayout>
  );
}