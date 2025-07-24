'use client';

import { useState } from 'react';
import Link from 'next/link';
import {
  PlusIcon,
  MagnifyingGlassIcon,
  FolderIcon,
  StarIcon,
  LockClosedIcon,
  CodeBracketIcon as GitBranchIcon,
} from '@heroicons/react/24/outline';
import {
  Card,
  CardContent,
  Button,
  Input,
  Badge,
  Dropdown,
} from '@/components/ui';
import { AppLayout } from '@/components/layout/AppLayout';
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
    owner: { username: 'user', name: 'User Name' },
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
    owner: { username: 'user', name: 'User Name' },
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
    owner: { username: 'user', name: 'User Name' },
  },
  {
    id: '4',
    name: 'data-analysis',
    full_name: 'user/data-analysis',
    description: 'Python scripts for data analysis and visualization',
    private: false,
    language: 'Python',
    stargazers_count: 67,
    forks_count: 12,
    updated_at: '2024-07-17T14:20:00Z',
    owner: { username: 'user', name: 'User Name' },
  },
  {
    id: '5',
    name: 'web-scraper',
    full_name: 'user/web-scraper',
    description: 'A robust web scraping tool with multiple backends',
    private: true,
    language: 'JavaScript',
    stargazers_count: 23,
    forks_count: 5,
    updated_at: '2024-07-16T11:45:00Z',
    owner: { username: 'user', name: 'User Name' },
  },
];

const filterMenuItems = [
  { label: 'All', onClick: () => {} },
  { label: 'Public', onClick: () => {} },
  { label: 'Private', onClick: () => {} },
  { label: 'Forks', onClick: () => {} },
  { label: 'Archived', onClick: () => {} },
];

const sortMenuItems = [
  { label: 'Recently updated', onClick: () => {} },
  { label: 'Name', onClick: () => {} },
  { label: 'Stars', onClick: () => {} },
];

export default function RepositoriesPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [filteredRepos, setFilteredRepos] = useState(mockRepositories);

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    if (query.trim() === '') {
      setFilteredRepos(mockRepositories);
    } else {
      const filtered = mockRepositories.filter(
        (repo) =>
          repo.name.toLowerCase().includes(query.toLowerCase()) ||
          repo.description?.toLowerCase().includes(query.toLowerCase())
      );
      setFilteredRepos(filtered);
    }
  };

  return (
    <AppLayout>
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Repositories</h1>
            <p className="text-muted-foreground mt-1">
              Manage and explore your repositories
            </p>
          </div>
          <Button asChild>
            <Link href="/repositories/new">
              <PlusIcon className="h-4 w-4 mr-2" />
              New repository
            </Link>
          </Button>
        </div>

        {/* Search and filters */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="flex-1 relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Find a repository..."
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex space-x-2">
            <Dropdown
              trigger={<Button variant="outline">Type</Button>}
              items={filterMenuItems}
            />
            <Dropdown
              trigger={<Button variant="outline">Sort</Button>}
              items={sortMenuItems}
            />
          </div>
        </div>

        {/* Repository list */}
        <div className="space-y-4">
          {filteredRepos.length === 0 ? (
            <Card>
              <CardContent className="p-8 text-center">
                <FolderIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-medium text-foreground mb-2">
                  No repositories found
                </h3>
                <p className="text-muted-foreground mb-4">
                  {searchQuery
                    ? `No repositories match "${searchQuery}"`
                    : "You don't have any repositories yet"}
                </p>
                {!searchQuery && (
                  <Button asChild>
                    <Link href="/repositories/new">
                      <PlusIcon className="h-4 w-4 mr-2" />
                      Create your first repository
                    </Link>
                  </Button>
                )}
              </CardContent>
            </Card>
          ) : (
            filteredRepos.map((repo) => (
              <Card key={repo.id} className="hover:shadow-md transition-shadow">
                <CardContent className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <Link
                          href={`/repositories/${repo.full_name}`}
                          className="text-xl font-semibold text-primary hover:underline"
                        >
                          {repo.name}
                        </Link>
                        {repo.private && (
                          <Badge variant="outline" size="sm" className="flex items-center">
                            <LockClosedIcon className="h-3 w-3 mr-1" />
                            Private
                          </Badge>
                        )}
                      </div>
                      
                      {repo.description && (
                        <p className="text-muted-foreground mb-4">
                          {repo.description}
                        </p>
                      )}

                      <div className="flex items-center space-x-6 text-sm text-muted-foreground">
                        {repo.language && (
                          <span className="flex items-center">
                            <span className="w-3 h-3 rounded-full bg-primary mr-2"></span>
                            {repo.language}
                          </span>
                        )}
                        <span className="flex items-center">
                          <StarIcon className="h-4 w-4 mr-1" />
                          {repo.stargazers_count}
                        </span>
                        <span className="flex items-center">
                          <GitBranchIcon className="h-4 w-4 mr-1" />
                          {repo.forks_count}
                        </span>
                        <span>
                          Updated {formatRelativeTime(repo.updated_at)}
                        </span>
                      </div>
                    </div>

                    <div className="flex space-x-2">
                      <Button variant="outline" size="sm">
                        <StarIcon className="h-4 w-4" />
                      </Button>
                      <Button variant="outline" size="sm" asChild>
                        <Link href={`/repositories/${repo.full_name}`}>
                          View
                        </Link>
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>

        {/* Pagination would go here */}
        {filteredRepos.length > 0 && (
          <div className="flex justify-center mt-8">
            <div className="flex space-x-2">
              <Button variant="outline" size="sm" disabled>
                Previous
              </Button>
              <Button variant="outline" size="sm" className="bg-primary text-primary-foreground">
                1
              </Button>
              <Button variant="outline" size="sm">
                2
              </Button>
              <Button variant="outline" size="sm">
                3
              </Button>
              <Button variant="outline" size="sm">
                Next
              </Button>
            </div>
          </div>
        )}
      </div>
    </AppLayout>
  );
}