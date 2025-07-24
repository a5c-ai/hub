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
  user: User;
  assignees: User[];
  labels: Label[];
  milestone?: Milestone;
  comments: number;
  created_at: string;
  updated_at: string;
  closed_at?: string;
}

export interface Label {
  id: string;
  name: string;
  color: string;
  description?: string;
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
  token: string;
  expires_at: string;
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