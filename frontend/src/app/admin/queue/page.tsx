 'use client';

 import React, { useEffect, useState } from 'react';
 import { AppLayout } from '@/components/layout/AppLayout';
 import { Card } from '@/components/ui/Card';
 import api from '@/lib/api';
 import { AdminQueueMetricsDashboard } from '@/components/queue/AdminQueueMetricsDashboard';
 import { QueueManagementPanel } from '@/components/queue/QueueManagementPanel';

 interface QueueStatus {
   redis: boolean;
   fallback: boolean;
 }

 export default function AdminQueuePage() {
   const [queueStatus, setQueueStatus] = useState<QueueStatus | null>(null);
   const [loading, setLoading] = useState(true);
   const [error, setError] = useState<string | null>(null);

   useEffect(() => {
     async function fetchStatus() {
       try {
         const res = await api.get('/api/v1/admin/queue/status');
         setQueueStatus(res.data);
       } catch (err: any) {
         setError(err.response?.data?.error || 'Failed to load queue status');
       } finally {
         setLoading(false);
       }
     }
     fetchStatus();
   }, []);

   if (loading) {
     return (
       <AppLayout>
         <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
           <div className="text-muted-foreground">Loading queue administration...</div>
         </div>
       </AppLayout>
     );
   }

   return (
     <AppLayout>
       <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
         <h1 className="text-3xl font-bold mb-4">Queue Monitoring & Management</h1>
         {error && <div className="text-red-500 mb-4">{error}</div>}
         {queueStatus && (
           <Card className="mb-6 p-6">
             <h2 className="text-xl font-semibold mb-2">Queue Health</h2>
             <p>
               Redis Connected: <strong>{queueStatus.redis ? 'Yes' : 'No'}</strong>
             </p>
             <p>
               Fallback Mode: <strong>{queueStatus.fallback ? 'Yes' : 'No'}</strong>
             </p>
           </Card>
         )}
         <AdminQueueMetricsDashboard />
         <QueueManagementPanel />
       </div>
     </AppLayout>
   );
 }
