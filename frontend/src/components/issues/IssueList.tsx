'use client';

import { useEffect } from 'react';
import { IssueCard } from './IssueCard';
import { useIssueStore } from '@/store/issues';

interface IssueListProps {
  repositoryOwner: string;
  repositoryName: string;
}

export function IssueList({ repositoryOwner, repositoryName }: IssueListProps) {
  const {
    issues,
    isLoadingIssues,
    issuesError,
    issuesTotal,
    currentPage,
    totalPages,
    filters,
    fetchIssues,
    setFilters,
  } = useIssueStore();

  useEffect(() => {
    fetchIssues(repositoryOwner, repositoryName);
  }, [repositoryOwner, repositoryName, fetchIssues]);

  const handleStateFilter = (state: 'open' | 'closed') => {
    setFilters({ state, page: 1 });
    fetchIssues(repositoryOwner, repositoryName, { ...filters, state, page: 1 });
  };

  const handlePageChange = (page: number) => {
    setFilters({ page });
    fetchIssues(repositoryOwner, repositoryName, { ...filters, page });
  };

  if (isLoadingIssues) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (issuesError) {
    return (
      <div className="text-center py-12">
        <div className="text-destructive mb-4">Error loading issues</div>
        <div className="text-muted-foreground text-sm">{issuesError}</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Filter tabs */}
              <div className="border-b border-border">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => handleStateFilter('open')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              filters.state === 'open'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground'
            }`}
          >
            Open Issues
            {filters.state === 'open' && (
              <span className="ml-2 bg-muted text-foreground py-0.5 px-2 rounded-full text-xs">
                {issuesTotal}
              </span>
            )}
          </button>
          <button
            onClick={() => handleStateFilter('closed')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              filters.state === 'closed'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground'
            }`}
          >
            Closed Issues
            {filters.state === 'closed' && (
              <span className="ml-2 bg-muted text-foreground py-0.5 px-2 rounded-full text-xs">
                {issuesTotal}
              </span>
            )}
          </button>
        </nav>
      </div>

      {/* Issue list */}
      {issues.length === 0 ? (
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-muted-foreground"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-foreground">No issues</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {filters.state === 'open'
              ? 'There are no open issues in this repository.'
              : 'There are no closed issues in this repository.'}
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {issues.map((issue) => (
            <IssueCard
              key={issue.id}
              issue={issue}
              repositoryOwner={repositoryOwner}
              repositoryName={repositoryName}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-foreground">
            Showing {issues.length} of {issuesTotal} issues
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => handlePageChange(Math.max(1, currentPage - 1))}
              disabled={currentPage === 1}
              className="px-3 py-2 text-sm font-medium text-muted-foreground bg-background border border-input rounded-md hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="px-3 py-2 text-sm font-medium text-foreground">
              Page {currentPage} of {totalPages}
            </span>
            <button
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
              className="px-3 py-2 text-sm font-medium text-muted-foreground bg-background border border-input rounded-md hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}