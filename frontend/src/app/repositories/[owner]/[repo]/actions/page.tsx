'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';

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

  useEffect(() => {
    fetchWorkflows();
    fetchRecentRuns();
  }, [owner, repo]);

  const fetchWorkflows = async () => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/workflows`);
      if (response.ok) {
        const data = await response.json();
        setWorkflows(data.workflows || []);
      } else {
        throw new Error('Failed to fetch workflows');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };

  const fetchRecentRuns = async () => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/runs?limit=10`);
      if (response.ok) {
        const data = await response.json();
        setRecentRuns(data.workflow_runs || []);
      } else {
        throw new Error('Failed to fetch workflow runs');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

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
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-24 bg-gray-200 rounded"></div>
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
            <h3 className="text-lg font-medium mb-2">Get started with GitHub Actions</h3>
            <p className="text-gray-600 mb-4">
              Automate your workflow from idea to production with GitHub Actions.
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
                      <p className="text-sm text-gray-500">{workflow.path}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={workflow.enabled ? "default" : "secondary"}>
                      {workflow.enabled ? 'Active' : 'Disabled'}
                    </Badge>
                  </div>
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
                      <div className="flex items-center gap-2 text-sm text-gray-500">
                        <span>{run.event}</span>
                        <span>•</span>
                        <span>{run.head_branch}</span>
                        <span>•</span>
                        <span>{run.head_sha.substring(0, 7)}</span>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <Badge variant={getStatusColor(run.status, run.conclusion) as 'default' | 'secondary' | 'destructive' | 'outline'}>
                      {run.conclusion || run.status}
                    </Badge>
                    <p className="text-sm text-gray-500 mt-1">
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