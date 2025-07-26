'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface BlacklistedToken {
  id: string;
  user_id: string;
  username?: string;
  token_hash: string;
  expires_at: string;
  reason: string;
  blacklisted_by: string;
  blacklisted_by_username?: string;
  created_at: string;
}

interface BlacklistStats {
  total_blacklisted: number;
  active_blacklisted: number;
  expired_count: number;
  recent_blacklists: number;
}

interface TokenBlacklistManagerProps {
  userId?: string;
  isAdminView?: boolean;
}

export function TokenBlacklistManager({ userId, isAdminView = false }: TokenBlacklistManagerProps) {
  const [blacklistedTokens, setBlacklistedTokens] = useState<BlacklistedToken[]>([]);
  const [stats, setStats] = useState<BlacklistStats>({
    total_blacklisted: 0,
    active_blacklisted: 0,
    expired_count: 0,
    recent_blacklists: 0
  });
  
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTokens, setSelectedTokens] = useState<string[]>([]);
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'expired'>('active');
  const [showAddModal, setShowAddModal] = useState(false);
  const [newBlacklistReason, setNewBlacklistReason] = useState('');
  const [newTokenToBlacklist, setNewTokenToBlacklist] = useState('');

  useEffect(() => {
    fetchBlacklistedTokens();
    fetchStats();
  }, [userId, statusFilter]);

  const fetchBlacklistedTokens = async () => {
    setLoading(true);
    try {
      const endpoint = isAdminView 
        ? `/admin/auth/tokens/blacklist${userId ? `?user_id=${userId}` : ''}${statusFilter !== 'all' ? `&status=${statusFilter}` : ''}`
        : `/user/auth/tokens/blacklist?status=${statusFilter}`;
      
      const response = await api.get(endpoint);
      setBlacklistedTokens(response.data.tokens || []);
    } catch (error) {
      console.error('Failed to fetch blacklisted tokens:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const endpoint = isAdminView 
        ? `/admin/auth/tokens/blacklist/stats${userId ? `?user_id=${userId}` : ''}`
        : '/user/auth/tokens/blacklist/stats';
      
      const response = await api.get(endpoint);
      setStats(response.data);
    } catch (error) {
      console.error('Failed to fetch blacklist stats:', error);
    }
  };

  const handleAddToBlacklist = async () => {
    if (!newTokenToBlacklist.trim() || !newBlacklistReason.trim()) return;
    
    try {
      const endpoint = isAdminView 
        ? '/admin/auth/tokens/blacklist'
        : '/user/auth/tokens/blacklist';
      
      await api.post(endpoint, {
        token: newTokenToBlacklist,
        reason: newBlacklistReason,
        user_id: userId
      });
      
      setNewTokenToBlacklist('');
      setNewBlacklistReason('');
      setShowAddModal(false);
      fetchBlacklistedTokens();
      fetchStats();
    } catch (error) {
      console.error('Failed to add token to blacklist:', error);
    }
  };

  const handleRemoveFromBlacklist = async (tokenId: string) => {
    if (!confirm('Are you sure you want to remove this token from the blacklist? This will restore the token if it hasn\'t expired.')) {
      return;
    }

    try {
      const endpoint = isAdminView 
        ? `/admin/auth/tokens/blacklist/${tokenId}/remove`
        : `/user/auth/tokens/blacklist/${tokenId}/remove`;
      
      await api.delete(endpoint);
      setBlacklistedTokens(prev => prev.filter(t => t.id !== tokenId));
      fetchStats();
    } catch (error) {
      console.error('Failed to remove token from blacklist:', error);
    }
  };

  const handleRemoveSelected = async () => {
    if (selectedTokens.length === 0) return;
    
    if (!confirm(`Are you sure you want to remove ${selectedTokens.length} tokens from the blacklist?`)) {
      return;
    }
    
    try {
      const endpoint = isAdminView 
        ? '/admin/auth/tokens/blacklist/remove-multiple' 
        : '/user/auth/tokens/blacklist/remove-multiple';
      
      await api.post(endpoint, { token_ids: selectedTokens });
      setBlacklistedTokens(prev => prev.filter(t => !selectedTokens.includes(t.id)));
      setSelectedTokens([]);
      fetchStats();
    } catch (error) {
      console.error('Failed to remove tokens from blacklist:', error);
    }
  };

  const handleCleanupExpired = async () => {
    if (!confirm('Are you sure you want to clean up all expired blacklist entries?')) {
      return;
    }

    try {
      const endpoint = isAdminView 
        ? '/admin/auth/tokens/blacklist/cleanup-expired' 
        : '/user/auth/tokens/blacklist/cleanup-expired';
      
      await api.post(endpoint, userId ? { user_id: userId } : {});
      fetchBlacklistedTokens();
      fetchStats();
    } catch (error) {
      console.error('Failed to cleanup expired tokens:', error);
    }
  };

  const toggleTokenSelection = (tokenId: string) => {
    setSelectedTokens(prev => 
      prev.includes(tokenId)
        ? prev.filter(id => id !== tokenId)
        : [...prev, tokenId]
    );
  };

  const filteredTokens = blacklistedTokens.filter(token => {
    const matchesSearch = !searchQuery || 
      token.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      token.reason.toLowerCase().includes(searchQuery.toLowerCase()) ||
      token.blacklisted_by_username?.toLowerCase().includes(searchQuery.toLowerCase());
    
    return matchesSearch;
  });

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const isTokenExpired = (expiresAt: string) => {
    return new Date(expiresAt) < new Date();
  };

  const getReasonBadgeColor = (reason: string) => {
    const lowerReason = reason.toLowerCase();
    if (lowerReason.includes('suspicious') || lowerReason.includes('compromised')) {
      return 'bg-red-100 text-red-800';
    }
    if (lowerReason.includes('logout') || lowerReason.includes('manual')) {
      return 'bg-blue-100 text-blue-800';
    }
    if (lowerReason.includes('expired') || lowerReason.includes('inactive')) {
      return 'bg-orange-100 text-orange-800';
    }
    return 'bg-gray-100 text-gray-800';
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Total Blacklisted</p>
              <p className="text-2xl font-bold text-foreground">{stats.total_blacklisted}</p>
            </div>
            <div className="h-8 w-8 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-red-600 text-sm">🚫</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Active Blacklisted</p>
              <p className="text-2xl font-bold text-red-600">{stats.active_blacklisted}</p>
            </div>
            <div className="h-8 w-8 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-red-600 text-sm">🔒</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Expired Entries</p>
              <p className="text-2xl font-bold text-orange-600">{stats.expired_count}</p>
            </div>
            <div className="h-8 w-8 bg-orange-100 rounded-full flex items-center justify-center">
              <span className="text-orange-600 text-sm">⏰</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Recent (24h)</p>
              <p className="text-2xl font-bold text-purple-600">{stats.recent_blacklists}</p>
            </div>
            <div className="h-8 w-8 bg-purple-100 rounded-full flex items-center justify-center">
              <span className="text-purple-600 text-sm">📊</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Token Blacklist Management */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">Token Blacklist Management</h3>
            <div className="flex space-x-2">
              {isAdminView && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowAddModal(true)}
                >
                  Add to Blacklist
                </Button>
              )}
              {selectedTokens.length > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRemoveSelected}
                  className="text-green-600 hover:text-green-700"
                >
                  Remove Selected ({selectedTokens.length})
                </Button>
              )}
              {stats.expired_count > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleCleanupExpired}
                  className="text-orange-600 hover:text-orange-700"
                >
                  Cleanup Expired
                </Button>
              )}
            </div>
          </div>

          <div className="flex flex-col md:flex-row gap-4 mb-4">
            <Input
              placeholder="Search by username, reason, or blacklisted by..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-md"
            />
            
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
              className="px-3 py-1 border rounded-md text-sm"
            >
              <option value="all">All Entries</option>
              <option value="active">Active</option>
              <option value="expired">Expired</option>
            </select>
          </div>

          {loading ? (
            <div className="text-center py-8">Loading blacklisted tokens...</div>
          ) : filteredTokens.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No blacklisted tokens found
            </div>
          ) : (
            <div className="space-y-3">
              {filteredTokens.map((token) => (
                <div
                  key={token.id}
                  className={`border rounded-lg p-4 ${
                    isTokenExpired(token.expires_at) 
                      ? 'border-orange-200 bg-orange-50' 
                      : 'border-red-200 bg-red-50'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <input
                        type="checkbox"
                        checked={selectedTokens.includes(token.id)}
                        onChange={() => toggleTokenSelection(token.id)}
                        className="rounded"
                      />
                      
                      <div className="flex-1">
                        <div className="flex items-center space-x-2">
                          {isAdminView && token.username && (
                            <span className="font-medium text-foreground">{token.username}</span>
                          )}
                          
                          <div className="flex items-center space-x-1">
                            {!isTokenExpired(token.expires_at) && (
                              <Badge variant="destructive">
                                Blacklisted
                              </Badge>
                            )}
                            {isTokenExpired(token.expires_at) && (
                              <Badge variant="outline" className="bg-orange-100 text-orange-800">
                                Expired
                              </Badge>
                            )}
                            <Badge className={getReasonBadgeColor(token.reason)}>
                              {token.reason}
                            </Badge>
                          </div>
                        </div>
                        
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>Token Hash: {token.token_hash.substring(0, 16)}...</span>
                            {token.blacklisted_by_username && (
                              <span>Blacklisted by: {token.blacklisted_by_username}</span>
                            )}
                          </div>
                          <div className="mt-1 flex items-center space-x-4">
                            <span>Blacklisted: {formatTimestamp(token.created_at)}</span>
                            <span>Expires: {formatTimestamp(token.expires_at)}</span>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRemoveFromBlacklist(token.id)}
                        className="text-green-600 hover:text-green-700"
                      >
                        Remove
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>

      {/* Add to Blacklist Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-md">
            <div className="p-6">
              <h3 className="text-lg font-semibold text-foreground mb-4">Add Token to Blacklist</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Token (or Token Hash)
                  </label>
                  <Input
                    value={newTokenToBlacklist}
                    onChange={(e) => setNewTokenToBlacklist(e.target.value)}
                    placeholder="Enter token or token hash"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Reason
                  </label>
                  <Input
                    value={newBlacklistReason}
                    onChange={(e) => setNewBlacklistReason(e.target.value)}
                    placeholder="e.g., Suspicious activity, Compromised token"
                  />
                </div>
              </div>
              
              <div className="flex justify-end space-x-2 mt-6">
                <Button
                  variant="outline"
                  onClick={() => setShowAddModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleAddToBlacklist}
                  disabled={!newTokenToBlacklist.trim() || !newBlacklistReason.trim()}
                  className="bg-red-600 hover:bg-red-700 text-white"
                >
                  Add to Blacklist
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}