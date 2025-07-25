'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import api from '@/lib/api';
import { Organization, Repository, User } from '@/types';

export default function OrganizationOverviewPage() {
  const params = useParams();
  const org = params.org as string;
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [members, setMembers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'repositories' | 'people' | 'teams'>('repositories');

  useEffect(() => {
    const fetchOrganizationData = async () => {
      try {
        setLoading(true);
        const [orgResponse, reposResponse, membersResponse] = await Promise.all([
          api.get(`/organizations/${org}`),
          api.get(`/organizations/${org}/repositories`),
          api.get(`/organizations/${org}/members`)
        ]);
        setOrganization(orgResponse.data);
        setRepositories(reposResponse.data);
        setMembers(membersResponse.data);
      } catch (err: unknown) {
        setError(
          (err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
          'Failed to fetch organization data'
        );
      } finally {
        setLoading(false);
      }
    };

    fetchOrganizationData();
  }, [org]);

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="flex items-start space-x-6 mb-8">
              <div className="w-24 h-24 bg-muted rounded-full"></div>
              <div className="flex-1">
                <div className="h-8 bg-muted rounded w-1/3 mb-2"></div>
                <div className="h-4 bg-muted rounded w-1/4 mb-4"></div>
                <div className="h-4 bg-muted rounded w-1/2"></div>
              </div>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
              <div className="lg:col-span-2">
                <div className="h-96 bg-muted rounded"></div>
              </div>
              <div>
                <div className="h-64 bg-muted rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error}</div>
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (!organization) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-gray-500 text-lg">Organization not found</div>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Organization Header */}
        <div className="flex flex-col lg:flex-row lg:items-start lg:space-x-8 mb-8">
          <div className="flex items-start space-x-6 mb-6 lg:mb-0">
            <Avatar
              src={organization.avatar_url}
              alt={organization.name}
              size="2xl"
            />
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">{organization.name}</h1>
              <p className="text-xl text-gray-600 mb-2">@{organization.login}</p>
              {organization.description && (
                <p className="text-gray-600 mb-4">{organization.description}</p>
              )}
              <div className="flex items-center text-sm text-gray-500 space-x-4 mb-4">
                {organization.location && (
                  <div className="flex items-center">
                    <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    {organization.location}
                  </div>
                )}
                {organization.website && (
                  <div className="flex items-center">
                    <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                    </svg>
                    <a href={organization.website} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">
                      {organization.website}
                    </a>
                  </div>
                )}
                <div className="flex items-center">
                  <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3a2 2 0 012-2h4a2 2 0 012 2v4m0 0V7a2 2 0 012 2v6a2 2 0 01-2 2H6a2 2 0 01-2-2V9a2 2 0 012-2m0 0V7a2 2 0 012-2h4a2 2 0 012 2v4" />
                  </svg>
                  Created {new Date(organization.created_at).toLocaleDateString()}
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <Button>
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
                  </svg>
                  Follow
                </Button>
                <Link href={`/organizations/${org}/settings`}>
                  <Button variant="outline">
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Settings
                  </Button>
                </Link>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-3">
            {/* Navigation Tabs */}
            <div className="border-b border-border mb-6">
              <nav className="-mb-px flex space-x-8">
                <button
                  onClick={() => setActiveTab('repositories')}
                  className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === 'repositories'
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                  }`}
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 7a2 2 0 012-2h10a2 2 0 012 2v2M7 7h10" />
                  </svg>
                  Repositories
                  <Badge variant="secondary" className="ml-2">
                    {repositories.length}
                  </Badge>
                </button>
                
                <button
                  onClick={() => setActiveTab('people')}
                  className={`py-2 px-1 border-b-2 font-medium text-sm transition-colors ${
                    activeTab === 'people'
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                  }`}
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                  </svg>
                  People
                  <Badge variant="secondary" className="ml-2">
                    {members.length}
                  </Badge>
                </button>
                
                <Link
                  href={`/organizations/${org}/teams`}
                  className="py-2 px-1 border-b-2 border-transparent text-muted-foreground hover:text-foreground hover:border-border font-medium text-sm transition-colors"
                >
                  <svg className="w-4 h-4 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  Teams
                </Link>
              </nav>
            </div>

            {/* Tab Content */}
            {activeTab === 'repositories' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-4">
                    <input
                      type="text"
                      placeholder="Search repositories..."
                      className="px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                    <select className="px-3 py-2 border border-gray-300 rounded-md">
                      <option>All repositories</option>
                      <option>Public</option>
                      <option>Private</option>
                      <option>Forks</option>
                    </select>
                  </div>
                  <Button>
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    New repository
                  </Button>
                </div>

                {repositories.length > 0 ? (
                  repositories.map((repository) => (
                    <Card key={repository.id}>
                      <div className="p-6">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center space-x-2 mb-2">
                              <Link 
                                href={`/repositories/${repository.owner?.username}/${repository.name}`}
                                className="text-lg font-semibold text-blue-600 hover:text-blue-800"
                              >
                                {repository.name}
                              </Link>
                              <Badge variant={repository.private ? 'secondary' : 'default'}>
                                {repository.private ? 'Private' : 'Public'}
                              </Badge>
                              {repository.fork && (
                                <Badge variant="outline">Fork</Badge>
                              )}
                            </div>
                            {repository.description && (
                              <p className="text-gray-600 mb-3">{repository.description}</p>
                            )}
                            <div className="flex items-center space-x-4 text-sm text-gray-500">
                              {repository.language && (
                                <div className="flex items-center">
                                  <div className="w-3 h-3 rounded-full bg-blue-500 mr-2"></div>
                                  {repository.language}
                                </div>
                              )}
                              <div className="flex items-center">
                                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                                </svg>
                                {repository.stargazers_count}
                              </div>
                              <div className="flex items-center">
                                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z" />
                                </svg>
                                {repository.forks_count}
                              </div>
                              <span>Updated {new Date(repository.updated_at).toLocaleDateString()}</span>
                            </div>
                          </div>
                          <Button size="sm" variant="outline">
                            <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                            </svg>
                            Star
                          </Button>
                        </div>
                      </div>
                    </Card>
                  ))
                ) : (
                  <Card>
                    <div className="p-12 text-center">
                      <svg className="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 7a2 2 0 012-2h10a2 2 0 012 2v2M7 7h10" />
                      </svg>
                      <h3 className="text-lg font-medium text-gray-900 mb-2">No repositories yet</h3>
                      <p className="text-gray-600 mb-4">This organization hasn&apos;t created any repositories yet.</p>
                      <Button>Create your first repository</Button>
                    </div>
                  </Card>
                )}
              </div>
            )}

            {activeTab === 'people' && (
              <div className="space-y-4">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-4">
                    <input
                      type="text"
                      placeholder="Search members..."
                      className="px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>
                  <Button>
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                    Invite member
                  </Button>
                </div>

                {members.length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {members.map((member) => (
                      <Card key={member.id}>
                        <div className="p-6 text-center">
                          <Avatar
                            src={member.avatar_url}
                            alt={member.username}
                            size="lg"
                            className="mx-auto mb-4"
                          />
                          <h3 className="text-lg font-semibold text-gray-900 mb-1">{member.name}</h3>
                          <p className="text-gray-600 mb-3">@{member.username}</p>
                          <Link href={`/users/${member.username}`}>
                            <Button size="sm" variant="outline" className="w-full">
                              View Profile
                            </Button>
                          </Link>
                        </div>
                      </Card>
                    ))}
                  </div>
                ) : (
                  <Card>
                    <div className="p-12 text-center">
                      <svg className="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                      </svg>
                      <h3 className="text-lg font-medium text-gray-900 mb-2">No members yet</h3>
                      <p className="text-gray-600 mb-4">Invite people to join this organization.</p>
                      <Button>Invite your first member</Button>
                    </div>
                  </Card>
                )}
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Quick Stats */}
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Organization Stats</h3>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Public repositories</span>
                    <span className="font-medium">{organization.public_repos}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Members</span>
                    <span className="font-medium">{members.length}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Followers</span>
                    <span className="font-medium">{organization.followers}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Following</span>
                    <span className="font-medium">{organization.following}</span>
                  </div>
                </div>
              </div>
            </Card>

            {/* Recent Activity */}
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h3>
                <div className="text-center py-4">
                  <svg className="w-8 h-8 mx-auto mb-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <p className="text-sm text-gray-600">No recent activity</p>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}