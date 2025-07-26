'use client';

import React from 'react';
import { AppLayout } from '@/components/layout/AppLayout';
import { AuthAuditLogViewer } from '@/components/auth/AuthAuditLogViewer';

export default function AdminAuditLogsPage() {
  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">Audit Logs</h1>
          <p className="text-muted-foreground mt-2">
            Monitor and review all administrative actions and authentication events
          </p>
        </div>

        <AuthAuditLogViewer />
      </div>
    </AppLayout>
  );
}