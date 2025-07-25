'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { IssueList } from '@/components/issues/IssueList';
import { Button } from '@/components/ui/Button';

export default function IssuesPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-2">
                <Link 
                  href="/repositories" 
                  className="hover:text-foreground transition-colors"
                >
                  Repositories
                </Link>
                <span>/</span>
                <Link 
                  href={`/repositories/${owner}/${repo}`}
                  className="hover:text-foreground transition-colors"
                >
                  {owner}/{repo}
                </Link>
                <span>/</span>
                <span className="text-foreground font-medium">Issues</span>
              </nav>
              <h1 className="text-2xl font-bold text-foreground">Issues</h1>
            </div>
            <div className="flex items-center space-x-3">
              <Link href={`/repositories/${owner}/${repo}/issues/labels`}>
                <Button variant="outline" size="sm">
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                  </svg>
                  Labels
                </Button>
              </Link>
              <Link href={`/repositories/${owner}/${repo}/issues/milestones`}>
                <Button variant="outline" size="sm">
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                  Milestones
                </Button>
              </Link>
              <Link href={`/repositories/${owner}/${repo}/issues/new`}>
                <Button>
                  New Issue
                </Button>
              </Link>
            </div>
          </div>
        </div>

        {/* Issue list */}
        <IssueList repositoryOwner={owner} repositoryName={repo} />
      </div>
    </AppLayout>
  );
}