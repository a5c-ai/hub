import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { User, AuthUser } from '@/types';
import { authApi } from '@/lib/api';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

interface AuthActions {
  login: (email: string, password: string) => Promise<void>;
  register: (userData: { username: string; email: string; password: string; full_name: string }) => Promise<void>;
  logout: () => void;
  setUser: (user: User) => void;
  clearError: () => void;
  checkAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set) => ({
      // State
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (email: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.login(email, password);
          if (response.success && response.data) {
            const authData = response.data as AuthUser;
            set({
              user: authData.user,
              token: authData.token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
            // Store token in localStorage for API requests
            localStorage.setItem('auth_token', authData.token);
          }
        } catch (error: unknown) {
          const errorMessage = error instanceof Error && 'response' in error && 
            typeof error.response === 'object' && error.response !== null &&
            'data' in error.response && typeof error.response.data === 'object' &&
            error.response.data !== null && 'message' in error.response.data
            ? String(error.response.data.message)
            : 'Login failed';
          
          set({
            isLoading: false,
            error: errorMessage,
          });
          throw error;
        }
      },

      register: async (userData: { username: string; email: string; password: string; full_name: string }) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.register(userData);
          if (response.success && response.data) {
            const authData = response.data as AuthUser;
            set({
              user: authData.user,
              token: authData.token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
            // Store token in localStorage for API requests
            localStorage.setItem('auth_token', authData.token);
          }
        } catch (error: unknown) {
          const errorMessage = error instanceof Error && 'response' in error && 
            typeof error.response === 'object' && error.response !== null &&
            'data' in error.response && typeof error.response.data === 'object' &&
            error.response.data !== null && 'message' in error.response.data
            ? String(error.response.data.message)
            : 'Registration failed';
          
          set({
            isLoading: false,
            error: errorMessage,
          });
          throw error;
        }
      },

      logout: () => {
        // Clear localStorage
        localStorage.removeItem('auth_token');
        // Reset state
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          error: null,
        });
        // Call logout API endpoint
        authApi.logout().catch(() => {
          // Ignore errors on logout API call
        });
      },

      setUser: (user: User) => {
        set({ user });
      },

      clearError: () => {
        set({ error: null });
      },

      checkAuth: async () => {
        const token = localStorage.getItem('auth_token');
        if (!token) {
          set({ isAuthenticated: false, user: null, token: null });
          return;
        }

        set({ isLoading: true });
        try {
          const response = await authApi.getProfile();
          if (response.success && response.data) {
            set({
              user: response.data as User,
              token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
          }
        } catch {
          // Token is invalid, clear auth state
          localStorage.removeItem('auth_token');
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
            error: null,
          });
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);