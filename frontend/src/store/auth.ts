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
  login: (email: string, password: string, mfaCode?: string) => Promise<void>;
  register: (userData: { username: string; email: string; password: string; full_name: string }) => Promise<void>;
  logout: () => void;
  setUser: (user: User) => void;
  clearError: () => void;
  checkAuth: () => Promise<void>;
  refreshToken: () => Promise<void>;
  forgotPassword: (email: string) => Promise<void>;
  resetPassword: (token: string, password: string) => Promise<void>;
}

export const useAuthStore = create<AuthState & AuthActions>()(
  persist(
    (set, get) => ({
      // State
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (email: string, password: string, mfaCode?: string) => {
        set({ isLoading: true, error: null });
        try {
          const response = await authApi.login(email, password, mfaCode);
          if (response.success && response.data) {
            const authData = response.data as AuthUser;
            set({
              user: authData.user,
              token: authData.access_token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
            // Store tokens in localStorage
            localStorage.setItem('auth_token', authData.access_token);
            if (authData.refresh_token) {
              localStorage.setItem('refresh_token', authData.refresh_token);
            }
          }
        } catch (error: unknown) {
          const errorMessage = error instanceof Error && 'response' in error && 
            typeof error.response === 'object' && error.response !== null &&
            'data' in error.response && typeof error.response.data === 'object' &&
            error.response.data !== null && 'error' in error.response.data
            ? String(error.response.data.error)
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
            // Registration returns a user object, not an auth response
            const user = response.data as User;
            set({
              user: user,
              token: null, // No auto-login after registration
              isAuthenticated: false,
              isLoading: false,
              error: null,
            });
          }
        } catch (error: unknown) {
          const errorMessage = error instanceof Error && 'response' in error && 
            typeof error.response === 'object' && error.response !== null &&
            'data' in error.response && typeof error.response.data === 'object' &&
            error.response.data !== null && 'error' in error.response.data
            ? String(error.response.data.error)
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
        localStorage.removeItem('refresh_token');
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

      refreshToken: async () => {
        const refreshToken = localStorage.getItem('refresh_token');
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        try {
          const response = await authApi.refreshToken(refreshToken);
          if (response.success && response.data) {
            const authData = response.data as AuthUser;
            set({
              token: authData.access_token,
            });
            localStorage.setItem('auth_token', authData.access_token);
          }
        } catch (error) {
          // Refresh token is invalid, logout user
          set({
            user: null,
            token: null,
            isAuthenticated: false,
            error: null,
          });
          localStorage.removeItem('auth_token');
          localStorage.removeItem('refresh_token');
          throw error;
        }
      },

      forgotPassword: async (email: string) => {
        set({ isLoading: true, error: null });
        try {
          await authApi.forgotPassword(email);
          set({ isLoading: false });
        } catch (error: unknown) {
          const errorMessage = error instanceof Error && 'response' in error && 
            typeof error.response === 'object' && error.response !== null &&
            'data' in error.response && typeof error.response.data === 'object' &&
            error.response.data !== null && 'error' in error.response.data
            ? String(error.response.data.error)
            : 'Failed to send password reset email';
          
          set({
            isLoading: false,
            error: errorMessage,
          });
          throw error;
        }
      },

      resetPassword: async (token: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          await authApi.resetPassword(token, password);
          set({ isLoading: false });
        } catch (error: unknown) {
          const errorMessage = error instanceof Error && 'response' in error && 
            typeof error.response === 'object' && error.response !== null &&
            'data' in error.response && typeof error.response.data === 'object' &&
            error.response.data !== null && 'error' in error.response.data
            ? String(error.response.data.error)
            : 'Failed to reset password';
          
          set({
            isLoading: false,
            error: errorMessage,
          });
          throw error;
        }
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