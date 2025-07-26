'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Modal } from '@/components/ui/Modal';
import { PatternInput } from '@/components/ui/PatternInput';
import { StatusChecksConfig } from '@/components/ui/StatusChecksConfig';
import { BranchRestrictionsConfig } from '@/components/ui/BranchRestrictionsConfig';
import { ProtectionRulePreview } from '@/components/ui/ProtectionRulePreview';
import { ProtectionRuleMigration } from '@/components/ui/ProtectionRuleMigration';
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
  pattern?: string;
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
  require_linear_history?: boolean;
  require_conversation_resolution?: boolean;
}

interface ProtectionRule {
  id: string;
  pattern: string;
  repository_id: string;
  required_status_checks?: string;
  enforce_admins: boolean;
  required_pull_request_reviews?: string;
  restrictions?: string;
  created_at: string;
  updated_at: string;
}

export default function BranchProtectionPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  
  const [branches, setBranches] = useState<Branch[]>([]);
  const [protectionRules, setProtectionRules] = useState<ProtectionRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showProtectionModal, setShowProtectionModal] = useState(false);
  const [showRuleModal, setShowRuleModal] = useState(false);
  const [selectedBranch, setSelectedBranch] = useState<string>('');
  const [selectedRule, setSelectedRule] = useState<ProtectionRule | null>(null);
  const [protectionLoading, setProtectionLoading] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  
  const [protectionData, setProtectionData] = useState<BranchProtection>({
    pattern: '',
    enforce_admins: false,
    allow_force_pushes: false,
    allow_deletions: false,
    require_linear_history: false,
    require_conversation_resolution: false,
    required_status_checks: {
      strict: false,
      contexts: []
    },
    required_pull_request_reviews: {
      required_approving_review_count: 1,
      dismiss_stale_reviews: false,
      require_code_owner_reviews: false,
      require_last_push_approval: false
    },
    restrictions: {
      users: [],
      teams: []
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

  const fetchProtectionRules = async () => {
    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      const response = await api.get(`/repositories/${owner}/${repo}/branches/protection-rules`);
      return response.data;
    };

    const result = await handleError(operation);
    if (result) {
      setProtectionRules(result);
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
    Promise.all([
      fetchBranches(),
      fetchProtectionRules()
    ]);
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
      fetchProtectionRules(); // Refresh rules list
    }
  };

  const handleSaveProtectionRule = async () => {
    if (!protectionData.pattern) return;

    const handleError = createErrorHandler(setError, setProtectionLoading);
    
    const operation = async () => {
      const payload = {
        pattern: protectionData.pattern,
        required_status_checks: protectionData.required_status_checks?.contexts?.length || protectionData.required_status_checks?.strict 
          ? protectionData.required_status_checks 
          : null,
        required_pull_request_reviews: protectionData.required_pull_request_reviews,
        enforce_admins: protectionData.enforce_admins,
        restrictions: (protectionData.restrictions?.users?.length || protectionData.restrictions?.teams?.length) 
          ? protectionData.restrictions 
          : null,
        allow_force_pushes: protectionData.allow_force_pushes,
        allow_deletions: protectionData.allow_deletions,
        require_linear_history: protectionData.require_linear_history,
        require_conversation_resolution: protectionData.require_conversation_resolution
      };

      if (selectedRule) {
        // Update existing rule
        const response = await api.put(`/repositories/${owner}/${repo}/branches/protection-rules/${selectedRule.id}`, payload);
        return response.data;
      } else {
        // Create new rule
        const response = await api.post(`/repositories/${owner}/${repo}/branches/protection-rules`, payload);
        return response.data;
      }
    };

    const result = await handleError(operation);
    if (result) {
      await fetchProtectionRules();
      await fetchBranches(); // Refresh branches to update protection status
      setShowRuleModal(false);
      resetProtectionData();
    }
  };

  const handleDeleteProtectionRule = async (ruleId: string) => {
    if (!confirm('Are you sure you want to delete this protection rule?')) return;

    const handleError = createErrorHandler(setError);
    
    const operation = async () => {
      await api.delete(`/repositories/${owner}/${repo}/branches/protection-rules/${ruleId}`);
    };

    const result = await handleError(operation);
    if (result !== null) {
      await fetchProtectionRules();
      await fetchBranches();
    }
  };

  const resetProtectionData = () => {
    setProtectionData({
      pattern: '',
      enforce_admins: false,
      allow_force_pushes: false,
      allow_deletions: false,
      require_linear_history: false,
      require_conversation_resolution: false,
      required_status_checks: {
        strict: false,
        contexts: []
      },
      required_pull_request_reviews: {
        required_approving_review_count: 1,
        dismiss_stale_reviews: false,
        require_code_owner_reviews: false,
        require_last_push_approval: false
      },
      restrictions: {
        users: [],
        teams: []
      }
    });
    setSelectedRule(null);
  };

  const openCreateRuleModal = () => {
    resetProtectionData();
    setShowRuleModal(true);
  };

  const openEditRuleModal = (rule: ProtectionRule) => {
    try {
      // Parse stored JSON data
      const requiredStatusChecks = rule.required_status_checks 
        ? JSON.parse(rule.required_status_checks) 
        : { strict: false, contexts: [] };
      
      const requiredPRReviews = rule.required_pull_request_reviews 
        ? JSON.parse(rule.required_pull_request_reviews) 
        : {
            required_approving_review_count: 1,
            dismiss_stale_reviews: false,
            require_code_owner_reviews: false,
            require_last_push_approval: false
          };
      
      const restrictions = rule.restrictions 
        ? JSON.parse(rule.restrictions) 
        : { users: [], teams: [] };

      setProtectionData({
        pattern: rule.pattern,
        enforce_admins: rule.enforce_admins,
        allow_force_pushes: false, // These need to be added to backend model
        allow_deletions: false,
        require_linear_history: false,
        require_conversation_resolution: false,
        required_status_checks: requiredStatusChecks,
        required_pull_request_reviews: requiredPRReviews,
        restrictions: restrictions
      });
      
      setSelectedRule(rule);
      setShowRuleModal(true);
    } catch (error) {
      console.error('Error parsing rule data:', error);
      // Fall back to default values if parsing fails
      resetProtectionData();
      setProtectionData(prev => ({ ...prev, pattern: rule.pattern }));
      setSelectedRule(rule);
      setShowRuleModal(true);
    }
  };

  const handleMigrateProtectionRules = async (rules: Array<{ pattern: string; branches: string[] }>) => {
    for (const rule of rules) {
      const payload = {
        pattern: rule.pattern,
        required_status_checks: {
          strict: false,
          contexts: []
        },
        required_pull_request_reviews: {
          required_approving_review_count: 1,
          dismiss_stale_reviews: false,
          require_code_owner_reviews: false,
          require_last_push_approval: false
        },
        enforce_admins: false,
        restrictions: null,
        allow_force_pushes: false,
        allow_deletions: false
      };

      await api.post(`/repositories/${owner}/${repo}/branches/protection-rules`, payload);
    }
    
    // Refresh data after migration
    await fetchProtectionRules();
    await fetchBranches();
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
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-foreground">Branch Protection Rules</h1>
              <p className="text-muted-foreground mt-2">
                Protect important branches by requiring status checks, reviews, or restrictions
              </p>
            </div>
            <Button onClick={openCreateRuleModal}>
              Add Protection Rule
            </Button>
          </div>
        </div>

        {/* Migration Tool */}
        <ProtectionRuleMigration
          branches={branches}
          onMigrate={handleMigrateProtectionRules}
          className="mb-8"
        />

        {/* Protection Rules Section */}
        {protectionRules.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-foreground mb-4">Active Protection Rules</h2>
            <div className="space-y-4">
              {protectionRules.map((rule) => (
                <Card key={rule.id}>
                  <div className="p-6">
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3 mb-2">
                          <h3 className="text-lg font-semibold text-foreground">
                            Pattern: <code className="bg-muted px-2 py-1 rounded text-sm">{rule.pattern}</code>
                          </h3>
                          <Badge variant="outline" className="text-xs">
                            {rule.pattern === '*' ? 'All Branches' : 
                             rule.pattern.includes('*') ? 'Wildcard Pattern' : 'Exact Match'}
                          </Badge>
                          {rule.enforce_admins && (
                            <Badge variant="default" className="text-xs bg-orange-100 text-orange-800">
                              Admin Enforced
                            </Badge>
                          )}
                        </div>
                        
                        <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                          {/* Status Checks */}
                          {rule.required_status_checks && (() => {
                            try {
                              const statusChecks = JSON.parse(rule.required_status_checks);
                              return (
                                <div className="flex items-center space-x-1">
                                  <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                  </svg>
                                  <span>Status checks ({statusChecks.contexts?.length || 0})</span>
                                </div>
                              );
                            } catch (e) {
                              return null;
                            }
                          })()}
                          
                          {/* PR Reviews */}
                          {rule.required_pull_request_reviews && (() => {
                            try {
                              const prReviews = JSON.parse(rule.required_pull_request_reviews);
                              return (
                                <div className="flex items-center space-x-1">
                                  <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                                  </svg>
                                  <span>{prReviews.required_approving_review_count} review{prReviews.required_approving_review_count !== 1 ? 's' : ''}</span>
                                </div>
                              );
                            } catch (e) {
                              return null;
                            }
                          })()}
                          
                          {/* Restrictions */}
                          {rule.restrictions && (() => {
                            try {
                              const restrictions = JSON.parse(rule.restrictions);
                              const totalRestrictions = (restrictions.users?.length || 0) + (restrictions.teams?.length || 0);
                              if (totalRestrictions > 0) {
                                return (
                                  <div className="flex items-center space-x-1">
                                    <svg className="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                                    </svg>
                                    <span>Restricted ({totalRestrictions})</span>
                                  </div>
                                );
                              }
                              return null;
                            } catch (e) {
                              return null;
                            }
                          })()}
                        </div>
                        
                        <p className="text-xs text-muted-foreground mt-2">
                          Created {new Date(rule.created_at).toLocaleDateString()} • 
                          Updated {new Date(rule.updated_at).toLocaleDateString()}
                        </p>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => openEditRuleModal(rule)}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDeleteProtectionRule(rule.id)}
                          className="text-red-600 hover:bg-red-50"
                        >
                          Delete
                        </Button>
                      </div>
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          </div>
        )}

        {/* Branches Section */}
        <div className="mb-6">
          <h2 className="text-xl font-semibold text-foreground mb-4">
            Repository Branches
            {protectionRules.length === 0 && (
              <span className="text-sm text-muted-foreground font-normal ml-2">
                (Legacy branch-specific protection)
              </span>
            )}
          </h2>
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

        {/* Enhanced Protection Rule Modal */}
        <Modal 
          open={showRuleModal} 
          onClose={() => setShowRuleModal(false)}
          title={selectedRule ? 'Edit Protection Rule' : 'Create Protection Rule'}
          size="xl"
        >
          <div className="space-y-6 max-h-[80vh] overflow-y-auto">
            {/* Pattern Configuration */}
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-3">Branch Pattern</h3>
              <PatternInput
                value={protectionData.pattern || ''}
                onChange={(pattern) => setProtectionData({
                  ...protectionData,
                  pattern
                })}
                placeholder="Enter branch pattern (e.g., main, feature/*, *)"
                disabled={protectionLoading}
              />
            </div>

            {/* Status Checks */}
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-3">Status Checks</h3>
              <StatusChecksConfig
                strict={protectionData.required_status_checks?.strict || false}
                contexts={protectionData.required_status_checks?.contexts || []}
                onStrictChange={(strict) => setProtectionData({
                  ...protectionData,
                  required_status_checks: {
                    ...protectionData.required_status_checks!,
                    strict
                  }
                })}
                onContextsChange={(contexts) => setProtectionData({
                  ...protectionData,
                  required_status_checks: {
                    ...protectionData.required_status_checks!,
                    contexts
                  }
                })}
                disabled={protectionLoading}
              />
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
                    disabled={protectionLoading}
                    className="w-32 px-3 py-2 border border-input rounded-md bg-background text-foreground disabled:opacity-50"
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
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
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
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
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
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
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

            {/* Branch Restrictions */}
            <BranchRestrictionsConfig
              users={protectionData.restrictions?.users || []}
              teams={protectionData.restrictions?.teams || []}
              onUsersChange={(users) => setProtectionData({
                ...protectionData,
                restrictions: {
                  ...protectionData.restrictions!,
                  users
                }
              })}
              onTeamsChange={(teams) => setProtectionData({
                ...protectionData,
                restrictions: {
                  ...protectionData.restrictions!,
                  teams
                }
              })}
              owner={owner}
              repo={repo}
              disabled={protectionLoading}
            />

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
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Include administrators</span>
                    <p className="text-xs text-muted-foreground">
                      Enforce all configured restrictions for administrators
                    </p>
                    {protectionData.enforce_admins && (
                      <p className="text-xs text-orange-600 mt-1">
                        ⚠️ This will apply protection rules to repository administrators
                      </p>
                    )}
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
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Allow force pushes</span>
                    <p className="text-xs text-muted-foreground">
                      Permit force pushes to this branch
                    </p>
                    {protectionData.allow_force_pushes && (
                      <p className="text-xs text-orange-600 mt-1">
                        ⚠️ Force pushes can rewrite history and cause issues
                      </p>
                    )}
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
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Allow deletions</span>
                    <p className="text-xs text-muted-foreground">
                      Allow users with push access to delete this branch
                    </p>
                    {protectionData.allow_deletions && (
                      <p className="text-xs text-red-600 mt-1">
                        ⚠️ Branch deletions are permanent and cannot be undone
                      </p>
                    )}
                  </div>
                </label>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.require_linear_history || false}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      require_linear_history: e.target.checked
                    })}
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Require linear history</span>
                    <p className="text-xs text-muted-foreground">
                      Prevent merge commits from being pushed to this branch
                    </p>
                  </div>
                </label>

                <label className="flex items-center space-x-3">
                  <input
                    type="checkbox"
                    checked={protectionData.require_conversation_resolution || false}
                    onChange={(e) => setProtectionData({
                      ...protectionData,
                      require_conversation_resolution: e.target.checked
                    })}
                    disabled={protectionLoading}
                    className="rounded border-border disabled:opacity-50"
                  />
                  <div>
                    <span className="text-sm font-medium text-foreground">Require conversation resolution</span>
                    <p className="text-xs text-muted-foreground">
                      All conversations on code must be resolved before merging
                    </p>
                  </div>
                </label>
              </div>
            </div>

            {/* Preview Section */}
            {protectionData.pattern && (
              <div>
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-lg font-semibold text-foreground">Preview</h3>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowPreview(!showPreview)}
                  >
                    {showPreview ? 'Hide Preview' : 'Show Preview'}
                  </Button>
                </div>
                {showPreview && (
                  <ProtectionRulePreview
                    rule={{
                      pattern: protectionData.pattern,
                      required_status_checks: protectionData.required_status_checks,
                      required_pull_request_reviews: protectionData.required_pull_request_reviews,
                      enforce_admins: protectionData.enforce_admins,
                      restrictions: protectionData.restrictions,
                      allow_force_pushes: protectionData.allow_force_pushes,
                      allow_deletions: protectionData.allow_deletions
                    }}
                    testBranches={branches.map(b => b.name)}
                  />
                )}
              </div>
            )}

            <div className="flex justify-end space-x-3 pt-4 border-t">
              <Button variant="outline" onClick={() => setShowRuleModal(false)}>
                Cancel
              </Button>
              <Button 
                onClick={handleSaveProtectionRule} 
                disabled={protectionLoading || !protectionData.pattern}
              >
                {protectionLoading ? 'Saving...' : (selectedRule ? 'Update Rule' : 'Create Rule')}
              </Button>
            </div>
          </div>
        </Modal>

        {/* Legacy Branch Protection Modal */}
        <Modal 
          open={showProtectionModal} 
          onClose={() => setShowProtectionModal(false)}
          title={`Branch Protection: ${selectedBranch}`}
          size="lg"
        >
          <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-md mb-4">
            <div className="flex items-center">
              <svg className="w-5 h-5 text-yellow-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
              <p className="text-sm text-yellow-800">
                This creates a simple protection rule for the specific branch. 
                Use &quot;Add Protection Rule&quot; for advanced pattern-based rules.
              </p>
            </div>
          </div>
          <div className="text-center py-8">
            <Button 
              onClick={() => {
                setProtectionData({
                  ...protectionData,
                  pattern: selectedBranch
                });
                setShowProtectionModal(false);
                setShowRuleModal(true);
              }}
            >
              Configure Advanced Protection
            </Button>
          </div>
        </Modal>
      </div>
    </AppLayout>
  );
}