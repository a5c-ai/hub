'use client';

import { useState } from 'react';
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
import { formatRelativeTime } from '@/lib/utils';

// Mock data - replace with real API calls
const mockRepositories = [
  {
    id: '1',
    name: 'awesome-project',
    full_name: 'user/awesome-project',
    description: 'An awesome project built with modern technologies',
    private: false,
    language: 'TypeScript',
    stargazers_count: 42,
    forks_count: 8,
    updated_at: '2024-07-20T10:00:00Z',
  },
  {
    id: '2',
    name: 'api-service',
    full_name: 'user/api-service',
    description: 'RESTful API service for the application',
    private: true,
    language: 'Go',
    stargazers_count: 15,
    forks_count: 3,
    updated_at: '2024-07-19T15:30:00Z',
  },
  {
    id: '3',
    name: 'mobile-app',
    full_name: 'user/mobile-app',
    description: 'Cross-platform mobile application',
    private: false,
    language: 'React Native',
    stargazers_count: 128,
    forks_count: 24,
    updated_at: '2024-07-18T09:15:00Z',
  },
];

const mockActivity = [
  {
    id: '1',
    type: 'push',
    repository: 'user/awesome-project',
    message: 'Added new authentication middleware',
    timestamp: '2024-07-20T10:00:00Z',
  },
  {
    id: '2',
    type: 'pull_request',
    repository: 'user/api-service',
    message: 'Opened pull request: Implement user management endpoints',
    timestamp: '2024-07-19T15:30:00Z',
  },
  {
    id: '3',
    type: 'issue',
    repository: 'user/mobile-app',
    message: 'Created issue: Fix login screen layout on tablet',
    timestamp: '2024-07-18T09:15:00Z',
  },
];

export default function DashboardPage() {
  const { user } = useAuthStore();
  const [repositories] = useState(mockRepositories);
  const [activity] = useState(mockActivity);

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

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          {/* Quick stats */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Repositories</CardTitle>
              <FolderIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">12</div>
              <p className="text-xs text-muted-foreground">+2 from last month</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Stars</CardTitle>
              <StarIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">185</div>
              <p className="text-xs text-muted-foreground">+12 from last week</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Pull Requests</CardTitle>
              <GitBranchIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">7</div>
              <p className="text-xs text-muted-foreground">3 waiting for review</p>
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
              {repositories.map((repo) => (
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
                      {repo.description}
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
                        {repo.stargazers_count}
                      </span>
                      <span className="flex items-center">
                        <GitBranchIcon className="h-3 w-3 mr-1" />
                        {repo.forks_count}
                      </span>
                      <span>Updated {formatRelativeTime(repo.updated_at)}</span>
                    </div>
                  </div>
                </div>
              ))}
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
              {activity.map((item) => (
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
            </CardContent>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
}