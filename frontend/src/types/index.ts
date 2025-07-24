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
  number: number;
  title: string;
  body?: string;
  state: 'open' | 'closed' | 'merged';
  draft: boolean;
  user: User;
  assignees: User[];
  reviewers: User[];
  head: {
    ref: string;
    sha: string;
  };
  base: {
    ref: string;
    sha: string;
  };
  mergeable: boolean;
  merged_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Issue {
  id: string;
  number: number;
  title: string;
  body?: string;
  state: 'open' | 'closed';
  state_reason?: string;
  user?: User;
  assignee?: User;
  assignees?: User[];
  labels: Label[];
  milestone?: Milestone;
  comments_count: number;
  locked: boolean;
  repository?: Repository;
  created_at: string;
  updated_at: string;
  closed_at?: string;
}

export interface Comment {
  id: string;
  issue_id: string;
  user?: User;
  body: string;
  created_at: string;
  updated_at: string;
}

export interface Label {
  id: string;
  name: string;
  color: string;
  description?: string;
}

export interface Milestone {
  id: string;
  number: number;
  title: string;
  description?: string;
  state: 'open' | 'closed';
  due_on?: string;
  closed_at?: string;
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

// Issue-specific request/response types
export interface CreateIssueRequest {
  title: string;
  body?: string;
  assignee_id?: string;
  milestone_id?: string;
  label_ids?: string[];
}

export interface UpdateIssueRequest {
  title?: string;
  body?: string;
  state?: 'open' | 'closed';
  state_reason?: string;
  assignee_id?: string;
  milestone_id?: string;
  label_ids?: string[];
}

export interface IssueFilters {
  state?: 'open' | 'closed';
  sort?: 'created' | 'updated' | 'comments';
  direction?: 'asc' | 'desc';
  assignee?: string;
  creator?: string;
  milestone?: string;
  labels?: string;
  since?: string;
  page?: number;
  per_page?: number;
}

export interface CreateCommentRequest {
  body: string;
}

export interface UpdateCommentRequest {
  body: string;
}

export interface CreateLabelRequest {
  name: string;
  color: string;
  description?: string;
}

export interface UpdateLabelRequest {
  name?: string;
  color?: string;
  description?: string;
}

export interface CreateMilestoneRequest {
  title: string;
  description?: string;
  due_on?: string;
}

export interface UpdateMilestoneRequest {
  title?: string;
  description?: string;
  state?: 'open' | 'closed';
  due_on?: string;
}