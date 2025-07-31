'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { PlusIcon, UsersIcon, FolderIcon, CalendarIcon } from '@heroicons/react/24/outline';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Avatar, Badge } from '@/components/ui';
import { orgApi } from '@/lib/api';
import { Organization } from '@/types';
import { formatRelativeTime, createErrorHandler } from '@/lib/utils';

export default function OrganizationsPage() {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchOrganizations = async () => {
      const handleError = createErrorHandler(setError, setLoading);
      
      const result = await handleError(async () => {
        const response = await orgApi.getOrganizations({ per_page: 50 });
        return response.data || [];
      });
      
      if (result) {
        setOrganizations(result);
      }
    };

    fetchOrganizations();
  }, []);

  return (
    <AppLayout>
      <div className="p-6">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Organizations</h1>
            <p className="text-muted-foreground mt-2">
              Manage your organizations and collaborate with teams.
            </p>
          </div>
          <Button asChild>
            <Link href="/organizations/new">
              <PlusIcon className="h-4 w-4 mr-2" />
              New Organization
            </Link>
          </Button>
        </div>

        {error && (
          <div className="mb-6 rounded-md bg-destructive/10 p-4">
            <div className="text-sm text-destructive">
              Failed to load organizations: {error}
            </div>
          </div>
        )}

        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <Card key={i} className="animate-pulse">
                <CardHeader className="flex flex-row items-center space-y-0 pb-3">
                  <div className="w-12 h-12 bg-muted rounded-full"></div>
                  <div className="ml-4 flex-1">
                    <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
                    <div className="h-3 bg-muted rounded w-1/2"></div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="h-3 bg-muted rounded w-full"></div>
                    <div className="h-3 bg-muted rounded w-2/3"></div>
                    <div className="flex space-x-4 mt-4">
                      <div className="h-3 bg-muted rounded w-16"></div>
                      <div className="h-3 bg-muted rounded w-20"></div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : organizations.length === 0 ? (
          <div className="text-center py-12">
            <UsersIcon className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-xl font-medium text-foreground mb-2">
              No organizations yet
            </h3>
            <p className="text-muted-foreground mb-6 max-w-md mx-auto">
              Organizations are shared accounts where teams can collaborate across many projects. 
              Create your first organization to get started.
            </p>
            <Button asChild>
              <Link href="/organizations/new">
                <PlusIcon className="h-4 w-4 mr-2" />
                Create your first organization
              </Link>
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {organizations.map((org) => (
              <Card key={org.id} className="hover:shadow-md transition-shadow">
                <CardHeader className="flex flex-row items-center space-y-0 pb-3">
                  <Avatar
                    size="lg"
                    name={org.name}
                    src={org.avatar_url}
                    className="mr-4"
                  />
                  <div className="flex-1 min-w-0">
                    <Link 
                      href={`/organizations/${org.login}`}
                      className="text-lg font-semibold text-foreground hover:text-primary transition-colors line-clamp-1"
                    >
                      {org.name}
                    </Link>
                    <p className="text-sm text-muted-foreground">@{org.login}</p>
                  </div>
                </CardHeader>
                <CardContent>
                  {org.description && (
                    <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                      {org.description}
                    </p>
                  )}
                  
                  <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                    <div className="flex items-center">
                      <FolderIcon className="h-4 w-4 mr-1" />
                      <span>{org.public_repos} repos</span>
                    </div>
                    <div className="flex items-center">
                      <UsersIcon className="h-4 w-4 mr-1" />
                      <span>{org.followers} members</span>
                    </div>
                  </div>
                  
                  {org.location && (
                    <div className="mt-2 text-sm text-muted-foreground">
                      📍 {org.location}
                    </div>
                  )}
                  
                  {org.website && (
                    <div className="mt-2">
                      <a
                        href={org.website}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-primary hover:underline"
                      >
                        🔗 {org.website}
                      </a>
                    </div>
                  )}
                  
                  <div className="mt-4 pt-4 border-t border-border">
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground flex items-center">
                        <CalendarIcon className="h-3 w-3 mr-1" />
                        Created {formatRelativeTime(org.created_at)}
                      </span>
                      <Button variant="outline" size="sm" asChild>
                        <Link href={`/organizations/${org.login}`}>
                          View
                        </Link>
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  );
}