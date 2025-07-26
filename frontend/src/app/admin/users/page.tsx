'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Modal } from '@/components/ui/Modal';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import api from '@/lib/api';
import { formatDistance } from 'date-fns';

interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
  avatar_url: string;
  bio: string;
  location: string;
  website: string;
  company: string;
  email_verified: boolean;
  two_factor_enabled: boolean;
  is_active: boolean;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
  last_login_at: string | null;
  phone_number: string;
  type: string;
}

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

interface UserFormData {
  username: string;
  email: string;
  password: string;
  full_name: string;
  bio: string;
  location: string;
  website: string;
  company: string;
  phone_number: string;
  is_admin: boolean;
}

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [selectedUsers, setSelectedUsers] = useState<Set<string>>(new Set());
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [confirmUsername, setConfirmUsername] = useState('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalUsers, setTotalUsers] = useState(0);

  const [newUser, setNewUser] = useState<UserFormData>({
    username: '',
    email: '',
    password: '',
    full_name: '',
    bio: '',
    location: '',
    website: '',
    company: '',
    phone_number: '',
    is_admin: false,
  });

  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        page: page.toString(),
        per_page: '30',
        ...(searchTerm && { search: searchTerm }),
        ...(roleFilter && { role: roleFilter }),
        ...(statusFilter && { status: statusFilter }),
      });

      const response = await api.get(`/api/v1/admin/users?${params}`);
      setUsers(response.data.users);
      setTotalPages(response.data.pagination.total_pages);
      setTotalUsers(response.data.pagination.total);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch users');
    } finally {
      setLoading(false);
    }
  }, [page, searchTerm, roleFilter, statusFilter]);

  const fetchStats = useCallback(async () => {
    try {
      const response = await api.get('/api/v1/admin/users/stats');
      setStats(response.data);
    } catch (err: any) {
      console.error('Failed to fetch user stats:', err);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
    fetchUsers();
  };

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post('/api/v1/admin/users', newUser);
      setShowCreateModal(false);
      setNewUser({
        username: '',
        email: '',
        password: '',
        full_name: '',
        bio: '',
        location: '',
        website: '',
        company: '',
        phone_number: '',
        is_admin: false,
      });
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create user');
    }
  };

  const handleUpdateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingUser) return;

    try {
      const updateData = {
        full_name: editingUser.full_name,
        email: editingUser.email,
        bio: editingUser.bio,
        location: editingUser.location,
        website: editingUser.website,
        company: editingUser.company,
        phone_number: editingUser.phone_number,
        is_active: editingUser.is_active,
        is_admin: editingUser.is_admin,
      };

      await api.patch(`/api/v1/admin/users/${editingUser.id}`, updateData);
      setShowEditModal(false);
      setEditingUser(null);
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update user');
    }
  };

  const handleDeleteUser = async () => {
    if (!deletingUser || confirmUsername !== deletingUser.username) return;

    try {
      await api.delete(`/api/v1/admin/users/${deletingUser.id}`);
      setShowDeleteModal(false);
      setDeletingUser(null);
      setConfirmUsername('');
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete user');
    }
  };

  const toggleUserStatus = async (user: User) => {
    try {
      const endpoint = user.is_active ? 'disable' : 'enable';
      await api.post(`/api/v1/admin/users/${user.id}/${endpoint}`);
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to ${user.is_active ? 'disable' : 'enable'} user`);
    }
  };

  const toggleUserRole = async (user: User) => {
    try {
      await api.patch(`/api/v1/admin/users/${user.id}/role`, {
        is_admin: !user.is_admin,
      });
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update user role');
    }
  };

  const handleSelectUser = (userId: string) => {
    const newSelection = new Set(selectedUsers);
    if (newSelection.has(userId)) {
      newSelection.delete(userId);
    } else {
      newSelection.add(userId);
    }
    setSelectedUsers(newSelection);
  };

  const handleSelectAll = () => {
    if (selectedUsers.size === users.length) {
      setSelectedUsers(new Set());
    } else {
      setSelectedUsers(new Set(users.map(u => u.id)));
    }
  };

  const handleBulkAction = async (action: 'enable' | 'disable' | 'promote' | 'demote') => {
    try {
      const promises = Array.from(selectedUsers).map(userId => {
        const user = users.find(u => u.id === userId);
        if (!user) return Promise.resolve();

        switch (action) {
          case 'enable':
            return api.post(`/api/v1/admin/users/${userId}/enable`);
          case 'disable':
            return api.post(`/api/v1/admin/users/${userId}/disable`);
          case 'promote':
            return api.patch(`/api/v1/admin/users/${userId}/role`, { is_admin: true });
          case 'demote':
            return api.patch(`/api/v1/admin/users/${userId}/role`, { is_admin: false });
          default:
            return Promise.resolve();
        }
      });

      await Promise.all(promises);
      setSelectedUsers(new Set());
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError(err.response?.data?.error || `Failed to perform bulk ${action}`);
    }
  };

  if (loading && users.length === 0) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-center h-64">
            <div className="text-muted-foreground">Loading users...</div>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">User Management</h1>
          <p className="text-muted-foreground mt-2">Manage user accounts, roles, and permissions</p>
        </div>

        {/* Error Display */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md" data-testid="error-container">
            <div className="text-red-800">{error}</div>
            <button 
              onClick={() => setError(null)} 
              className="mt-2 text-red-600 hover:text-red-800"
              data-testid="retry-button"
            >
              Dismiss
            </button>
          </div>
        )}

        {/* Stats Cards */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <Card>
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-muted-foreground">Total Users</p>
                    <p className="text-2xl font-bold text-foreground">{stats.total_users}</p>
                  </div>
                  <div className="text-blue-500">👥</div>
                </div>
              </div>
            </Card>
            <Card>
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-muted-foreground">Active Users</p>
                    <p className="text-2xl font-bold text-green-600">{stats.active_users}</p>
                  </div>
                  <div className="text-green-500">✅</div>
                </div>
              </div>
            </Card>
            <Card>
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-muted-foreground">Admin Users</p>
                    <p className="text-2xl font-bold text-purple-600">{stats.admin_users}</p>
                  </div>
                  <div className="text-purple-500">⚡</div>
                </div>
              </div>
            </Card>
            <Card>
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-muted-foreground">This Month</p>
                    <p className="text-2xl font-bold text-orange-600">{stats.users_this_month}</p>
                  </div>
                  <div className="text-orange-500">📈</div>
                </div>
              </div>
            </Card>
          </div>
        )}

        {/* Controls */}
        <Card className="mb-6">
          <div className="p-6">
            <div className="flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between">
              <div className="flex flex-col sm:flex-row gap-4 flex-1">
                <form onSubmit={handleSearch} className="flex-1">
                  <Input
                    type="text"
                    placeholder="Search users by name, email, username, or company..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full"
                  />
                </form>
                <div className="flex gap-2">
                  <select
                    value={roleFilter}
                    onChange={(e) => {
                      setRoleFilter(e.target.value);
                      setPage(1);
                    }}
                    className="px-3 py-2 border border-border rounded-md text-sm"
                    data-testid="role-filter"
                  >
                    <option value="">All Roles</option>
                    <option value="admin">Admin</option>
                    <option value="user">User</option>
                  </select>
                  <select
                    value={statusFilter}
                    onChange={(e) => {
                      setStatusFilter(e.target.value);
                      setPage(1);
                    }}
                    className="px-3 py-2 border border-border rounded-md text-sm"
                    data-testid="status-filter"
                  >
                    <option value="">All Status</option>
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </div>
              </div>
              <Button onClick={() => setShowCreateModal(true)} data-testid="add-user-btn">
                Add User
              </Button>
            </div>
          </div>
        </Card>

        {/* Bulk Actions */}
        {selectedUsers.size > 0 && (
          <Card className="mb-6">
            <div className="p-4 bg-blue-50 border-blue-200" data-testid="bulk-actions-bar">
              <div className="flex items-center justify-between">
                <span className="text-sm text-blue-800">
                  {selectedUsers.size} user{selectedUsers.size > 1 ? 's' : ''} selected
                </span>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleBulkAction('enable')}
                  >
                    Enable
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleBulkAction('disable')}
                    data-testid="bulk-deactivate-btn"
                  >
                    Disable
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleBulkAction('promote')}
                    data-testid="bulk-role-change-btn"
                  >
                    Promote to Admin
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleBulkAction('demote')}
                  >
                    Demote to User
                  </Button>
                </div>
              </div>
            </div>
          </Card>
        )}

        {/* Users Table */}
        <Card>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-muted/50">
                <tr>
                  <th className="text-left p-4 font-medium text-foreground">
                    <input
                      type="checkbox"
                      checked={selectedUsers.size === users.length && users.length > 0}
                      onChange={handleSelectAll}
                      className="rounded"
                    />
                  </th>
                  <th className="text-left p-4 font-medium text-foreground">User</th>
                  <th className="text-left p-4 font-medium text-foreground">Email</th>
                  <th className="text-left p-4 font-medium text-foreground">Status</th>
                  <th className="text-left p-4 font-medium text-foreground">Role</th>
                  <th className="text-left p-4 font-medium text-foreground">Created</th>
                  <th className="text-left p-4 font-medium text-foreground">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id} className="border-t border-border" data-testid={`user-row-${user.id}`}>
                    <td className="p-4">
                      <input
                        type="checkbox"
                        checked={selectedUsers.has(user.id)}
                        onChange={() => handleSelectUser(user.id)}
                        className="rounded"
                        data-testid={`user-checkbox-${user.id}`}
                      />
                    </td>
                    <td className="p-4">
                      <div className="flex items-center space-x-3">
                        <Avatar src={user.avatar_url} alt={user.full_name} size="sm" />
                        <div>
                          <div className="font-medium text-foreground">{user.full_name}</div>
                          <div className="text-sm text-muted-foreground">@{user.username}</div>
                          {user.company && (
                            <div className="text-xs text-muted-foreground">{user.company}</div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="p-4">
                      <div className="text-sm text-foreground">{user.email}</div>
                      {user.email_verified && (
                        <Badge variant="success" size="sm">Verified</Badge>
                      )}
                      {user.two_factor_enabled && (
                        <Badge variant="outline" size="sm">2FA</Badge>
                      )}
                    </td>
                    <td className="p-4">
                      <Badge 
                        variant={user.is_active ? 'success' : 'secondary'}
                        data-testid="user-status"
                      >
                        {user.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </td>
                    <td className="p-4">
                      <Badge 
                        variant={user.is_admin ? 'default' : 'outline'}
                        data-testid="user-role"
                      >
                        {user.is_admin ? 'Admin' : 'User'}
                      </Badge>
                    </td>
                    <td className="p-4">
                      <div className="text-sm text-muted-foreground">
                        {formatDistance(new Date(user.created_at), new Date(), { addSuffix: true })}
                      </div>
                      {user.last_login_at && (
                        <div className="text-xs text-muted-foreground">
                          Last login: {formatDistance(new Date(user.last_login_at), new Date(), { addSuffix: true })}
                        </div>
                      )}
                    </td>
                    <td className="p-4">
                      <div className="flex space-x-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            setEditingUser(user);
                            setShowEditModal(true);
                          }}
                          data-testid="edit-user-btn"
                        >
                          Edit
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => toggleUserStatus(user)}
                          data-testid={user.is_active ? "deactivate-user-btn" : "activate-user-btn"}
                        >
                          {user.is_active ? 'Disable' : 'Enable'}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => toggleUserRole(user)}
                        >
                          {user.is_admin ? 'Demote' : 'Promote'}
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => {
                            setDeletingUser(user);
                            setShowDeleteModal(true);
                          }}
                          data-testid="delete-user-btn"
                        >
                          Delete
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="p-4 border-t border-border">
              <div className="flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                  Showing {((page - 1) * 30) + 1} to {Math.min(page * 30, totalUsers)} of {totalUsers} users
                </div>
                <div className="flex space-x-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page <= 1}
                  >
                    Previous
                  </Button>
                  <span className="px-3 py-1 text-sm text-foreground">
                    Page {page} of {totalPages}
                  </span>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={page >= totalPages}
                  >
                    Next
                  </Button>
                </div>
              </div>
            </div>
          )}
        </Card>

        {/* Create User Modal */}
        <Modal
          open={showCreateModal}
          onClose={() => setShowCreateModal(false)}
          title="Create New User"
          data-testid="add-user-modal"
        >
          <form onSubmit={handleCreateUser} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Username *
                </label>
                <Input
                  type="text"
                  value={newUser.username}
                  onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                  required
                  minLength={3}
                  maxLength={50}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Full Name *
                </label>
                <Input
                  type="text"
                  value={newUser.full_name}
                  onChange={(e) => setNewUser({ ...newUser, full_name: e.target.value })}
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Email *
                </label>
                <Input
                  type="email"
                  value={newUser.email}
                  onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
                  required
                  data-testid="email-input"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Password *
                </label>
                <Input
                  type="password"
                  value={newUser.password}
                  onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
                  required
                  minLength={12}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Company
                </label>
                <Input
                  type="text"
                  value={newUser.company}
                  onChange={(e) => setNewUser({ ...newUser, company: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Location
                </label>
                <Input
                  type="text"
                  value={newUser.location}
                  onChange={(e) => setNewUser({ ...newUser, location: e.target.value })}
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Bio
              </label>
              <textarea
                value={newUser.bio}
                onChange={(e) => setNewUser({ ...newUser, bio: e.target.value })}
                className="w-full px-3 py-2 border border-border rounded-md text-sm"
                rows={3}
              />
            </div>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="is_admin"
                checked={newUser.is_admin}
                onChange={(e) => setNewUser({ ...newUser, is_admin: e.target.checked })}
                className="rounded"
              />
              <label htmlFor="is_admin" className="ml-2 text-sm text-foreground">
                Grant admin privileges
              </label>
            </div>
            <div className="flex justify-end space-x-3">
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                Cancel
              </Button>
              <Button type="submit" data-testid="save-new-user">
                Create User
              </Button>
            </div>
          </form>
        </Modal>

        {/* Edit User Modal */}
        <Modal
          open={showEditModal}
          onClose={() => setShowEditModal(false)}
          title="Edit User"
          data-testid="edit-user-modal"
        >
          {editingUser && (
            <form onSubmit={handleUpdateUser} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Full Name
                  </label>
                  <Input
                    type="text"
                    value={editingUser.full_name}
                    onChange={(e) => setEditingUser({ ...editingUser, full_name: e.target.value })}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Email
                  </label>
                  <Input
                    type="email"
                    value={editingUser.email}
                    onChange={(e) => setEditingUser({ ...editingUser, email: e.target.value })}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Company
                  </label>
                  <Input
                    type="text"
                    value={editingUser.company}
                    onChange={(e) => setEditingUser({ ...editingUser, company: e.target.value })}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Location
                  </label>
                  <Input
                    type="text"
                    value={editingUser.location}
                    onChange={(e) => setEditingUser({ ...editingUser, location: e.target.value })}
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Bio
                </label>
                <textarea
                  value={editingUser.bio}
                  onChange={(e) => setEditingUser({ ...editingUser, bio: e.target.value })}
                  className="w-full px-3 py-2 border border-border rounded-md text-sm"
                  rows={3}
                />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="edit_is_active"
                    checked={editingUser.is_active}
                    onChange={(e) => setEditingUser({ ...editingUser, is_active: e.target.checked })}
                    className="rounded"
                  />
                  <label htmlFor="edit_is_active" className="ml-2 text-sm text-foreground">
                    Account is active
                  </label>
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="edit_is_admin"
                    checked={editingUser.is_admin}
                    onChange={(e) => setEditingUser({ ...editingUser, is_admin: e.target.checked })}
                    className="rounded"
                    data-testid="user-role-select"
                  />
                  <label htmlFor="edit_is_admin" className="ml-2 text-sm text-foreground">
                    Admin privileges
                  </label>
                </div>
              </div>
              <div className="flex justify-end space-x-3">
                <Button type="button" variant="outline" onClick={() => setShowEditModal(false)}>
                  Cancel
                </Button>
                <Button type="submit" data-testid="save-user-changes">
                  Save Changes
                </Button>
              </div>
            </form>
          )}
        </Modal>

        {/* Delete User Modal */}
        <Modal
          open={showDeleteModal}
          onClose={() => {
            setShowDeleteModal(false);
            setDeletingUser(null);
            setConfirmUsername('');
          }}
          title="Delete User"
          data-testid="delete-confirmation-modal"
        >
          {deletingUser && (
            <div className="space-y-4">
              <div className="p-4 bg-red-50 border border-red-200 rounded-md">
                <p className="text-red-800 font-medium">
                  Are you sure you want to delete this user?
                </p>
                <p className="text-red-700 text-sm mt-1">
                  This action cannot be undone. The user&apos;s data will be permanently removed.
                </p>
              </div>
              <div>
                <p className="text-sm text-foreground mb-2">
                  Type the username <strong>{deletingUser.username}</strong> to confirm:
                </p>
                <Input
                  type="text"
                  value={confirmUsername}
                  onChange={(e) => setConfirmUsername(e.target.value)}
                  placeholder={deletingUser.username}
                  data-testid="confirm-username-input"
                />
              </div>
              <div className="flex justify-end space-x-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setShowDeleteModal(false);
                    setDeletingUser(null);
                    setConfirmUsername('');
                  }}
                >
                  Cancel
                </Button>
                <Button
                  type="button"
                  variant="destructive"
                  onClick={handleDeleteUser}
                  disabled={confirmUsername !== deletingUser.username}
                  data-testid="confirm-delete-btn"
                >
                  Delete User
                </Button>
              </div>
            </div>
          )}
        </Modal>
      </div>
    </AppLayout>
  );
}
