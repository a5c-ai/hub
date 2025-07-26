'use client';

import { useState } from 'react';
import { Input } from './Input';
import { Badge } from './Badge';

interface PatternInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  className?: string;
}

interface PatternSuggestion {
  pattern: string;
  description: string;
  example: string;
}

const PATTERN_SUGGESTIONS: PatternSuggestion[] = [
  {
    pattern: '*',
    description: 'All branches',
    example: 'Matches all branches'
  },
  {
    pattern: 'main',
    description: 'Exact match',
    example: 'Matches only "main"'
  },
  {
    pattern: 'feature/*',
    description: 'Feature branches',
    example: 'Matches "feature/login", "feature/auth"'
  },
  {
    pattern: 'release/*',
    description: 'Release branches',
    example: 'Matches "release/v1.0", "release/v2.0"'
  },
  {
    pattern: 'hotfix/*',
    description: 'Hotfix branches',
    example: 'Matches "hotfix/urgent-fix"'
  }
];

export function PatternInput({ value, onChange, placeholder = "Enter branch pattern...", disabled, className }: PatternInputProps) {
  const [showSuggestions, setShowSuggestions] = useState(false);

  const handlePatternSelect = (pattern: string) => {
    onChange(pattern);
    setShowSuggestions(false);
  };

  const validatePattern = (pattern: string): { isValid: boolean; message?: string } => {
    if (!pattern.trim()) {
      return { isValid: false, message: 'Pattern cannot be empty' };
    }
    
    // Basic validation - could be expanded
    if (pattern.includes('**')) {
      return { isValid: false, message: 'Double wildcards (**) are not supported' };
    }
    
    return { isValid: true };
  };

  const validation = validatePattern(value);

  return (
    <div className="space-y-2">
      <div className="relative">
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          disabled={disabled}
          className={className}
          onFocus={() => setShowSuggestions(true)}
        />
        
        {/* Pattern suggestions dropdown */}
        {showSuggestions && !disabled && (
          <div className="absolute z-10 w-full mt-1 bg-background border border-border rounded-md shadow-lg max-h-60 overflow-auto">
            <div className="p-2 border-b border-border">
              <p className="text-sm font-medium text-foreground">Pattern Examples</p>
            </div>
            {PATTERN_SUGGESTIONS.map((suggestion) => (
              <button
                key={suggestion.pattern}
                type="button"
                onClick={() => handlePatternSelect(suggestion.pattern)}
                className="w-full px-3 py-2 text-left hover:bg-muted/50 focus:bg-muted/50 focus:outline-none"
              >
                <div className="flex items-center justify-between">
                  <code className="text-sm font-mono bg-muted px-2 py-1 rounded text-foreground">
                    {suggestion.pattern}
                  </code>
                  <Badge variant="secondary" className="text-xs">
                    {suggestion.description}
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {suggestion.example}
                </p>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Validation message */}
      {!validation.isValid && (
        <p className="text-sm text-red-600">{validation.message}</p>
      )}

      {/* Pattern explanation */}
      {value && validation.isValid && (
        <div className="text-sm text-muted-foreground">
          <p>
            <strong>Pattern:</strong> <code className="bg-muted px-1 rounded">{value}</code>
          </p>
          {value === '*' && <p>This will match all branches in the repository.</p>}
          {value.endsWith('/*') && <p>This will match all branches starting with &quot;{value.slice(0, -2)}&quot;.</p>}
          {!value.includes('*') && <p>This will match only the exact branch name &quot;{value}&quot;.</p>}
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
  );
}