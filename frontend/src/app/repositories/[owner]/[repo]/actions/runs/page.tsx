'use client';

import { useState, useEffect } from 'react';
import { useParams, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Input } from '@/components/ui/Input';

interface WorkflowRun {
  id: string;
  number: number;
  status: 'queued' | 'in_progress' | 'completed' | 'cancelled';
  conclusion?: 'success' | 'failure' | 'cancelled' | 'skipped';
  head_sha: string;
  head_branch?: string;
  event: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  actor?: {
    id: string;
    login: string;
    avatar_url: string;
  };
  workflow: {
    id: string;
    name: string;
  };
}

export default function WorkflowRunsPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const { owner, repo } = params;
  
  const [runs, setRuns] = useState<WorkflowRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [eventFilter, setEventFilter] = useState('all');

  useEffect(() => {
    fetchRuns();
  }, [owner, repo, statusFilter, eventFilter]);

  const fetchRuns = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      params.append('limit', '50');
      params.append('offset', '0');
      
      if (statusFilter !== 'all') {
        params.append('status', statusFilter);
      }
      
      if (eventFilter !== 'all') {
        params.append('event', eventFilter);
      }

      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/runs?${params}`);
      if (response.ok) {
        const data = await response.json();
        setRuns(data.workflow_runs || []);
      } else {
        throw new Error('Failed to fetch workflow runs');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string, conclusion?: string): 'default' | 'secondary' | 'destructive' | 'success' | 'warning' | 'outline' => {
    if (status === 'in_progress') return 'warning';
    if (status === 'queued') return 'secondary';
    if (status === 'completed') {
      if (conclusion === 'success') return 'success';
      if (conclusion === 'failure') return 'destructive';
      if (conclusion === 'cancelled') return 'secondary';
    }
    return 'default';
  };

  const getStatusIcon = (status: string, conclusion?: string) => {
    if (status === 'in_progress') return '🔄';
    if (status === 'queued') return '⏳';
    if (status === 'completed') {
      if (conclusion === 'success') return '✅';
      if (conclusion === 'failure') return '❌';
      if (conclusion === 'cancelled') return '⭕';
    }
    return '❓';
  };

  const formatDuration = (startedAt?: string, completedAt?: string) => {
    if (!startedAt) return 'Not started';
    if (!completedAt) return 'Running...';
    
    const start = new Date(startedAt);
    const end = new Date(completedAt);
    const diffMs = end.getTime() - start.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    
    if (diffMins > 0) {
      return `${diffMins}m ${diffSecs % 60}s`;
    }
    return `${diffSecs}s`;
  };

  const filteredRuns = runs.filter(run => {
    if (searchQuery && !run.workflow.name.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    return true;
  });

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-20 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <h3 className="text-lg font-medium text-red-800">Error</h3>
          <p className="text-red-700">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">Workflow runs</h1>
        <div className="flex gap-2">
          <Link href={`/repositories/${owner}/${repo}/actions`}>
            <Button variant="outline">Back to Actions</Button>
          </Link>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-4 mb-6">
        <div className="flex-1 max-w-sm">
          <Input
            placeholder="Search workflows..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm"
        >
          <option value="all">All statuses</option>
          <option value="queued">Queued</option>
          <option value="in_progress">In progress</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        
        <select
          value={eventFilter}
          onChange={(e) => setEventFilter(e.target.value)}
          className="px-3 py-2 border border-gray-300 rounded-md text-sm"
        >
          <option value="all">All events</option>
          <option value="push">Push</option>
          <option value="pull_request">Pull request</option>
          <option value="schedule">Schedule</option>
          <option value="workflow_dispatch">Manual</option>
        </select>
      </div>

      {/* Workflow Runs List */}
      {filteredRuns.length === 0 ? (
        <Card className="p-8 text-center">
          <h3 className="text-lg font-medium mb-2">No workflow runs found</h3>
          <p className="text-gray-600">
            {searchQuery || statusFilter !== 'all' || eventFilter !== 'all'
              ? 'Try adjusting your filters to see workflow runs.'
              : 'No workflows have been run yet.'}
          </p>
        </Card>
      ) : (
        <div className="space-y-4">
          {filteredRuns.map((run) => (
            <Card key={run.id} className="p-4 hover:shadow-md transition-shadow">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <span className="text-2xl">
                    {getStatusIcon(run.status, run.conclusion)}
                  </span>
                  
                  <div>
                    <div className="flex items-center gap-2">
                      <Link
                        href={`/repositories/${owner}/${repo}/actions/runs/${run.id}`}
                        className="font-medium text-blue-600 hover:text-blue-800"
                      >
                        {run.workflow.name}
                      </Link>
                      <span className="text-gray-500">#{run.number}</span>
                    </div>
                    
                    <div className="flex items-center gap-2 text-sm text-gray-500 mt-1">
                      <Badge variant="outline" className="px-2 py-0 text-xs">
                        {run.event}
                      </Badge>
                      {run.head_branch && (
                        <>
                          <span>•</span>
                          <span className="font-mono">{run.head_branch}</span>
                        </>
                      )}
                      <span>•</span>
                      <span className="font-mono">{run.head_sha.substring(0, 7)}</span>
                      {run.actor && (
                        <>
                          <span>•</span>
                          <span>by {run.actor.login}</span>
                        </>
                      )}
                    </div>
                  </div>
                </div>
                
                <div className="text-right">
                  <Badge variant={getStatusColor(run.status, run.conclusion)}>
                    {run.conclusion || run.status}
                  </Badge>
                  
                  <div className="text-sm text-gray-500 mt-1">
                    <div>{formatDuration(run.started_at, run.completed_at)}</div>
                    <div>{new Date(run.created_at).toLocaleString()}</div>
                  </div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}