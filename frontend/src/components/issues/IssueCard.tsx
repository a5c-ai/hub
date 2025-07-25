'use client';

import Link from 'next/link';
import { Issue } from '@/types';
import { IssueStateButton } from './IssueStateButton';
import { LabelBadge } from '@/components/labels/LabelBadge';
import { formatRelativeTime } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';

interface IssueCardProps {
  issue: Issue;
  repositoryOwner: string;
  repositoryName: string;
}

export function IssueCard({ issue, repositoryOwner, repositoryName }: IssueCardProps) {
  return (
    <div className="border border-border rounded-lg p-4 hover:border-muted-foreground/50 transition-colors bg-card">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            <IssueStateButton state={issue.state} />
            <Link
              href={`/repositories/${repositoryOwner}/${repositoryName}/issues/${issue.number}`}
              className="text-lg font-semibold text-foreground hover:text-primary transition-colors"
            >
              {issue.title}
            </Link>
          </div>
          
          {issue.body && (
            <p className="text-muted-foreground text-sm mb-3 line-clamp-2">
              {issue.body}
            </p>
          )}
          
          {issue.labels && issue.labels.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-3">
              {issue.labels.map((label) => (
                <LabelBadge key={label.id} label={label} />
              ))}
            </div>
          )}
          
          <div className="flex items-center text-sm text-muted-foreground space-x-4">
            <span>#{issue.number}</span>
            <span>
              opened {formatRelativeTime(issue.created_at)}
              {issue.user && ` by ${issue.user.username}`}
            </span>
            {issue.assignee && (
              <span>assigned to {issue.assignee.username}</span>
            )}
            {issue.milestone && (
              <span>milestone: {issue.milestone.title}</span>
            )}
          </div>
        </div>
        
        <div className="flex items-center text-sm text-muted-foreground ml-4">
          <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clipRule="evenodd" />
          </svg>
          {formatDistanceToNow(new Date(issue.created_at), { addSuffix: true })}
        </div>
      </div>
    </div>
  );
}