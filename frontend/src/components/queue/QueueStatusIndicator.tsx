 'use client';

 import React, { useEffect, useState } from 'react';
 import { apiClient } from '@/lib/api';

 interface Props {
   runId: string;
 }

 export function QueueStatusIndicator({ runId }: Props) {
   const [position, setPosition] = useState<number | null>(null);
   const [waitTime, setWaitTime] = useState<string>('');
   const [error, setError] = useState<string | null>(null);

   useEffect(() => {
     async function fetchPosition() {
       try {
        const res = await apiClient.get<{ position: number; estimated_wait_time: string }>(
          `/actions/runs/${runId}/queue-position`
        );
        setPosition(res.data.position);
        setWaitTime(res.data.estimated_wait_time);
       } catch (err: any) {
         setError(err.response?.data?.error || 'Failed to load queue status');
       }
     }
     fetchPosition();
   }, [runId]);

   if (error) {
     return <div className="text-red-500">Queue status unavailable</div>;
   }
   if (position === null) {
     return <div>Loading queue status...</div>;
   }
   return (
     <div className="queue-status-indicator text-sm text-muted-foreground">
       {position > 0 ? (
         <span>
           Position in queue: <strong>{position}</strong>, estimated wait: <strong>{waitTime}</strong>
         </span>
       ) : (
         <span>Your job is running</span>
       )}
     </div>
   );
 }
