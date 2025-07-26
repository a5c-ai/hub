'use client';

import { useState } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';

interface SearchResult {
  message?: string;
  query?: string;
  note?: string;
  job_id?: string;
  workflow_name?: string;
  line_number?: number;
  content?: string;
  timestamp?: string;
}

interface SearchResponse {
  results: SearchResult[];
  query: string;
  total: number;
}

interface BuildLogSearchProps {
  owner: string;
  repo: string;
}

export default function BuildLogSearch({ owner, repo }: BuildLogSearchProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [limit, setLimit] = useState(50);

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      setError('Please enter a search query');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({
        q: searchQuery.trim(),
        limit: limit.toString(),
      });

      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/logs/search?${params}`);
      
      if (!response.ok) {
        throw new Error('Failed to search build logs');
      }

      const data: SearchResponse = await response.json();
      setResults(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const clearSearch = () => {
    setSearchQuery('');
    setResults(null);
    setError(null);
  };

  const highlightText = (text: string, query: string) => {
    if (!query || !text) return text;
    
    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);
    
    return parts.map((part, index) =>
      regex.test(part) ? (
        <mark key={index} className="bg-yellow-200 px-1 rounded">
          {part}
        </mark>
      ) : (
        part
      )
    );
  };

  return (
    <Card>
      <div className="p-4 border-b">
        <h4 className="font-medium mb-4">Build Log Search</h4>
        
        <div className="space-y-4">
          <div className="flex gap-2">
            <div className="flex-1">
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Search build logs... (e.g., error, failed, timeout)"
                className="w-full"
              />
            </div>
            <div className="w-20">
              <Input
                type="number"
                value={limit}
                onChange={(e) => setLimit(parseInt(e.target.value) || 50)}
                min="1"
                max="100"
                placeholder="50"
                className="w-full text-center"
              />
            </div>
            <Button 
              onClick={handleSearch}
              disabled={loading || !searchQuery.trim()}
            >
              {loading ? 'Searching...' : 'Search'}
            </Button>
            {(results || error) && (
              <Button 
                variant="outline"
                onClick={clearSearch}
              >
                Clear
              </Button>
            )}
          </div>
          
          <div className="text-sm text-muted-foreground">
            <p>Search through build logs for specific terms, errors, or patterns.</p>
            <p className="mt-1">
              <strong>Tips:</strong> Try searching for &quot;error&quot;, &quot;failed&quot;, &quot;timeout&quot;, or specific function names.
            </p>
          </div>
        </div>
      </div>

      {error && (
        <div className="p-4 bg-red-50 border-b border-red-200">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {results && (
        <div className="divide-y">
          <div className="p-4 bg-blue-50 border-b">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <Badge variant="outline">
                  {results.total} result{results.total !== 1 ? 's' : ''}
                </Badge>
                <span className="text-sm text-muted-foreground">
                  Query: &quot;{results.query}&quot;
                </span>
              </div>
              {results.total > results.results.length && (
                <span className="text-sm text-muted-foreground">
                  Showing first {results.results.length} results
                </span>
              )}
            </div>
          </div>

          {results.results.length === 0 ? (
            <div className="p-8 text-center">
              <div className="text-muted-foreground mb-4">
                <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p className="text-sm">No results found</p>
              </div>
              <p className="text-sm text-muted-foreground">
                Try different search terms or check if there are any build logs available for this repository.
              </p>
            </div>
          ) : (
            results.results.map((result, index) => (
              <div key={index} className="p-4">
                {result.note ? (
                  // Implementation note result
                  <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
                    <div className="flex items-start gap-3">
                      <svg className="w-5 h-5 text-yellow-600 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <div>
                        <h6 className="font-medium text-yellow-800 mb-1">Implementation Note</h6>
                        <p className="text-sm text-yellow-700 mb-2">{result.note}</p>
                        {result.message && (
                          <p className="text-sm text-yellow-600">{result.message}</p>
                        )}
                      </div>
                    </div>
                  </div>
                ) : (
                  // Actual search result
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {result.workflow_name && (
                          <Badge variant="outline" className="text-xs">
                            {result.workflow_name}
                          </Badge>
                        )}
                        {result.job_id && (
                          <span className="text-xs text-muted-foreground font-mono">
                            Job: {result.job_id}
                          </span>
                        )}
                        {result.line_number && (
                          <span className="text-xs text-muted-foreground">
                            Line {result.line_number}
                          </span>
                        )}
                      </div>
                      {result.timestamp && (
                        <span className="text-xs text-muted-foreground">
                          {new Date(result.timestamp).toLocaleString()}
                        </span>
                      )}
                    </div>
                    
                    {result.content && (
                      <div className="bg-gray-50 rounded-md p-3">
                        <pre className="text-sm font-mono whitespace-pre-wrap break-words">
                          {highlightText(result.content, searchQuery)}
                        </pre>
                      </div>
                    )}
                    
                    {result.message && !result.content && (
                      <p className="text-sm">
                        {highlightText(result.message, searchQuery)}
                      </p>
                    )}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      )}

      {!results && !error && !loading && (
        <div className="p-8 text-center">
          <div className="text-muted-foreground mb-4">
            <svg className="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <p className="text-sm">Search through build logs</p>
          </div>
          <p className="text-sm text-muted-foreground">
            Enter a search term above to find specific content in your build logs.
          </p>
        </div>
      )}
    </Card>
  );
}