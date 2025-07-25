'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';

interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
  source: 'kubernetes' | 'runner' | 'system';
  step_id?: string;
}

interface LogStreamingComponentProps {
  jobId: string;
  owner: string;
  repo: string;
  onError?: (error: string) => void;
}

export default function LogStreamingComponent({ 
  jobId, 
  owner, 
  repo, 
  onError 
}: LogStreamingComponentProps) {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isStreaming, setIsStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);

  const scrollToBottom = useCallback(() => {
    if (autoScroll && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [autoScroll]);

  const connectSSE = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const eventSource = new EventSource(
      `/api/v1/repos/${owner}/${repo}/actions/jobs/${jobId}/logs/stream`
    );

    eventSource.onopen = () => {
      setIsConnected(true);
      setIsStreaming(true);
      setError(null);
    };

    eventSource.onmessage = (event) => {
      try {
        const logEntry: LogEntry = JSON.parse(event.data);
        setLogs(prevLogs => [...prevLogs, logEntry]);
      } catch (err) {
        console.error('Failed to parse log entry:', err);
      }
    };

    eventSource.onerror = (event) => {
      setIsConnected(false);
      setIsStreaming(false);
      const errorMsg = 'Failed to connect to log stream';
      setError(errorMsg);
      onError?.(errorMsg);
      
      // Auto-reconnect after 5 seconds
      setTimeout(() => {
        if (!eventSourceRef.current || eventSourceRef.current.readyState === EventSource.CLOSED) {
          connectSSE();
        }
      }, 5000);
    };

    eventSourceRef.current = eventSource;
  }, [jobId, owner, repo, onError]);

  const disconnectSSE = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsConnected(false);
    setIsStreaming(false);
  }, []);

  const fetchHistoricalLogs = useCallback(async () => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/jobs/${jobId}/logs`);
      if (response.ok) {
        const data = await response.json();
        setLogs(data.logs || []);
      }
    } catch (err) {
      console.error('Failed to fetch historical logs:', err);
    }
  }, [jobId, owner, repo]);

  useEffect(() => {
    // Fetch historical logs first
    fetchHistoricalLogs();
    
    // Then connect to streaming
    connectSSE();

    return () => {
      disconnectSSE();
    };
  }, [fetchHistoricalLogs, connectSSE, disconnectSSE]);

  useEffect(() => {
    scrollToBottom();
  }, [logs, scrollToBottom]);

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;
    const isAtBottom = scrollHeight - scrollTop === clientHeight;
    setAutoScroll(isAtBottom);
  };

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error': return 'text-red-600';
      case 'warn': return 'text-yellow-600';
      case 'debug': return 'text-gray-500';
      default: return 'text-gray-300';
    }
  };

  const getSourceIcon = (source: string) => {
    switch (source) {
      case 'kubernetes': return '☸️';
      case 'runner': return '🏃';
      case 'system': return '⚙️';
      default: return '📝';
    }
  };

  const clearLogs = () => {
    setLogs([]);
  };

  const downloadLogs = () => {
    const logText = logs.map(log => 
      `[${log.timestamp}] [${log.level.toUpperCase()}] [${log.source}] ${log.message}`
    ).join('\n');
    
    const blob = new Blob([logText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `job-${jobId}-logs.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <Card className="h-96 flex flex-col">
      {/* Header */}
      <div className="p-4 border-b flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h4 className="font-medium">Live Logs</h4>
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
            <span className="text-sm text-muted-foreground">
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
            {isStreaming && (
              <Badge variant="outline" className="text-xs">
                Streaming
              </Badge>
            )}
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={autoScroll ? () => setAutoScroll(false) : () => setAutoScroll(true)}
          >
            {autoScroll ? '📌' : '📌'} Auto-scroll
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={clearLogs}
          >
            Clear
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={downloadLogs}
            disabled={logs.length === 0}
          >
            Download
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={isConnected ? disconnectSSE : connectSSE}
          >
            {isConnected ? 'Disconnect' : 'Connect'}
          </Button>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="bg-red-50 border-b border-red-200 p-2">
          <p className="text-sm text-red-700">{error}</p>
        </div>
      )}

      {/* Logs Container */}
      <div 
        className="flex-1 overflow-y-auto bg-gray-900 text-gray-100 p-4 font-mono text-sm"
        onScroll={handleScroll}
      >
        {logs.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            <p>No logs available</p>
            <p className="text-xs mt-1">
              {isConnected ? 'Waiting for logs...' : 'Connect to start streaming'}
            </p>
          </div>
        ) : (
          <div className="space-y-1">
            {logs.map((log, index) => (
              <div key={index} className="flex items-start gap-2 hover:bg-gray-800 px-2 py-1 rounded">
                <span className="text-gray-500 text-xs shrink-0 w-20">
                  {new Date(log.timestamp).toLocaleTimeString()}
                </span>
                <span className="text-xs shrink-0">
                  {getSourceIcon(log.source)}
                </span>
                <span className={`text-xs shrink-0 w-12 ${getLevelColor(log.level)}`}>
                  {log.level.toUpperCase()}
                </span>
                <span className="text-gray-300 break-all">
                  {log.message}
                </span>
              </div>
            ))}
            <div ref={logsEndRef} />
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-2 border-t bg-gray-50 text-xs text-muted-foreground flex justify-between items-center">
        <span>{logs.length} log entries</span>
        {isStreaming && (
          <span className="flex items-center gap-1">
            <div className="w-1 h-1 bg-green-500 rounded-full animate-pulse"></div>
            Live streaming active
          </span>
        )}
      </div>
    </Card>
  );
}