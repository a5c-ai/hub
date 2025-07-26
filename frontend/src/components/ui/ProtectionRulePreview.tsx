'use client';

import { Card } from './Card';
import { Badge } from './Badge';

interface ProtectionRule {
  pattern: string;
  required_status_checks?: {
    strict: boolean;
    contexts: string[];
  };
  required_pull_request_reviews?: {
    required_approving_review_count: number;
    dismiss_stale_reviews: boolean;
    require_code_owner_reviews: boolean;
    require_last_push_approval: boolean;
  };
  enforce_admins: boolean;
  restrictions?: {
    users: string[];
    teams: string[];
  };
  allow_force_pushes?: boolean;
  allow_deletions?: boolean;
}

interface ProtectionRulePreviewProps {
  rule: ProtectionRule;
  testBranches?: string[];
  className?: string;
}

export function ProtectionRulePreview({ rule, testBranches = [], className }: ProtectionRulePreviewProps) {
  const matchesPattern = (pattern: string, branchName: string): boolean => {
    if (pattern === '*') return true;
    if (pattern === branchName) return true;
    if (pattern.endsWith('/*')) {
      const prefix = pattern.slice(0, -2);
      return branchName.startsWith(prefix + '/');
    }
    return false;
  };

  const getMatchingBranches = () => {
    return testBranches.filter(branch => matchesPattern(rule.pattern, branch));
  };

  const getNonMatchingBranches = () => {
    return testBranches.filter(branch => !matchesPattern(rule.pattern, branch));
  };

  return (
    <Card className={className}>
      <div className="p-4 space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-foreground">Protection Rule Preview</h3>
          <Badge variant="outline">
            Pattern: <code className="ml-1">{rule.pattern}</code>
          </Badge>
        </div>

        {/* Pattern matching test */}
        {testBranches.length > 0 && (
          <div className="space-y-3">
            <div>
              <h4 className="text-sm font-medium text-foreground mb-2">Branch Pattern Matching</h4>
              
              {getMatchingBranches().length > 0 && (
                <div className="mb-2">
                  <p className="text-xs text-muted-foreground mb-1">
                    Protected branches ({getMatchingBranches().length}):
                  </p>
                  <div className="flex flex-wrap gap-1">
                    {getMatchingBranches().map(branch => (
                      <Badge key={branch} variant="default" className="text-xs bg-green-100 text-green-800">
                        {branch}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              {getNonMatchingBranches().length > 0 && (
                <div>
                  <p className="text-xs text-muted-foreground mb-1">
                    Unprotected branches ({getNonMatchingBranches().length}):
                  </p>
                  <div className="flex flex-wrap gap-1">
                    {getNonMatchingBranches().map(branch => (
                      <Badge key={branch} variant="secondary" className="text-xs">
                        {branch}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Protection settings summary */}
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-foreground">Protection Settings</h4>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            {/* Status Checks */}
            <div className="space-y-2">
              <h5 className="font-medium text-foreground">Status Checks</h5>
              {rule.required_status_checks ? (
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    <span className="text-xs">Required status checks enabled</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {rule.required_status_checks.strict ? (
                      <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    )}
                    <span className="text-xs">Require up-to-date branches</span>
                  </div>
                  {rule.required_status_checks.contexts.length > 0 && (
                    <div className="mt-2">
                      <p className="text-xs text-muted-foreground mb-1">Required contexts:</p>
                      <div className="flex flex-wrap gap-1">
                        {rule.required_status_checks.contexts.map(context => (
                          <Badge key={context} variant="outline" className="text-xs">
                            {context}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <span className="text-xs text-muted-foreground">No status checks required</span>
                </div>
              )}
            </div>

            {/* Pull Request Reviews */}
            <div className="space-y-2">
              <h5 className="font-medium text-foreground">Pull Request Reviews</h5>
              {rule.required_pull_request_reviews ? (
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    <span className="text-xs">
                      {rule.required_pull_request_reviews.required_approving_review_count} approving review{rule.required_pull_request_reviews.required_approving_review_count !== 1 ? 's' : ''} required
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    {rule.required_pull_request_reviews.dismiss_stale_reviews ? (
                      <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    )}
                    <span className="text-xs">Dismiss stale reviews</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {rule.required_pull_request_reviews.require_code_owner_reviews ? (
                      <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    )}
                    <span className="text-xs">Code owner reviews required</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {rule.required_pull_request_reviews.require_last_push_approval ? (
                      <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    )}
                    <span className="text-xs">Require approval for latest push</span>
                  </div>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <span className="text-xs text-muted-foreground">No review requirements</span>
                </div>
              )}
            </div>

            {/* Administrative Settings */}
            <div className="space-y-2">
              <h5 className="font-medium text-foreground">Administrative</h5>
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  {rule.enforce_admins ? (
                    <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  )}
                  <span className="text-xs">Include administrators</span>
                </div>
                <div className="flex items-center gap-2">
                  {rule.allow_force_pushes ? (
                    <svg className="w-4 h-4 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.728-.833-2.498 0L4.316 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                  ) : (
                    <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  )}
                  <span className="text-xs">Allow force pushes</span>
                </div>
                <div className="flex items-center gap-2">
                  {rule.allow_deletions ? (
                    <svg className="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.728-.833-2.498 0L4.316 16.5c-.77.833.192 2.5 1.732 2.5z" />
                    </svg>
                  ) : (
                    <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  )}
                  <span className="text-xs">Allow deletions</span>
                </div>
              </div>
            </div>

            {/* Branch Restrictions */}
            <div className="space-y-2">
              <h5 className="font-medium text-foreground">Push Restrictions</h5>
              {rule.restrictions && (rule.restrictions.users.length > 0 || rule.restrictions.teams.length > 0) ? (
                <div className="space-y-2">
                  {rule.restrictions.users.length > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Restricted to users:</p>
                      <div className="flex flex-wrap gap-1">
                        {rule.restrictions.users.map(user => (
                          <Badge key={user} variant="outline" className="text-xs">
                            {user}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                  {rule.restrictions.teams.length > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Restricted to teams:</p>
                      <div className="flex flex-wrap gap-1">
                        {rule.restrictions.teams.map(team => (
                          <Badge key={team} variant="outline" className="text-xs">
                            {team}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <span className="text-xs text-muted-foreground">No push restrictions</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}