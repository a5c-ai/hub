'use client';

import { useState, useEffect } from 'react';
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
import { useAuthStore } from '@/store/auth';
import { useRepositoryStore } from '@/store/repository';
import { formatRelativeTime } from '@/lib/utils';

export default function RepositoriesPage() {
  const { isAuthenticated } = useAuthStore();
  const { 
    repositories, 
    isLoading, 
    error, 
    totalCount,
    currentPage,
    totalPages,
    fetchRepositories,
    clearError 
  } = useRepositoryStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('All');
  const [sortBy, setSortBy] = useState('Recently updated');

  // Filter and search logic
  const filteredRepos = repositories.filter((repo) => {
    const matchesSearch = searchQuery.trim() === '' || 
      repo.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      repo.description?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      repo.owner?.username?.toLowerCase().includes(searchQuery.toLowerCase()) || false;
    
    const matchesType = typeFilter === 'All' ||
      (typeFilter === 'Public' && !repo.private) ||
      (typeFilter === 'Private' && repo.private) ||
      (typeFilter === 'Forks' && repo.fork);
    
    return matchesSearch && matchesType;
  });

  const filterMenuItems = [
    { label: 'All', onClick: () => setTypeFilter('All') },
    { label: 'Public', onClick: () => setTypeFilter('Public') },
    { label: 'Private', onClick: () => setTypeFilter('Private') },
    { label: 'Forks', onClick: () => setTypeFilter('Forks') },
  ];

  const sortMenuItems = [
    { label: 'Recently updated', onClick: () => setSortBy('Recently updated') },
    { label: 'Name', onClick: () => setSortBy('Name') },
    { label: 'Stars', onClick: () => setSortBy('Stars') },
  ];

  useEffect(() => {
    if (isAuthenticated) {
      const sortParam = sortBy === 'Recently updated' ? 'updated' : 
                        sortBy === 'Name' ? 'name' : 'stars';
      fetchRepositories({ 
        page: currentPage, 
        per_page: 10, 
        sort: sortParam 
      });
    }
  }, [isAuthenticated, currentPage, sortBy, fetchRepositories]);

  useEffect(() => {
    if (error) {
      console.error('Repository error:', error);
      const timer = setTimeout(() => {
        clearError();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [error, clearError]);

  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages) {
      fetchRepositories({ 
        page, 
        per_page: 10, 
        sort: sortBy === 'Recently updated' ? 'updated' : 
                     sortBy === 'Name' ? 'name' : 'stars'
      });
    }
  };

  return (
    <AppLayout>
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Repositories</h1>
            <p className="text-muted-foreground mt-1">
              {isLoading ? 'Loading repositories...' : `${totalCount} repositories`}
            </p>
          </div>
          <Button asChild>
            <Link href="/repositories/new">
              <PlusIcon className="h-4 w-4 mr-2" />
              New repository
            </Link>
          </Button>
        </div>

        {error && (
          <div className="mb-6 rounded-md bg-destructive/10 p-4">
            <div className="text-sm text-destructive">
              Failed to load repositories: {error}
            </div>
          </div>
        )}

        {/* Search and filters */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="flex-1 relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Find a repository..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex space-x-2">
            <Dropdown
              trigger={<Button variant="outline">{typeFilter}</Button>}
              items={filterMenuItems}
            />
            <Dropdown
              trigger={<Button variant="outline">{sortBy}</Button>}
              items={sortMenuItems}
            />
          </div>
        </div>

        {/* Repository list */}
        <div className="space-y-4">
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <Card key={i} className="animate-pulse">
                  <CardContent className="p-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="h-6 bg-muted rounded w-1/3 mb-2"></div>
                        <div className="h-4 bg-muted rounded w-2/3 mb-4"></div>
                        <div className="flex space-x-6">
                          <div className="h-4 bg-muted rounded w-16"></div>
                          <div className="h-4 bg-muted rounded w-12"></div>
                          <div className="h-4 bg-muted rounded w-20"></div>
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <div className="h-8 w-8 bg-muted rounded"></div>
                        <div className="h-8 w-16 bg-muted rounded"></div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : error ? (
            <Card>
              <CardContent className="p-8 text-center">
                <div className="text-destructive mb-4">{error}</div>
                <Button onClick={() => fetchRepositories()}>
                  Retry
                </Button>
              </CardContent>
            </Card>
          ) : filteredRepos.length === 0 ? (
            <Card>
              <CardContent className="p-8 text-center">
                <FolderIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-medium text-foreground mb-2">
                  No repositories found
                </h3>
                <p className="text-muted-foreground mb-4">
                  {searchQuery || typeFilter !== 'All'
                    ? `No repositories match your current filters`
                    : repositories.length === 0
                    ? "You don't have any repositories yet"
                    : "No repositories match your current filters"}
                </p>
                {!searchQuery && typeFilter === 'All' && repositories.length === 0 && (
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
                        {repo.fork && (
                          <Badge variant="secondary" size="sm">
                            Fork
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

        {/* Pagination */}
        {!isLoading && totalPages > 1 && (
          <div className="flex justify-center mt-8">
            <div className="flex space-x-2">
              <Button 
                variant="outline" 
                size="sm" 
                disabled={currentPage <= 1}
                onClick={() => handlePageChange(currentPage - 1)}
              >
                Previous
              </Button>
              
              {/* Page numbers */}
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const pageNum = Math.max(1, Math.min(totalPages - 4, currentPage - 2)) + i;
                return (
                  <Button
                    key={pageNum}
                    variant={pageNum === currentPage ? "default" : "outline"}
                    size="sm"
                    onClick={() => handlePageChange(pageNum)}
                  >
                    {pageNum}
                  </Button>
                );
              })}
              
              <Button 
                variant="outline" 
                size="sm" 
                disabled={currentPage >= totalPages}
                onClick={() => handlePageChange(currentPage + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}

        {/* Pagination info */}
        {!isLoading && totalCount > 0 && (
          <div className="text-center mt-4 text-sm text-muted-foreground">
            Showing {((currentPage - 1) * 10) + 1} to{' '}
            {Math.min(currentPage * 10, totalCount)} of{' '}
            {totalCount} repositories
          </div>
        )}
      </div>
    </AppLayout>
  );
}