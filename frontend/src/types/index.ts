export interface User {
  id: string;
  username: string;
  email: string;
  name: string;
  avatar_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Repository {
  id: string;
  name: string;
  full_name: string;
  description?: string;
  private: boolean;
  fork: boolean;
  owner: User;
  default_branch: string;
  clone_url: string;
  ssh_url: string;
  size: number;
  language?: string;
  forks_count: number;
  stargazers_count: number;
  watchers_count: number;
  open_issues_count: number;
  created_at: string;
  updated_at: string;
  pushed_at: string;
}

export interface Organization {
  id: string;
  login: string;
  name: string;
  description?: string;
  avatar_url?: string;
  location?: string;
  website?: string;
  public_repos: number;
  followers: number;
  following: number;
  created_at: string;
  updated_at: string;
}

export interface PullRequest {
  id: string;
  issue_id: string;
  head_repository_id?: string;
  head_ref: string;
  base_repository_id: string;
  base_ref: string;
  merge_commit_sha?: string;
  merged: boolean;
  merged_at?: string;
  merged_by_id?: string;
  draft: boolean;
  mergeable?: boolean;
  mergeable_state: string;
  additions: number;
  deletions: number;
  changed_files: number;
  created_at: string;
  updated_at: string;
  issue: Issue;
  head_repository?: Repository;
  base_repository: Repository;
  merged_by?: User;
}

export interface Issue {
  id: string;
  repository_id: string;
  number: number;
  title: string;
  body: string;
  user_id?: string;
  assignee_id?: string;
  milestone_id?: string;
  state: 'open' | 'closed';
  state_reason: string;
  locked: boolean;
  comments_count: number;
  closed_at?: string;
  created_at: string;
  updated_at: string;
  user?: User;
  assignee?: User;
  labels: Label[];
  pull_request?: PullRequest;
}

export interface Label {
  id: string;
  repository_id: string;
  name: string;
  color: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Milestone {
  id: string;
  title: string;
  description?: string;
  state: 'open' | 'closed';
  due_on?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthUser {
  user: User;
  access_token: string;
  refresh_token?: string;
  expires_in: number;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export interface Review {
  id: string;
  pull_request_id: string;
  user_id?: string;
  commit_sha: string;
  state: 'pending' | 'approved' | 'request_changes' | 'commented' | 'dismissed';
  body: string;
  submitted_at?: string;
  created_at: string;
  updated_at: string;
  user?: User;
  review_comments: ReviewComment[];
}

export interface ReviewComment {
  id: string;
  review_id?: string;
  pull_request_id: string;
  user_id?: string;
  commit_sha: string;
  path: string;
  position?: number;
  original_position?: number;
  line?: number;
  original_line?: number;
  side: 'LEFT' | 'RIGHT';
  start_line?: number;
  start_side: 'LEFT' | 'RIGHT';
  body: string;
  in_reply_to_id?: string;
  created_at: string;
  updated_at: string;
  user?: User;
  review?: Review;
  in_reply_to?: ReviewComment;
  replies: ReviewComment[];
}

export interface PullRequestFile {
  id: string;
  pull_request_id: string;
  filename: string;
  status: 'added' | 'deleted' | 'modified' | 'renamed' | 'copied';
  additions: number;
  deletions: number;
  changes: number;
  patch: string;
  previous_filename?: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePullRequestRequest {
  title: string;
  body: string;
  head: string;
  base: string;
  head_repository_id?: string;
  draft: boolean;
  maintainer_can_modify: boolean;
}

export interface CreateReviewRequest {
  body: string;
  event: 'APPROVE' | 'REQUEST_CHANGES' | 'COMMENT';
  comments: CreateReviewCommentRequest[];
  commit_sha: string;
}

export interface CreateReviewCommentRequest {
  path: string;
  position?: number;
  original_position?: number;
  line?: number;
  original_line?: number;
  side: 'LEFT' | 'RIGHT';
  start_line?: number;
  start_side: 'LEFT' | 'RIGHT';
  body: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
}