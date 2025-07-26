 'use client';

 import React, { useEffect, useState } from 'react';
 import { Card } from '@/components/ui/Card';
 import { Button } from '@/components/ui/Button';
 import api from '@/lib/api';

 interface Job {
   id: string;
   name: string;
   priority: number;
 }

 export function QueueManagementPanel() {
   const [jobs, setJobs] = useState<Job[]>([]);
   const [error, setError] = useState<string | null>(null);

   useEffect(() => {
     async function fetchJobs() {
       try {
         const res = await api.get('/api/v1/admin/queue/jobs?limit=10');
         setJobs(res.data.jobs);
       } catch (err: any) {
         setError(err.response?.data?.error || 'Failed to load queued jobs');
       }
     }
     fetchJobs();
   }, []);

   async function updatePriority(jobId: string, newPriority: number) {
     try {
       await api.post(`/api/v1/admin/queue/jobs/${jobId}/priority`, { priority: newPriority });
       setJobs(jobs.map(j => (j.id === jobId ? { ...j, priority: newPriority } : j)));
     } catch (err: any) {
       setError(err.response?.data?.error || 'Failed to update job priority');
     }
   }

   if (error) {
     return <div className="text-red-500">{error}</div>;
   }
   if (jobs.length === 0) {
     return <div>Loading queued jobs...</div>;
   }
   return (
     <Card className="p-6">
       <h2 className="text-xl font-semibold mb-4">Job Priority Management</h2>
       <table className="w-full">
         <thead>
           <tr className="text-left">
             <th>ID</th>
             <th>Name</th>
             <th>Priority</th>
             <th>Action</th>
           </tr>
         </thead>
         <tbody>
           {jobs.map(job => (
             <tr key={job.id}>
               <td className="py-2">{job.id}</td>
               <td className="py-2">{job.name}</td>
               <td className="py-2">{job.priority}</td>
               <td className="py-2 space-x-2">
                 <Button size="sm" onClick={() => updatePriority(job.id, job.priority + 1)}>
                   ↑
                 </Button>
                 <Button size="sm" onClick={() => updatePriority(job.id, job.priority - 1)}>
                   ↓
                 </Button>
               </td>
             </tr>
           ))}
         </tbody>
       </table>
     </Card>
   );
 }
