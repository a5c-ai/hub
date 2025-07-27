'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  Bars3Icon,
  MagnifyingGlassIcon,
  BellIcon,
  PlusIcon,
} from '@heroicons/react/24/outline';
import {
  Button,
  Avatar,
  Dropdown,
  Input,
} from '@/components/ui';
import { ThemeToggle } from '@/components/ui/ThemeToggle';
import { OfflineIndicator } from '@/components/ui/OfflineIndicator';
import { useAuthStore } from '@/store/auth';
import { useAppStore } from '@/store/app';

export function Header() {
  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState('');
  const { user, logout } = useAuthStore();
  const { toggleSidebar } = useAppStore();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      router.push(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  const userMenuItems = [
    {
      label: 'Your profile',
      onClick: () => router.push(`/${user?.username}`),
    },
    {
      label: 'Your repositories',
      onClick: () => router.push('/repositories'),
    },
    {
      label: 'Your organizations',
      onClick: () => router.push('/organizations'),
    },
    {
      label: 'Settings',
      onClick: () => router.push('/settings'),
    },
    {
      label: 'Sign out',
      onClick: handleLogout,
      destructive: true,
    },
  ];

  const createMenuItems = [
    {
      label: 'New repository',
      onClick: () => router.push('/repositories/new'),
    },
    {
      label: 'New organization',
      onClick: () => router.push('/organizations/new'),
    },
  ];

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60" data-testid="main-header">
      <div className="container mx-auto flex h-16 items-center justify-between px-4">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleSidebar}
            className="md:hidden"
            data-testid="mobile-menu-button"
          >
            <Bars3Icon className="h-5 w-5" />
          </Button>

          <Link href="/dashboard" className="flex items-center space-x-2">
            <div className="h-8 w-8 rounded-md bg-primary flex items-center justify-center">
              <span className="text-sm font-bold text-primary-foreground">H</span>
            </div>
            <span className="hidden font-bold sm:inline-block">Hub</span>
          </Link>
        </div>

        <div className="flex flex-1 items-center justify-center px-2 lg:ml-6 lg:justify-end">
          <div className="w-full max-w-lg lg:max-w-xs">
            <form onSubmit={handleSearch} className="relative">
              <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                <MagnifyingGlassIcon className="h-5 w-5 text-muted-foreground" />
              </div>
              <Input
                className="block w-full rounded-md border-0 bg-muted py-1.5 pl-10 pr-3 text-foreground placeholder:text-muted-foreground focus:ring-2 focus:ring-inset focus:ring-primary sm:text-sm sm:leading-6"
                placeholder="Search repositories..."
                type="search"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </form>
          </div>
        </div>

        <div className="flex items-center space-x-4">
          <OfflineIndicator className="hidden sm:flex" />
          
          <Dropdown
            trigger={
              <Button variant="ghost" size="icon" asChild>
                <span>
                  <PlusIcon className="h-5 w-5" />
                </span>
              </Button>
            }
            items={createMenuItems}
          />

          <Button variant="ghost" size="icon" className="relative">
            <BellIcon className="h-5 w-5" />
            <span className="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-destructive text-xs text-destructive-foreground flex items-center justify-center">
              3
            </span>
          </Button>

          <ThemeToggle variant="button" size="md" />

          <Dropdown
            trigger={
              <div className="flex items-center space-x-2" data-testid="user-menu">
                <Avatar
                  src={user?.avatar_url}
                  name={user?.name || user?.username}
                  size="sm"
                  data-testid="user-avatar"
                />
              </div>
            }
            items={userMenuItems.map(item => ({ ...item, ...(item.label === 'Sign out' ? { 'data-testid': 'logout-button' } : {}) }))}
          />
        </div>
      </div>
    </header>
  );
}