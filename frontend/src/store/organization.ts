import { create } from 'zustand';
import { Organization, PaginatedResponse } from '@/types';
import { orgApi } from '@/lib/api';

interface OrganizationState {
  organizations: Organization[];
  currentOrganization: Organization | null;
  isLoading: boolean;
  error: string | null;
  pagination: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  } | null;
}

interface OrganizationActions {
  fetchOrganizations: (params?: { page?: number; per_page?: number }) => Promise<void>;
  fetchOrganization: (org: string) => Promise<void>;
  createOrganization: (data: { login: string; name: string; description?: string }) => Promise<Organization>;
  updateOrganization: (org: string, data: Partial<{
    name: string;
    description: string;
    location: string;
    website: string;
  }>) => Promise<void>;
  clearError: () => void;
  clearOrganizations: () => void;
}

export const useOrganizationStore = create<OrganizationState & OrganizationActions>((set, get) => ({
  // State
  organizations: [],
  currentOrganization: null,
  isLoading: false,
  error: null,
  pagination: null,

  // Actions
  fetchOrganizations: async (params = {}) => {
    set({ isLoading: true, error: null });
    try {
      const paginatedData = await orgApi.getOrganizations(params) as PaginatedResponse<Organization>;
      set({
        organizations: paginatedData.data,
        pagination: paginatedData.pagination,
        isLoading: false,
      });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'message' in error.response.data
        ? String(error.response.data.message)
        : 'Failed to fetch organizations';
      
      set({
        isLoading: false,
        error: errorMessage,
        organizations: [],
        pagination: null,
      });
    }
  },

  fetchOrganization: async (org: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orgApi.getOrganization(org);
      if (response.success && response.data) {
        set({
          currentOrganization: response.data as Organization,
          isLoading: false,
        });
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'message' in error.response.data
        ? String(error.response.data.message)
        : 'Failed to fetch organization';
      
      set({
        isLoading: false,
        error: errorMessage,
        currentOrganization: null,
      });
    }
  },

  createOrganization: async (data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orgApi.createOrganization(data);
      if (response.success && response.data) {
        const newOrg = response.data as Organization;
        set((state) => ({
          organizations: [newOrg, ...state.organizations],
          isLoading: false,
        }));
        return newOrg;
      }
      throw new Error('Failed to create organization');
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'message' in error.response.data
        ? String(error.response.data.message)
        : 'Failed to create organization';
      
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error;
    }
  },

  updateOrganization: async (org: string, data) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orgApi.updateOrganization(org, data);
      if (response.success && response.data) {
        const updatedOrg = response.data as Organization;
        set((state) => ({
          organizations: state.organizations.map(o => 
            o.login === updatedOrg.login ? updatedOrg : o
          ),
          currentOrganization: state.currentOrganization?.login === updatedOrg.login 
            ? updatedOrg 
            : state.currentOrganization,
          isLoading: false,
        }));
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error && 'response' in error && 
        typeof error.response === 'object' && error.response !== null &&
        'data' in error.response && typeof error.response.data === 'object' &&
        error.response.data !== null && 'message' in error.response.data
        ? String(error.response.data.message)
        : 'Failed to update organization';
      
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

  clearOrganizations: () => {
    set({ 
      organizations: [], 
      currentOrganization: null, 
      pagination: null,
      error: null 
    });
  },
}));