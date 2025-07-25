'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams, useParams } from 'next/navigation';
import { Card, Input, Button } from '@/components/ui';
import { MagnifyingGlassIcon, ChevronDownIcon, FolderIcon, DocumentTextIcon, CodeBracketIcon, ChatBubbleLeftRightIcon } from '@heroicons/react/24/outline';
import { searchApi } from '@/lib/api';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';

type SearchType = 'all' | 'code' | 'issues' | 'commits';

interface SearchResult {
  code?: CodeResult[];
  issues?: Issue[];
  commits?: Commit[];
  total_count: number;
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

interface CodeResult {
  id: string;
  repository_id: string;
  repository_name: string;
  file_path: string;
  file_name: string;
  language: string;
  content: string;
  line_count: number;
  branch: string;
  highlighted_content?: string;
}

function RepositorySearchContent() {
  const searchParams = useSearchParams();
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const [query, setQuery] = useState(searchParams.get('q') || '');
  const [type, setType] = useState<SearchType>('all');
  const [results, setResults] = useState<SearchResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showFilters, setShowFilters] = useState(false);
  const [filters, setFilters] = useState({
    language: '',
    path: '',
    extension: ''
  });

  const searchTypes = [
    { key: 'all', label: 'All', icon: DocumentTextIcon },
    { key: 'code', label: 'Code', icon: CodeBracketIcon },
    { key: 'issues', label: 'Issues', icon: ChatBubbleLeftRightIcon },
    { key: 'commits', label: 'Commits', icon: FolderIcon },
  ];

  const performSearch = async (searchQuery: string, searchType: SearchType) => {
    if (!searchQuery.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const searchParams: Record<string, string | number | undefined> = {
        type: searchType === 'all' ? undefined : searchType,
        page: 1,
        per_page: 30,
      };

      // Add repository-specific filters
      if (filters.language) searchParams.language = filters.language;
      if (filters.path) searchParams.path = filters.path;
      if (filters.extension) searchParams.extension = filters.extension;

      const result = await searchApi.searchInRepository(owner, repo, searchQuery, searchParams);
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

  const handleFilterChange = (filterName: string, value: string) => {
    const newFilters = { ...filters, [filterName]: value };
    setFilters(newFilters);
    if (query.trim()) {
      performSearch(query, type);
    }
  };

  const getLanguageFromFileName = (fileName: string): string => {
    const ext = fileName.split('.').pop()?.toLowerCase();
    const languageMap: { [key: string]: string } = {
      'js': 'javascript',
      'jsx': 'javascript',
      'ts': 'typescript',
      'tsx': 'typescript',
      'py': 'python',
      'rb': 'ruby',
      'go': 'go',
      'java': 'java',
      'c': 'c',
      'cpp': 'cpp',
      'css': 'css',
      'html': 'html',
      'json': 'json',
      'xml': 'xml',
      'yaml': 'yaml',
      'yml': 'yaml',
      'sh': 'bash',
      'sql': 'sql',
      'php': 'php'
    };
    return languageMap[ext || ''] || 'text';
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center gap-2 mb-2">
            <FolderIcon className="h-5 w-5 text-muted-foreground" />
            <span className="text-muted-foreground">
              <a href={`/repositories/${owner}`} className="hover:underline">{owner}</a>
              {' / '}
              <a href={`/repositories/${owner}/${repo}`} className="hover:underline">{repo}</a>
            </span>
          </div>
          <h1 className="text-2xl font-bold">Search in this repository</h1>
          <p className="text-muted-foreground mt-1">
            Find code, issues, and commits in {owner}/{repo}
          </p>
        </div>

        {/* Search Form */}
        <form onSubmit={handleSearch} className="mb-6">
          <div className="relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
            <Input
              type="search"
              placeholder={`Search in ${owner}/${repo}...`}
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

        {/* Search Controls */}
        <div className="flex flex-wrap gap-3 mb-6">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowFilters(!showFilters)}
          >
            <ChevronDownIcon className={`h-4 w-4 mr-1 transition-transform ${showFilters ? 'rotate-180' : ''}`} />
            Filters
          </Button>
        </div>

        {/* Repository-specific Filters */}
        {showFilters && (
          <Card className="p-4 mb-6">
            <h3 className="font-semibold mb-3">Repository Filters</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Language</label>
                <select
                  value={filters.language}
                  onChange={(e) => handleFilterChange('language', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                >
                  <option value="">Any</option>
                  <option value="javascript">JavaScript</option>
                  <option value="typescript">TypeScript</option>
                  <option value="python">Python</option>
                  <option value="java">Java</option>
                  <option value="go">Go</option>
                  <option value="rust">Rust</option>
                  <option value="c">C</option>
                  <option value="cpp">C++</option>
                  <option value="php">PHP</option>
                  <option value="ruby">Ruby</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Path</label>
                <Input
                  type="text"
                  placeholder="e.g., src/"
                  value={filters.path}
                  onChange={(e) => handleFilterChange('path', e.target.value)}
                  className="text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Extension</label>
                <Input
                  type="text"
                  placeholder="e.g., .js, .py"
                  value={filters.extension}
                  onChange={(e) => handleFilterChange('extension', e.target.value)}
                  className="text-sm"
                />
              </div>
            </div>
          </Card>
        )}

        {/* Search Type Filters */}
        <div className="flex flex-wrap gap-2 mb-6">
          {searchTypes.map((searchType) => {
            const IconComponent = searchType.icon;
            return (
              <Button
                key={searchType.key}
                variant={type === searchType.key ? 'default' : 'outline'}
                size="sm"
                onClick={() => handleTypeChange(searchType.key as SearchType)}
                className="flex items-center gap-2"
              >
                <IconComponent className="h-4 w-4" />
                {searchType.label}
                {results && type === 'all' && (
                  <span className="ml-1 text-xs bg-muted px-1.5 py-0.5 rounded">
                    {searchType.key === 'code' && results.code?.length}
                    {searchType.key === 'issues' && results.issues?.length}
                    {searchType.key === 'commits' && results.commits?.length}
                    {searchType.key === 'all' && results.total_count}
                  </span>
                )}
              </Button>
            );
          })}
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
            {/* Code Results */}
            {(type === 'all' || type === 'code') && results.code && results.code.length > 0 && (
              <div>
                <h2 className="text-xl font-semibold mb-4">
                  Code
                  {type === 'all' && ` (${results.code.length})`}
                </h2>
                <div className="space-y-4">
                  {results.code.map((code) => (
                    <Card key={code.id} className="overflow-hidden hover:shadow-md transition-shadow">
                      <div className="p-4 border-b bg-gray-50">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <h3 className="text-lg font-medium text-primary">
                              <a href={`/repositories/${owner}/${repo}/blob/${code.branch}/${code.file_path}`} className="hover:underline">
                                {code.file_name}
                              </a>
                            </h3>
                            {code.language && (
                              <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded font-mono">
                                {code.language}
                              </span>
                            )}
                            <span className="text-xs bg-gray-200 text-gray-700 px-2 py-1 rounded">
                              {code.line_count} lines
                            </span>
                          </div>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                          {code.file_path}
                        </p>
                      </div>
                      <div className="max-h-96 overflow-auto">
                        <SyntaxHighlighter
                          language={getLanguageFromFileName(code.file_name)}
                          style={oneLight}
                          showLineNumbers={true}
                          customStyle={{
                            margin: 0,
                            padding: '1rem',
                            background: 'white',
                            fontSize: '0.875rem'
                          }}
                          lineNumberStyle={{
                            minWidth: '3em',
                            paddingRight: '1em',
                            color: '#666',
                            borderRight: '1px solid #e5e7eb',
                            marginRight: '1em'
                          }}
                        >
                          {code.content.length > 1000 ? code.content.substring(0, 1000) + '\\n\\n// ... truncated ...' : code.content}
                        </SyntaxHighlighter>
                      </div>
                      <div className="p-3 bg-gray-50 border-t">
                        <div className="flex items-center justify-between text-sm text-muted-foreground">
                          <span>Branch: {code.branch}</span>
                          <a 
                            href={`/repositories/${owner}/${repo}/blob/${code.branch}/${code.file_path}`}
                            className="text-primary hover:underline"
                          >
                            View full file →
                          </a>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              </div>
            )}

            {/* Issues Results */}
            {(type === 'all' || type === 'issues') && results.issues && results.issues.length > 0 && (
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
                            <a href={`/repositories/${owner}/${repo}/issues/${issue.number}`} className="hover:underline">
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

            {/* Commits Results */}
            {(type === 'all' || type === 'commits') && results.commits && results.commits.length > 0 && (
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
                            <a href={`/repositories/${owner}/${repo}/commit/${commit.sha}`} className="hover:underline">
                              {commit.message.split('\\n')[0]}
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

export default function RepositorySearchPage() {
  return (
    <Suspense fallback={
      <div className="text-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p className="mt-4 text-muted-foreground">Loading search...</p>
      </div>
    }>
      <RepositorySearchContent />
    </Suspense>
  );
}