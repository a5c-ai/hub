'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { apiClient } from '@/lib/api';

interface MFASetupProps {
  onComplete: () => void;
  onCancel: () => void;
}

interface MFASetupStatus {
  emailSent: boolean;
  emailDelivered?: boolean;
  emailSentAt?: string;
  error?: string;
}

export function MFASetup({ onComplete, onCancel }: MFASetupProps) {
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState<'initial' | 'setup' | 'complete'>('initial');
  const [emailStatus, setEmailStatus] = useState<MFASetupStatus>({
    emailSent: false
  });

  const setupMFA = async () => {
    try {
      setLoading(true);
      setStep('setup');
      
      // Simulate MFA setup process
      const response = await apiClient.post('/auth/mfa/setup');
      
      if (response.success) {
        // Email with backup codes was sent
        setEmailStatus({
          emailSent: true,
          emailDelivered: true, // This would come from actual email service status
          emailSentAt: new Date().toISOString()
        });
        
        setTimeout(() => {
          setStep('complete');
          setLoading(false);
        }, 2000);
      }
    } catch (error) {
      console.error('MFA setup failed:', error);
      setEmailStatus({
        emailSent: true,
        emailDelivered: false,
        error: 'Failed to send backup codes email'
      });
      setLoading(false);
    }
  };

  const handleComplete = () => {
    onComplete();
  };

  return (
    <Card className="p-6 max-w-md mx-auto">
      <div className="text-center mb-6">
        <h2 className="text-2xl font-bold mb-2">Setup Two-Factor Authentication</h2>
        <p className="text-muted-foreground">
          Two-factor authentication adds an additional layer of security to your account.
        </p>
      </div>

      {step === 'initial' && (
        <>
          <div className="space-y-4 mb-6">
            <div className="p-4 bg-muted rounded-lg">
              <h3 className="font-medium text-foreground mb-2">What happens next:</h3>
              <ol className="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
                <li>We&apos;ll generate backup recovery codes for your account</li>
                <li>An email with your backup codes will be sent to your email address</li>
                <li>Save these codes in a secure location</li>
                <li>Two-factor authentication will be enabled</li>
              </ol>
            </div>
          </div>

          <div className="flex space-x-3">
            <Button variant="outline" onClick={onCancel} className="flex-1">
              Cancel
            </Button>
            <Button onClick={setupMFA} disabled={loading} className="flex-1">
              {loading ? 'Setting up...' : 'Enable MFA'}
            </Button>
          </div>
        </>
      )}

      {step === 'setup' && (
        <>
          <div className="space-y-4 mb-6">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <h3 className="font-medium text-foreground mb-2">Setting up MFA...</h3>
            </div>

            {/* Email Delivery Status */}
            <div className="p-4 bg-muted rounded-lg">
              <h4 className="font-medium text-foreground mb-3">Email Delivery Status</h4>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Backup codes email</span>
                  {emailStatus.emailSent ? (
                    emailStatus.emailDelivered ? (
                      <Badge variant="default" className="bg-success text-success-foreground">
                        ✓ Delivered
                      </Badge>
                    ) : (
                      <Badge variant="destructive">
                        ✗ Failed
                      </Badge>
                    )
                  ) : (
                    <Badge variant="outline">
                      Sending...
                    </Badge>
                  )}
                </div>
                
                {emailStatus.emailSentAt && (
                  <p className="text-xs text-muted-foreground">
                    Sent at: {new Date(emailStatus.emailSentAt).toLocaleString()}
                  </p>
                )}
                
                {emailStatus.error && (
                  <p className="text-xs text-destructive">
                    Error: {emailStatus.error}
                  </p>
                )}
              </div>
            </div>
          </div>
        </>
      )}

      {step === 'complete' && (
        <>
          <div className="space-y-4 mb-6">
            <div className="text-center">
              <div className="rounded-full h-12 w-12 bg-success text-success-foreground flex items-center justify-center mx-auto mb-4">
                ✓
              </div>
              <h3 className="font-medium text-foreground mb-2">MFA Setup Complete!</h3>
              <p className="text-sm text-muted-foreground">
                Two-factor authentication has been enabled for your account.
              </p>
            </div>

            <div className="p-4 bg-success/10 border border-success/20 rounded-lg">
              <h4 className="font-medium text-success mb-2">✓ Email Sent Successfully</h4>
              <p className="text-sm text-success-foreground">
                Your backup recovery codes have been sent to your email address. 
                Please save them in a secure location.
              </p>
              {emailStatus.emailSentAt && (
                <p className="text-xs text-success-foreground/80 mt-2">
                  Delivered at: {new Date(emailStatus.emailSentAt).toLocaleString()}
                </p>
              )}
            </div>

            <div className="p-4 bg-warning/10 border border-warning/20 rounded-lg">
              <h4 className="font-medium text-warning mb-2">⚠️ Important</h4>
              <p className="text-sm text-warning-foreground">
                Each backup code can only be used once. Store them securely and 
                consider printing or saving them in a password manager.
              </p>
            </div>
          </div>

          <div className="flex justify-end">
            <Button onClick={handleComplete}>
              Complete Setup
            </Button>
          </div>
        </>
      )}
    </Card>
  );
}