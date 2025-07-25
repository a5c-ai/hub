'use client';

import { useState, useRef, useEffect } from 'react';
import {
  DocumentDuplicateIcon,
  ShareIcon,
  MagnifyingGlassIcon,
  ChevronUpIcon,
  ChevronDownIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/utils';

interface MobileCodeViewerProps {
  content: string;
  filename: string;
  language?: string;
  lineNumbers?: boolean;
  onCopy?: () => void;
  onShare?: () => void;
}

export function MobileCodeViewer({
  content,
  filename,
  language = 'text',
  lineNumbers = true,
  onCopy,
  onShare,
}: MobileCodeViewerProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchVisible, setSearchVisible] = useState(false);
  const [currentMatch, setCurrentMatch] = useState(0);
  const [totalMatches, setTotalMatches] = useState(0);
  const [fontSize, setFontSize] = useState(14);
  const codeRef = useRef<HTMLDivElement>(null);

  const lines = content.split('\n');

  // Search functionality
  useEffect(() => {
    if (searchQuery) {
      const matches = lines.reduce((acc, line) => {
        const lineMatches = (line.toLowerCase().match(new RegExp(searchQuery.toLowerCase(), 'g')) || []).length;
        return acc + lineMatches;
      }, 0);
      setTotalMatches(matches);
      setCurrentMatch(matches > 0 ? 1 : 0);
    } else {
      setTotalMatches(0);
      setCurrentMatch(0);
    }
  }, [searchQuery, lines]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(content);
      onCopy?.();
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  };

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: filename,
          text: content,
        });
      } catch (error) {
        console.error('Failed to share:', error);
      }
    } else {
      onShare?.();
    }
  };

  const highlightSearchTerm = (text: string) => {
    if (!searchQuery) return text;
    
    const regex = new RegExp(`(${searchQuery})`, 'gi');
    return text.replace(regex, '<mark class="bg-yellow-300 text-black">$1</mark>');
  };

  const goToNextMatch = () => {
    if (totalMatches > 0) {
      setCurrentMatch(currentMatch < totalMatches ? currentMatch + 1 : 1);
    }
  };

  const goToPrevMatch = () => {
    if (totalMatches > 0) {
      setCurrentMatch(currentMatch > 1 ? currentMatch - 1 : totalMatches);
    }
  };

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Header */}
      <div className="sticky top-0 z-10 bg-background border-b border-border">
        <div className="flex items-center justify-between p-4">
          <div className="flex-1 min-w-0">
            <h2 className="text-lg font-medium text-foreground truncate">
              {filename}
            </h2>
            <p className="text-sm text-muted-foreground">
              {language} • {lines.length} lines
            </p>
          </div>
          
          <div className="flex items-center space-x-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setSearchVisible(!searchVisible)}
              className={cn(searchVisible && 'bg-muted')}
            >
              <MagnifyingGlassIcon className="h-4 w-4" />
            </Button>
            
            <Button variant="ghost" size="icon" onClick={handleCopy}>
              <DocumentDuplicateIcon className="h-4 w-4" />
            </Button>
            
            <Button variant="ghost" size="icon" onClick={handleShare}>
              <ShareIcon className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Search bar */}
        {searchVisible && (
          <div className="border-t border-border p-4 space-y-3">
            <div className="flex items-center space-x-2">
              <div className="flex-1">
                <Input
                  type="search"
                  placeholder="Search in file..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full"
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSearchVisible(false)}
              >
                <XMarkIcon className="h-4 w-4" />
              </Button>
            </div>
            
            {totalMatches > 0 && (
              <div className="flex items-center justify-between text-sm text-muted-foreground">
                <span>{currentMatch} of {totalMatches} matches</span>
                <div className="flex items-center space-x-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={goToPrevMatch}
                    className="h-8 w-8"
                  >
                    <ChevronUpIcon className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={goToNextMatch}
                    className="h-8 w-8"
                  >
                    <ChevronDownIcon className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Font size controls */}
        <div className="border-t border-border p-2">
          <div className="flex items-center justify-center space-x-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setFontSize(Math.max(10, fontSize - 1))}
              disabled={fontSize <= 10}
            >
              A-
            </Button>
            <span className="text-sm text-muted-foreground">
              {fontSize}px
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setFontSize(Math.min(24, fontSize + 1))}
              disabled={fontSize >= 24}
            >
              A+
            </Button>
          </div>
        </div>
      </div>

      {/* Code content */}
      <div className="flex-1 overflow-auto">
        <div
          ref={codeRef}
          className="p-4"
          style={{ fontSize: `${fontSize}px` }}
        >
          <pre className="text-sm">
            <code className="block">
              {lines.map((line, index) => (
                <div
                  key={index}
                  className="flex min-h-[1.25rem] hover:bg-muted/50"
                >
                  {lineNumbers && (
                    <span className="inline-block w-12 text-right text-muted-foreground pr-4 select-none">
                      {index + 1}
                    </span>
                  )}
                  <span
                    className="flex-1 break-all whitespace-pre-wrap"
                    dangerouslySetInnerHTML={{
                      __html: highlightSearchTerm(line || ' '),
                    }}
                  />
                </div>
              ))}
            </code>
          </pre>
        </div>
      </div>
    </div>
  );
}