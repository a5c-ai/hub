import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  context: { params: { runnerId: string } }
) {
  const { runnerId } = await context.params;

  // Mock health data
  const healthStats = {
    id: runnerId,
    name: `runner-${runnerId.slice(-8)}`,
    status: Math.random() > 0.3 ? 'online' : Math.random() > 0.5 ? 'busy' : 'offline',
    cpu_usage: Math.random() * 100,
    memory_usage: Math.random() * 100,
    disk_usage: Math.random() * 100,
    current_job: Math.random() > 0.7 ? {
      id: 'job-123',
      workflow_name: 'CI/CD Pipeline',
      started_at: new Date(Date.now() - Math.random() * 3600000).toISOString(),
    } : undefined,
    queue_length: Math.floor(Math.random() * 5),
    last_heartbeat: new Date(Date.now() - Math.random() * 300000).toISOString(),
  };

  return NextResponse.json(healthStats);
}