'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import api from '@/lib/api';

interface RefreshToken {
  id: string;
  user_id: string;
  username?: string;
  device_name: string;
  ip_address: string;
  user_agent: string;
  location_info?: string;
  created_at: string;
  expires_at: string;
  last_used_at: string;
  is_active: boolean;
  is_remembered: boolean;
  security_flags: number;
}

interface RefreshTokenStats {
  total_tokens: number;
  active_tokens: number;
  expired_tokens: number;
  suspicious_tokens: number;
}

interface RefreshTokenManagerProps {
  userId?: string;
  isAdminView?: boolean;
}

export function RefreshTokenManager({ userId, isAdminView = false }: RefreshTokenManagerProps) {
  const [tokens, setTokens] = useState<RefreshToken[]>([]);
  const [stats, setStats] = useState<RefreshTokenStats>({
    total_tokens: 0,
    active_tokens: 0,
    expired_tokens: 0,
    suspicious_tokens: 0
  });
  
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTokens, setSelectedTokens] = useState<string[]>([]);
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'expired' | 'suspicious'>('all');

  useEffect(() => {
    fetchTokens();
    fetchStats();
  }, [userId, statusFilter]);

  const fetchTokens = async () => {
    setLoading(true);
    try {
      const endpoint = isAdminView 
        ? `/admin/auth/tokens/refresh${userId ? `?user_id=${userId}` : ''}&status=${statusFilter === 'all' ? '' : statusFilter}`
        : `/user/auth/tokens/refresh?status=${statusFilter === 'all' ? '' : statusFilter}`;
      
      const response = await api.get(endpoint);
      setTokens(response.data.tokens || []);
    } catch (error) {
      console.error('Failed to fetch refresh tokens:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const endpoint = isAdminView 
        ? `/admin/auth/tokens/refresh/stats${userId ? `?user_id=${userId}` : ''}`
        : '/user/auth/tokens/refresh/stats';
      
      const response = await api.get(endpoint);
      setStats(response.data);
    } catch (error) {
      console.error('Failed to fetch token stats:', error);
    }
  };

  const handleRevokeToken = async (tokenId: string) => {
    try {
      const endpoint = isAdminView 
        ? `/admin/auth/tokens/refresh/${tokenId}/revoke`
        : `/user/auth/tokens/refresh/${tokenId}/revoke`;
      
      await api.post(endpoint);
      setTokens(prev => prev.filter(t => t.id !== tokenId));
      fetchStats();
    } catch (error) {
      console.error('Failed to revoke token:', error);
    }
  };

  const handleRevokeSelected = async () => {
    if (selectedTokens.length === 0) return;
    
    try {
      const endpoint = isAdminView 
        ? '/admin/auth/tokens/refresh/revoke-multiple' 
        : '/user/auth/tokens/refresh/revoke-multiple';
      
      await api.post(endpoint, { token_ids: selectedTokens });
      setTokens(prev => prev.filter(t => !selectedTokens.includes(t.id)));
      setSelectedTokens([]);
      fetchStats();
    } catch (error) {
      console.error('Failed to revoke tokens:', error);
    }
  };

  const handleRevokeExpired = async () => {
    if (!confirm('Are you sure you want to revoke all expired tokens?')) {
      return;
    }

    try {
      const endpoint = isAdminView 
        ? '/admin/auth/tokens/refresh/revoke-expired' 
        : '/user/auth/tokens/refresh/revoke-expired';
      
      await api.post(endpoint, userId ? { user_id: userId } : {});
      fetchTokens();
      fetchStats();
    } catch (error) {
      console.error('Failed to revoke expired tokens:', error);
    }
  };

  const toggleTokenSelection = (tokenId: string) => {
    setSelectedTokens(prev => 
      prev.includes(tokenId)
        ? prev.filter(id => id !== tokenId)
        : [...prev, tokenId]
    );
  };

  const filteredTokens = tokens.filter(token => {
    const matchesSearch = !searchQuery || 
      token.username?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      token.device_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      token.ip_address.includes(searchQuery) ||
      token.location_info?.toLowerCase().includes(searchQuery.toLowerCase());
    
    return matchesSearch;
  });

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    
    if (diffDays > 0) return `${diffDays}d ago`;
    if (diffHours > 0) return `${diffHours}h ago`;
    return 'Recently';
  };

  const isTokenExpired = (expiresAt: string) => {
    return new Date(expiresAt) < new Date();
  };

  const hasSecurityFlags = (flags: number) => {
    return flags > 0;
  };

  return (
    <div className="space-y-6">
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Total Tokens</p>
              <p className="text-2xl font-bold text-foreground">{stats.total_tokens}</p>
            </div>
            <div className="h-8 w-8 bg-blue-100 rounded-full flex items-center justify-center">
              <span className="text-blue-600 text-sm">🔄</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Active Tokens</p>
              <p className="text-2xl font-bold text-green-600">{stats.active_tokens}</p>
            </div>
            <div className="h-8 w-8 bg-green-100 rounded-full flex items-center justify-center">
              <span className="text-green-600 text-sm">✅</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Expired Tokens</p>
              <p className="text-2xl font-bold text-orange-600">{stats.expired_tokens}</p>
            </div>
            <div className="h-8 w-8 bg-orange-100 rounded-full flex items-center justify-center">
              <span className="text-orange-600 text-sm">⏰</span>
            </div>
          </div>
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Suspicious Tokens</p>
              <p className="text-2xl font-bold text-red-600">{stats.suspicious_tokens}</p>
            </div>
            <div className="h-8 w-8 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-red-600 text-sm">⚠️</span>
            </div>
          </div>
        </Card>
      </div>

      {/* Refresh Token Management */}
      <Card>
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-foreground">Refresh Token Management</h3>
            <div className="flex space-x-2">
              {selectedTokens.length > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRevokeSelected}
                  className="text-red-600 hover:text-red-700"
                >
                  Revoke Selected ({selectedTokens.length})
                </Button>
              )}
              {stats.expired_tokens > 0 && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRevokeExpired}
                  className="text-orange-600 hover:text-orange-700"
                >
                  Clean Expired
                </Button>
              )}
            </div>
          </div>

          <div className="flex flex-col md:flex-row gap-4 mb-4">
            <Input
              placeholder="Search by username, device, IP, or location..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="max-w-md"
            />
            
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
              className="px-3 py-1 border rounded-md text-sm"
            >
              <option value="all">All Tokens</option>
              <option value="active">Active</option>
              <option value="expired">Expired</option>
              <option value="suspicious">Suspicious</option>
            </select>
          </div>

          {loading ? (
            <div className="text-center py-8">Loading refresh tokens...</div>
          ) : filteredTokens.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No refresh tokens found
            </div>
          ) : (
            <div className="space-y-3">
              {filteredTokens.map((token) => (
                <div
                  key={token.id}
                  className={`border rounded-lg p-4 ${
                    isTokenExpired(token.expires_at) 
                      ? 'border-orange-200 bg-orange-50' 
                      : hasSecurityFlags(token.security_flags)
                        ? 'border-red-200 bg-red-50'
                        : 'border-border'
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
                            {token.is_active && !isTokenExpired(token.expires_at) && (
                              <Badge variant="default" className="bg-green-100 text-green-800">
                                Active
                              </Badge>
                            )}
                            {isTokenExpired(token.expires_at) && (
                              <Badge variant="destructive" className="bg-orange-100 text-orange-800">
                                Expired
                              </Badge>
                            )}
                            {hasSecurityFlags(token.security_flags) && (
                              <Badge variant="destructive">
                                Suspicious
                              </Badge>
                            )}
                            {token.is_remembered && (
                              <Badge variant="outline">
                                Remember Me
                              </Badge>
                            )}
                          </div>
                        </div>
                        
                        <div className="text-sm text-muted-foreground mt-1">
                          <div className="flex items-center space-x-4">
                            <span>{token.device_name}</span>
                            <span>{token.ip_address}</span>
                            {token.location_info && (
                              <span>{token.location_info}</span>
                            )}
                          </div>
                          <div className="mt-1 flex items-center space-x-4">
                            <span>Created: {formatTimestamp(token.created_at)}</span>
                            <span>Last used: {formatTimestamp(token.last_used_at)}</span>
                            <span>Expires: {new Date(token.expires_at).toLocaleDateString()}</span>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleRevokeToken(token.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        Revoke
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}