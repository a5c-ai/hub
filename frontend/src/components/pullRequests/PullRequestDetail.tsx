'use client'

import React, { useState } from 'react'
import { PullRequest } from '../../types'
import { Badge } from '../ui/Badge'
import { Card } from '../ui/Card'
import { formatDistanceToNow } from 'date-fns'

interface PullRequestDetailProps {
  pullRequest: PullRequest
}

export function PullRequestDetail({ pullRequest }: PullRequestDetailProps) {
  const [activeTab, setActiveTab] = useState<'conversation' | 'files' | 'commits'>('conversation')

  const getStateColor = (state: string, merged: boolean) => {
    if (merged) return 'bg-purple-100 text-purple-800'
    if (state === 'open') return 'bg-green-100 text-green-800'
    return 'bg-red-100 text-red-800'
  }

  const getStateText = (state: string, merged: boolean) => {
    if (merged) return 'Merged'
    return state === 'open' ? 'Open' : 'Closed'
  }

  return (
    <div className="space-y-6">
      {/* Pull Request Header */}
      <Card>
        <div className="p-6">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center space-x-3 mb-3">
                <Badge className={getStateColor(pullRequest.issue.state, pullRequest.merged)}>
                  {getStateText(pullRequest.issue.state, pullRequest.merged)}
                </Badge>
                <span className="text-sm text-muted-foreground">
                  #{pullRequest.issue.number} opened {formatDistanceToNow(new Date(pullRequest.issue.created_at), { addSuffix: true })} by {pullRequest.issue.user?.username || 'Unknown'}
                </span>
              </div>
              
              <h2 className="text-xl font-semibold text-foreground mb-2">
                {pullRequest.issue.title}
              </h2>
              
              <div className="text-sm text-muted-foreground mb-4">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  {pullRequest.merged ? 'Merged' : pullRequest.issue.state === 'open' ? 'Open' : 'Closed'}
                </span>
              </div>

              {pullRequest.issue.body && (
                <div className="prose max-w-none">
                  <p className="text-foreground">{pullRequest.issue.body}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </Card>

      {/* Tabs */}
              <div className="border-b border-border">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('conversation')}
            className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'conversation'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
            }`}
          >
            <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            Conversation
            <Badge variant="secondary" className="ml-2">
              {pullRequest.issue.comments_count}
            </Badge>
          </button>
          
          <button
            onClick={() => setActiveTab('files')}
            className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'files'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
            }`}
          >
            <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            Files changed
            <Badge variant="secondary" className="ml-2">
              {pullRequest.changed_files}
            </Badge>
          </button>
          
          <button
            onClick={() => setActiveTab('commits')}
            className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'commits'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
            }`}
          >
            <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
            </svg>
            Commits
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      <div className="mt-6">
        {activeTab === 'conversation' && (
          <Card>
            <div className="p-6">
              <div className="text-center py-8 text-muted-foreground">
                <svg className="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <p className="text-lg font-medium mb-2">Comments and reviews</p>
                <p>Comments and review discussions would be displayed here</p>
              </div>
            </div>
          </Card>
        )}

        {activeTab === 'files' && (
          <Card>
            <div className="p-6">
              <div className="text-center py-8 text-muted-foreground">
                <svg className="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p className="text-lg font-medium mb-2">File changes</p>
                <p>
                  {pullRequest.additions > 0 && (
                    <span className="text-green-600">+{pullRequest.additions} additions</span>
                  )}
                  {pullRequest.additions > 0 && pullRequest.deletions > 0 && ', '}
                  {pullRequest.deletions > 0 && (
                    <span className="text-red-600">-{pullRequest.deletions} deletions</span>
                  )}
                </p>
                <p className="text-sm mt-2">File diff would be displayed here</p>
              </div>
            </div>
          </Card>
        )}

        {activeTab === 'commits' && (
          <Card>
            <div className="p-6">
              <div className="text-center py-8 text-muted-foreground">
                <svg className="w-12 h-12 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
                </svg>
                <p className="text-lg font-medium mb-2">Commit history</p>
                <p>List of commits in this pull request would be displayed here</p>
              </div>
            </div>
          </Card>
        )}
      </div>
    </div>
  )
}