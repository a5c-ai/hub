'use client';

import { useState, useEffect } from 'react';
import { Input } from './Input';
import { Badge } from './Badge';
import { Card } from './Card';
import api from '@/lib/api';

interface User {
  login: string;
  avatar_url?: string;
  type: 'User';
}

interface Team {
  name: string;
  slug: string;
  members_count?: number;
  type: 'Team';
}

interface BranchRestrictionsConfigProps {
  users: string[];
  teams: string[];
  onUsersChange: (users: string[]) => void;
  onTeamsChange: (teams: string[]) => void;
  owner: string;
  repo: string;
  disabled?: boolean;
}

export function BranchRestrictionsConfig({ 
  users, 
  teams, 
  onUsersChange, 
  onTeamsChange,
  owner,
  repo,
  disabled 
}: BranchRestrictionsConfigProps) {
  const [userSearch, setUserSearch] = useState('');
  const [teamSearch, setTeamSearch] = useState('');
  const [searchResults, setSearchResults] = useState<{ users: User[], teams: Team[] }>({ users: [], teams: [] });
  const [isSearching, setIsSearching] = useState(false);
  const [showUserDropdown, setShowUserDropdown] = useState(false);
  const [showTeamDropdown, setShowTeamDropdown] = useState(false);

  // Search for users and teams
  useEffect(() => {
    const searchDelay = setTimeout(async () => {
      if ((userSearch.length > 2 && showUserDropdown) || (teamSearch.length > 2 && showTeamDropdown)) {
        setIsSearching(true);
        try {
          const searchPromises = [];
          
          if (userSearch.length > 2 && showUserDropdown) {
            searchPromises.push(
              api.get(`/search/users?q=${encodeURIComponent(userSearch)}&per_page=10`)
                .then(response => ({ users: response.data.items || [] }))
                .catch(() => ({ users: [] }))
            );
          }
          
          if (teamSearch.length > 2 && showTeamDropdown) {
            // For organizations, search teams
            searchPromises.push(
              api.get(`/orgs/${owner}/teams?per_page=10`)
                .then(response => ({ 
                  teams: (response.data || []).filter((team: Team) => 
                    team.name.toLowerCase().includes(teamSearch.toLowerCase())
                  )
                }))
                .catch(() => ({ teams: [] }))
            );
          }

          const results = await Promise.all(searchPromises);
          const combinedResults = results.reduce((acc, result) => ({ ...acc, ...result }), { users: [], teams: [] });
          setSearchResults(combinedResults);
        } catch (error) {
          console.error('Search error:', error);
        } finally {
          setIsSearching(false);
        }
      } else {
        setSearchResults({ users: [], teams: [] });
      }
    }, 300);

    return () => clearTimeout(searchDelay);
  }, [userSearch, teamSearch, showUserDropdown, showTeamDropdown, owner]);

  const addUser = (userLogin: string) => {
    if (!users.includes(userLogin)) {
      onUsersChange([...users, userLogin]);
    }
    setUserSearch('');
    setShowUserDropdown(false);
  };

  const removeUser = (userLogin: string) => {
    onUsersChange(users.filter(user => user !== userLogin));
  };

  const addTeam = (teamName: string) => {
    if (!teams.includes(teamName)) {
      onTeamsChange([...teams, teamName]);
    }
    setTeamSearch('');
    setShowTeamDropdown(false);
  };

  const removeTeam = (teamName: string) => {
    onTeamsChange(teams.filter(team => team !== teamName));
  };

  return (
    <div className="space-y-6">
      <div>
        <h4 className="text-sm font-medium text-foreground mb-2">Push Access Restrictions</h4>
        <p className="text-xs text-muted-foreground mb-4">
          Restrict pushes to this branch to specific users and teams. 
          Leave empty to allow all users with push access to the repository.
        </p>
      </div>

      {/* Users Section */}
      <Card>
        <div className="p-4">
          <div className="flex items-center justify-between mb-3">
            <label className="text-sm font-medium text-foreground">
              Restrict pushes to users
            </label>
            <span className="text-xs text-muted-foreground">
              {users.length} user{users.length !== 1 ? 's' : ''}
            </span>
          </div>

          {/* Existing users */}
          {users.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-3">
              {users.map((user) => (
                <Badge
                  key={user}
                  variant="secondary"
                  className="flex items-center gap-2 pr-1"
                >
                  <div className="flex items-center gap-1">
                    <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
                    </svg>
                    <span className="text-xs">{user}</span>
                  </div>
                  {!disabled && (
                    <button
                      type="button"
                      onClick={() => removeUser(user)}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  )}
                </Badge>
              ))}
            </div>
          )}

          {/* Add user */}
          {!disabled && (
            <div className="relative">
              <Input
                value={userSearch}
                onChange={(e) => setUserSearch(e.target.value)}
                onFocus={() => setShowUserDropdown(true)}
                placeholder="Search for users..."
                className="w-full"
              />

              {/* User search dropdown */}
              {showUserDropdown && (
                <div className="absolute z-10 w-full mt-1 bg-background border border-border rounded-md shadow-lg max-h-48 overflow-auto">
                  {isSearching ? (
                    <div className="p-3 text-center text-sm text-muted-foreground">
                      Searching...
                    </div>
                  ) : searchResults.users.length > 0 ? (
                    searchResults.users.map((user) => (
                      <button
                        key={user.login}
                        type="button"
                        onClick={() => addUser(user.login)}
                        disabled={users.includes(user.login)}
                        className="w-full px-3 py-2 text-left hover:bg-muted/50 focus:bg-muted/50 focus:outline-none disabled:opacity-50"
                      >
                        <div className="flex items-center gap-2">
                          {user.avatar_url && (
                            <img 
                              src={user.avatar_url} 
                              alt={user.login}
                              className="w-5 h-5 rounded-full"
                            />
                          )}
                          <span className="text-sm">{user.login}</span>
                        </div>
                      </button>
                    ))
                  ) : userSearch.length > 2 ? (
                    <div className="p-3 text-center text-sm text-muted-foreground">
                      No users found for &quot;{userSearch}&quot;
                    </div>
                  ) : userSearch.length > 0 ? (
                    <div className="p-3 text-center text-sm text-muted-foreground">
                      Type at least 3 characters to search
                    </div>
                  ) : null}
                </div>
              )}

              {/* Click outside handler for users */}
              {showUserDropdown && (
                <div 
                  className="fixed inset-0 z-0" 
                  onClick={() => setShowUserDropdown(false)}
                />
              )}
            </div>
          )}
        </div>
      </Card>

      {/* Teams Section */}
      <Card>
        <div className="p-4">
          <div className="flex items-center justify-between mb-3">
            <label className="text-sm font-medium text-foreground">
              Restrict pushes to teams
            </label>
            <span className="text-xs text-muted-foreground">
              {teams.length} team{teams.length !== 1 ? 's' : ''}
            </span>
          </div>

          {/* Existing teams */}
          {teams.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-3">
              {teams.map((team) => (
                <Badge
                  key={team}
                  variant="secondary"
                  className="flex items-center gap-2 pr-1"
                >
                  <div className="flex items-center gap-1">
                    <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3z" />
                      <path d="M6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" />
                    </svg>
                    <span className="text-xs">{team}</span>
                  </div>
                  {!disabled && (
                    <button
                      type="button"
                      onClick={() => removeTeam(team)}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  )}
                </Badge>
              ))}
            </div>
          )}

          {/* Add team */}
          {!disabled && (
            <div className="relative">
              <Input
                value={teamSearch}
                onChange={(e) => setTeamSearch(e.target.value)}
                onFocus={() => setShowTeamDropdown(true)}
                placeholder="Search for teams..."
                className="w-full"
              />

              {/* Team search dropdown */}
              {showTeamDropdown && (
                <div className="absolute z-10 w-full mt-1 bg-background border border-border rounded-md shadow-lg max-h-48 overflow-auto">
                  {isSearching ? (
                    <div className="p-3 text-center text-sm text-muted-foreground">
                      Searching...
                    </div>
                  ) : searchResults.teams.length > 0 ? (
                    searchResults.teams.map((team) => (
                      <button
                        key={team.slug}
                        type="button"
                        onClick={() => addTeam(team.name)}
                        disabled={teams.includes(team.name)}
                        className="w-full px-3 py-2 text-left hover:bg-muted/50 focus:bg-muted/50 focus:outline-none disabled:opacity-50"
                      >
                        <div className="flex items-center justify-between">
                          <span className="text-sm">{team.name}</span>
                          {team.members_count && (
                            <span className="text-xs text-muted-foreground">
                              {team.members_count} members
                            </span>
                          )}
                        </div>
                      </button>
                    ))
                  ) : teamSearch.length > 2 ? (
                    <div className="p-3 text-center text-sm text-muted-foreground">
                      No teams found for &quot;{teamSearch}&quot;
                    </div>
                  ) : teamSearch.length > 0 ? (
                    <div className="p-3 text-center text-sm text-muted-foreground">
                      Type at least 3 characters to search
                    </div>
                  ) : null}
                </div>
              )}

              {/* Click outside handler for teams */}
              {showTeamDropdown && (
                <div 
                  className="fixed inset-0 z-0" 
                  onClick={() => setShowTeamDropdown(false)}
                />
              )}
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}