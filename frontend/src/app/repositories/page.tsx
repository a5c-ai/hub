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
    pagination,
    fetchRepositories 
  } = useRepositoryStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('All');
  const [sortBy, setSortBy] = useState('Recently updated');
  const [currentPage, setCurrentPage] = useState(1);

  // Filter and search logic
  const filteredRepos = repositories.filter((repo) => {
    const matchesSearch = searchQuery.trim() === '' || 
      repo.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      repo.description?.toLowerCase().includes(searchQuery.toLowerCase());
    
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

  const handleSearch = (query: string) => {
    setSearchQuery(query);
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
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
            <div className="text-center py-12">
              <div className="text-muted-foreground">Loading repositories...</div>
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
                    : "You don't have any repositories yet"}
                </p>
                {!searchQuery && typeFilter === 'All' && (
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
        {pagination && pagination.total_pages > 1 && (
          <div className="flex justify-center mt-8">
            <div className="flex space-x-2">
              <Button 
                variant="outline" 
                size="sm" 
                disabled={currentPage === 1}
                onClick={() => handlePageChange(currentPage - 1)}
              >
                Previous
              </Button>
              
              {/* Page numbers */}
              {Array.from({ length: Math.min(5, pagination.total_pages) }, (_, i) => {
                const pageNum = i + Math.max(1, currentPage - 2);
                if (pageNum > pagination.total_pages) return null;
                
                return (
                  <Button
                    key={pageNum}
                    variant="outline"
                    size="sm"
                    className={currentPage === pageNum ? "bg-primary text-primary-foreground" : ""}
                    onClick={() => handlePageChange(pageNum)}
                  >
                    {pageNum}
                  </Button>
                );
              })}
              
              <Button 
                variant="outline" 
                size="sm" 
                disabled={currentPage === pagination.total_pages}
                onClick={() => handlePageChange(currentPage + 1)}
              >
                Next
              </Button>
            </div>
          </div>
        )}

        {/* Pagination info */}
        {pagination && (
          <div className="text-center mt-4 text-sm text-muted-foreground">
            Showing {((currentPage - 1) * (pagination.per_page || 10)) + 1} to{' '}
            {Math.min(currentPage * (pagination.per_page || 10), pagination.total)} of{' '}
            {pagination.total} repositories
          </div>
        )}
      </div>
    </AppLayout>
  );
}