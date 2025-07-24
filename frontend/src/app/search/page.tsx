'use client';

import { useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import { Card, Input, Button } from '@/components/ui';
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { searchApi } from '@/lib/api';

type SearchType = 'all' | 'repositories' | 'issues' | 'users' | 'commits';

interface SearchResult {
  users: User[];
  repositories: Repository[];
  issues: Issue[];
  organizations: Organization[];
  commits: Commit[];
  total_count: number;
}

interface User {
  id: string;
  username: string;
  full_name: string;
  email: string;
  bio: string;
  avatar_url?: string;
  company?: string;
  location?: string;
}

interface Repository {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  owner_type: string;
  visibility: string;
  stars_count: number;
  forks_count: number;
  primary_language?: string;
  created_at: string;
  updated_at: string;
}

interface Issue {
  id: string;
  number: number;
  title: string;
  body: string;
  state: string;
  repository_id: string;
  user_id: string;
  created_at: string;
  updated_at: string;
}

interface Organization {
  id: string;
  name: string;
  display_name: string;
  description: string;
  location?: string;
  website?: string;
  created_at: string;
}

interface Commit {
  id: string;
  sha: string;
  message: string;
  author_name: string;
  author_email: string;
  committer_name: string;
  committer_date: string;
  repository_id: string;
}

export default function SearchPage() {
  const searchParams = useSearchParams();
  const [query, setQuery] = useState(searchParams.get('q') || '');
  const [type, setType] = useState<SearchType>('all');
  const [results, setResults] = useState<SearchResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const searchTypes = [
    { key: 'all', label: 'All' },
    { key: 'repositories', label: 'Repositories' },
    { key: 'issues', label: 'Issues' },
    { key: 'users', label: 'Users' },
    { key: 'commits', label: 'Commits' },
  ];

  const performSearch = async (searchQuery: string, searchType: SearchType) => {
    if (!searchQuery.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const result = await searchApi.globalSearch(searchQuery, {
        type: searchType === 'all' ? undefined : searchType,
        page: 1,
        per_page: 30,
      });

      setResults(result.data as SearchResult);
    } catch (err) {
      setError('Failed to perform search. Please try again.');
      console.error('Search error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const initialQuery = searchParams.get('q');
    if (initialQuery) {
      setQuery(initialQuery);
      performSearch(initialQuery, type);
    }
  }, [searchParams, type]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    performSearch(query, type);
  };

  const handleTypeChange = (newType: SearchType) => {
    setType(newType);
    if (query.trim()) {
      performSearch(query, newType);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        {/* Search Form */}
        <form onSubmit={handleSearch} className="mb-8">
          <div className="relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search repositories, issues, users, and commits..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="pl-10 pr-20 py-3 text-lg"
            />
            <Button
              type="submit"
              className="absolute right-2 top-1/2 transform -translate-y-1/2"
              disabled={loading}
            >
              {loading ? 'Searching...' : 'Search'}
            </Button>
          </div>
        </form>

        {/* Search Type Filters */}
        <div className="flex flex-wrap gap-2 mb-6">
          {searchTypes.map((searchType) => (
            <Button
              key={searchType.key}
              variant={type === searchType.key ? 'default' : 'outline'}
              size="sm"
              onClick={() => handleTypeChange(searchType.key as SearchType)}
            >
              {searchType.label}
              {results && type === 'all' && (
                <span className="ml-2 text-xs bg-muted px-1.5 py-0.5 rounded">
                  {searchType.key === 'repositories' && results.repositories.length}
                  {searchType.key === 'issues' && results.issues.length}
                  {searchType.key === 'users' && results.users.length}
                  {searchType.key === 'commits' && results.commits.length}
                  {searchType.key === 'all' && results.total_count}
                </span>
              )}
            </Button>
          ))}
        </div>

        {/* Error Message */}
        {error && (
          <Card className="p-4 mb-6 border-destructive">
            <p className="text-destructive">{error}</p>
          </Card>
        )}

        {/* Search Results */}
        {results && !loading && (
          <div className="space-y-6">
            {/* Repositories */}
            {(type === 'all' || type === 'repositories') && results.repositories.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold mb-4">
                  Repositories
                  {type === 'all' && ` (${results.repositories.length})`}
                </h2>
                <div className="space-y-3">
                  {results.repositories.map((repo) => (
                    <Card key={repo.id} className="p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h3 className="text-lg font-medium text-primary">
                            <a href={`/repositories/${repo.name}`} className="hover:underline">
                              {repo.name}
                            </a>
                          </h3>
                          {repo.description && (
                            <p className="text-muted-foreground mt-1">{repo.description}</p>
                          )}
                          <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                            {repo.primary_language && (
                              <span>{repo.primary_language}</span>
                            )}
                            <span>⭐ {repo.stars_count}</span>
                            <span>🍴 {repo.forks_count}</span>
                            <span>Updated {new Date(repo.updated_at).toLocaleDateString()}</span>
                          </div>
                        </div>
                        <div className="ml-4">
                          <span className={`px-2 py-1 text-xs rounded ${
                            repo.visibility === 'public' 
                              ? 'bg-green-100 text-green-800' 
                              : 'bg-yellow-100 text-yellow-800'
                          }`}>
                            {repo.visibility}
                          </span>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            {/* Issues */}
            {(type === 'all' || type === 'issues') && results.issues.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold mb-4">
                  Issues
                  {type === 'all' && ` (${results.issues.length})`}
                </h2>
                <div className="space-y-3">
                  {results.issues.map((issue) => (
                    <Card key={issue.id} className="p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-start gap-3">
                        <div className={`w-2 h-2 rounded-full mt-2 ${
                          issue.state === 'open' ? 'bg-green-500' : 'bg-purple-500'
                        }`} />
                        <div className="flex-1">
                          <h3 className="text-lg font-medium text-primary">
                            <a href={`/issues/${issue.id}`} className="hover:underline">
                              {issue.title}
                            </a>
                          </h3>
                          <p className="text-sm text-muted-foreground mt-1">
                            #{issue.number} opened on {new Date(issue.created_at).toLocaleDateString()}
                          </p>
                          {issue.body && (
                            <p className="text-muted-foreground mt-2 line-clamp-2">
                              {issue.body.substring(0, 200)}...
                            </p>
                          )}
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            {/* Users */}
            {(type === 'all' || type === 'users') && results.users.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold mb-4">
                  Users
                  {type === 'all' && ` (${results.users.length})`}
                </h2>
                <div className="space-y-3">
                  {results.users.map((user) => (
                    <Card key={user.id} className="p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-center gap-4">
                        <img
                          src={user.avatar_url || '/api/placeholder/50/50'}
                          alt={user.username}
                          className="w-12 h-12 rounded-full"
                        />
                        <div className="flex-1">
                          <h3 className="text-lg font-medium text-primary">
                            <a href={`/users/${user.username}`} className="hover:underline">
                              {user.full_name || user.username}
                            </a>
                          </h3>
                          <p className="text-sm text-muted-foreground">@{user.username}</p>
                          {user.bio && (
                            <p className="text-muted-foreground mt-1">{user.bio}</p>
                          )}
                          <div className="flex items-center gap-3 mt-2 text-sm text-muted-foreground">
                            {user.company && <span>🏢 {user.company}</span>}
                            {user.location && <span>📍 {user.location}</span>}
                          </div>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            {/* Commits */}
            {(type === 'all' || type === 'commits') && results.commits.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold mb-4">
                  Commits
                  {type === 'all' && ` (${results.commits.length})`}
                </h2>
                <div className="space-y-3">
                  {results.commits.map((commit) => (
                    <Card key={commit.id} className="p-4 hover:shadow-md transition-shadow">
                      <div className="flex items-start gap-3">
                        <div className="font-mono text-sm bg-muted px-2 py-1 rounded">
                          {commit.sha.substring(0, 7)}
                        </div>
                        <div className="flex-1">
                          <h3 className="text-lg font-medium">
                            <a href={`/commits/${commit.sha}`} className="hover:underline">
                              {commit.message.split('\n')[0]}
                            </a>
                          </h3>
                          <p className="text-sm text-muted-foreground mt-1">
                            by {commit.author_name} on {new Date(commit.committer_date).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            {/* No Results */}
            {results.total_count === 0 && (
              <Card className="p-8 text-center">
                <h3 className="text-lg font-medium mb-2">No results found</h3>
                <p className="text-muted-foreground">
                  Try adjusting your search query or search in a different category.
                </p>
              </Card>
            )}
          </div>
        )}

        {/* Loading State */}
        {loading && (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p className="mt-4 text-muted-foreground">Searching...</p>
          </div>
        )}
      </div>
    </div>
  );
}