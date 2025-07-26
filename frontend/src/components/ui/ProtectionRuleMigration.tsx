'use client';

import { useState } from 'react';
import { Button } from './Button';
import { Card } from './Card';
import { Badge } from './Badge';
import { Modal } from './Modal';

interface Branch {
  name: string;
  protected: boolean;
}

interface MigrationPlan {
  createRules: Array<{
    pattern: string;
    branches: string[];
    description: string;
  }>;
  conflicts: Array<{
    branch: string;
    issue: string;
  }>;
}

interface ProtectionRuleMigrationProps {
  branches: Branch[];
  onMigrate: (rules: Array<{ pattern: string; branches: string[] }>) => Promise<void>;
  className?: string;
}

export function ProtectionRuleMigration({ branches, onMigrate, className }: ProtectionRuleMigrationProps) {
  const [showModal, setShowModal] = useState(false);
  const [migrationPlan, setMigrationPlan] = useState<MigrationPlan | null>(null);
  const [isMigrating, setIsMigrating] = useState(false);

  const protectedBranches = branches.filter(b => b.protected);

  const generateMigrationPlan = (): MigrationPlan => {
    const rules: MigrationPlan['createRules'] = [];
    const conflicts: MigrationPlan['conflicts'] = [];

    // Group branches by common patterns
    const branchGroups = new Map<string, string[]>();
    
    // Check for feature/ pattern
    const featureBranches = protectedBranches.filter(b => b.name.startsWith('feature/'));
    if (featureBranches.length > 1) {
      branchGroups.set('feature/*', featureBranches.map(b => b.name));
    }

    // Check for release/ pattern
    const releaseBranches = protectedBranches.filter(b => b.name.startsWith('release/'));
    if (releaseBranches.length > 1) {
      branchGroups.set('release/*', releaseBranches.map(b => b.name));
    }

    // Check for hotfix/ pattern
    const hotfixBranches = protectedBranches.filter(b => b.name.startsWith('hotfix/'));
    if (hotfixBranches.length > 1) {
      branchGroups.set('hotfix/*', hotfixBranches.map(b => b.name));
    }

    // Add grouped rules
    branchGroups.forEach((branches, pattern) => {
      rules.push({
        pattern,
        branches,
        description: `Migrate ${branches.length} branches matching ${pattern} pattern`
      });
    });

    // Add individual branch rules for non-grouped branches
    const groupedBranchNames = new Set([...branchGroups.values()].flat());
    const individualBranches = protectedBranches.filter(b => !groupedBranchNames.has(b.name));
    
    individualBranches.forEach(branch => {
      rules.push({
        pattern: branch.name,
        branches: [branch.name],
        description: `Migrate exact match rule for ${branch.name}`
      });
    });

    // Check for potential conflicts
    if (rules.length === 0 && protectedBranches.length > 0) {
      conflicts.push({
        branch: 'Multiple branches',
        issue: 'No clear pattern detected for migration'
      });
    }

    return { createRules: rules, conflicts };
  };

  const handleStartMigration = () => {
    const plan = generateMigrationPlan();
    setMigrationPlan(plan);
    setShowModal(true);
  };

  const handleExecuteMigration = async () => {
    if (!migrationPlan) return;

    setIsMigrating(true);
    try {
      await onMigrate(migrationPlan.createRules);
      setShowModal(false);
      setMigrationPlan(null);
    } catch (error) {
      console.error('Migration failed:', error);
    } finally {
      setIsMigrating(false);
    }
  };

  if (protectedBranches.length === 0) {
    return null;
  }

  return (
    <>
      <Card className={className}>
        <div className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-foreground mb-2">
                Legacy Branch Protection Detected
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                You have {protectedBranches.length} branches with legacy protection rules. 
                Migrate them to the new pattern-based system for better management.
              </p>
              <div className="flex flex-wrap gap-2 mb-4">
                {protectedBranches.slice(0, 5).map(branch => (
                  <Badge key={branch.name} variant="secondary" className="text-xs">
                    {branch.name}
                  </Badge>
                ))}
                {protectedBranches.length > 5 && (
                  <Badge variant="outline" className="text-xs">
                    +{protectedBranches.length - 5} more
                  </Badge>
                )}
              </div>
            </div>
            <Button onClick={handleStartMigration}>
              Plan Migration
            </Button>
          </div>
        </div>
      </Card>

      <Modal
        open={showModal}
        onClose={() => setShowModal(false)}
        title="Branch Protection Migration Plan"
        size="lg"
      >
        <div className="space-y-6">
          {migrationPlan && (
            <>
              <div>
                <h3 className="text-lg font-semibold text-foreground mb-3">Migration Summary</h3>
                <div className="bg-blue-50 border border-blue-200 rounded-md p-4 mb-4">
                  <p className="text-sm text-blue-800">
                    This migration will create {migrationPlan.createRules.length} new protection rule{migrationPlan.createRules.length !== 1 ? 's' : ''} 
                    and migrate {protectedBranches.length} protected branch{protectedBranches.length !== 1 ? 'es' : ''}.
                  </p>
                </div>
              </div>

              {/* Proposed Rules */}
              <div>
                <h4 className="text-md font-medium text-foreground mb-3">Proposed Protection Rules</h4>
                <div className="space-y-3">
                  {migrationPlan.createRules.map((rule, index) => (
                    <div key={index} className="border border-border rounded-md p-4">
                      <div className="flex items-center justify-between mb-2">
                        <code className="text-sm bg-muted px-2 py-1 rounded">{rule.pattern}</code>
                        <Badge variant="outline" className="text-xs">
                          {rule.branches.length} branch{rule.branches.length !== 1 ? 'es' : ''}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mb-2">{rule.description}</p>
                      <div className="flex flex-wrap gap-1">
                        {rule.branches.map(branch => (
                          <Badge key={branch} variant="secondary" className="text-xs">
                            {branch}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Conflicts */}
              {migrationPlan.conflicts.length > 0 && (
                <div>
                  <h4 className="text-md font-medium text-foreground mb-3">Potential Issues</h4>
                  <div className="space-y-2">
                    {migrationPlan.conflicts.map((conflict, index) => (
                      <div key={index} className="bg-yellow-50 border border-yellow-200 rounded-md p-3">
                        <div className="flex items-center">
                          <svg className="w-4 h-4 text-yellow-600 mr-2" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                          </svg>
                          <div>
                            <p className="text-sm font-medium text-yellow-800">{conflict.branch}</p>
                            <p className="text-xs text-yellow-700">{conflict.issue}</p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Migration Notes */}
              <div className="bg-gray-50 border border-gray-200 rounded-md p-4">
                <h4 className="text-sm font-semibold text-foreground mb-2">Migration Notes</h4>
                <ul className="text-xs text-muted-foreground space-y-1">
                  <li>• Existing protection settings will be preserved</li>
                  <li>• Legacy branch-specific rules will remain until manually removed</li>
                  <li>• New pattern-based rules take precedence over legacy rules</li>
                  <li>• You can modify the generated rules after migration</li>
                </ul>
              </div>

              <div className="flex justify-end space-x-3 pt-4 border-t">
                <Button variant="outline" onClick={() => setShowModal(false)}>
                  Cancel
                </Button>
                <Button 
                  onClick={handleExecuteMigration}
                  disabled={isMigrating || migrationPlan.createRules.length === 0}
                >
                  {isMigrating ? 'Migrating...' : 'Execute Migration'}
                </Button>
              </div>
            </>
          )}
        </div>
      </Modal>
    </>
  );
}