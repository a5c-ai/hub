'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { CreatePullRequestForm } from '@/components/pullRequests/CreatePullRequestForm';

export default function CreatePullRequestPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <Link href="/repositories" className="hover:text-foreground transition-colors">
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
          <Link 
            href={`/repositories/${owner}/${repo}/pulls`}
            className="hover:text-foreground transition-colors"
          >
            Pull requests
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">New</span>
        </nav>
        
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Create a pull request</h1>
            <p className="text-muted-foreground mt-2">
              Create a pull request to propose and collaborate on changes to a repository.
            </p>
          </div>
        </div>

        <CreatePullRequestForm repositoryOwner={owner} repositoryName={repo} />
      </div>
    </AppLayout>
  );
}