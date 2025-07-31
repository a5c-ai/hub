import { create } from 'zustand';
import { Repository } from '@/types';
import { repoApi } from '@/lib/api';

interface RepositoryState {
  repositories: Repository[];
  currentRepository: Repository | null;
  isLoading: boolean;
  error: string | null;
  totalCount: number;
  currentPage: number;
  totalPages: number;
}

interface RepositoryActions {
  fetchRepositories: (params?: { page?: number; per_page?: number; type?: string; sort?: string }) => Promise<void>;
  fetchRepository: (owner: string, repo: string) => Promise<void>;
  createRepository: (data: {
    name: string;
    description?: string;
    private: boolean;
    auto_init?: boolean;
  }) => Promise<Repository>;
  updateRepository: (owner: string, repo: string, data: Partial<{
    name: string;
    description: string;
    private: boolean;
    default_branch: string;
  }>) => Promise<void>;
  deleteRepository: (owner: string, repo: string) => Promise<void>;
  clearError: () => void;
  resetRepositories: () => void;
}

export const useRepositoryStore = create<RepositoryState & RepositoryActions>((set) => ({
  // State
  repositories: [],
  currentRepository: null,
  isLoading: false,
  error: null,
  totalCount: 0,
  currentPage: 1,
  totalPages: 1,

  // Actions
  fetchRepositories: async (params = {}) => {
    set({ isLoading: true, error: null });
    try {
      const response = await repoApi.getRepositories(params);
      console.log('Repository store: Fetch response received:', response);
      
      // Handle different response formats
      let repositories: Repository[] = [];
      let totalCount = 0;
      let currentPage = 1;
      let totalPages = 1;
      
      if (Array.isArray(response)) {
        // Direct array response
        repositories = response as unknown as Repository[];
        totalCount = response.length;
      } else if (response.data && Array.isArray(response.data)) {
        // Wrapped array response
        repositories = response.data as unknown as Repository[];
        totalCount = response.pagination?.total || response.data.length;
        currentPage = response.pagination?.page || 1;
        totalPages = response.pagination?.total_pages || 1;
      } else if (response && response.data) {
        // Other wrapped response
        repositories = Array.isArray(response.data) ? response.data as unknown as Repository[] : [];
        totalCount = response.pagination?.total || repositories.length;
        currentPage = response.pagination?.page || 1;
        totalPages = response.pagination?.total_pages || 1;
      }
      
      console.log('Repository store: Processed repositories:', { 
        count: repositories.length, 
        totalCount, 
        repositories: repositories.slice(0, 2) // Log first 2 for debugging
      });
      
      set({
        repositories,
        totalCount,
        currentPage,
        totalPages,
        isLoading: false,
      });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'error' in error.response.data
        ? String(error.response.data.error)
        : 'Failed to fetch repositories';
      
      set({
        isLoading: false,
        error: errorMessage,
        repositories: [],
        totalCount: 0,
        currentPage: 1,
        totalPages: 1,
      });
    }
  },

  fetchRepository: async (owner: string, repo: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await repoApi.getRepository(owner, repo);
      
      // Handle response - apiClient returns data directly or wrapped response
      let repository: Repository | null = null;
      
      if (response && typeof response === 'object') {
        // Check if it's a direct repository object
        if ('id' in response && 'name' in response) {
          // Intentional cast: convert via unknown to satisfy TypeScript
          // Intentional cast: convert via unknown to satisfy TypeScript
          repository = response as unknown as Repository;
        }
        // Check if it's wrapped in a success response
        else if ('success' in response && response.success && 'data' in response && response.data) {
          repository = response.data as unknown as Repository;
        }
        // Check if it's wrapped in just data
        else if ('data' in response && response.data) {
          repository = response.data as unknown as Repository;
        }
      }
      
      set({
        currentRepository: repository,
        isLoading: false,
      });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'error' in error.response.data
        ? String(error.response.data.error)
        : 'Failed to fetch repository';
      
      set({
        isLoading: false,
        error: errorMessage,
        currentRepository: null,
      });
    }
  },

  createRepository: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await repoApi.createRepository(data);
      console.log('Repository store: API response received:', response);
      
      // Handle response - apiClient returns data directly or wrapped response
      let newRepo: Repository | null = null;
      
      if (response && typeof response === 'object') {
        // Check if it's a direct repository object
        if ('id' in response && 'name' in response) {
          // Intentional cast: convert via unknown to satisfy TypeScript
          newRepo = response as unknown as Repository;
        }
        // Check if it's wrapped in a success response
        else if ('success' in response && response.success && 'data' in response && response.data && typeof response.data === 'object' && 'id' in response.data && 'name' in response.data) {
          newRepo = response.data as unknown as Repository;
        }
        // Check if it's wrapped in just data
        else if ('data' in response && response.data && typeof response.data === 'object' && 'id' in response.data && 'name' in response.data) {
          newRepo = response.data as unknown as Repository;
        }
      }
      
      if (newRepo) {
        console.log('Repository store: Created repository:', newRepo);
        set((state) => ({
          repositories: [newRepo, ...state.repositories],
          isLoading: false,
        }));
        return newRepo;
      }
      
      console.error('Repository store: Unexpected response structure:', response);
      throw new Error('Invalid response format from server');
    } catch (error: unknown) {
      console.error('Repository store: Create repository error:', error);
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'error' in error.response.data
        ? String(error.response.data.error)
        : error instanceof Error ? error.message : 'Failed to create repository';
      
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  updateRepository: async (owner: string, repo: string, data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await repoApi.updateRepository(owner, repo, data);
      
      // Handle response - apiClient returns data directly or wrapped response
      let updatedRepo: Repository | null = null;
      
      if (response && typeof response === 'object') {
        // Check if it's a direct repository object
        if ('id' in response && 'name' in response) {
          // Intentional cast: convert via unknown to satisfy TypeScript
          updatedRepo = response as unknown as Repository;
        }
        // Check if it's wrapped in a success response
        else if ('success' in response && response.success && 'data' in response && response.data) {
          updatedRepo = response.data as unknown as Repository;
        }
        // Check if it's wrapped in just data
        else if ('data' in response && response.data) {
          updatedRepo = response.data as unknown as Repository;
        }
      }
      
      if (updatedRepo) {
        set((state) => ({
          repositories: state.repositories.map(r => 
            r.full_name === updatedRepo.full_name ? updatedRepo : r
          ),
          currentRepository: state.currentRepository?.full_name === updatedRepo.full_name
            ? updatedRepo 
            : state.currentRepository,
          isLoading: false,
        }));
      } else {
        set({ isLoading: false });
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'error' in error.response.data
        ? String(error.response.data.error)
        : 'Failed to update repository';
      
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  deleteRepository: async (owner: string, repo: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await repoApi.deleteRepository(owner, repo);
      
      // Delete is successful if no error is thrown
      const fullName = `${owner}/${repo}`;
      set((state) => ({
        repositories: state.repositories.filter(r => r.full_name !== fullName),
        currentRepository: state.currentRepository?.full_name === fullName
          ? null 
          : state.currentRepository,
        isLoading: false,
      }));
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'error' in error.response.data
        ? String(error.response.data.error)
        : 'Failed to delete repository';
      
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  clearError: () => {
    set({ error: null });
  },

  resetRepositories: () => {
    set({
      repositories: [],
      currentRepository: null,
      error: null,
      totalCount: 0,
      currentPage: 1,
      totalPages: 1,
    });
  },
}));
