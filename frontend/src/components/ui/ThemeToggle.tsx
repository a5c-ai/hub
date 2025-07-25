'use client';

import { useEffect, useState } from 'react';
import { useThemeStore, Theme } from '@/store/theme';
import { Dropdown } from './Dropdown';

const themeOptions = [
  { value: 'light', label: 'Light', icon: '☀️' },
  { value: 'dark', label: 'Dark', icon: '🌙' },
  { value: 'system', label: 'System', icon: '💻' },
] as const;

interface ThemeToggleProps {
  variant?: 'button' | 'dropdown';
  size?: 'sm' | 'md' | 'lg';
}

export function ThemeToggle({ variant = 'dropdown', size = 'md' }: ThemeToggleProps) {
  const { theme, setTheme, toggleTheme } = useThemeStore();
  const [mounted, setMounted] = useState(false);

  // Avoid hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return null;
  }

  if (variant === 'button') {
    const currentOption = themeOptions.find(option => option.value === theme);
    
    return (
      <button
        onClick={toggleTheme}
        className={`
          inline-flex items-center justify-center rounded-md border border-border
          bg-card text-card-foreground hover:bg-muted transition-colors
          focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2
          ${size === 'sm' ? 'h-8 w-8 text-sm' : size === 'lg' ? 'h-12 w-12 text-lg' : 'h-10 w-10'}
        `}
        title={`Current theme: ${currentOption?.label || 'Unknown'}`}
        aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
      >
        {currentOption?.icon || '🌓'}
      </button>
    );
  }

  return (
    <Dropdown
      trigger={
        <button
          className={`
            inline-flex items-center justify-center rounded-md border border-border
            bg-card text-card-foreground hover:bg-muted transition-colors
            focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2
            ${size === 'sm' ? 'h-8 px-2 text-sm' : size === 'lg' ? 'h-12 px-4 text-lg' : 'h-10 px-3'}
          `}
          aria-label="Select theme"
        >
          <span className="mr-2">
            {themeOptions.find(option => option.value === theme)?.icon || '🌓'}
          </span>
          <span className="capitalize">
            {themeOptions.find(option => option.value === theme)?.label || 'Theme'}
          </span>
          <svg
            className="ml-2 h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 9l-7 7-7-7"
            />
          </svg>
        </button>
      }
      items={themeOptions.map(option => ({
        label: `${option.icon} ${option.label}${theme === option.value ? ' ✓' : ''}`,
        onClick: () => setTheme(option.value as Theme),
      }))}
    />
  );
}