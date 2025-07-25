'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

interface MFASetupProps {
  onComplete: () => void;
  onCancel: () => void;
}

export function MFASetup({ onCancel }: MFASetupProps) {
  const [] = useState(false);

  return (
    <Card className="p-6 max-w-md mx-auto">
      <div className="text-center mb-6">
        <h2 className="text-2xl font-bold mb-2">Setup Two-Factor Authentication</h2>
        <p className="text-gray-600">
          MFA setup is not yet implemented in this version
        </p>
      </div>

      <div className="bg-blue-50 border border-blue-200 text-blue-700 px-4 py-3 rounded mb-4">
        MFA functionality will be implemented in a future release. The authentication 
        system currently supports basic JWT-based authentication with OAuth providers.
      </div>

      <div className="flex space-x-3">
        <Button variant="outline" onClick={onCancel} className="flex-1">
          Close
        </Button>
      </div>
    </Card>
  );
}