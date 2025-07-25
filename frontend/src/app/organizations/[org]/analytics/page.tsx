'use client';

import { useParams } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { OrganizationAnalyticsDashboard } from '@/components/organization/OrganizationAnalyticsDashboard';

export default function OrganizationAnalyticsPage() {
  const params = useParams();
  const org = params.org as string;

  return (
    <AppLayout>
      <OrganizationAnalyticsDashboard orgName={org} />
    </AppLayout>
  );
}