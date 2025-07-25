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
        <nav className="flex items-center space-x-2 text-sm text-gray-500 mb-6">
          <Link href="/repositories" className="hover:text-gray-700 transition-colors">
            Repositories
          </Link>
          <span>/</span>
          <Link 
            href={`/repositories/${owner}/${repo}`}
            className="hover:text-gray-700 transition-colors"
          >
            {owner}/{repo}
          </Link>
          <span>/</span>
          <Link 
            href={`/repositories/${owner}/${repo}/pulls`}
            className="hover:text-gray-700 transition-colors"
          >
            Pull requests
          </Link>
          <span>/</span>
          <span className="text-gray-900 font-medium">New</span>
        </nav>

        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Create a pull request</h1>
          <p className="text-gray-600 mt-2">
            Create a new pull request by comparing changes across two branches. 
            If you need to, you can also compare across forks.
          </p>
        </div>

        <CreatePullRequestForm repositoryOwner={owner} repositoryName={repo} />
      </div>
    </AppLayout>
  );
}