'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  HomeIcon,
  FolderIcon,
  MagnifyingGlassIcon,
  BellIcon,
  UserIcon,
} from '@heroicons/react/24/outline';
import {
  HomeIcon as HomeIconSolid,
  FolderIcon as FolderIconSolid,
  MagnifyingGlassIcon as MagnifyingGlassIconSolid,
  BellIcon as BellIconSolid,
  UserIcon as UserIconSolid,
} from '@heroicons/react/24/solid';
import { cn } from '@/lib/utils';
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
  {
    name: 'Search',
    href: '/search',
    icon: MagnifyingGlassIcon,
    iconSolid: MagnifyingGlassIconSolid,
  },
  {
    name: 'Notifications',
    href: '/notifications',
    icon: BellIcon,
    iconSolid: BellIconSolid,
  },
  {
    name: 'Profile',
    href: '/settings',
    icon: UserIcon,
    iconSolid: UserIconSolid,
  },
];

export function MobileBottomNav() {
  const pathname = usePathname();
  const { user } = useAuthStore();

  // Don't show on auth pages
  if (pathname.includes('/login') || pathname.includes('/register')) {
    return null;
  }

  return (
    <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-t border-border">
      <div className="flex items-center justify-around px-2 py-2 safe-area-inset-bottom">
        {navigation.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
          const Icon = isActive ? item.iconSolid : item.icon;
          
          return (
            <Link
              key={item.name}
              href={item.name === 'Profile' ? `/${user?.username}` : item.href}
              className={cn(
                'flex flex-col items-center justify-center px-2 py-2 text-xs font-medium transition-colors min-w-0 flex-1',
                isActive
                  ? 'text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              <Icon className="h-5 w-5 mb-1" />
              <span className="truncate">{item.name}</span>
              
              {/* Notification badge for notifications tab */}
              {item.name === 'Notifications' && (
                <div className="absolute top-1 right-1/2 transform translate-x-2 -translate-y-1">
                  <div className="h-2 w-2 rounded-full bg-destructive"></div>
                </div>
              )}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}