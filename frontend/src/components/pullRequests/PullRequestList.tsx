'use client'

import React, { useState, useEffect } from 'react'
import Link from 'next/link'
import { pullRequestApi, PullRequestListOptions } from '../../lib/pullRequestApi'
import { PullRequest } from '../../types'
import { Button } from '../ui/Button'
import { Badge } from '../ui/Badge'

interface PullRequestListProps {
  repositoryOwner: string
  repositoryName: string
  state?: 'open' | 'closed'
}

export function PullRequestList({ repositoryOwner, repositoryName, state = 'open' }: PullRequestListProps) {
  const [pullRequests, setPullRequests] = useState<PullRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filters, setFilters] = useState<PullRequestListOptions>({
    state: state,
    sort: 'created',
    direction: 'desc',
    page: 1,
    per_page: 25
  })
  const [totalCount, setTotalCount] = useState(0)

  useEffect(() => {
    loadPullRequests()
  }, [repositoryOwner, repositoryName, filters])

  const loadPullRequests = async () => {
    try {
      setLoading(true)
      const response = await pullRequestApi.listPullRequests(repositoryOwner, repositoryName, filters)
      setPullRequests(response.pull_requests)
      setTotalCount(response.total_count)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pull requests')
    } finally {
      setLoading(false)
    }
  }

  const handleFilterChange = (newFilters: Partial<PullRequestListOptions>) => {
    setFilters(prev => ({ ...prev, ...newFilters, page: 1 }))
  }

  const getStateColor = (state: string, merged: boolean) => {
    if (merged) return 'bg-purple-100 text-purple-800'
    if (state === 'open') return 'bg-green-100 text-green-800'
    return 'bg-red-100 text-red-800'
  }

  const getStateText = (state: string, merged: boolean) => {
    if (merged) return 'Merged'
    return state === 'open' ? 'Open' : 'Closed'
  }

  if (loading && pullRequests.length === 0) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <p className="text-red-800">{error}</p>
        <Button onClick={loadPullRequests} className="mt-2">
          Try Again
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header and Filters */}
      <div className="flex justify-between items-center">
        <div className="flex items-center space-x-4">
          <h2 className="text-xl font-semibold">Pull Requests</h2>
          <span className="text-gray-500">{totalCount} total</span>
        </div>
        <Link href={`/${repositoryOwner}/${repositoryName}/pulls/new`}>
          <Button>New Pull Request</Button>
        </Link>
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center space-x-4 border-b border-gray-200">
        <button
          onClick={() => handleFilterChange({ state: 'open' })}
          className={`py-2 px-1 border-b-2 font-medium text-sm ${
            filters.state === 'open'
              ? 'border-blue-500 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          }`}
        >
          Open ({pullRequests.filter(pr => pr.issue.state === 'open').length})
        </button>
        <button
          onClick={() => handleFilterChange({ state: 'closed' })}
          className={`py-2 px-1 border-b-2 font-medium text-sm ${
            filters.state === 'closed'
              ? 'border-blue-500 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          }`}
        >
          Closed
        </button>
        <button
          onClick={() => handleFilterChange({ state: 'all' })}
          className={`py-2 px-1 border-b-2 font-medium text-sm ${
            filters.state === 'all'
              ? 'border-blue-500 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
          }`}
        >
          All
        </button>
      </div>

      {/* Pull Request List */}
      {pullRequests.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-gray-500">
            <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            <p className="mt-2 text-sm text-gray-500">No pull requests found</p>
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          {pullRequests.map((pr) => (
            <div key={pr.id} className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors">
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2 mb-2">
                    <Badge className={getStateColor(pr.issue.state, pr.merged)}>
                      {getStateText(pr.issue.state, pr.merged)}
                    </Badge>
                    {pr.draft && (
                      <Badge className="bg-gray-100 text-gray-800">Draft</Badge>
                    )}
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    <Link 
                      href={`/${repositoryOwner}/${repositoryName}/pull/${pr.issue.number}`}
                      className="text-lg font-medium text-blue-600 hover:underline truncate"
                    >
                      {pr.issue.title}
                    </Link>
                    <span className="text-gray-500">#{pr.issue.number}</span>
                  </div>
                  
                  <div className="flex items-center space-x-4 mt-2 text-sm text-gray-600">
                    <span>
                      {pr.head_ref} → {pr.base_ref}
                    </span>
                    <span>
                      by {pr.issue.user?.username || 'Unknown'}
                    </span>
                    <span>
                      {new Date(pr.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  
                  <div className="flex items-center space-x-4 mt-2 text-sm text-gray-500">
                    <span className="flex items-center">
                      <svg className="w-4 h-4 mr-1 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
                      </svg>
                      +{pr.additions}
                    </span>
                    <span className="flex items-center">
                      <svg className="w-4 h-4 mr-1 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clipRule="evenodd" />
                      </svg>
                      -{pr.deletions}
                    </span>
                    <span>
                      {pr.changed_files} {pr.changed_files === 1 ? 'file' : 'files'} changed
                    </span>
                    <span>
                      {pr.issue.comments_count} {pr.issue.comments_count === 1 ? 'comment' : 'comments'}
                    </span>
                  </div>
                </div>
                
                <div className="flex items-center space-x-2 ml-4">
                  {pr.mergeable === false && (
                    <Badge className="bg-yellow-100 text-yellow-800">
                      Conflicts
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalCount > (filters.per_page || 25) && (
        <div className="flex justify-between items-center pt-4">
          <div className="text-sm text-gray-500">
            Showing {Math.min((filters.page || 1) * (filters.per_page || 25), totalCount)} of {totalCount} pull requests
          </div>
          <div className="flex space-x-2">
            <Button
              disabled={filters.page === 1}
              onClick={() => handleFilterChange({ page: (filters.page || 1) - 1 })}
              variant="secondary"
            >
              Previous
            </Button>
            <Button
              disabled={(filters.page || 1) * (filters.per_page || 25) >= totalCount}
              onClick={() => handleFilterChange({ page: (filters.page || 1) + 1 })}
              variant="secondary"
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}