import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export type Theme = 'light' | 'dark' | 'system';

interface ThemeState {
  theme: Theme;
  resolvedTheme: 'light' | 'dark';
}

interface ThemeActions {
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

const getSystemTheme = (): 'light' | 'dark' => {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
};

const resolveTheme = (theme: Theme): 'light' | 'dark' => {
  if (theme === 'system') {
    return getSystemTheme();
  }
  return theme;
};

export const useThemeStore = create<ThemeState & ThemeActions>()(
  persist(
    (set, get) => ({
      // State
      theme: 'system',
      resolvedTheme: 'light', // Always start with light to prevent hydration mismatch

      // Actions
      setTheme: (theme: Theme) => {
        const resolved = resolveTheme(theme);
        set({ theme, resolvedTheme: resolved });
        
        // Apply theme to document only after hydration is complete
        if (typeof window !== 'undefined') {
          // Use requestAnimationFrame to ensure DOM is ready and hydration is complete
          requestAnimationFrame(() => {
            const root = window.document.documentElement;
            root.classList.remove('light', 'dark');
            root.classList.add(resolved);
          });
        }
      },

      toggleTheme: () => {
        const { theme } = get();
        // If currently system or light, switch to dark; if dark, switch to light
        const newTheme = theme === 'dark' ? 'light' : 'dark';
        get().setTheme(newTheme);
      },
    }),
    {
      name: 'theme-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        theme: state.theme,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // Re-resolve theme after rehydration
          const resolved = resolveTheme(state.theme);
          state.resolvedTheme = resolved;
          
          // Apply theme to document after hydration is complete
          if (typeof window !== 'undefined') {
            requestAnimationFrame(() => {
              const root = window.document.documentElement;
              root.classList.remove('light', 'dark');
              root.classList.add(resolved);
            });
          }
        }
      },
    }
  )
);

// Initialize theme on the client side
if (typeof window !== 'undefined') {
  // Listen for system theme changes
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleChange = () => {
    const { theme, setTheme } = useThemeStore.getState();
    if (theme === 'system') {
      setTheme('system'); // This will re-resolve the system theme
    }
  };
  
  mediaQuery.addEventListener('change', handleChange);
}