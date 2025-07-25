'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import api from '@/lib/api';
import { createErrorHandler } from '@/lib/utils';

interface Branch {
  name: string;
  protected: boolean;
  commit: {
    sha: string;
    message: string;
  };
}

interface BranchProtection {
  required_status_checks?: {
    strict: boolean;
    contexts: string[];
  };
  enforce_admins: boolean;
  required_pull_request_reviews?: {
    required_approving_review_count: number;
    dismiss_stale_reviews: boolean;
    require_code_owner_reviews: boolean;
    require_last_push_approval: boolean;
  };
  restrictions?: {
    users: string[];
    teams: string[];
  };
  allow_force_pushes: boolean;
  allow_deletions: boolean;
}

export default function BranchProtectionPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showProtectionModal, setShowProtectionModal] = useState(false);
  const [selectedBranch, setSelectedBranch] = useState<string>('');
  const [protectionLoading, setProtectionLoading] = useState(false);
  
  const [protectionData, setProtectionData] = useState<BranchProtection>({
    enforce_admins: false,
    allow_force_pushes: false,
    allow_deletions: false,
    required_status_checks: {
      strict: false,
      contexts: []
    },
    required_pull_request_reviews: {
      required_approving_review_count: 1,
      dismiss_stale_reviews: false,
      require_code_owner_reviews: false,
      require_last_push_approval: false
    }
  });

  const fetchBranches = async () => {
    const handleError = createErrorHandler(setError, setLoading);
    
    const operation = async () => {
      const response = await api.get(`/repositories/${owner}/${repo}/branches`);
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setBranches(result);
    }
  };

  const fetchBranchProtection = async (branchName: string) => {
    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      const response = await api.get(`/repositories/${owner}/${repo}/branches/${branchName}/protection`);
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setProtectionData(result);
    }
  };

  useEffect(() => {
    fetchBranches();
  }, [owner, repo]);

  const handleProtectBranch = async () => {
    if (!selectedBranch) return;

    const handleError = createErrorHandler(setError, setProtectionLoading);
    
    const operation = async () => {
      const response = await api.put(
        `/repositories/${owner}/${repo}/branches/${selectedBranch}/protection`,
        protectionData
      );
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      // Update branch in the list
      setBranches(branches.map(branch => 
        branch.name === selectedBranch 
          ? { ...branch, protected: true }
          : branch
      ));
      setShowProtectionModal(false);
      setSelectedBranch('');
    }
  };

  const handleRemoveProtection = async (branchName: string) => {
    if (!confirm(`Are you sure you want to remove protection from ${branchName}?`)) return;

    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      await api.delete(`/repositories/${owner}/${repo}/branches/${branchName}/protection`);
    };

    const result = await handleError(operation);
    if (result !== null) {
      setBranches(branches.map(branch => 
        branch.name === branchName 
          ? { ...branch, protected: false }
          : branch
      ));
    }
  };

  const openProtectionModal = async (branchName: string) => {
    setSelectedBranch(branchName);
    
    // If branch is already protected, fetch current protection settings
    const branch = branches.find(b => b.name === branchName);
    if (branch?.protected) {
      await fetchBranchProtection(branchName);
    } else {
      // Reset to default protection settings
      setProtectionData({
        enforce_admins: false,
        allow_force_pushes: false,
        allow_deletions: false,
        required_status_checks: {
          strict: false,
          contexts: []
        },
        required_pull_request_reviews: {
          required_approving_review_count: 1,
          dismiss_stale_reviews: false,
          require_code_owner_reviews: false,
          require_last_push_approval: false
        }
      });
    }
    
    setShowProtectionModal(true);
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse space-y-4">
            <div className="h-8 bg-muted rounded w-1/3"></div>
            <div className="h-32 bg-muted rounded"></div>
            <div className="h-32 bg-muted rounded"></div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error}</div>
            <Button onClick={fetchBranches} disabled={loading}>
              {loading ? 'Retrying...' : 'Try Again'}
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <Link href="/repositories" className="hover:text-foreground transition-colors">
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
            href={`/repositories/${owner}/${repo}/settings`}
            className="hover:text-foreground transition-colors"
          >
            Settings
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">Branches</span>
        </nav>

        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Branch Protection Rules</h1>
          <p className="text-muted-foreground mt-2">
            Protect important branches by requiring status checks, reviews, or restrictions
          </p>
        </div>

        {/* Branches List */}
        <div className="space-y-4">
          {branches.length === 0 ? (
            <Card>
              <div className="p-8 text-center">
                <svg className="w-12 h-12 mx-auto mb-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
                <h3 className="text-lg font-medium text-foreground mb-2">No branches found</h3>
                <p className="text-muted-foreground">
                  Create some branches to configure protection rules
                </p>
              </div>
            </Card>
          ) : (
            branches.map((branch) => (
              <Card key={branch.name}>
                <div className="p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <h3 className="text-lg font-semibold text-foreground">{branch.name}</h3>
                        {branch.protected ? (
                          <Badge variant="default" className="bg-green-100 text-green-800">
                            Protected
                          </Badge>
                        ) : (
                          <Badge variant="secondary">
                            Unprotected
                          </Badge>
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground">
                        Latest commit: {branch.commit.message.slice(0, 60)}
                        {branch.commit.message.length > 60 ? '...' : ''}
                      </p>
                      <p className="text-xs text-muted-foreground mt-1">
                        {branch.commit.sha.slice(0, 7)}
                      </p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => openProtectionModal(branch.name)}
                      >
                        {branch.protected ? 'Edit Protection' : 'Add Protection'}
                      </Button>
                      {branch.protected && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleRemoveProtection(branch.name)}
                          className="text-red-600 hover:bg-red-50"
                        >
                          Remove Protection
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              </Card>
            ))
          )}
        </div>

        {/* Protection Configuration Modal */}
        <Modal 
          isOpen={showProtectionModal} 
          onClose={() => setShowProtectionModal(false)}
          title={`Branch Protection: ${selectedBranch}`}
          size="large"
        >
          <div className="space-y-6">
            {/* Status Checks */}
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-3">Status Checks</h3>
              <div className="space-y-3">
                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.required_status_checks?.strict || false}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      required_status_checks: {
                        ...protectionData.required_status_checks!,
                        strict: e.target.checked
                      }
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Require branches to be up to date before merging</span>
                    <p className="text-xs text-muted-foreground">
                      Ensure pull request branch is up to date with the base branch
                    </p>
                  </div>
                </label>
              </div>
            </div>

            {/* Pull Request Reviews */}
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-3">Pull Request Reviews</h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    Required approving reviews
                  </label>
                  <select
                    value={protectionData.required_pull_request_reviews?.required_approving_review_count || 1}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      required_pull_request_reviews: {
                        ...protectionData.required_pull_request_reviews!,
                        required_approving_review_count: parseInt(e.target.value)
                      }
                    })}
                    className="w-32 px-3 py-2 border border-input rounded-md bg-background text-foreground"
                  >
                    {[1, 2, 3, 4, 5, 6].map(num => (
                      <option key={num} value={num}>{num}</option>
                    ))}
                  </select>
                </div>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.required_pull_request_reviews?.dismiss_stale_reviews || false}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      required_pull_request_reviews: {
                        ...protectionData.required_pull_request_reviews!,
                        dismiss_stale_reviews: e.target.checked
                      }
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Dismiss stale reviews</span>
                    <p className="text-xs text-muted-foreground">
                      Automatically dismiss old reviews when new commits are pushed
                    </p>
                  </div>
                </label>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.required_pull_request_reviews?.require_code_owner_reviews || false}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      required_pull_request_reviews: {
                        ...protectionData.required_pull_request_reviews!,
                        require_code_owner_reviews: e.target.checked
                      }
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Require review from code owners</span>
                    <p className="text-xs text-muted-foreground">
                      Require review from users or teams listed as code owners
                    </p>
                  </div>
                </label>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.required_pull_request_reviews?.require_last_push_approval || false}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      required_pull_request_reviews: {
                        ...protectionData.required_pull_request_reviews!,
                        require_last_push_approval: e.target.checked
                      }
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Require approval of the most recent push</span>
                    <p className="text-xs text-muted-foreground">
                      Require approval even if the reviewer is the author of the last commit
                    </p>
                  </div>
                </label>
              </div>
            </div>

            {/* Administrative Settings */}
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-3">Administrative Settings</h3>
              <div className="space-y-3">
                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.enforce_admins}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      enforce_admins: e.target.checked
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Include administrators</span>
                    <p className="text-xs text-muted-foreground">
                      Enforce all configured restrictions for administrators
                    </p>
                  </div>
                </label>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.allow_force_pushes}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      allow_force_pushes: e.target.checked
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Allow force pushes</span>
                    <p className="text-xs text-muted-foreground">
                      Permit force pushes to this branch
                    </p>
                  </div>
                </label>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.allow_deletions}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      allow_deletions: e.target.checked
                    })}
                    className="rounded border-border"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Allow deletions</span>
                    <p className="text-xs text-muted-foreground">
                      Allow users with push access to delete this branch
                    </p>
                  </div>
                </label>
              </div>
            </div>

            <div className="flex justify-end space-x-3 pt-4 border-t">
              <Button variant="outline" onClick={() => setShowProtectionModal(false)}>
                Cancel
              </Button>
              <Button 
                onClick={handleProtectBranch} 
                disabled={protectionLoading}
              >
                {protectionLoading ? 'Saving...' : 'Save Protection Rules'}
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </AppLayout>
  );
}