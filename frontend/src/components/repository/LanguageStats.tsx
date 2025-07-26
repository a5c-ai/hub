'use client';

import React from 'react';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { RepositoryLanguage } from '@/types';

interface LanguageStatsProps {
  languages: RepositoryLanguage[];
  primaryLanguage?: string;
  showPercentages?: boolean;
  showBytes?: boolean;
  compact?: boolean;
}

const LanguageStats: React.FC<LanguageStatsProps> = ({
  languages,
  primaryLanguage,
  showPercentages = true,
  showBytes = false,
  compact = false,
}) => {
  // Language color mapping (commonly used colors for popular languages)
  const getLanguageColor = (language: string): string => {
    const colors: Record<string, string> = {
      'TypeScript': '#3178c6',
      'JavaScript': '#f1e05a',
      'Python': '#3572A5',
      'Java': '#b07219',
      'Go': '#00ADD8',
      'Rust': '#dea584',
      'C': '#555555',
      'C++': '#f34b7d',
      'C#': '#239120',
      'PHP': '#4F5D95',
      'Ruby': '#701516',
      'Swift': '#fa7343',
      'Kotlin': '#A97BFF',
      'Dart': '#00B4AB',
      'HTML': '#e34c26',
      'CSS': '#1572B6',
      'SCSS': '#c6538c',
      'Shell': '#89e051',
      'Dockerfile': '#384d54',
      'YAML': '#cb171e',
      'JSON': '#292929',
      'Markdown': '#083fa1',
      'SQL': '#4169e1',
    };
    return colors[language] || '#6b7280';
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (!languages || languages.length === 0) {
    return (
      <Card className="p-4">
        <div className="text-center text-muted-foreground">
          <div className="text-sm">No language data available</div>
        </div>
      </Card>
    );
  }

  // Sort languages by percentage (descending)
  const sortedLanguages = [...languages].sort((a, b) => b.percentage - a.percentage);

  if (compact) {
    return (
      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          {primaryLanguage && (
            <div className="flex items-center space-x-1">
              <div 
                className="w-3 h-3 rounded-full" 
                style={{ backgroundColor: getLanguageColor(primaryLanguage) }}
              ></div>
              <span className="text-sm font-medium">{primaryLanguage}</span>
            </div>
          )}
          {sortedLanguages.length > 1 && (
            <span className="text-xs text-muted-foreground">
              +{sortedLanguages.length - 1} more
            </span>
          )}
        </div>
        
        {/* Language bar */}
        <div className="w-full bg-muted rounded-full h-2 overflow-hidden">
          <div className="flex h-full">
            {sortedLanguages.map((lang) => (
              <div
                key={lang.id}
                className="h-full"
                style={{
                  width: `${lang.percentage}%`,
                  backgroundColor: getLanguageColor(lang.language),
                }}
                title={`${lang.language}: ${lang.percentage.toFixed(1)}%`}
              />
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <Card>
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-foreground">Languages</h3>
          {primaryLanguage && (
            <Badge variant="secondary">
              Primary: {primaryLanguage}
            </Badge>
          )}
        </div>
        
        <div className="space-y-3">
          {sortedLanguages.map((lang) => (
            <div key={lang.id} className="flex items-center justify-between">
              <div className="flex items-center space-x-3 flex-1">
                <div 
                  className="w-3 h-3 rounded-full flex-shrink-0" 
                  style={{ backgroundColor: getLanguageColor(lang.language) }}
                ></div>
                <span className="font-medium text-foreground text-sm">
                  {lang.language}
                </span>
                <div className="flex-1 mx-3">
                  <div className="bg-muted rounded-full h-2">
                    <div
                      className="h-2 rounded-full"
                      style={{
                        width: `${lang.percentage}%`,
                        backgroundColor: getLanguageColor(lang.language),
                      }}
                    ></div>
                  </div>
                </div>
              </div>
              
              <div className="flex items-center space-x-3 text-sm text-muted-foreground">
                {showPercentages && (
                  <span className="w-12 text-right">
                    {lang.percentage.toFixed(1)}%
                  </span>
                )}
                {showBytes && (
                  <span className="w-16 text-right">
                    {formatBytes(lang.bytes)}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>

        {/* Summary */}
        <div className="mt-4 pt-4 border-t border-border">
          <div className="text-xs text-muted-foreground">
            {sortedLanguages.length} language{sortedLanguages.length !== 1 ? 's' : ''} detected
            {showBytes && (
              <span> • {formatBytes(sortedLanguages.reduce((sum, lang) => sum + lang.bytes, 0))} total</span>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
};

export default LanguageStats;