'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import { Card } from '@/components/ui/Card';
import { PullRequestDetail } from '@/components/pullRequests/PullRequestDetail';
import api from '@/lib/api';
import { PullRequest } from '@/types';

export default function PullRequestDetailPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const number = params.number as string;
  const [pullRequest, setPullRequest] = useState<PullRequest | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPullRequest = async () => {
      try {
        setLoading(true);
        const response = await api.get(`/repositories/${owner}/${repo}/pulls/${number}`);
        setPullRequest(response.data);
      } catch (err: unknown) {
        setError(
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          'Failed to fetch pull request'
        );
      } finally {
        setLoading(false);
      }
    };

    fetchPullRequest();
  }, [owner, repo, number]);

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-4 bg-gray-200 rounded w-1/4 mb-6"></div>
            <div className="h-8 bg-gray-200 rounded w-3/4 mb-2"></div>
            <div className="h-4 bg-gray-200 rounded w-1/2 mb-8"></div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2">
                <div className="h-96 bg-gray-200 rounded"></div>
              </div>
              <div>
                <div className="h-64 bg-gray-200 rounded"></div>
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
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (!pullRequest) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-gray-500 text-lg">Pull request not found</div>
          </div>
        </div>
      </AppLayout>
    );
  }

  const getStateColor = (state: string, merged: boolean) => {
    if (merged) return 'bg-purple-100 text-purple-800';
    if (state === 'open') return 'bg-green-100 text-green-800';
    return 'bg-red-100 text-red-800';
  };

  const getStateIcon = (state: string, merged: boolean) => {
    if (merged) {
      return (
        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
          <path d="M21 7L9 19l-5.5-5.5 1.41-1.41L9 16.17 19.59 5.59 21 7z" />
        </svg>
      );
    }
    if (state === 'open') {
      return (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
        </svg>
      );
    }
    return (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    );
  };

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-gray-500 mb-6">
          <Link href="/repositories" className="hover:text-gray-700 transition-colors">
            Repositories
          </Link>
          <span>/</span>
          <Link 
            href={`/repositories/${owner}/${repo}`}
            className="hover:text-gray-700 transition-colors"
          >
            {owner}/{repo}
          </Link>
          <span>/</span>
          <Link 
            href={`/repositories/${owner}/${repo}/pulls`}
            className="hover:text-gray-700 transition-colors"
          >
            Pull requests
          </Link>
          <span>/</span>
          <span className="text-gray-900 font-medium">#{number}</span>
        </nav>

        {/* Header */}
        <div className="mb-8">
          <div className="flex items-start space-x-3">
            <div className={`flex items-center px-2 py-1 text-xs font-medium rounded-full ${getStateColor(pullRequest.issue.state, pullRequest.merged)}`}>
              {getStateIcon(pullRequest.issue.state, pullRequest.merged)}
              <span className="ml-1">
                {pullRequest.merged ? 'Merged' : pullRequest.issue.state === 'open' ? 'Open' : 'Closed'}
              </span>
            </div>
            <div className="flex-1">
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                {pullRequest.issue.title}
                <span className="text-gray-500 font-normal ml-2">#{pullRequest.issue.number}</span>
              </h1>
              <div className="flex items-center text-sm text-gray-600 space-x-4">
                <div className="flex items-center">
                  <Avatar
                    src={pullRequest.issue.user?.avatar_url}
                    alt={pullRequest.issue.user?.username || 'User'}
                    size="sm"
                    className="mr-2"
                  />
                  <span>
                    <strong>{pullRequest.issue.user?.username}</strong> wants to merge{' '}
                    <Badge variant="outline" className="mx-1">
                      {pullRequest.changed_files} {pullRequest.changed_files === 1 ? 'file' : 'files'}
                    </Badge>
                    into{' '}
                    <Badge variant="outline" className="mx-1">
                      {pullRequest.base_ref}
                    </Badge>
                    from{' '}
                    <Badge variant="outline" className="mx-1">
                      {pullRequest.head_ref}
                    </Badge>
                  </span>
                </div>
                <span>•</span>
                <span>{new Date(pullRequest.created_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            <PullRequestDetail pullRequest={pullRequest} />
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Actions */}
            <Card>
              <div className="p-4">
                <div className="space-y-3">
                  {pullRequest.issue.state === 'open' && !pullRequest.merged && (
                    <>
                      <Button className="w-full">
                        <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                        </svg>
                        Merge pull request
                      </Button>
                      <div className="flex space-x-2">
                        <Button variant="outline" size="sm" className="flex-1">
                          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                          </svg>
                          Edit
                        </Button>
                        <Button variant="outline" size="sm" className="flex-1">
                          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                          </svg>
                          Close
                        </Button>
                      </div>
                    </>
                  )}
                  
                  {pullRequest.merged && (
                    <Button disabled className="w-full">
                      <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M21 7L9 19l-5.5-5.5 1.41-1.41L9 16.17 19.59 5.59 21 7z" />
                      </svg>
                      Merged
                    </Button>
                  )}
                </div>
              </div>
            </Card>

            {/* Reviewers */}
            <Card>
              <div className="p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-3">Reviewers</h3>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">No reviewers yet</span>
                    <Button size="sm" variant="outline">
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                      </svg>
                      Add
                    </Button>
                  </div>
                </div>
              </div>
            </Card>

            {/* Assignees */}
            <Card>
              <div className="p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-3">Assignees</h3>
                <div className="space-y-2">
                  {pullRequest.issue.assignees && pullRequest.issue.assignees.length > 0 ? (
                    pullRequest.issue.assignees.map((assignee) => (
                      <div key={assignee.id} className="flex items-center">
                        <Avatar
                          src={assignee.avatar_url}
                          alt={assignee.username}
                          size="sm"
                          className="mr-2"
                        />
                        <span className="text-sm text-gray-900">{assignee.username}</span>
                      </div>
                    ))
                  ) : (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">No one assigned</span>
                      <Button size="sm" variant="outline">
                        <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                        Add
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </Card>

            {/* Labels */}
            <Card>
              <div className="p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-3">Labels</h3>
                <div className="space-y-2">
                  {pullRequest.issue.labels && pullRequest.issue.labels.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                      {pullRequest.issue.labels.map((label) => (
                        <Badge
                          key={label.id}
                          style={{ backgroundColor: `#${label.color}` }}
                          className="text-white"
                        >
                          {label.name}
                        </Badge>
                      ))}
                    </div>
                  ) : (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">None yet</span>
                      <Button size="sm" variant="outline">
                        <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                        Add
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </Card>

            {/* Milestone */}
            <Card>
              <div className="p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-3">Milestone</h3>
                <div className="space-y-2">
                  {pullRequest.issue.milestone ? (
                    <div className="flex items-center">
                      <svg className="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                      </svg>
                      <span className="text-sm text-gray-900">{pullRequest.issue.milestone.title}</span>
                    </div>
                  ) : (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">No milestone</span>
                      <Button size="sm" variant="outline">
                        <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                        Add
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </Card>

            {/* Changes Summary */}
            <Card>
              <div className="p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-3">Changes</h3>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Files changed</span>
                    <span className="font-medium">{pullRequest.changed_files}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-green-600">Additions</span>
                    <span className="font-medium text-green-600">+{pullRequest.additions}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-red-600">Deletions</span>
                    <span className="font-medium text-red-600">-{pullRequest.deletions}</span>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}