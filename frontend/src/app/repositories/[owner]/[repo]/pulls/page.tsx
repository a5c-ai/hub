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
        <nav className="flex items-center space-x-2 text-sm mb-6">
          <Link 
            href="/repositories" 
            className="text-blue-600 hover:text-blue-800 transition-colors font-medium"
          >
            Repositories
          </Link>
          <span className="text-gray-400">/</span>
          <Link 
            href={`/repositories/${owner}/${repo}`}
            className="text-blue-600 hover:text-blue-800 transition-colors font-medium"
          >
            {owner}/{repo}
          </Link>
          <span className="text-gray-400">/</span>
          <span className="text-gray-900 font-semibold">Pull requests</span>
        </nav>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Pull requests</h1>
            <p className="text-gray-600 mt-2 text-lg">Propose changes to the repository</p>
          </div>
          
          <div className="flex items-center space-x-3">
            <Link href={`/repositories/${owner}/${repo}/pulls/new`}>
              <Button size="sm">
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
            <div className="flex items-center px-6 py-4">
              <div className="flex space-x-8">
                <button
                  onClick={() => setActiveTab('open')}
                  className={`flex items-center text-sm font-medium pb-3 transition-colors ${
                    activeTab === 'open'
                      ? 'text-gray-900 border-b-2 border-blue-500'
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
                  className={`flex items-center text-sm font-medium pb-3 transition-colors ${
                    activeTab === 'closed'
                      ? 'text-gray-900 border-b-2 border-blue-500'
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

              <div className="flex items-center ml-auto space-x-3">
                {/* Filters */}
                <div className="flex items-center space-x-3">
                  <select 
                    className="text-sm bg-gray-50 border border-gray-300 rounded-md px-3 py-2 text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                    title="Filter by author"
                  >
                    <option value="">Author</option>
                    <option value="anyone">Anyone</option>
                  </select>
                  
                  <select 
                    className="text-sm bg-gray-50 border border-gray-300 rounded-md px-3 py-2 text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                    title="Filter by label"
                  >
                    <option value="">Label</option>
                    <option value="bug">bug</option>
                    <option value="enhancement">enhancement</option>
                  </select>
                  
                  <select 
                    className="text-sm bg-gray-50 border border-gray-300 rounded-md px-3 py-2 text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                    title="Filter by projects"
                  >
                    <option value="">Projects</option>
                  </select>
                  
                  <select 
                    className="text-sm bg-gray-50 border border-gray-300 rounded-md px-3 py-2 text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                    title="Filter by milestones"
                  >
                    <option value="">Milestones</option>
                  </select>
                  
                  <select 
                    className="text-sm bg-gray-50 border border-gray-300 rounded-md px-3 py-2 text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                    title="Filter by assignee"
                  >
                    <option value="">Assignee</option>
                  </select>
                  
                  <select 
                    className="text-sm bg-gray-50 border border-gray-300 rounded-md px-3 py-2 text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
                    title="Sort pull requests"
                  >
                    <option value="">Sort</option>
                    <option value="newest">Newest</option>
                    <option value="oldest">Oldest</option>
                    <option value="updated">Recently updated</option>
                    <option value="least-updated">Least recently updated</option>
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
                  className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent text-gray-900 bg-white"
                />
              </div>
            </div>

            {/* Pull Request List */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-semibold text-gray-900">
                  Pull Requests <span className="text-gray-500">(0 total)</span>
                </h3>
              </div>
              
              <PullRequestList 
                repositoryOwner={owner} 
                repositoryName={repo} 
                state={activeTab}
              />
            </div>
          </div>
        </Card>
      </div>
    </AppLayout>
  );
}