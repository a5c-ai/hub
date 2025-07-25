import { api } from './api'
import { 
  PullRequest, 
  PullRequestFile, 
  Review, 
  ReviewComment, 
  CreatePullRequestRequest,
  CreateReviewRequest,
  CreateReviewCommentRequest
} from '../types'

export interface PullRequestListResponse {
  pull_requests: PullRequest[]
  total_count: number
  page: number
  per_page: number
}

export interface PullRequestListOptions {
  state?: 'open' | 'closed' | 'all'
  head?: string
  base?: string
  sort?: 'created' | 'updated'
  direction?: 'asc' | 'desc'
  page?: number
  per_page?: number
}

export const pullRequestApi = {
  // Pull Request CRUD operations
  async listPullRequests(owner: string, repo: string, options: PullRequestListOptions = {}): Promise<PullRequestListResponse> {
    const params = new URLSearchParams()
    if (options.state) params.append('state', options.state)
    if (options.head) params.append('head', options.head)
    if (options.base) params.append('base', options.base)
    if (options.sort) params.append('sort', options.sort)
    if (options.direction) params.append('direction', options.direction)
    if (options.page) params.append('page', options.page.toString())
    if (options.per_page) params.append('per_page', options.per_page.toString())
    
    const response = await api.get(`/repositories/${owner}/${repo}/pulls?${params.toString()}`)
    return response.data
  },

  async getPullRequest(owner: string, repo: string, number: number): Promise<PullRequest> {
    const response = await api.get(`/repositories/${owner}/${repo}/pulls/${number}`)
    return response.data
  },

  async createPullRequest(owner: string, repo: string, request: CreatePullRequestRequest): Promise<PullRequest> {
    const response = await api.post(`/repositories/${owner}/${repo}/pulls`, request)
    return response.data
  },

  async updatePullRequest(owner: string, repo: string, number: number, updates: Partial<PullRequest>): Promise<PullRequest> {
    const response = await api.patch(`/repositories/${owner}/${repo}/pulls/${number}`, updates)
    return response.data
  },

  async mergePullRequest(
    owner: string, 
    repo: string, 
    number: number, 
    options: {
      merge_method?: 'merge' | 'squash' | 'rebase'
      commit_title?: string
      commit_message?: string
    } = {}
  ): Promise<{ sha: string; merged: boolean; message: string }> {
    const response = await api.put(`/repositories/${owner}/${repo}/pulls/${number}/merge`, {
      merge_method: options.merge_method || 'merge',
      commit_title: options.commit_title,
      commit_message: options.commit_message
    })
    return response.data
  },

  // Pull Request Files
  async getPullRequestFiles(owner: string, repo: string, number: number): Promise<PullRequestFile[]> {
    const response = await api.get(`/repositories/${owner}/${repo}/pulls/${number}/files`)
    return response.data
  },

  // Reviews
  async listReviews(owner: string, repo: string, number: number): Promise<Review[]> {
    const response = await api.get(`/repositories/${owner}/${repo}/pulls/${number}/reviews`)
    return response.data
  },

  async createReview(owner: string, repo: string, number: number, request: CreateReviewRequest): Promise<Review> {
    const response = await api.post(`/repositories/${owner}/${repo}/pulls/${number}/reviews`, request)
    return response.data
  },

  // Review Comments
  async listReviewComments(owner: string, repo: string, number: number): Promise<ReviewComment[]> {
    const response = await api.get(`/repositories/${owner}/${repo}/pulls/${number}/comments`)
    return response.data
  },

  async createReviewComment(owner: string, repo: string, number: number, request: CreateReviewCommentRequest): Promise<ReviewComment> {
    const response = await api.post(`/repositories/${owner}/${repo}/pulls/${number}/comments`, request)
    return response.data
  }
}