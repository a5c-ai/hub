'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { apiClient } from '@/lib/api';

interface Workflow {
  id: string;
  name: string;
  path: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

interface WorkflowRun {
  id: string;
  number: number;
  status: 'queued' | 'in_progress' | 'completed' | 'cancelled';
  conclusion?: 'success' | 'failure' | 'cancelled' | 'skipped';
  head_sha: string;
  head_branch?: string;
  event: string;
  created_at: string;
  actor?: {
    login: string;
    avatar_url: string;
  };
}

export default function ActionsPage() {
  const params = useParams();
  const { owner, repo } = params;
  
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [recentRuns, setRecentRuns] = useState<WorkflowRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchWorkflows = useCallback(async () => {
    try {
      const response = await apiClient.get<{ workflows: Workflow[] }>(`/repositories/${owner}/${repo}/actions/workflows`);
      setWorkflows(response.data.workflows || []);
    } catch (err: unknown) {
      if (err instanceof Error && 'response' in err) {
        const axiosErr = err as { response?: { status?: number; data?: { error?: string } }; message?: string };
        if (axiosErr.response?.status === 401) {
          setError('Please log in to access GitHub Actions');
        } else if (axiosErr.response?.status === 404) {
          setError('Repository not found or you do not have access to it');
        } else if (axiosErr.response?.status === 400) {
          setError(`Bad request: ${axiosErr.response?.data?.error || axiosErr.message}`);
        } else {
          setError(axiosErr instanceof Error ? axiosErr.message : 'Failed to fetch workflows');
        }
      } else {
        setError('Failed to fetch workflows');
      }
    }
  }, [owner, repo]);

  const fetchRecentRuns = useCallback(async () => {
    try {
      const response = await apiClient.get<{ workflow_runs: WorkflowRun[] }>(`/repositories/${owner}/${repo}/actions/runs?limit=10`);
      setRecentRuns(response.data.workflow_runs || []);
    } catch (err: unknown) {
      if (err instanceof Error && 'response' in err) {
        const axiosErr = err as { response?: { status?: number; data?: { error?: string } }; message?: string };
        if (axiosErr.response?.status === 401) {
          setError('Please log in to access GitHub Actions');
        } else if (axiosErr.response?.status === 404) {
          setError('Repository not found or you do not have access to it');
        } else {
          setError(axiosErr instanceof Error ? axiosErr.message : 'Failed to fetch workflow runs');
        }
      } else {
        setError('Failed to fetch workflow runs');
      }
    } finally {
      setLoading(false);
    }
  }, [owner, repo]);

  useEffect(() => {
    fetchWorkflows();
    fetchRecentRuns();
  }, [fetchWorkflows, fetchRecentRuns]);


  const getStatusColor = (status: string, conclusion?: string) => {
    if (status === 'in_progress') return 'yellow';
    if (status === 'queued') return 'gray';
    if (status === 'completed') {
      if (conclusion === 'success') return 'green';
      if (conclusion === 'failure') return 'red';
      if (conclusion === 'cancelled') return 'gray';
    }
    return 'gray';
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

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/4 mb-4"></div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-24 bg-muted rounded"></div>
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
        <h1 className="text-3xl font-bold">Actions</h1>
        <div className="flex gap-2">
          <Link href={`/repositories/${owner}/${repo}/actions/new`}>
            <Button>New workflow</Button>
          </Link>
          <Link href={`/repositories/${owner}/${repo}/settings/secrets`}>
            <Button variant="outline">Manage secrets</Button>
          </Link>
          <Link href={`/repositories/${owner}/${repo}/settings/runners`}>
            <Button variant="outline">Runners</Button>
          </Link>
        </div>
      </div>

      {/* Workflows Section */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">All workflows</h2>
        
        {workflows.length === 0 ? (
          <Card className="p-8 text-center">
            <h3 className="text-lg font-medium mb-2">Get started with Hub Actions</h3>
            <p className="text-muted-foreground mb-4">
              Workflows help you automate your software development workflows with CI/CD.
            </p>
            <Link href={`/repositories/${owner}/${repo}/actions/new`}>
              <Button>Set up a workflow yourself</Button>
            </Link>
          </Card>
        ) : (
          <div className="space-y-4">
            {workflows.map((workflow) => (
              <Card key={workflow.id} className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-6 h-6 bg-blue-100 rounded-full flex items-center justify-center">
                      <span className="text-xs">🔧</span>
                    </div>
                    <div>
                      <Link
                        href={`/repositories/${owner}/${repo}/actions/workflows/${workflow.id}`}
                        className="font-medium text-blue-600 hover:text-blue-800"
                      >
                        {workflow.name}
                      </Link>
                      <p className="text-sm text-muted-foreground">{workflow.path}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={workflow.enabled ? "default" : "secondary"}>
                      {workflow.enabled ? 'Active' : 'Disabled'}
                    </Badge>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <span>{new Date(workflow.created_at).toLocaleDateString()}</span>
                  {workflow.updated_at !== workflow.created_at && (
                    <span>• Updated {new Date(workflow.updated_at).toLocaleDateString()}</span>
                  )}
                </div>
                
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Recent Workflow Runs */}
      {recentRuns.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Recent workflow runs</h2>
            <Link
              href={`/repositories/${owner}/${repo}/actions/runs`}
              className="text-blue-600 hover:text-blue-800"
            >
              View all runs
            </Link>
          </div>
          
          <div className="space-y-2">
            {recentRuns.map((run) => (
              <Card key={run.id} className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className="text-xl">
                      {getStatusIcon(run.status, run.conclusion)}
                    </span>
                    <div>
                      <Link
                        href={`/repositories/${owner}/${repo}/actions/runs/${run.id}`}
                        className="font-medium text-blue-600 hover:text-blue-800"
                      >
                        #{run.number}
                      </Link>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span>{run.event}</span>
                        <span>•</span>
                        <span>{run.head_branch}</span>
                        <span>•</span>
                        <span>{run.head_sha.substring(0, 7)}</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge variant={getStatusColor(run.status, run.conclusion) as 'default' | 'secondary' | 'outline' | 'destructive'}>
                      {run.conclusion || run.status}
                    </Badge>
                    <p className="text-sm text-muted-foreground mt-1">
                      {new Date(run.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}