 'use client';

 import React, { useEffect, useState } from 'react';
 import { Card } from '@/components/ui/Card';
 import api from '@/lib/api';

 interface Metrics {
   throughput: number;
   avg_wait_time: string;
   success_rate: number;
 }

 export function AdminQueueMetricsDashboard() {
   const [metrics, setMetrics] = useState<Metrics | null>(null);
   const [error, setError] = useState<string | null>(null);

   useEffect(() => {
     async function fetchMetrics() {
       try {
         const res = await api.get('/api/v1/admin/queue/metrics');
         setMetrics(res.data);
       } catch (err: any) {
         setError(err.response?.data?.error || 'Failed to load queue metrics');
       }
     }
     fetchMetrics();
   }, []);

   if (error) {
     return <div className="text-red-500">{error}</div>;
   }
   if (!metrics) {
     return <div>Loading queue metrics...</div>;
   }
   return (
     <Card className="mb-6 p-6">
       <h2 className="text-xl font-semibold mb-4">Queue Performance Metrics</h2>
       <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
         <div>
           <p className="text-sm text-muted-foreground">Throughput (jobs/sec)</p>
           <p className="text-2xl font-bold">{metrics.throughput}</p>
         </div>
         <div>
           <p className="text-sm text-muted-foreground">Avg Wait Time</p>
           <p className="text-2xl font-bold">{metrics.avg_wait_time}</p>
         </div>
         <div>
           <p className="text-sm text-muted-foreground">Success Rate (%)</p>
           <p className="text-2xl font-bold">{metrics.success_rate}%</p>
         </div>
       </div>
     </Card>
   );
 }
