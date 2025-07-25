'use client';

import Link from 'next/link';
import { AppLayout } from '@/components/layout/AppLayout';
import { SSHKeyManagement } from '@/components/settings/SSHKeyManagement';

export default function SSHKeysPage() {
  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
          <Link href="/settings" className="hover:text-foreground transition-colors">
            Settings
          </Link>
          <span>/</span>
          <span className="text-foreground font-medium">SSH Keys</span>
        </nav>

        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground">SSH Keys</h1>
          <p className="text-muted-foreground mt-2">
            Manage your SSH keys for secure access to repositories
          </p>
        </div>

        {/* SSH Key Management Component */}
        <SSHKeyManagement />
      </div>
    </AppLayout>
  );
}