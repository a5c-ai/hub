'use client';

import Link from 'next/link';
import { Issue } from '@/types';
import { IssueStateButton } from './IssueStateButton';
import { LabelBadge } from '@/components/labels/LabelBadge';
import { formatRelativeTime } from '@/lib/utils';

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
              className="text-lg font-semibold text-gray-900 hover:text-blue-600 transition-colors"
            >
              {issue.title}
            </Link>
          </div>
          
          {issue.body && (
            <p className="text-gray-600 text-sm mb-3 line-clamp-2">
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
          
          <div className="flex items-center text-sm text-gray-500 space-x-4">
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
        
        <div className="flex items-center text-sm text-gray-500 ml-4">
          {issue.comments_count > 0 && (
            <div className="flex items-center">
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={2} 
                  d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" 
                />
              </svg>
              {issue.comments_count}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}