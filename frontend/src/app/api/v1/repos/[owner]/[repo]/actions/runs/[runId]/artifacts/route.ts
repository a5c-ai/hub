import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ owner: string; repo: string; runId: string }> }
) {
  const { owner: _owner, repo: _repo, runId: _runId } = await context.params;

  // Mock artifacts data
  const artifacts = [
    {
      id: 'artifact-1',
      name: 'build-artifacts.zip',
      size_bytes: 1024000,
      created_at: new Date(Date.now() - 3600000).toISOString(),
      expires_at: new Date(Date.now() + 7776000000).toISOString(), // 90 days
      expired: false,
    },
    {
      id: 'artifact-2',
      name: 'test-results.xml',
      size_bytes: 52000,
      created_at: new Date(Date.now() - 1800000).toISOString(),
      expires_at: new Date(Date.now() + 7776000000).toISOString(),
      expired: false,
    },
  ];

  return NextResponse.json({ artifacts });
}

export async function POST(
  request: NextRequest,
  context: { params: Promise<{ owner: string; repo: string; runId: string }> }
) {
  const { owner: _owner, repo: _repo, runId: _runId } = await context.params;

  try {
    const formData = await request.formData();
    const file = formData.get('artifact') as File;
    const name = formData.get('name') as string;

    if (!file || !name) {
      return NextResponse.json(
        { error: 'Missing file or name' },
        { status: 400 }
      );
    }

    // In a real implementation, you would:
    // 1. Validate the file
    // 2. Upload to storage backend
    // 3. Save metadata to database
    // 4. Return the created artifact

    const artifact = {
      id: `artifact-${Date.now()}`,
      name,
      size_bytes: file.size,
      created_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 7776000000).toISOString(),
      expired: false,
    };

    return NextResponse.json({ artifact }, { status: 201 });
  } catch (_error) {
    return NextResponse.json(
      { error: 'Failed to upload artifact' },
      { status: 500 }
    );
  }
}