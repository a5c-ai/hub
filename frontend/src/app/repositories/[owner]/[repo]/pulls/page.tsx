'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Card } from '@/components/ui/Card';
import { PullRequestList } from '@/components/pullRequests/PullRequestList';

export default function PullRequestsPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const [activeTab, setActiveTab] = useState<'open' | 'closed'>('open');

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
          <span className="text-gray-900 font-medium">Pull requests</span>
        </nav>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Pull requests</h1>
            <p className="text-gray-600 mt-1">Propose changes to the repository</p>
          </div>
          
          <div className="flex items-center space-x-3">
            <Link href={`/repositories/${owner}/${repo}/pulls/new`}>
              <Button>
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                New pull request
              </Button>
            </Link>
          </div>
        </div>

        <Card>
          <div className="border-b border-gray-200">
            {/* Tabs */}
            <div className="flex items-center px-6 py-3">
              <div className="flex space-x-8">
                <button
                  onClick={() => setActiveTab('open')}
                  className={`flex items-center text-sm font-medium transition-colors ${
                    activeTab === 'open'
                      ? 'text-gray-900 border-b-2 border-blue-500 pb-3'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                  </svg>
                  Open
                  <Badge variant="secondary" className="ml-2">0</Badge>
                </button>
                
                <button
                  onClick={() => setActiveTab('closed')}
                  className={`flex items-center text-sm font-medium transition-colors ${
                    activeTab === 'closed'
                      ? 'text-gray-900 border-b-2 border-blue-500 pb-3'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Closed
                  <Badge variant="secondary" className="ml-2">0</Badge>
                </button>
              </div>

              <div className="flex items-center ml-auto space-x-4">
                {/* Filters */}
                <div className="flex items-center space-x-2">
                  <select className="text-sm bg-white border border-gray-300 rounded-md px-3 py-1">
                    <option>Author</option>
                    <option>Anyone</option>
                  </select>
                  
                  <select className="text-sm bg-white border border-gray-300 rounded-md px-3 py-1">
                    <option>Label</option>
                    <option>bug</option>
                    <option>enhancement</option>
                  </select>
                  
                  <select className="text-sm bg-white border border-gray-300 rounded-md px-3 py-1">
                    <option>Projects</option>
                  </select>
                  
                  <select className="text-sm bg-white border border-gray-300 rounded-md px-3 py-1">
                    <option>Milestones</option>
                  </select>
                  
                  <select className="text-sm bg-white border border-gray-300 rounded-md px-3 py-1">
                    <option>Assignee</option>
                  </select>
                  
                  <select className="text-sm bg-white border border-gray-300 rounded-md px-3 py-1">
                    <option>Sort</option>
                    <option>Newest</option>
                    <option>Oldest</option>
                    <option>Recently updated</option>
                    <option>Least recently updated</option>
                  </select>
                </div>
              </div>
            </div>
          </div>

          <div className="p-6">
            {/* Search bar */}
            <div className="mb-6">
              <div className="relative">
                <svg className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                  type="text"
                  placeholder="Search pull requests..."
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* Pull Request List */}
            <PullRequestList 
              repositoryOwner={owner} 
              repositoryName={repo} 
              state={activeTab}
            />
          </div>
        </Card>
      </div>
    </AppLayout>
  );
}