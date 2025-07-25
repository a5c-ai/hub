'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { useIssueStore } from '@/store/issues';
import { Button } from '@/components/ui/Button';
import { IssueStateButton } from '@/components/issues/IssueStateButton';
import { LabelBadge } from '@/components/labels/LabelBadge';
import { formatRelativeTime } from '@/lib/utils';

export default function IssuePage() {
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
    closeIssue,
    reopenIssue,
    isUpdating,
    operationError,
    comments,
    isLoadingComments,
    commentsError,
    fetchComments,
    createComment,
  } = useIssueStore();

  const [showCommentForm, setShowCommentForm] = useState(false);
  const [commentBody, setCommentBody] = useState('');
  const [isSubmittingComment, setIsSubmittingComment] = useState(false);

  useEffect(() => {
    if (issueNumber) {
      fetchIssue(owner, repo, issueNumber);
      fetchComments(owner, repo, issueNumber);
    }
  }, [owner, repo, issueNumber, fetchIssue, fetchComments]);

  const handleCloseIssue = async () => {
    if (!currentIssue) return;
    try {
      await closeIssue(owner, repo, currentIssue.number);
    } catch (error) {
      console.error('Failed to close issue:', error);
    }
  };

  const handleReopenIssue = async () => {
    if (!currentIssue) return;
    try {
      await reopenIssue(owner, repo, currentIssue.number);
    } catch (error) {
      console.error('Failed to reopen issue:', error);
    }
  };

  const handleEditIssue = () => {
    router.push(`/repositories/${owner}/${repo}/issues/${currentIssue.number}/edit`);
  };

  const handleSubmitComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentBody.trim() || !currentIssue) return;

    setIsSubmittingComment(true);
    try {
      await createComment(owner, repo, currentIssue.number, commentBody.trim());
      setCommentBody('');
      setShowCommentForm(false);
    } catch (error) {
      console.error('Failed to create comment:', error);
    } finally {
      setIsSubmittingComment(false);
    }
  };

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
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
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
            <span className="text-foreground font-medium">#{currentIssue.number}</span>
          </nav>

          <div className="flex items-center gap-4 mb-4">
            <IssueStateButton state={currentIssue.state} />
            <h1 className="text-3xl font-bold text-foreground">{currentIssue.title}</h1>
          </div>

          <div className="flex items-center text-muted-foreground text-sm space-x-4">
            <span>#{currentIssue.number}</span>
            <span>
              opened {formatRelativeTime(currentIssue.created_at)}
              {currentIssue.user && ` by ${currentIssue.user.username}`}
            </span>
            {currentIssue.assignee && (
              <span>assigned to {currentIssue.assignee.username}</span>
            )}
            {currentIssue.milestone && (
              <span>milestone: {currentIssue.milestone.title}</span>
            )}
          </div>
        </div>

        {/* Labels */}
        {currentIssue.labels && currentIssue.labels.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-6">
            {currentIssue.labels.map((label) => (
              <LabelBadge key={label.id} label={label} />
            ))}
          </div>
        )}

        {/* Issue body */}
        <div className="border border-border rounded-lg">
          <div className="border-b border-border p-4 bg-muted/30">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-medium">
                  {currentIssue.user?.username?.charAt(0).toUpperCase() || 'U'}
                </div>
                <div>
                  <div className="font-medium text-foreground">
                    {currentIssue.user?.username || 'Unknown User'}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    commented {formatRelativeTime(currentIssue.created_at)}
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="p-6">
            {currentIssue.body ? (
              <div className="prose dark:prose-invert max-w-none">
                <p className="text-foreground whitespace-pre-wrap">{currentIssue.body}</p>
              </div>
            ) : (
              <p className="text-muted-foreground italic">No description provided.</p>
            )}
          </div>
        </div>

        {/* Comments Section */}
        <div className="mt-8 space-y-6">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-foreground">
              Comments ({comments.length})
            </h3>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowCommentForm(!showCommentForm)}
            >
              {showCommentForm ? 'Cancel' : 'Add Comment'}
            </Button>
          </div>

          {/* Add Comment Form */}
          {showCommentForm && (
            <div className="border border-border rounded-lg">
              <div className="border-b border-border p-4 bg-muted/30">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-medium">
                    U
                  </div>
                  <div className="font-medium text-foreground">Add a comment</div>
                </div>
              </div>
              <form onSubmit={handleSubmitComment} className="p-6">
                <textarea
                  value={commentBody}
                  onChange={(e) => setCommentBody(e.target.value)}
                  placeholder="Leave a comment..."
                  rows={6}
                  className="w-full px-3 py-2 border border-input bg-background rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-ring focus:border-input text-foreground"
                  disabled={isSubmittingComment}
                />
                <div className="flex items-center justify-end space-x-3 mt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setShowCommentForm(false);
                      setCommentBody('');
                    }}
                    disabled={isSubmittingComment}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={isSubmittingComment || !commentBody.trim()}
                  >
                    {isSubmittingComment ? (
                      <div className="flex items-center">
                        <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2"></div>
                        Commenting...
                      </div>
                    ) : (
                      'Comment'
                    )}
                  </Button>
                </div>
              </form>
            </div>
          )}

          {/* Comments List */}
          {isLoadingComments ? (
            <div className="flex justify-center items-center py-8">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
            </div>
          ) : commentsError ? (
            <div className="text-center py-8">
              <div className="text-destructive mb-2">Failed to load comments</div>
              <div className="text-muted-foreground text-sm">{commentsError}</div>
            </div>
          ) : comments.length === 0 ? (
            <div className="text-center py-8">
              <div className="text-muted-foreground">No comments yet.</div>
            </div>
          ) : (
            <div className="space-y-4">
              {comments.map((comment) => (
                <div key={comment.id} className="border border-border rounded-lg">
                  <div className="border-b border-border p-4 bg-muted/30">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-medium">
                          {comment.user?.username?.charAt(0).toUpperCase() || 'U'}
                        </div>
                        <div>
                          <div className="font-medium text-foreground">
                            {comment.user?.username || 'Unknown User'}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            commented {formatRelativeTime(comment.created_at)}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div className="p-6">
                    <div className="prose dark:prose-invert max-w-none">
                      <p className="text-foreground whitespace-pre-wrap">{comment.body}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Action buttons */}
        <div className="mt-6 flex items-center space-x-3">
          {operationError && (
            <div className="text-destructive text-sm mb-4">
              {operationError}
            </div>
          )}
          <Link href={`/repositories/${owner}/${repo}/issues`}>
            <Button variant="outline">Back to Issues</Button>
          </Link>
          <Button variant="outline" onClick={handleEditIssue}>
            Edit Issue
          </Button>
          {currentIssue.state === 'open' ? (
            <Button 
              variant="outline" 
              onClick={handleCloseIssue}
              disabled={isUpdating}
            >
              {isUpdating ? 'Closing...' : 'Close Issue'}
            </Button>
          ) : (
            <Button 
              variant="outline" 
              onClick={handleReopenIssue}
              disabled={isUpdating}
            >
              {isUpdating ? 'Reopening...' : 'Reopen Issue'}
            </Button>
          )}
        </div>
      </div>
    </AppLayout>
  );
} 