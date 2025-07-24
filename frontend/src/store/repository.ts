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

export const useRepositoryStore = create<RepositoryState & RepositoryActions>((set, get) => ({
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
      if (response.data) {
        set({
          repositories: response.data as Repository[],
          totalCount: response.pagination?.total || 0,
          currentPage: response.pagination?.page || 1,
          totalPages: response.pagination?.total_pages || 1,
          isLoading: false,
        });
      }
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
      if (response.success && response.data) {
        set({
          currentRepository: response.data as Repository,
          isLoading: false,
        });
      }
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
      if (response.success && response.data) {
        const newRepo = response.data as Repository;
        set((state) => ({
          repositories: [newRepo, ...state.repositories],
          isLoading: false,
        }));
        return newRepo;
      }
      throw new Error('Failed to create repository');
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'error' in error.response.data
        ? String(error.response.data.error)
        : 'Failed to create repository';
      
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
      if (response.success && response.data) {
        const updatedRepo = response.data as Repository;
        set((state) => ({
          repositories: state.repositories.map(r => 
            r.full_name === updatedRepo.full_name ? updatedRepo : r
          ),
          currentRepository: state.currentRepository?.full_name === updatedRepo.full_name
            ? updatedRepo 
            : state.currentRepository,
          isLoading: false,
        }));
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
      if (response.success) {
        const fullName = `${owner}/${repo}`;
        set((state) => ({
          repositories: state.repositories.filter(r => r.full_name !== fullName),
          currentRepository: state.currentRepository?.full_name === fullName
            ? null 
            : state.currentRepository,
          isLoading: false,
        }));
      }
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