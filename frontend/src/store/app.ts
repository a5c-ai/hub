import { create } from 'zustand';

interface AppState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark' | 'system';
  currentRepository: string | null;
  currentOrganization: string | null;
}

interface AppActions {
  setSidebarOpen: (open: boolean) => void;
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  setCurrentRepository: (repo: string | null) => void;
  setCurrentOrganization: (org: string | null) => void;
}

export const useAppStore = create<AppState & AppActions>((set) => ({
  // State
  sidebarOpen: true,
  theme: 'system',
  currentRepository: null,
  currentOrganization: null,

  // Actions
  setSidebarOpen: (open: boolean) => set({ sidebarOpen: open }),
  
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  
  setTheme: (theme: 'light' | 'dark' | 'system') => {
    set({ theme });
    // Apply theme to document
    if (typeof window !== 'undefined') {
      const root = window.document.documentElement;
      root.classList.remove('light', 'dark');
      
      if (theme === 'system') {
        const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
        root.classList.add(systemTheme);
      } else {
        root.classList.add(theme);
      }
    }
  },
  
  setCurrentRepository: (repo: string | null) => set({ currentRepository: repo }),
  
  setCurrentOrganization: (org: string | null) => set({ currentOrganization: org }),
}));