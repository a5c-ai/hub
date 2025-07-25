'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { IssueForm } from '@/components/issues/IssueForm';
import { MobileIssueForm } from '@/components/mobile/MobileIssueForm';
import { useMobile } from '@/hooks/useDevice';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useIssueStore } from '@/store/issues';

export default function NewIssuePage() {
  const params = useParams();
  const router = useRouter();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const isMobile = useMobile();
  const { createIssue, isCreating, operationError } = useIssueStore();

  const handleMobileSubmit = async (data: any) => {
    try {
      const issue = await createIssue(owner, repo, {
        title: data.title,
        body: data.description,
        assignee_id: data.assignees?.[0],
        label_ids: data.labels || [],
      });
      router.push(`/repositories/${owner}/${repo}/issues/${issue.number}`);
    } catch (error) {
      console.error('Failed to create issue:', error);
    }
  };

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-4">
            <Link 
              href="/repositories" 
              className="hover:text-foreground transition-colors"
            >
              Repositories
            </Link>
            <span>/</span>
            <Link 
              href={`/repositories/${owner}/${repo}`}
              className="hover:text-foreground transition-colors"
            >
              {owner}/{repo}
            </Link>
            <span>/</span>
            <Link 
              href={`/repositories/${owner}/${repo}/issues`}
              className="hover:text-foreground transition-colors"
            >
              Issues
            </Link>
            <span>/</span>
            <span className="text-foreground font-medium">New Issue</span>
          </nav>
          <h1 className="text-2xl font-bold text-foreground">Create New Issue</h1>
        </div>

        {/* Issue form */}
        {isMobile ? (
          <div className="h-full">
            <MobileIssueForm
              onSubmit={handleMobileSubmit}
              loading={isCreating}
              availableLabels={[]}
              availableAssignees={[]}
            />
          </div>
        ) : (
          <div className="bg-card shadow-sm rounded-lg border border-border p-6">
            <IssueForm
              repositoryOwner={owner}
              repositoryName={repo}
              mode="create"
            />
          </div>
        )}
      </div>
    </AppLayout>
  );
}