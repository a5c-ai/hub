'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';

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
    path: string;
  };
  jobs: Job[];
}

interface Job {
  id: string;
  name: string;
  status: 'queued' | 'in_progress' | 'completed' | 'cancelled';
  conclusion?: 'success' | 'failure' | 'cancelled' | 'skipped';
  started_at?: string;
  completed_at?: string;
  runner_name?: string;
  steps: Step[];
}

interface Step {
  id: string;
  number: number;
  name: string;
  status: 'queued' | 'in_progress' | 'completed' | 'cancelled';
  conclusion?: 'success' | 'failure' | 'cancelled' | 'skipped';
  started_at?: string;
  completed_at?: string;
  output?: string;
}

export default function WorkflowRunDetailPage() {
  const params = useParams();
  const { owner, repo, runId } = params;
  
  const [run, setRun] = useState<WorkflowRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedJob, setSelectedJob] = useState<string | null>(null);
  const [logs, setLogs] = useState<Record<string, string>>({});

  useEffect(() => {
    fetchWorkflowRun();
  }, [owner, repo, runId]);

  useEffect(() => {
    if (run && run.jobs.length > 0 && !selectedJob) {
      setSelectedJob(run.jobs[0].id);
    }
  }, [run, selectedJob]);

  const fetchWorkflowRun = async () => {
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/runs/${runId}`);
      if (response.ok) {
        const data = await response.json();
        setRun(data);
      } else {
        throw new Error('Failed to fetch workflow run');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const fetchJobLogs = async (jobId: string) => {
    if (logs[jobId]) return; // Already fetched
    
    try {
      const response = await fetch(`/api/v1/repos/${owner}/${repo}/actions/jobs/${jobId}/logs`);
      if (response.ok) {
        const logText = await response.text();
        setLogs(prev => ({ ...prev, [jobId]: logText }));
      }
    } catch (err) {
      console.error('Failed to fetch job logs:', err);
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

  const selectedJobData = run?.jobs.find(job => job.id === selectedJob);

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="grid grid-cols-4 gap-6">
            <div className="h-64 bg-gray-200 rounded"></div>
            <div className="col-span-3 h-64 bg-gray-200 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  if (error || !run) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <h3 className="text-lg font-medium text-red-800">Error</h3>
          <p className="text-red-700">{error || 'Workflow run not found'}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Link
              href={`/repositories/${owner}/${repo}/actions/runs`}
              className="text-blue-600 hover:text-blue-800"
            >
              ← Back to runs
            </Link>
          </div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <span className="text-2xl">
              {getStatusIcon(run.status, run.conclusion)}
            </span>
            {run.workflow.name} #{run.number}
          </h1>
        </div>
        
        <div className="flex gap-2">
          <Button 
            variant="outline" 
            disabled={run.status !== 'in_progress'}
          >
            Cancel run
          </Button>
          <Button variant="outline">
            Re-run jobs
          </Button>
        </div>
      </div>

      {/* Run Info */}
      <Card className="p-4 mb-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <span className="text-sm text-gray-500">Status</span>
            <div className="flex items-center gap-2 mt-1">
              <Badge variant={getStatusColor(run.status, run.conclusion) as any}>
                {run.conclusion || run.status}
              </Badge>
            </div>
          </div>
          
          <div>
            <span className="text-sm text-gray-500">Duration</span>
            <p className="mt-1">{formatDuration(run.started_at, run.completed_at)}</p>
          </div>
          
          <div>
            <span className="text-sm text-gray-500">Event</span>
            <p className="mt-1">
              <Badge variant="outline">{run.event}</Badge>
            </p>
          </div>
          
          <div>
            <span className="text-sm text-gray-500">Commit</span>
            <p className="mt-1 font-mono text-sm">
              {run.head_sha.substring(0, 7)}
              {run.head_branch && (
                <span className="ml-2 text-gray-500">on {run.head_branch}</span>
              )}
            </p>
          </div>
        </div>
      </Card>

      {/* Jobs and Logs */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Jobs Sidebar */}
        <div className="space-y-2">
          <h3 className="font-semibold mb-3">Jobs ({run.jobs.length})</h3>
          {run.jobs.map((job) => (
            <Card
              key={job.id}
              className={`p-3 cursor-pointer transition-colors ${
                selectedJob === job.id ? 'bg-blue-50 border-blue-200' : 'hover:bg-gray-50'
              }`}
              onClick={() => {
                setSelectedJob(job.id);
                fetchJobLogs(job.id);
              }}
            >
              <div className="flex items-center gap-2">
                <span className="text-lg">
                  {getStatusIcon(job.status, job.conclusion)}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{job.name}</p>
                  <p className="text-xs text-gray-500">
                    {formatDuration(job.started_at, job.completed_at)}
                  </p>
                </div>
              </div>
            </Card>
          ))}
        </div>

        {/* Job Details and Logs */}
        <div className="lg:col-span-3">
          {selectedJobData ? (
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-xl font-semibold">{selectedJobData.name}</h3>
                {selectedJobData.runner_name && (
                  <Badge variant="outline">
                    Runner: {selectedJobData.runner_name}
                  </Badge>
                )}
              </div>

              {/* Steps */}
              <Card className="mb-4">
                <div className="p-4 border-b">
                  <h4 className="font-medium">Steps</h4>
                </div>
                <div className="divide-y">
                  {selectedJobData.steps.map((step) => (
                    <div key={step.id} className="p-4">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          <span className="text-lg">
                            {getStatusIcon(step.status, step.conclusion)}
                          </span>
                          <div>
                            <p className="font-medium">{step.name}</p>
                            <p className="text-sm text-gray-500">
                              Step {step.number}
                            </p>
                          </div>
                        </div>
                        <div className="text-right text-sm text-gray-500">
                          {formatDuration(step.started_at, step.completed_at)}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </Card>

              {/* Logs */}
              <Card>
                <div className="p-4 border-b">
                  <h4 className="font-medium">Logs</h4>
                </div>
                <div className="p-4">
                  {logs[selectedJob] ? (
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded text-sm overflow-x-auto">
                      {logs[selectedJob]}
                    </pre>
                  ) : (
                    <div className="text-center py-8 text-gray-500">
                      <p>Click "View logs" to see the build output</p>
                      <Button
                        variant="outline"
                        className="mt-2"
                        onClick={() => fetchJobLogs(selectedJob!)}
                      >
                        View logs
                      </Button>
                    </div>
                  )}
                </div>
              </Card>
            </div>
          ) : (
            <Card className="p-8 text-center">
              <p className="text-gray-500">Select a job to view details</p>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}