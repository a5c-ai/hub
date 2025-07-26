import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { ApiResponse, PaginatedResponse } from '@/types';

// Create axios instance with default configuration
const api: AxiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    // Get token from localStorage or auth store
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      // Clear auth token and redirect to login
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Generic API methods
export const apiClient = {
  get: async <T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    const response = await api.get(url, config);
    return response.data;
  },

  post: async <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    const response = await api.post(url, data, config);
    return response.data;
  },

  put: async <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    const response = await api.put(url, data, config);
    return response.data;
  },

  patch: async <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    const response = await api.patch(url, data, config);
    return response.data;
  },

  delete: async <T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
    const response = await api.delete(url, config);
    return response.data;
  },

  // Paginated request helper
  getPaginated: async <T>(
    url: string, 
    params?: { page?: number; per_page?: number; [key: string]: unknown }
  ): Promise<PaginatedResponse<T>> => {
    const response = await api.get(url, { params });
    console.log('API Client: getPaginated response for', url, ':', response.data);
    return response.data;
  },
};

// Auth API methods
export const authApi = {
  // Basic authentication
  login: (email: string, password: string, mfaCode?: string) =>
    apiClient.post('/auth/login', { email, password, mfa_code: mfaCode }),

  register: (userData: { username: string; email: string; password: string; full_name: string }) =>
    apiClient.post('/auth/register', userData),

  logout: () => apiClient.post('/auth/logout'),

  refreshToken: (refreshToken: string) => 
    apiClient.post('/auth/refresh', { refresh_token: refreshToken }),

  getProfile: () => apiClient.get('/profile'),

  // Password reset
  forgotPassword: (email: string) =>
    apiClient.post('/auth/forgot-password', { email }),

  resetPassword: (token: string, password: string) =>
    apiClient.post('/auth/reset-password', { token, password }),

  // Email verification
  verifyEmail: (token: string) =>
    apiClient.get(`/auth/verify-email?token=${token}`),

  // OAuth
  getOAuthURL: (provider: string, state?: string) =>
    `/auth/oauth/${provider}?state=${state || ''}`,

  // OAuth callback handling
  handleOAuthCallback: (provider: string, code: string, state?: string) =>
    apiClient.get(`/auth/oauth/${provider}/callback?code=${code}&state=${state || ''}`),
};

// User API methods
export const userApi = {
  getUsers: (params?: { page?: number; per_page?: number; search?: string }) =>
    apiClient.getPaginated('/users', params),

  getUser: (username: string) => apiClient.get(`/users/${username}`),

  updateProfile: (data: Partial<{ name: string; bio: string; location: string; website: string }>) =>
    apiClient.patch('/user', data),
};

// Repository API methods
export const repoApi = {
  getRepositories: (params?: { page?: number; per_page?: number; type?: string; sort?: string }) =>
    apiClient.getPaginated('/repositories', params),

  getRepository: (owner: string, repo: string) =>
    apiClient.get(`/repositories/${owner}/${repo}`),

  createRepository: (data: {
    name: string;
    description?: string;
    private: boolean;
    auto_init?: boolean;
  }) => {
    // Transform private boolean to visibility string for backend
    const payload = {
      name: data.name,
      description: data.description,
      visibility: data.private ? 'private' : 'public',
      auto_init: data.auto_init,
      // Set sensible defaults for repository features
      has_issues: true,
      has_projects: true,
      has_wiki: true,
      has_downloads: true,
      allow_merge_commit: true,
      allow_squash_merge: true,
      allow_rebase_merge: true,
      delete_branch_on_merge: false,
    };
    return apiClient.post('/repositories', payload);
  },

  updateRepository: (owner: string, repo: string, data: Partial<{
    name: string;
    description: string;
    private: boolean;
    default_branch: string;
  }>) => apiClient.patch(`/repositories/${owner}/${repo}`, data),

  deleteRepository: (owner: string, repo: string) =>
    apiClient.delete(`/repositories/${owner}/${repo}`),

  forkRepository: (owner: string, repo: string, data?: { name?: string; organization?: string }) =>
    apiClient.post(`/repositories/${owner}/${repo}/fork`, data),

  // Repository starring
  starRepository: (owner: string, repo: string) =>
    apiClient.put(`/repositories/${owner}/${repo}/star`),

  unstarRepository: (owner: string, repo: string) =>
    apiClient.delete(`/repositories/${owner}/${repo}/star`),

  checkStarred: (owner: string, repo: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/star`),

  // Repository content methods
  getTree: (owner: string, repo: string, path: string = '', ref?: string) => {
    const params = ref ? { ref } : {};
    return api.get(`/repositories/${owner}/${repo}/contents/${path}`, { params });
  },

  getFile: (owner: string, repo: string, path: string, ref?: string) => {
    const params = ref ? { ref } : {};
    return api.get(`/repositories/${owner}/${repo}/contents/${path}`, { params });
  },

  getRepositoryInfo: (owner: string, repo: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/info`),

  getBranches: (owner: string, repo: string) => 
    api.get(`/repositories/${owner}/${repo}/branches`),

  getCommits: (owner: string, repo: string, ref?: string, path?: string) => {
    const params: Record<string, string> = {};
    if (ref) params.ref = ref;
    if (path) params.path = path;
    return apiClient.get(`/repositories/${owner}/${repo}/commits`, { params });
  },

  getCommit: (owner: string, repo: string, sha: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/commits/${sha}`),

  // Repository statistics methods
  getRepositoryStats: (owner: string, repo: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/stats`),

  getRepositoryStatistics: (owner: string, repo: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/stats`),

  getRepositoryLanguages: (owner: string, repo: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/languages`),

  updateRepositoryStats: (owner: string, repo: string) =>
    apiClient.post(`/repositories/${owner}/${repo}/stats/update`),
};

// Organization API methods
export const orgApi = {
  getOrganizations: (params?: { page?: number; per_page?: number }) =>
    apiClient.getPaginated('/organizations', params),

  getOrganization: (org: string) => apiClient.get(`/organizations/${org}`),

  createOrganization: (data: { login: string; name: string; description?: string }) =>
    apiClient.post('/organizations', data),

  updateOrganization: (org: string, data: Partial<{
    name: string;
    description: string;
    location: string;
    website: string;
  }>) => apiClient.patch(`/organizations/${org}`, data),
};

// Search API methods
export const searchApi = {
  // Global search across all content types
  globalSearch: (query: string, params?: {
    type?: string;
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.type) searchParams.append('type', params.type);
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/search?${searchParams.toString()}`);
  },

  // Search repositories
  searchRepositories: (query: string, params?: {
    user?: string;
    language?: string;
    visibility?: string;
    stars?: string;
    forks?: string;
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.user) searchParams.append('user', params.user);
    if (params?.language) searchParams.append('language', params.language);
    if (params?.visibility) searchParams.append('visibility', params.visibility);
    if (params?.stars) searchParams.append('stars', params.stars);
    if (params?.forks) searchParams.append('forks', params.forks);
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/search/repositories?${searchParams.toString()}`);
  },

  // Search issues and pull requests
  searchIssues: (query: string, params?: {
    state?: string;
    author?: string;
    assignee?: string;
    labels?: string;
    milestone?: string;
    is_pr?: boolean;
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.state) searchParams.append('state', params.state);
    if (params?.author) searchParams.append('author', params.author);
    if (params?.assignee) searchParams.append('assignee', params.assignee);
    if (params?.labels) searchParams.append('labels', params.labels);
    if (params?.milestone) searchParams.append('milestone', params.milestone);
    if (params?.is_pr !== undefined) searchParams.append('is_pr', params.is_pr.toString());
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/search/issues?${searchParams.toString()}`);
  },

  // Search users
  searchUsers: (query: string, params?: {
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/search/users?${searchParams.toString()}`);
  },

  // Search commits
  searchCommits: (query: string, params?: {
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/search/commits?${searchParams.toString()}`);
  },

  // Search code
  searchCode: (query: string, params?: {
    repo?: string;
    language?: string;
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.repo) searchParams.append('repo', params.repo);
    if (params?.language) searchParams.append('language', params.language);
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/search/code?${searchParams.toString()}`);
  },

  // Search within a specific repository
  searchInRepository: (owner: string, repo: string, query: string, params?: {
    type?: string;
    sort?: string;
    order?: string;
    page?: number;
    per_page?: number;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.type) searchParams.append('type', params.type);
    if (params?.sort) searchParams.append('sort', params.sort);
    if (params?.order) searchParams.append('order', params.order);
    if (params?.page) searchParams.append('page', params.page.toString());
    if (params?.per_page) searchParams.append('per_page', params.per_page.toString());
    
    return apiClient.get(`/repositories/${owner}/${repo}/search?${searchParams.toString()}`);
  },
};

// Issue API methods
export const issueApi = {
  // List issues
  getIssues: (owner: string, repo: string, params?: {
    state?: 'open' | 'closed';
    sort?: 'created' | 'updated' | 'comments';
    direction?: 'asc' | 'desc';
    page?: number;
    per_page?: number;
    assignee?: string;
    creator?: string;
    milestone?: string;
    labels?: string;
    since?: string;
  }) => apiClient.getPaginated(`/repositories/${owner}/${repo}/issues`, params),

  // Search issues
  searchIssues: (owner: string, repo: string, query: string, params?: {
    state?: 'open' | 'closed';
    sort?: 'created' | 'updated' | 'comments';
    direction?: 'asc' | 'desc';
    page?: number;
    per_page?: number;
  }) => apiClient.getPaginated(`/repositories/${owner}/${repo}/issues/search`, { q: query, ...params }),

  // Get specific issue
  getIssue: (owner: string, repo: string, number: number) =>
    apiClient.get(`/repositories/${owner}/${repo}/issues/${number}`),

  // Create issue
  createIssue: (owner: string, repo: string, data: {
    title: string;
    body?: string;
    assignee_id?: string;
    milestone_id?: string;
    label_ids?: string[];
  }) => apiClient.post(`/repositories/${owner}/${repo}/issues`, data),

  // Update issue
  updateIssue: (owner: string, repo: string, number: number, data: {
    title?: string;
    body?: string;
    state?: 'open' | 'closed';
    state_reason?: string;
    assignee_id?: string;
    milestone_id?: string;
    label_ids?: string[];
  }) => apiClient.patch(`/repositories/${owner}/${repo}/issues/${number}`, data),

  // Close issue
  closeIssue: (owner: string, repo: string, number: number, reason?: string) =>
    apiClient.post(`/repositories/${owner}/${repo}/issues/${number}/close`, { reason: reason || '' }),

  // Reopen issue
  reopenIssue: (owner: string, repo: string, number: number) =>
    apiClient.post(`/repositories/${owner}/${repo}/issues/${number}/reopen`),

  // Lock issue
  lockIssue: (owner: string, repo: string, number: number, reason?: string) =>
    apiClient.post(`/repositories/${owner}/${repo}/issues/${number}/lock`, { reason: reason || '' }),

  // Unlock issue
  unlockIssue: (owner: string, repo: string, number: number) =>
    apiClient.post(`/repositories/${owner}/${repo}/issues/${number}/unlock`),
};

// Comment API methods
export const commentApi = {
  // Get comments for an issue
  getComments: (owner: string, repo: string, issueNumber: number, params?: {
    page?: number;
    per_page?: number;
  }) => apiClient.getPaginated(`/repositories/${owner}/${repo}/issues/${issueNumber}/comments`, params),

  // Get specific comment
  getComment: (owner: string, repo: string, issueNumber: number, commentId: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/issues/${issueNumber}/comments/${commentId}`),

  // Create comment
  createComment: (owner: string, repo: string, issueNumber: number, body: string) =>
    apiClient.post(`/repositories/${owner}/${repo}/issues/${issueNumber}/comments`, { body }),

  // Update comment
  updateComment: (owner: string, repo: string, issueNumber: number, commentId: string, body: string) =>
    apiClient.patch(`/repositories/${owner}/${repo}/issues/${issueNumber}/comments/${commentId}`, { body }),

  // Delete comment
  deleteComment: (owner: string, repo: string, issueNumber: number, commentId: string) =>
    apiClient.delete(`/repositories/${owner}/${repo}/issues/${issueNumber}/comments/${commentId}`),
};

// Label API methods
export const labelApi = {
  // Get all labels
  getLabels: (owner: string, repo: string, params?: {
    page?: number;
    per_page?: number;
  }) => apiClient.getPaginated(`/repositories/${owner}/${repo}/labels`, params),

  // Get specific label
  getLabel: (owner: string, repo: string, name: string) =>
    apiClient.get(`/repositories/${owner}/${repo}/labels/${name}`),

  // Create label
  createLabel: (owner: string, repo: string, data: {
    name: string;
    color: string;
    description?: string;
  }) => apiClient.post(`/repositories/${owner}/${repo}/labels`, data),

  // Update label
  updateLabel: (owner: string, repo: string, name: string, data: {
    name?: string;
    color?: string;
    description?: string;
  }) => apiClient.patch(`/repositories/${owner}/${repo}/labels/${name}`, data),

  // Delete label
  deleteLabel: (owner: string, repo: string, name: string) =>
    apiClient.delete(`/repositories/${owner}/${repo}/labels/${name}`),
};

// Milestone API methods
export const milestoneApi = {
  // Get all milestones
  getMilestones: (owner: string, repo: string, params?: {
    state?: 'open' | 'closed';
    page?: number;
    per_page?: number;
  }) => apiClient.getPaginated(`/repositories/${owner}/${repo}/milestones`, params),

  // Get specific milestone
  getMilestone: (owner: string, repo: string, number: number) =>
    apiClient.get(`/repositories/${owner}/${repo}/milestones/${number}`),

  // Create milestone
  createMilestone: (owner: string, repo: string, data: {
    title: string;
    description?: string;
    due_on?: string;
  }) => apiClient.post(`/repositories/${owner}/${repo}/milestones`, data),

  // Update milestone
  updateMilestone: (owner: string, repo: string, number: number, data: {
    title?: string;
    description?: string;
    state?: 'open' | 'closed';
    due_on?: string;
  }) => apiClient.patch(`/repositories/${owner}/${repo}/milestones/${number}`, data),

  // Delete milestone
  deleteMilestone: (owner: string, repo: string, number: number) =>
    apiClient.delete(`/repositories/${owner}/${repo}/milestones/${number}`),

  // Close milestone
  closeMilestone: (owner: string, repo: string, number: number) =>
    apiClient.post(`/repositories/${owner}/${repo}/milestones/${number}/close`),

  // Reopen milestone
  reopenMilestone: (owner: string, repo: string, number: number) =>
    apiClient.post(`/repositories/${owner}/${repo}/milestones/${number}/reopen`),

};

// SSH Key API methods
export const sshKeyApi = {
  // Get all SSH keys for current user
  getSSHKeys: () => apiClient.get('/user/keys'),

  // Get specific SSH key
  getSSHKey: (keyId: string) => apiClient.get(`/user/keys/${keyId}`),

  // Create new SSH key
  createSSHKey: (data: { title: string; key_data: string }) =>
    apiClient.post('/user/keys', data),

  // Delete SSH key
  deleteSSHKey: (keyId: string) => apiClient.delete(`/user/keys/${keyId}`),
};

export default api;