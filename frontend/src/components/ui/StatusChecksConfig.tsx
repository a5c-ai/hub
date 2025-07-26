'use client';

import { useState } from 'react';
import { Input } from './Input';
import { Button } from './Button';
import { Badge } from './Badge';

interface StatusChecksConfigProps {
  strict: boolean;
  contexts: string[];
  onStrictChange: (strict: boolean) => void;
  onContextsChange: (contexts: string[]) => void;
  disabled?: boolean;
}

const COMMON_CONTEXTS = [
  'ci/build',
  'ci/test',
  'ci/lint',
  'ci/security-scan',
  'continuous-integration',
  'codecov/project',
  'codecov/patch',
  'sonarcloud',
  'renovate',
  'dependency-check'
];

export function StatusChecksConfig({ 
  strict, 
  contexts, 
  onStrictChange, 
  onContextsChange, 
  disabled 
}: StatusChecksConfigProps) {
  const [newContext, setNewContext] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);

  const addContext = (context: string) => {
    const trimmedContext = context.trim();
    if (trimmedContext && !contexts.includes(trimmedContext)) {
      onContextsChange([...contexts, trimmedContext]);
    }
    setNewContext('');
    setShowSuggestions(false);
  };

  const removeContext = (contextToRemove: string) => {
    onContextsChange(contexts.filter(context => context !== contextToRemove));
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      addContext(newContext);
    }
  };

  const filteredSuggestions = COMMON_CONTEXTS.filter(
    context => 
      !contexts.includes(context) && 
      context.toLowerCase().includes(newContext.toLowerCase())
  );

  return (
    <div className="space-y-4">
      {/* Strict mode toggle */}
      <label className="flex items-center space-x-3">
        <input
          type="checkbox"
          checked={strict}
          onChange={(e) => onStrictChange(e.target.checked)}
          disabled={disabled}
          className="rounded border-border"
        />
        <div>
          <span className="text-sm font-medium text-foreground">
            Require branches to be up to date before merging
          </span>
          <p className="text-xs text-muted-foreground">
            Ensure pull request branch is up to date with the base branch before merging
          </p>
        </div>
      </label>

      {/* Status check contexts */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium text-foreground">
            Required status checks
          </label>
          <span className="text-xs text-muted-foreground">
            {contexts.length} context{contexts.length !== 1 ? 's' : ''}
          </span>
        </div>

        {/* Existing contexts */}
        {contexts.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {contexts.map((context) => (
              <Badge
                key={context}
                variant="secondary"
                className="flex items-center gap-2 pr-1"
              >
                <code className="text-xs">{context}</code>
                {!disabled && (
                  <button
                    type="button"
                    onClick={() => removeContext(context)}
                    className="text-muted-foreground hover:text-foreground"
                  >
                    <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                )}
              </Badge>
            ))}
          </div>
        )}

        {/* Add new context */}
        {!disabled && (
          <div className="relative">
            <div className="flex gap-2">
              <Input
                value={newContext}
                onChange={(e) => setNewContext(e.target.value)}
                onKeyPress={handleKeyPress}
                onFocus={() => setShowSuggestions(true)}
                placeholder="Add status check context (e.g., ci/build)"
                className="flex-1"
              />
              <Button
                type="button"
                onClick={() => addContext(newContext)}
                disabled={!newContext.trim() || contexts.includes(newContext.trim())}
                size="sm"
              >
                Add
              </Button>
            </div>

            {/* Suggestions dropdown */}
            {showSuggestions && filteredSuggestions.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-background border border-border rounded-md shadow-lg max-h-48 overflow-auto">
                <div className="p-2 border-b border-border">
                  <p className="text-xs font-medium text-foreground">Common contexts</p>
                </div>
                {filteredSuggestions.map((context) => (
                  <button
                    key={context}
                    type="button"
                    onClick={() => addContext(context)}
                    className="w-full px-3 py-2 text-left hover:bg-muted/50 focus:bg-muted/50 focus:outline-none"
                  >
                    <code className="text-sm">{context}</code>
                  </button>
                ))}
              </div>
            )}

            {/* Click outside handler */}
            {showSuggestions && (
              <div 
                className="fixed inset-0 z-0" 
                onClick={() => setShowSuggestions(false)}
              />
            )}
          </div>
        )}

        {/* Help text */}
        <p className="text-xs text-muted-foreground">
          Status checks are external services that report the state of a commit. 
          Common examples include CI builds, tests, and security scans.
        </p>
      </div>
    </div>
  );
}