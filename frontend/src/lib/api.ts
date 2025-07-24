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

export default api;