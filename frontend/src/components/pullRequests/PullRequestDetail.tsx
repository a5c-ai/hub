'use client'

import React, { useState, useEffect } from 'react'
import { pullRequestApi } from '../../lib/pullRequestApi'
import { PullRequest, Review, ReviewComment, PullRequestFile } from '../../types'
import { Button } from '../ui/Button'
import { Badge } from '../ui/Badge'
import { Card } from '../ui/Card'
import { PullRequestFiles } from './PullRequestFiles'
import { ReviewSection } from './ReviewSection'

interface PullRequestDetailProps {
  owner: string
  repo: string
  number: number
}

export function PullRequestDetail({ owner, repo, number }: PullRequestDetailProps) {
  const [pullRequest, setPullRequest] = useState<PullRequest | null>(null)
  const [reviews, setReviews] = useState<Review[]>([])
  const [reviewComments, setReviewComments] = useState<ReviewComment[]>([])
  const [files, setFiles] = useState<PullRequestFile[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'conversation' | 'files' | 'commits'>('conversation')
  const [merging, setMerging] = useState(false)

  useEffect(() => {
    loadPullRequestData()
  }, [owner, repo, number])

  const loadPullRequestData = async () => {
    try {
      setLoading(true)
      setError(null)

      const [prData, reviewsData, commentsData, filesData] = await Promise.all([
        pullRequestApi.getPullRequest(owner, repo, number),
        pullRequestApi.listReviews(owner, repo, number),
        pullRequestApi.listReviewComments(owner, repo, number),
        pullRequestApi.getPullRequestFiles(owner, repo, number)
      ])

      setPullRequest(prData)
      setReviews(reviewsData)
      setReviewComments(commentsData)
      setFiles(filesData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load pull request')
    } finally {
      setLoading(false)
    }
  }

  const handleMerge = async (mergeMethod: 'merge' | 'squash' | 'rebase' = 'merge') => {
    if (!pullRequest) return

    try {
      setMerging(true)
      await pullRequestApi.mergePullRequest(owner, repo, number, { merge_method: mergeMethod })
      await loadPullRequestData() // Reload to get updated state
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to merge pull request')
    } finally {
      setMerging(false)
    }
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

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  if (error || !pullRequest) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <p className="text-red-800">{error || 'Pull request not found'}</p>
        <Button onClick={loadPullRequestData} className="mt-2">
          Try Again
        </Button>
      </div>
    )
  }

  const canMerge = pullRequest.issue.state === 'open' && !pullRequest.merged && pullRequest.mergeable !== false

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Header */}
      <div className="border-b border-gray-200 pb-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center space-x-3 mb-3">
              <h1 className="text-2xl font-bold text-gray-900">
                {pullRequest.issue.title}
              </h1>
              <span className="text-gray-500">#{pullRequest.issue.number}</span>
            </div>
            
            <div className="flex items-center space-x-4 mb-4">
              <Badge className={getStateColor(pullRequest.issue.state, pullRequest.merged)}>
                {getStateText(pullRequest.issue.state, pullRequest.merged)}
              </Badge>
              {pullRequest.draft && (
                <Badge className="bg-gray-100 text-gray-800">Draft</Badge>
              )}
              <span className="text-gray-600">
                {pullRequest.issue.user?.username || 'Unknown'} wants to merge {pullRequest.changed_files} commits into{' '}
                <code className="bg-gray-100 px-1 rounded text-sm">{pullRequest.base_ref}</code> from{' '}
                <code className="bg-gray-100 px-1 rounded text-sm">{pullRequest.head_ref}</code>
              </span>
            </div>

            <div className="flex items-center space-x-6 text-sm text-gray-600">
              <span className="flex items-center">
                <svg className="w-4 h-4 mr-1 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
                </svg>
                {pullRequest.additions} additions
              </span>
              <span className="flex items-center">
                <svg className="w-4 h-4 mr-1 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clipRule="evenodd" />
                </svg>
                {pullRequest.deletions} deletions
              </span>
              <span>{pullRequest.changed_files} files changed</span>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center space-x-3">
            {canMerge && (
              <div className="flex items-center space-x-2">
                <Button
                  onClick={() => handleMerge('merge')}
                  disabled={merging}
                  className="bg-green-600 hover:bg-green-700"
                >
                  {merging ? 'Merging...' : 'Merge pull request'}
                </Button>
                {/* Merge options dropdown could go here */}
              </div>
            )}
          </div>
        </div>

        {/* Merge status */}
        {pullRequest.mergeable === false && (
          <div className="mt-4 p-4 bg-yellow-50 border border-yellow-200 rounded-md">
            <div className="flex items-center">
              <svg className="w-5 h-5 text-yellow-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
              <p className="text-yellow-800">
                This pull request has conflicts that must be resolved before merging.
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('conversation')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'conversation'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Conversation
            <span className="ml-2 bg-gray-100 text-gray-900 py-0.5 px-2 rounded-full text-xs">
              {pullRequest.issue.comments_count}
            </span>
          </button>
          <button
            onClick={() => setActiveTab('files')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'files'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Files changed
            <span className="ml-2 bg-gray-100 text-gray-900 py-0.5 px-2 rounded-full text-xs">
              {files.length}
            </span>
          </button>
          <button
            onClick={() => setActiveTab('commits')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'commits'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Commits
            <span className="ml-2 bg-gray-100 text-gray-900 py-0.5 px-2 rounded-full text-xs">
              {pullRequest.changed_files}
            </span>
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          {activeTab === 'conversation' && (
            <div className="space-y-6">
              {/* Description */}
              {pullRequest.issue.body && (
                <Card className="p-6">
                  <div className="prose max-w-none">
                    <div className="whitespace-pre-wrap text-gray-700">
                      {pullRequest.issue.body}
                    </div>
                  </div>
                </Card>
              )}

              {/* Reviews and Comments */}
              <ReviewSection
                owner={owner}
                repo={repo}
                number={number}
                reviews={reviews}
                reviewComments={reviewComments}
                onReviewCreated={loadPullRequestData}
              />
            </div>
          )}

          {activeTab === 'files' && (
            <PullRequestFiles
              files={files}
              owner={owner}
              repo={repo}
              number={number}
            />
          )}

          {activeTab === 'commits' && (
            <Card className="p-6">
              <p className="text-gray-500">Commits view not yet implemented</p>
            </Card>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          {/* Review Status */}
          <Card className="p-4">
            <h3 className="font-medium text-gray-900 mb-3">Reviews</h3>
            {reviews.length === 0 ? (
              <p className="text-sm text-gray-500">No reviews yet</p>
            ) : (
              <div className="space-y-2">
                {reviews.map((review) => (
                  <div key={review.id} className="flex items-center space-x-2 text-sm">
                    <div className={`w-3 h-3 rounded-full ${
                      review.state === 'approved' ? 'bg-green-400' :
                      review.state === 'request_changes' ? 'bg-red-400' :
                      'bg-gray-400'
                    }`} />
                    <span className="font-medium">{review.user?.username || 'Unknown'}</span>
                    <span className="text-gray-500">
                      {review.state === 'approved' ? 'approved' :
                       review.state === 'request_changes' ? 'requested changes' :
                       'commented'}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </Card>

          {/* Merge Status */}
          <Card className="p-4">
            <h3 className="font-medium text-gray-900 mb-3">Merge status</h3>
            <div className="space-y-2 text-sm">
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${
                  pullRequest.mergeable === true ? 'bg-green-400' :
                  pullRequest.mergeable === false ? 'bg-red-400' :
                  'bg-yellow-400'
                }`} />
                <span className="text-gray-700">
                  {pullRequest.mergeable === true ? 'Ready to merge' :
                   pullRequest.mergeable === false ? 'Conflicts detected' :
                   'Checking mergeability...'}
                </span>
              </div>
              <p className="text-gray-500 ml-5">
                {pullRequest.mergeable_state}
              </p>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}