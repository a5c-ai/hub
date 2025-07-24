'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { IssueForm } from '@/components/issues/IssueForm';

export default function NewIssuePage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <nav className="flex items-center space-x-2 text-sm text-gray-500 mb-4">
            <Link 
              href="/repositories" 
              className="hover:text-gray-700 transition-colors"
            >
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
              href={`/repositories/${owner}/${repo}/issues`}
              className="hover:text-gray-700 transition-colors"
            >
              Issues
            </Link>
            <span>/</span>
            <span className="text-gray-900 font-medium">New Issue</span>
          </nav>
          <h1 className="text-2xl font-bold text-gray-900">Create New Issue</h1>
        </div>

        {/* Issue form */}
        <div className="bg-white shadow-sm rounded-lg border border-gray-200 p-6">
          <IssueForm
            repositoryOwner={owner}
            repositoryName={repo}
            mode="create"
          />
        </div>
      </div>
    </AppLayout>
  );
}