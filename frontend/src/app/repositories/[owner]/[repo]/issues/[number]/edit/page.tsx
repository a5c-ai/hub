'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect } from 'react';
import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { IssueForm } from '@/components/issues/IssueForm';
import { MobileIssueForm } from '@/components/mobile/MobileIssueForm';
import { useMobile } from '@/hooks/useDevice';
import { useIssueStore } from '@/store/issues';
import { Button } from '@/components/ui/Button';

export default function EditIssuePage() {
  const params = useParams();
  const router = useRouter();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const issueNumber = parseInt(params.number as string);

  const {
    currentIssue,
    isLoadingCurrentIssue,
    currentIssueError,
    fetchIssue,
    updateIssue,
    isUpdating,
  } = useIssueStore();
  
  const isMobile = useMobile();

  const handleMobileSubmit = async (data: {title: string; description: string; assignees?: string[]; labels?: string[]}) => {
    try {
      await updateIssue(owner, repo, issueNumber, {
        title: data.title,
        body: data.description,
        assignee_id: data.assignees?.[0],
        label_ids: data.labels || [],
      });
      router.push(`/repositories/${owner}/${repo}/issues/${issueNumber}`);
    } catch (error) {
      console.error('Failed to update issue:', error);
    }
  };

  useEffect(() => {
    if (issueNumber) {
      fetchIssue(owner, repo, issueNumber);
    }
  }, [owner, repo, issueNumber, fetchIssue]);

  if (isLoadingCurrentIssue) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (currentIssueError || !currentIssue) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center py-12">
            <div className="text-destructive mb-4">
              {currentIssueError || 'Issue not found'}
            </div>
            <Link href={`/repositories/${owner}/${repo}/issues`}>
              <Button variant="outline">Back to Issues</Button>
            </Link>
          </div>
        </div>
      </AppLayout>
    );
  }

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
            <Link 
              href={`/repositories/${owner}/${repo}/issues/${currentIssue.number}`}
              className="hover:text-foreground transition-colors"
            >
              #{currentIssue.number}
            </Link>
            <span>/</span>
            <span className="text-foreground font-medium">Edit</span>
          </nav>

          <h1 className="text-3xl font-bold text-foreground">
            Edit Issue #{currentIssue.number}
          </h1>
          <p className="text-muted-foreground mt-2">
            Update the details of this issue
          </p>
        </div>

        {/* Edit Form */}
        {isMobile ? (
          <div className="h-full">
            <MobileIssueForm
              onSubmit={handleMobileSubmit}
              loading={isUpdating}
              initialData={{
                title: currentIssue.title,
                description: currentIssue.body || '',
                assignees: currentIssue.assignee ? [currentIssue.assignee.id] : [],
                labels: currentIssue.labels?.map(label => label.id) || [],
              }}
              availableLabels={[]}
              availableAssignees={[]}
            />
          </div>
        ) : (
          <IssueForm
            repositoryOwner={owner}
            repositoryName={repo}
            mode="edit"
            issueNumber={currentIssue.number}
            initialData={{
              title: currentIssue.title,
              body: currentIssue.body || '',
              assignee_id: currentIssue.assignee?.id,
              milestone_id: currentIssue.milestone?.id,
              label_ids: currentIssue.labels?.map(label => label.id) || [],
            }}
          />
        )}
      </div>
    </AppLayout>
  );
} 