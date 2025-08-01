import { NextResponse } from 'next/server';

/**
 * Health check endpoint for the frontend application.
 * Responds with HTTP 200 if the application is running.
 */
export async function GET() {
  return NextResponse.json({ status: 'ok' });
}
