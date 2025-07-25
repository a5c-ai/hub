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
        console.log('AuthStore: Starting login...', { email });
        set({ isLoading: true, error: null });
        try {
          console.log('AuthStore: Making API call...');
          const response = await authApi.login(email, password, mfaCode);
          console.log('AuthStore: API response received:', response);
          console.log('AuthStore: Response structure check:', {
            hasSuccess: 'success' in response,
            hasData: 'data' in response,
            successValue: response.success,
            dataValue: response.data,
            responseKeys: Object.keys(response)
          });
          
          // Handle both wrapped ApiResponse and direct AuthResponse from backend
          let authData: AuthUser | null = null;
          
          // Check if response is wrapped in ApiResponse format
          if (response && 'success' in response && 'data' in response && response.success && response.data) {
            authData = response.data as AuthUser;
          }
          // Check if response is a direct AuthUser object
          else if (response && 'user' in response && 'access_token' in response) {
            authData = response as unknown as AuthUser;
          }
          
          if (authData && authData.user && authData.access_token) {
            console.log('AuthStore: Setting auth state...', { user: authData.user.username });
            
            // Store tokens in localStorage FIRST
            console.log('AuthStore: Storing tokens in localStorage...', {
              access_token: authData.access_token?.substring(0, 20) + '...',
              refresh_token: authData.refresh_token?.substring(0, 20) + '...'
            });
            localStorage.setItem('auth_token', authData.access_token);
            if (authData.refresh_token) {
              localStorage.setItem('refresh_token', authData.refresh_token);
            }
            
            // Verify storage worked
            const storedToken = localStorage.getItem('auth_token');
            console.log('AuthStore: Token storage verification:', {
              stored: !!storedToken,
              matches: storedToken === authData.access_token
            });
            
            // Then set the store state
            set({
              user: authData.user,
              token: authData.access_token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
            
            console.log('AuthStore: Login completed successfully');
          } else {
            console.error('AuthStore: Response does not have expected structure', {
              hasAuthData: !!authData,
              responseKeys: response ? Object.keys(response) : 'no response',
              responseType: response && 'success' in response ? 'ApiResponse' : 'direct'
            });
            set({
              isLoading: false,
              error: 'Invalid response format from server',
            });
          }
        } catch (error: unknown) {
          console.error('AuthStore: Login error:', error);
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
        const { token: storeToken, isAuthenticated } = get();
        console.log('AuthStore: checkAuth called', { storeToken: !!storeToken, isAuthenticated });
        
        // If already authenticated with a valid token, no need to check again
        if (isAuthenticated && storeToken) {
          console.log('AuthStore: Already authenticated, skipping check');
          return;
        }
        
        // Always check localStorage as the source of truth
        const storageToken = localStorage.getItem('auth_token');
        const token = storeToken || storageToken;
        console.log('AuthStore: Tokens found', { 
          storeToken: !!storeToken, 
          storageToken: !!storageToken,
          localStorageKeys: Object.keys(localStorage),
          actualStorageToken: storageToken?.substring(0, 20) + '...' || 'none'
        });
        
        if (!token) {
          console.log('AuthStore: No token found, clearing auth state');
          set({ isAuthenticated: false, user: null, token: null });
          return;
        }

        console.log('AuthStore: Token found, validating with API...');
        set({ isLoading: true });
        try {
          const response = await authApi.getProfile();
          console.log('AuthStore: Profile API response:', response);
          
          if (response.success && response.data) {
            console.log('AuthStore: Valid token, setting authenticated state');
            set({
              user: response.data as User,
              token,
              isAuthenticated: true,
              isLoading: false,
              error: null,
            });
            // Ensure localStorage is in sync for API interceptor
            localStorage.setItem('auth_token', token);
          }
        } catch (error) {
          console.error('AuthStore: Token validation failed:', error);
          // Token is invalid, clear all auth state
          localStorage.removeItem('auth_token');
          localStorage.removeItem('refresh_token');
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
      onRehydrateStorage: () => (state) => {
        console.log('AuthStore: Rehydrating from storage...', state);
        // Ensure localStorage token is in sync after rehydration
        if (state?.token) {
          localStorage.setItem('auth_token', state.token);
        }
      },
    }
  )
);