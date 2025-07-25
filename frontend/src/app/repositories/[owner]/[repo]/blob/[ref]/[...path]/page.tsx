'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useState, useEffect } from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import { repoApi } from '@/lib/api';
import api from '@/lib/api';
import { Repository, File } from '@/types';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import ReactMarkdown from 'react-markdown';
import { MobileCodeViewer } from '@/components/mobile/MobileCodeViewer';
import { useMobile } from '@/hooks/useDevice';

export default function BlobPage() {
  const params = useParams();
  const owner = params.owner as string;
  const repo = params.repo as string;
  const ref = params.ref as string;
  const pathArray = params.path as string[];
  const filePath = pathArray ? pathArray.join('/') : '';
  
  const [repository, setRepository] = useState<Repository | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isDarkMode, setIsDarkMode] = useState(false);
  const isMobile = useMobile();

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // Fetch repository info and file content in parallel
        const [repoResponse, fileResponse] = await Promise.all([
          api.get(`/repositories/${owner}/${repo}`),
          repoApi.getFile(owner, repo, filePath, ref)
        ]);
        
        setRepository(repoResponse.data);
        setFile(fileResponse.data);
      } catch (err: unknown) {
        const errorMessage = err instanceof Error && 'response' in err && 
          typeof err.response === 'object' && err.response && 
          'data' in err.response && 
          typeof err.response.data === 'object' && err.response.data &&
          'message' in err.response.data && 
          typeof err.response.data.message === 'string'
          ? err.response.data.message 
          : 'Failed to fetch file';
        setError(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [owner, repo, filePath, ref]);

  useEffect(() => {
    // Check for dark mode
    const checkDarkMode = () => {
      setIsDarkMode(document.documentElement.classList.contains('dark'));
    };
    
    checkDarkMode();
    
    // Watch for dark mode changes
    const observer = new MutationObserver(checkDarkMode);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });
    
    return () => observer.disconnect();
  }, []);

  const getLanguageFromFileName = (filename: string): string => {
    const ext = filename.split('.').pop()?.toLowerCase();
    const languageMap: Record<string, string> = {
      'js': 'javascript',
      'jsx': 'jsx',
      'ts': 'typescript',
      'tsx': 'tsx',
      'py': 'python',
      'rb': 'ruby',
      'go': 'go',
      'java': 'java',
      'c': 'c',
      'cpp': 'cpp',
      'cc': 'cpp',
      'cxx': 'cpp',
      'h': 'c',
      'hpp': 'cpp',
      'cs': 'csharp',
      'php': 'php',
      'html': 'html',
      'css': 'css',
      'scss': 'scss',
      'sass': 'sass',
      'less': 'less',
      'json': 'json',
      'xml': 'xml',
      'yaml': 'yaml',
      'yml': 'yaml',
      'md': 'markdown',
      'sh': 'bash',
      'bash': 'bash',
      'zsh': 'bash',
      'fish': 'bash',
      'ps1': 'powershell',
      'sql': 'sql',
      'r': 'r',
      'swift': 'swift',
      'kt': 'kotlin',
      'rs': 'rust',
      'dart': 'dart',
      'lua': 'lua',
      'vim': 'vim',
      'dockerfile': 'dockerfile',
      'makefile': 'makefile'
    };
    return languageMap[ext || ''] || 'text';
  };

  const downloadFile = () => {
    if (!file) return;
    
    const content = file.encoding === 'base64' 
      ? atob(file.content || '') 
      : file.content || '';
    
    const blob = new Blob([content], { type: 'application/octet-stream' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = file.name;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const viewRaw = () => {
    if (!file) return;
    
    const content = file.encoding === 'base64' 
      ? atob(file.content || '') 
      : file.content || '';
    
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
    URL.revokeObjectURL(url);
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
            <div className="h-64 bg-muted rounded"></div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error || !repository || !file) {
    return (
      <AppLayout>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <div className="text-red-600 text-lg mb-4">Error: {error || 'File not found'}</div>
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
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
          <span className="text-foreground font-medium">blob</span>
          <span>/</span>
          <span className="text-foreground font-medium">{ref}</span>
          <span>/</span>
          <span className="text-foreground font-medium">{filePath}</span>
        </nav>

        {/* Repository Header */}
        <div className="flex items-center space-x-4 mb-6">
          <Avatar
            src={repository.owner?.avatar_url}
            alt={repository.owner?.username || 'Repository owner'}
            size="md"
          />
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              {repository?.name || repo}
            </h1>
            <div className="flex items-center space-x-2">
              <Badge variant={repository.private ? 'secondary' : 'default'}>
                {repository.private ? 'Private' : 'Public'}
              </Badge>
            </div>
          </div>
        </div>

        {/* File Content */}
        {isMobile ? (
          <MobileCodeViewer
            content={file.encoding === 'base64' ? atob(file.content || '') : file.content || ''}
            filename={file.name}
            language={getLanguageFromFileName(file.name)}
            onCopy={() => {
              const content = file.encoding === 'base64' 
                ? atob(file.content || '') 
                : file.content || '';
              navigator.clipboard.writeText(content);
            }}
            onShare={() => {
              if (navigator.share) {
                navigator.share({
                  title: file.name,
                  text: file.encoding === 'base64' 
                    ? atob(file.content || '') 
                    : file.content || '',
                });
              }
            }}
          />
        ) : (
          <Card>
            <div className="border-b border-border px-6 py-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4">
                  <h2 className="text-lg font-semibold text-foreground">{file.name}</h2>
                  <span className="text-sm text-muted-foreground">{formatFileSize(file.size)}</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Link 
                    href={`/repositories/${owner}/${repo}/edit/${ref}/${filePath}`}
                    className="inline-flex"
                  >
                    <Button size="sm" variant="outline">
                      <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                      Edit
                    </Button>
                  </Link>
                  <Button size="sm" variant="outline" onClick={viewRaw}>
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                    Raw
                  </Button>
                  <Button size="sm" variant="outline" onClick={downloadFile}>
                    <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    Download
                  </Button>
                </div>
              </div>
            </div>
            <div className="p-6">
              {file.encoding === 'base64' ? (
                <div className="text-center py-8 text-muted-foreground">
                  <svg className="w-16 h-16 mx-auto mb-4 text-muted-foreground" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
                  </svg>
                  <h3 className="text-lg font-medium text-foreground mb-2">Binary file cannot be displayed</h3>
                  <p className="text-muted-foreground">Use the download button to save the file</p>
                </div>
              ) : file.name.endsWith('.md') ? (
                <div className="prose dark:prose-invert max-w-none">
                  <ReactMarkdown>{file.content}</ReactMarkdown>
                </div>
              ) : (
                <div className="relative">
                  <SyntaxHighlighter
                    language={getLanguageFromFileName(file.name)}
                    style={isDarkMode ? oneDark : oneLight}
                    showLineNumbers={true}
                    lineNumberStyle={{
                      minWidth: '3em',
                      paddingRight: '1em',
                      textAlign: 'right',
                      userSelect: 'none'
                    }}
                    customStyle={{
                      margin: 0,
                      borderRadius: '0.375rem',
                      fontSize: '0.875rem',
                      lineHeight: '1.25rem'
                    }}
                    wrapLines={true}
                    wrapLongLines={true}
                  >
                    {file.content || ''}
                  </SyntaxHighlighter>
                </div>
              )}
            </div>
          </Card>
        )}
      </div>
    </AppLayout>
  );
} 