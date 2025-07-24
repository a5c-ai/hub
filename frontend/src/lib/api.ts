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
  login: (email: string, password: string) =>
    apiClient.post('/auth/login', { email, password }),

  register: (userData: { username: string; email: string; password: string; full_name: string }) =>
    apiClient.post('/auth/register', userData),

  logout: () => apiClient.post('/auth/logout'),

  refreshToken: () => apiClient.post('/auth/refresh'),

  getProfile: () => apiClient.get('/profile'),

  forgotPassword: (email: string) =>
    apiClient.post('/auth/forgot-password', { email }),

  resetPassword: (token: string, password: string) =>
    apiClient.post('/auth/reset-password', { token, password }),

  verifyEmail: (token: string) =>
    apiClient.post(`/auth/verify-email?token=${token}`),
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

export default api;