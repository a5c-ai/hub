'use client';

import React from 'react';
import { Card } from '@/components/ui/Card';
import { Avatar } from '@/components/ui/Avatar';
import { Badge } from '@/components/ui/Badge';
import { ContributorStats } from '@/types';
import { UserGroupIcon } from '@heroicons/react/24/outline';

interface ContributorStatsProps {
  contributors: ContributorStats[];
  totalContributors?: number;
  showDetails?: boolean;
  maxDisplay?: number;
}

const ContributorStatsComponent: React.FC<ContributorStatsProps> = ({
  contributors,
  totalContributors,
  showDetails = true,
  maxDisplay = 10,
}) => {
  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  // Generate avatar URL based on email (using Gravatar-style hash)
  const getAvatarUrl = (email: string): string => {
    // Simple hash function for demo purposes
    let hash = 0;
    for (let i = 0; i < email.length; i++) {
      const char = email.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    const avatarId = Math.abs(hash) % 1000;
    return `https://i.pravatar.cc/40?img=${avatarId}`;
  };

  const getUsernameFromEmail = (email: string): string => {
    return email.split('@')[0];
  };

  if (!contributors || contributors.length === 0) {
    return (
      <Card className="p-6">
        <div className="text-center">
          <UserGroupIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-medium text-foreground mb-2">No Contributors</h3>
          <p className="text-muted-foreground">No contributor data available for this repository.</p>
        </div>
      </Card>
    );
  }

  // Sort contributors by commit count (descending)
  const sortedContributors = [...contributors].sort((a, b) => b.commits - a.commits);
  const displayedContributors = sortedContributors.slice(0, maxDisplay);

  return (
    <Card>
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <UserGroupIcon className="h-5 w-5 text-muted-foreground" />
            <h3 className="text-lg font-semibold text-foreground">Contributors</h3>
          </div>
          <Badge variant="secondary">
            {totalContributors || contributors.length}
          </Badge>
        </div>

        <div className="space-y-3">
          {displayedContributors.map((contributor, index) => (
            <div key={contributor.email} className="flex items-center justify-between p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors">
              <div className="flex items-center space-x-3">
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-muted-foreground w-6">
                    #{index + 1}
                  </span>
                  <Avatar
                    src={getAvatarUrl(contributor.email)}
                    alt={contributor.name || getUsernameFromEmail(contributor.email)}
                    size="sm"
                  />
                </div>
                <div className="flex-1">
                  <div className="font-medium text-foreground text-sm">
                    {contributor.name || getUsernameFromEmail(contributor.email)}
                  </div>
                  {showDetails && (
                    <div className="text-xs text-muted-foreground">
                      {contributor.email}
                    </div>
                  )}
                </div>
              </div>

              <div className="flex items-center space-x-4 text-sm">
                <div className="text-center">
                  <div className="font-medium text-foreground">
                    {formatNumber(contributor.commits)}
                  </div>
                  <div className="text-xs text-muted-foreground">commits</div>
                </div>
                
                {showDetails && contributor.additions !== undefined && contributor.deletions !== undefined && (
                  <>
                    <div className="text-center">
                      <div className="font-medium text-green-600">
                        +{formatNumber(contributor.additions)}
                      </div>
                      <div className="text-xs text-muted-foreground">added</div>
                    </div>
                    
                    <div className="text-center">
                      <div className="font-medium text-red-600">
                        -{formatNumber(contributor.deletions)}
                      </div>
                      <div className="text-xs text-muted-foreground">deleted</div>
                    </div>
                  </>
                )}

                {showDetails && contributor.last_commit_date && (
                  <div className="text-center">
                    <div className="text-xs text-muted-foreground">
                      Last: {formatDate(contributor.last_commit_date)}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        {contributors.length > maxDisplay && (
          <div className="mt-4 pt-4 border-t border-border text-center">
            <div className="text-sm text-muted-foreground">
              Showing {maxDisplay} of {contributors.length} contributors
            </div>
          </div>
        )}

        {/* Summary Stats */}
        <div className="mt-4 pt-4 border-t border-border">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
            <div>
              <div className="text-lg font-bold text-foreground">
                {contributors.length}
              </div>
              <div className="text-xs text-muted-foreground">Total Contributors</div>
            </div>
            <div>
              <div className="text-lg font-bold text-foreground">
                {formatNumber(contributors.reduce((sum, c) => sum + c.commits, 0))}
              </div>
              <div className="text-xs text-muted-foreground">Total Commits</div>
            </div>
            {showDetails && (
              <>
                <div>
                  <div className="text-lg font-bold text-green-600">
                    {formatNumber(contributors.reduce((sum, c) => sum + (c.additions || 0), 0))}
                  </div>
                  <div className="text-xs text-muted-foreground">Lines Added</div>
                </div>
                <div>
                  <div className="text-lg font-bold text-red-600">
                    {formatNumber(contributors.reduce((sum, c) => sum + (c.deletions || 0), 0))}
                  </div>
                  <div className="text-xs text-muted-foreground">Lines Deleted</div>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
};

export default ContributorStatsComponent;