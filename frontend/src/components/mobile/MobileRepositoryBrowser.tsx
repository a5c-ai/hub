'use client';

import { useState } from 'react';
import Link from 'next/link';
import {
  FolderIcon,
  DocumentTextIcon,
  CodeBracketIcon,
  ChevronRightIcon,
  ArrowPathIcon,
  EllipsisVerticalIcon,
} from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/utils';

interface FileItem {
  name: string;
  type: 'file' | 'directory';
  size?: string;
  lastModified?: string;
  url: string;
}

interface MobileRepositoryBrowserProps {
  files: FileItem[];
  currentPath: string;
  repositoryUrl: string;
  onRefresh?: () => void;
  loading?: boolean;
}

export function MobileRepositoryBrowser({
  files,
  currentPath,
  repositoryUrl,
  onRefresh,
  loading = false,
}: MobileRepositoryBrowserProps) {
  const [searchQuery, setSearchQuery] = useState('');

  const filteredFiles = files.filter(file =>
    file.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getFileIcon = (file: FileItem) => {
    if (file.type === 'directory') {
      return <FolderIcon className="h-5 w-5 text-blue-500" />;
    }
    
    const extension = file.name.split('.').pop()?.toLowerCase();
    
    switch (extension) {
      case 'js':
      case 'ts':
      case 'jsx':
      case 'tsx':
      case 'json':
        return <CodeBracketIcon className="h-5 w-5 text-yellow-500" />;
      case 'md':
      case 'txt':
      case 'readme':
        return <DocumentTextIcon className="h-5 w-5 text-green-500" />;
      default:
        return <DocumentTextIcon className="h-5 w-5 text-muted-foreground" />;
    }
  };

  const formatFileSize = (size: string | undefined) => {
    if (!size) return '';
    const bytes = parseInt(size);
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const formatDate = (date: string | undefined) => {
    if (!date) return '';
    return new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Header */}
      <div className="sticky top-0 z-10 bg-background border-b border-border">
        <div className="p-4 space-y-3">
          {/* Path breadcrumb */}
          <div className="flex items-center space-x-1 text-sm text-muted-foreground">
            <Link href={repositoryUrl} className="hover:text-foreground">
              Repository
            </Link>
            {currentPath.split('/').filter(Boolean).map((segment, index, array) => (
              <span key={index} className="flex items-center space-x-1">
                <ChevronRightIcon className="h-3 w-3" />
                <Link
                  href={`${repositoryUrl}/tree/${array.slice(0, index + 1).join('/')}`}
                  className="hover:text-foreground"
                >
                  {segment}
                </Link>
              </span>
            ))}
          </div>

          {/* Search and actions */}
          <div className="flex items-center space-x-2">
            <div className="flex-1">
              <Input
                type="search"
                placeholder="Search files..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full"
              />
            </div>
            <Button
              variant="ghost"
              size="icon"
              onClick={onRefresh}
              disabled={loading}
              className="flex-shrink-0"
            >
              <ArrowPathIcon className={cn("h-4 w-4", loading && "animate-spin")} />
            </Button>
          </div>
        </div>
      </div>

      {/* File list */}
      <div className="flex-1 overflow-y-auto">
        {filteredFiles.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-center p-4">
            <FolderIcon className="h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium text-foreground mb-2">
              {searchQuery ? 'No files found' : 'Empty directory'}
            </h3>
            <p className="text-muted-foreground">
              {searchQuery
                ? `No files match "${searchQuery}"`
                : 'This directory is empty or has no files to display.'}
            </p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {filteredFiles.map((file, index) => (
              <Link
                key={index}
                href={file.url}
                className="flex items-center justify-between p-4 hover:bg-muted/50 active:bg-muted transition-colors"
              >
                <div className="flex items-center space-x-3 flex-1 min-w-0">
                  {getFileIcon(file)}
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium text-foreground truncate">
                      {file.name}
                    </div>
                    {file.type === 'file' && (
                      <div className="flex items-center space-x-2 text-xs text-muted-foreground mt-1">
                        {file.size && (
                          <span>{formatFileSize(file.size)}</span>
                        )}
                        {file.lastModified && (
                          <span>{formatDate(file.lastModified)}</span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="flex items-center space-x-2 flex-shrink-0">
                  {file.type === 'directory' && (
                    <ChevronRightIcon className="h-4 w-4 text-muted-foreground" />
                  )}
                  <Button variant="ghost" size="icon" className="h-8 w-8">
                    <EllipsisVerticalIcon className="h-4 w-4" />
                  </Button>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}