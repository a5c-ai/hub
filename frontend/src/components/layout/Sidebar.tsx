'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  HomeIcon,
  // DocumentTextIcon,
  FolderIcon,
  UsersIcon,
  CogIcon,
  ShieldCheckIcon,
  // ChartBarIcon,
} from '@heroicons/react/24/outline';
import {
  HomeIcon as HomeIconSolid,
  // DocumentTextIcon as DocumentTextIconSolid,
  FolderIcon as FolderIconSolid,
  UsersIcon as UsersIconSolid,
  CogIcon as CogIconSolid,
  ShieldCheckIcon as ShieldCheckIconSolid,
  // ChartBarIcon as ChartBarIconSolid,
} from '@heroicons/react/24/solid';
import { cn } from '@/lib/utils';
import { useAppStore } from '@/store/app';
import { useAuthStore } from '@/store/auth';

const navigation = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: HomeIcon,
    iconSolid: HomeIconSolid,
  },
  {
    name: 'Repositories',
    href: '/repositories',
    icon: FolderIcon,
    iconSolid: FolderIconSolid,
  },
];

const adminNavigation = [
  {
    name: 'Admin Dashboard',
    href: '/admin',
    icon: ShieldCheckIcon,
    iconSolid: ShieldCheckIconSolid,
  },
  {
    name: 'User Management',
    href: '/admin/users',
    icon: UsersIcon,
    iconSolid: UsersIconSolid,
  },
];

const bottomNavigation = [
  {
    name: 'Settings',
    href: '/settings',
    icon: CogIcon,
    iconSolid: CogIconSolid,
  },
];

export function Sidebar() {
  const pathname = usePathname();
  const { sidebarOpen } = useAppStore();
  const { user } = useAuthStore();

  return (
    <div
      data-testid="sidebar"
      className={cn(
        'fixed inset-y-0 left-0 z-50 w-64 transform bg-background border-r border-border transition-transform lg:static lg:inset-0 lg:translate-x-0',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      )}
    >
      <div className="flex h-full flex-col">
        {/* Logo area */}
        <div className="flex h-16 shrink-0 items-center border-b border-border px-6">
          <Link href="/dashboard" className="flex items-center space-x-2">
            <div className="h-8 w-8 rounded-md bg-primary flex items-center justify-center">
              <span className="text-sm font-bold text-primary-foreground">H</span>
            </div>
            <span className="font-bold">Hub</span>
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 px-2 py-4">
          {navigation.map((item) => {
            const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
            const Icon = isActive ? item.iconSolid : item.icon;
            
            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  'group flex items-center rounded-md px-2 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <Icon
                  className={cn(
                    'mr-3 h-5 w-5 shrink-0',
                    isActive ? 'text-primary-foreground' : 'text-muted-foreground group-hover:text-foreground'
                  )}
                  aria-hidden="true"
                />
                {item.name}
              </Link>
            );
          })}

          {/* Admin Navigation */}
          {user?.is_admin && (
            <>
              <div className="pt-4">
                <h3 className="px-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                  Administration
                </h3>
                <div className="mt-2 space-y-1">
                  {adminNavigation.map((item) => {
                    const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
                    const Icon = isActive ? item.iconSolid : item.icon;
                    
                    return (
                      <Link
                        key={item.name}
                        href={item.href}
                        className={cn(
                          'group flex items-center rounded-md px-2 py-2 text-sm font-medium transition-colors',
                          isActive
                            ? 'bg-primary text-primary-foreground'
                            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                        )}
                      >
                        <Icon
                          className={cn(
                            'mr-3 h-5 w-5 shrink-0',
                            isActive ? 'text-primary-foreground' : 'text-muted-foreground group-hover:text-foreground'
                          )}
                          aria-hidden="true"
                        />
                        {item.name}
                      </Link>
                    );
                  })}
                </div>
              </div>
            </>
          )}
        </nav>

        {/* Recent repositories */}
        <div className="border-t border-border px-2 py-4">
          <h3 className="px-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Recent Repositories
          </h3>
          <div className="mt-2 space-y-1">
            {/* Placeholder for recent repositories */}
            <Link
              href="/repositories/user/example-repo"
              className="group flex items-center rounded-md px-2 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <FolderIcon className="mr-3 h-4 w-4 shrink-0" />
              example-repo
            </Link>
            <Link
              href="/repositories/user/another-repo"
              className="group flex items-center rounded-md px-2 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <FolderIcon className="mr-3 h-4 w-4 shrink-0" />
              another-repo
            </Link>
          </div>
        </div>

        {/* Bottom navigation */}
        <div className="border-t border-border px-2 py-4">
          {bottomNavigation.map((item) => {
            const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
            const Icon = isActive ? item.iconSolid : item.icon;
            
            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  'group flex items-center rounded-md px-2 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <Icon
                  className={cn(
                    'mr-3 h-5 w-5 shrink-0',
                    isActive ? 'text-primary-foreground' : 'text-muted-foreground group-hover:text-foreground'
                  )}
                  aria-hidden="true"
                />
                {item.name}
              </Link>
            );
          })}
        </div>

        {/* User info */}
        <div className="border-t border-border p-4">
          <div className="flex items-center">
            <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
              <span className="text-sm font-medium">
                {user?.name?.charAt(0) || user?.username?.charAt(0) || 'U'}
              </span>
            </div>
            <div className="ml-3">
              <p className="text-sm font-medium text-foreground">
                {user?.name || user?.username}
              </p>
              <p className="text-xs text-muted-foreground">{user?.email}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}