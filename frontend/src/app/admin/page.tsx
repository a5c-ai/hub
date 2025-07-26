'use client';

import React, { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import api from '@/lib/api';
import Link from 'next/link';

interface UserStats {
  total_users: number;
  active_users: number;
  inactive_users: number;
  admin_users: number;
  verified_users: number;
  mfa_enabled_users: number;
  users_this_month: number;
  users_last_month: number;
  logins_this_week: number;
}

interface RecentUser {
  id: string;
  username: string;
  email: string;
  full_name: string;
  avatar_url: string;
  created_at: string;
  is_admin: boolean;
  is_active: boolean;
}

export default function AdminDashboardPage() {
  const [userStats, setUserStats] = useState<UserStats | null>(null);
  const [recentUsers, setRecentUsers] = useState<RecentUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch user statistics
        const statsResponse = await api.get('/api/v1/admin/users/stats');
        setUserStats(statsResponse.data);

        // Fetch recent users (last 5)
        const usersResponse = await api.get('/api/v1/admin/users?page=1&per_page=5&sort_by=created_at&sort_dir=desc');
        setRecentUsers(usersResponse.data.users);

        setError(null);
      } catch (err: any) {
        setError(err.response?.data?.error || 'Failed to fetch admin data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-center h-64">
            <div className="text-muted-foreground">Loading admin dashboard...</div>
          </div>
        </div>
      </AppLayout>
    );
  }

  const growthRate = userStats && userStats.users_last_month > 0 
    ? ((userStats.users_this_month - userStats.users_last_month) / userStats.users_last_month * 100)
    : 0;

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Admin Dashboard</h1>
          <p className="text-muted-foreground mt-2">
            Overview of system statistics and recent activity
          </p>
        </div>

        {/* Error Display */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
            <div className="text-red-800">{error}</div>
            <button onClick={() => setError(null)} className="mt-2 text-red-600 hover:text-red-800">
              Dismiss
            </button>
          </div>
        )}

        {/* Quick Actions */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Link href="/admin/users">
            <Card className="p-6 hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">User Management</p>
                  <p className="text-lg font-semibold text-foreground">Manage Users</p>
                </div>
                <div className="text-blue-500 text-2xl">👥</div>
              </div>
            </Card>
          </Link>
          
          <Link href="/admin/auth">
            <Card className="p-6 hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">Authentication</p>
                  <p className="text-lg font-semibold text-foreground">SSO & Security</p>
                </div>
                <div className="text-green-500 text-2xl">🔐</div>
              </div>
            </Card>
          </Link>
          
          <Link href="/admin/analytics">
            <Card className="p-6 hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">Analytics</p>
                  <p className="text-lg font-semibold text-foreground">View Reports</p>
                </div>
                <div className="text-purple-500 text-2xl">📊</div>
              </div>
            </Card>
          </Link>
          
        <Link href="/admin/queue">
          <Card className="p-6 hover:shadow-lg transition-shadow cursor-pointer">
            <div className="flex items-center">
              <div className="flex-1">
                <p className="text-sm font-medium text-muted-foreground">Job Queue</p>
                <p className="text-lg font-semibold text-foreground">Monitoring & Management</p>
              </div>
              <div className="text-yellow-500 text-2xl">⚙️</div>
            </div>
          </Card>
        </Link>
        <Link href="/admin/email">
            <Card className="p-6 hover:shadow-lg transition-shadow cursor-pointer">
              <div className="flex items-center">
                <div className="flex-1">
                  <p className="text-sm font-medium text-muted-foreground">Email Service</p>
                  <p className="text-lg font-semibold text-foreground">Configuration</p>
                </div>
                <div className="text-orange-500 text-2xl">📧</div>
              </div>
            </Card>
          </Link>
        </div>

        {/* User Statistics */}
        {userStats && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
            <div className="lg:col-span-2">
              <Card>
                <div className="p-6">
                  <h3 className="text-lg font-semibold text-foreground mb-6">User Statistics</h3>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <div className="text-center">
                      <div className="text-3xl font-bold text-blue-600">{userStats.total_users}</div>
                      <div className="text-sm text-muted-foreground">Total Users</div>
                    </div>
                    
                    <div className="text-center">
                      <div className="text-3xl font-bold text-green-600">{userStats.active_users}</div>
                      <div className="text-sm text-muted-foreground">Active Users</div>
                      <div className="text-xs text-muted-foreground">
                        {userStats.total_users > 0 
                          ? `${Math.round((userStats.active_users / userStats.total_users) * 100)}%`
                          : '0%'} of total
                      </div>
                    </div>
                    
                    <div className="text-center">
                      <div className="text-3xl font-bold text-purple-600">{userStats.admin_users}</div>
                      <div className="text-sm text-muted-foreground">Admin Users</div>
                      <div className="text-xs text-muted-foreground">
                        {userStats.total_users > 0 
                          ? `${Math.round((userStats.admin_users / userStats.total_users) * 100)}%`
                          : '0%'} of total
                      </div>
                    </div>
                    
                    <div className="text-center">
                      <div className="text-3xl font-bold text-orange-600">{userStats.verified_users}</div>
                      <div className="text-sm text-muted-foreground">Verified Users</div>
                      <div className="text-xs text-muted-foreground">
                        {userStats.total_users > 0 
                          ? `${Math.round((userStats.verified_users / userStats.total_users) * 100)}%`
                          : '0%'} verified
                      </div>
                    </div>
                    
                    <div className="text-center">
                      <div className="text-3xl font-bold text-red-600">{userStats.mfa_enabled_users}</div>
                      <div className="text-sm text-muted-foreground">2FA Enabled</div>
                      <div className="text-xs text-muted-foreground">
                        {userStats.total_users > 0 
                          ? `${Math.round((userStats.mfa_enabled_users / userStats.total_users) * 100)}%`
                          : '0%'} secured
                      </div>
                    </div>
                    
                    <div className="text-center">
                      <div className="text-3xl font-bold text-cyan-600">{userStats.logins_this_week}</div>
                      <div className="text-sm text-muted-foreground">Weekly Logins</div>
                      <div className="text-xs text-muted-foreground">This week</div>
                    </div>
                  </div>
                </div>
              </Card>
            </div>

            {/* Growth Metrics */}
            <div>
              <Card>
                <div className="p-6">
                  <h3 className="text-lg font-semibold text-foreground mb-6">Growth Metrics</h3>
                  
                  <div className="space-y-6">
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-muted-foreground">This Month</span>
                        <span className="text-sm font-medium text-foreground">
                          {userStats.users_this_month} users
                        </span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                          className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                          style={{ 
                            width: userStats.users_this_month > 0 
                              ? `${Math.min(100, (userStats.users_this_month / Math.max(userStats.users_this_month, userStats.users_last_month)) * 100)}%`
                              : '0%'
                          }}
                        />
                      </div>
                    </div>
                    
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm text-muted-foreground">Last Month</span>
                        <span className="text-sm font-medium text-foreground">
                          {userStats.users_last_month} users
                        </span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                          className="bg-gray-400 h-2 rounded-full transition-all duration-300"
                          style={{ 
                            width: userStats.users_last_month > 0 
                              ? `${Math.min(100, (userStats.users_last_month / Math.max(userStats.users_this_month, userStats.users_last_month)) * 100)}%`
                              : '0%'
                          }}
                        />
                      </div>
                    </div>
                    
                    <div className="pt-4 border-t border-border">
                      <div className="text-center">
                        <div className={`text-2xl font-bold ${growthRate >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {growthRate >= 0 ? '+' : ''}{growthRate.toFixed(1)}%
                        </div>
                        <div className="text-sm text-muted-foreground">Growth Rate</div>
                        <div className="text-xs text-muted-foreground">vs. last month</div>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            </div>
          </div>
        )}

        {/* Recent Users */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <Card>
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-foreground">Recent Users</h3>
                <Link href="/admin/users">
                  <Button variant="outline" size="sm">
                    View All
                  </Button>
                </Link>
              </div>
              
              {recentUsers.length > 0 ? (
                <div className="space-y-4">
                  {recentUsers.map((user) => (
                    <div key={user.id} className="flex items-center space-x-3">
                      <div className="w-10 h-10 bg-gray-200 rounded-full flex items-center justify-center">
                        {user.avatar_url ? (
                          <img
                            src={user.avatar_url}
                            alt={user.full_name}
                            className="w-10 h-10 rounded-full"
                          />
                        ) : (
                          <span className="text-sm font-medium text-gray-600">
                            {user.full_name.split(' ').map(n => n[0]).join('').slice(0, 2)}
                          </span>
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium text-foreground truncate">
                          {user.full_name}
                        </div>
                        <div className="text-sm text-muted-foreground truncate">
                          @{user.username} • {user.email}
                        </div>
                      </div>
                      <div className="flex flex-col items-end space-y-1">
                        <div className="flex space-x-1">
                          {user.is_admin && (
                            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                              Admin
                            </span>
                          )}
                          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                            user.is_active 
                              ? 'bg-green-100 text-green-800' 
                              : 'bg-gray-100 text-gray-800'
                          }`}>
                            {user.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {new Date(user.created_at).toLocaleDateString()}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center text-muted-foreground py-8">
                  No recent users found
                </div>
              )}
            </div>
          </Card>

          {/* System Health */}
          <Card>
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-6">System Health</h3>
              
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">User Account Status</span>
                  <span className="text-sm font-medium text-green-600">Healthy</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Authentication System</span>
                  <span className="text-sm font-medium text-green-600">Online</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Email Service</span>
                  <span className="text-sm font-medium text-green-600">Operational</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Database</span>
                  <span className="text-sm font-medium text-green-600">Connected</span>
                </div>

                {userStats && (
                  <>
                    <div className="pt-4 border-t border-border">
                      <div className="text-sm text-muted-foreground mb-2">Account Security</div>
                      <div className="text-xs text-muted-foreground mb-1">
                        Email Verification: {userStats.total_users > 0 
                          ? `${Math.round((userStats.verified_users / userStats.total_users) * 100)}%`
                          : '0%'}
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-1.5 mb-2">
                        <div 
                          className="bg-orange-500 h-1.5 rounded-full"
                          style={{ 
                            width: userStats.total_users > 0 
                              ? `${(userStats.verified_users / userStats.total_users) * 100}%`
                              : '0%'
                          }}
                        />
                      </div>
                      
                      <div className="text-xs text-muted-foreground mb-1">
                        2FA Adoption: {userStats.total_users > 0 
                          ? `${Math.round((userStats.mfa_enabled_users / userStats.total_users) * 100)}%`
                          : '0%'}
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-1.5">
                        <div 
                          className="bg-red-500 h-1.5 rounded-full"
                          style={{ 
                            width: userStats.total_users > 0 
                              ? `${(userStats.mfa_enabled_users / userStats.total_users) * 100}%`
                              : '0%'
                          }}
                        />
                      </div>
                    </div>
                  </>
                )}
              </div>
            </div>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
}
