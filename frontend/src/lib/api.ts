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
  }) => apiClient.post('/repositories', data),

  updateRepository: (owner: string, repo: string, data: Partial<{
    name: string;
    description: string;
    private: boolean;
    default_branch: string;
  }>) => apiClient.patch(`/repositories/${owner}/${repo}`, data),

  deleteRepository: (owner: string, repo: string) =>
    apiClient.delete(`/repositories/${owner}/${repo}`),

  forkRepository: (owner: string, repo: string, organization?: string) =>
    apiClient.post(`/repositories/${owner}/${repo}/forks`, { organization }),
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

export default api;