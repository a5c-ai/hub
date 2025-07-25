'use client'

import React, { useState } from 'react'
import { pullRequestApi } from '../../lib/pullRequestApi'
import { Review, ReviewComment, CreateReviewRequest } from '../../types'
import { Button } from '../ui/Button'
import { Card } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface ReviewSectionProps {
  owner: string
  repo: string
  number: number
  reviews: Review[]
  reviewComments: ReviewComment[]
  onReviewCreated: () => void
}

export function ReviewSection({ 
  owner, 
  repo, 
  number, 
  reviews, 
  reviewComments, 
  onReviewCreated 
}: ReviewSectionProps) {
  const [showReviewForm, setShowReviewForm] = useState(false)
  const [reviewText, setReviewText] = useState('')
  const [reviewEvent, setReviewEvent] = useState<'APPROVE' | 'REQUEST_CHANGES' | 'COMMENT'>('COMMENT')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmitReview = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!reviewText.trim() && reviewEvent === 'COMMENT') {
      setError('Review comment is required')
      return
    }

    try {
      setSubmitting(true)
      setError(null)

      const reviewRequest: CreateReviewRequest = {
        body: reviewText,
        event: reviewEvent,
        comments: [], // Would include inline comments if implemented
        commit_sha: 'mock_commit_sha' // Would get actual SHA from PR
      }

      await pullRequestApi.createReview(owner, repo, number, reviewRequest)
      
      setReviewText('')
      setShowReviewForm(false)
      onReviewCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit review')
    } finally {
      setSubmitting(false)
    }
  }

  const getReviewStateIcon = (state: string) => {
    switch (state) {
      case 'approved':
        return (
          <div className="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          </div>
        )
      case 'request_changes':
        return (
          <div className="w-6 h-6 bg-red-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-red-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </div>
        )
      default:
        return (
          <div className="w-6 h-6 bg-gray-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zM9 9h2v2H9V9z" clipRule="evenodd" />
            </svg>
          </div>
        )
    }
  }

  const getReviewStateText = (state: string) => {
    switch (state) {
      case 'approved':
        return 'approved these changes'
      case 'request_changes':
        return 'requested changes'
      case 'commented':
        return 'reviewed'
      default:
        return 'commented'
    }
  }

  const getReviewStateBadge = (state: string) => {
    switch (state) {
      case 'approved':
        return <Badge className="bg-green-100 text-green-800">Approved</Badge>
      case 'request_changes':
        return <Badge className="bg-red-100 text-red-800">Changes requested</Badge>
      default:
        return <Badge className="bg-gray-100 text-gray-800">Comment</Badge>
    }
  }

  return (
    <div className="space-y-6">
      {/* Reviews */}
      {reviews.map((review) => (
        <Card key={review.id} className="p-6">
          <div className="flex items-start space-x-4">
            {getReviewStateIcon(review.state)}
            
            <div className="flex-1 min-w-0">
              <div className="flex items-center space-x-3 mb-2">
                <span className="font-medium text-gray-900">
                  {review.user?.username || 'Unknown'}
                </span>
                <span className="text-gray-500 text-sm">
                  {getReviewStateText(review.state)}
                </span>
                {getReviewStateBadge(review.state)}
                <span className="text-gray-500 text-sm">
                  {new Date(review.created_at).toLocaleDateString()}
                </span>
              </div>
              
              {review.body && (
                <div className="prose max-w-none text-gray-700 whitespace-pre-wrap">
                  {review.body}
                </div>
              )}

              {/* Review Comments */}
              {review.review_comments.length > 0 && (
                <div className="mt-4 space-y-3">
                  <h4 className="text-sm font-medium text-gray-900">
                    File comments ({review.review_comments.length})
                  </h4>
                  {review.review_comments.map((comment) => (
                    <div key={comment.id} className="bg-gray-50 rounded-md p-3">
                      <div className="flex items-center space-x-2 text-sm text-gray-600 mb-2">
                        <span className="font-medium">{comment.user?.username || 'Unknown'}</span>
                        <span>commented on</span>
                        <code className="bg-gray-200 px-1 rounded text-xs">
                          {comment.path}
                        </code>
                      </div>
                      <div className="text-sm text-gray-700 whitespace-pre-wrap">
                        {comment.body}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </Card>
      ))}

      {/* Standalone Review Comments */}
      {reviewComments.filter(comment => !comment.review_id).map((comment) => (
        <Card key={comment.id} className="p-6">
          <div className="flex items-start space-x-4">
            <div className="w-6 h-6 bg-blue-100 rounded-full flex items-center justify-center">
              <svg className="w-4 h-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zM9 9h2v2H9V9z" clipRule="evenodd" />
              </svg>
            </div>
            
            <div className="flex-1 min-w-0">
              <div className="flex items-center space-x-3 mb-2">
                <span className="font-medium text-gray-900">
                  {comment.user?.username || 'Unknown'}
                </span>
                <span className="text-gray-500 text-sm">
                  commented on
                </span>
                <code className="bg-gray-200 px-1 rounded text-xs">
                  {comment.path}
                </code>
                <span className="text-gray-500 text-sm">
                  {new Date(comment.created_at).toLocaleDateString()}
                </span>
              </div>
              
              <div className="prose max-w-none text-gray-700 whitespace-pre-wrap">
                {comment.body}
              </div>
            </div>
          </div>
        </Card>
      ))}

      {/* Add Review Form */}
      {!showReviewForm ? (
        <div className="text-center">
          <Button onClick={() => setShowReviewForm(true)}>
            Add a review
          </Button>
        </div>
      ) : (
        <Card className="p-6">
          <form onSubmit={handleSubmitReview}>
            <div className="space-y-4">
              <div>
                <label htmlFor="reviewText" className="block text-sm font-medium text-gray-700 mb-2">
                  Review comment
                </label>
                <textarea
                  id="reviewText"
                  value={reviewText}
                  onChange={(e) => setReviewText(e.target.value)}
                  placeholder="Leave a comment"
                  rows={6}
                  disabled={submitting}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-vertical"
                />
              </div>

              {error && (
                <div className="text-sm text-red-600">{error}</div>
              )}

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4">
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="reviewEvent"
                      value="COMMENT"
                      checked={reviewEvent === 'COMMENT'}
                      onChange={(e) => setReviewEvent(e.target.value as 'COMMENT' | 'APPROVE' | 'REQUEST_CHANGES')}
                      className="form-radio text-blue-600"
                    />
                    <span className="ml-2 text-sm text-gray-700">Comment</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="reviewEvent"
                      value="APPROVE"
                      checked={reviewEvent === 'APPROVE'}
                      onChange={(e) => setReviewEvent(e.target.value as 'COMMENT' | 'APPROVE' | 'REQUEST_CHANGES')}
                      className="form-radio text-green-600"
                    />
                    <span className="ml-2 text-sm text-gray-700">Approve</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="radio"
                      name="reviewEvent"
                      value="REQUEST_CHANGES"
                      checked={reviewEvent === 'REQUEST_CHANGES'}
                      onChange={(e) => setReviewEvent(e.target.value as 'COMMENT' | 'APPROVE' | 'REQUEST_CHANGES')}
                      className="form-radio text-red-600"
                    />
                    <span className="ml-2 text-sm text-gray-700">Request changes</span>
                  </label>
                </div>

                <div className="flex items-center space-x-3">
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => {
                      setShowReviewForm(false)
                      setReviewText('')
                      setError(null)
                    }}
                    disabled={submitting}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={submitting || (!reviewText.trim() && reviewEvent === 'COMMENT')}
                    className={
                      reviewEvent === 'APPROVE' ? 'bg-green-600 hover:bg-green-700' :
                      reviewEvent === 'REQUEST_CHANGES' ? 'bg-red-600 hover:bg-red-700' :
                      ''
                    }
                  >
                    {submitting ? 'Submitting...' : 
                     reviewEvent === 'APPROVE' ? 'Approve changes' :
                     reviewEvent === 'REQUEST_CHANGES' ? 'Request changes' :
                     'Comment'}
                  </Button>
                </div>
              </div>
            </div>
          </form>
        </Card>
      )}
    </div>
  )
}