'use client';

import React, { useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { useAuthStore } from '@/store/auth';

interface MFASetupProps {
  onComplete: () => void;
  onCancel: () => void;
}

export function MFASetup({ onComplete, onCancel }: MFASetupProps) {
  const [step, setStep] = useState<'generate' | 'verify'>('generate');
  const [mfaData, setMfaData] = useState<{
    secret: string;
    qrCodeUrl: string;
    backupCodes: string[];
  } | null>(null);
  const [verificationCode, setVerificationCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const { setupMFA, verifyMFA } = useAuthStore();

  const handleSetupMFA = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await setupMFA();
      setMfaData(data);
      setStep('verify');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to setup MFA');
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyMFA = async () => {
    if (!mfaData || !verificationCode) return;
    
    setIsLoading(true);
    setError(null);
    try {
      await verifyMFA(mfaData.secret, verificationCode);
      onComplete();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid verification code');
    } finally {
      setIsLoading(false);
    }
  };

  const downloadBackupCodes = () => {
    if (!mfaData) return;
    
    const content = `Hub MFA Backup Codes\n\nGenerated: ${new Date().toLocaleDateString()}\n\n${mfaData.backupCodes.join('\n')}\n\nKeep these codes safe. Each code can only be used once.`;
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'hub-mfa-backup-codes.txt';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <Card className="p-6 max-w-md mx-auto">
      <div className="text-center mb-6">
        <h2 className="text-2xl font-bold mb-2">Setup Two-Factor Authentication</h2>
        <p className="text-gray-600">
          Add an extra layer of security to your account
        </p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {step === 'generate' && (
        <div className="space-y-4">
          <div className="text-sm text-gray-600">
            <p className="mb-4">
              Two-factor authentication adds an extra layer of security to your account.
              You&apos;ll need to enter a code from your authenticator app each time you sign in.
            </p>
            <p className="mb-4">
              You&apos;ll need an authenticator app like Google Authenticator, Authy, or 1Password.
            </p>
          </div>
          
          <div className="flex space-x-3">
            <Button 
              onClick={handleSetupMFA} 
              disabled={isLoading}
              className="flex-1"
            >
              {isLoading ? 'Setting up...' : 'Setup MFA'}
            </Button>
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </div>
      )}

      {step === 'verify' && mfaData && (
        <div className="space-y-4">
          <div className="text-center">
            <p className="text-sm text-gray-600 mb-4">
              Scan this QR code with your authenticator app:
            </p>
            
            {/* QR Code */}
            <div className="bg-white p-4 border rounded-lg mb-4">
              <img 
                src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(mfaData.qrCodeUrl)}`}
                alt="MFA QR Code"
                className="mx-auto"
              />
            </div>
            
            <p className="text-xs text-gray-500 mb-4">
              Manual entry key: <code className="bg-gray-100 px-2 py-1 rounded">{mfaData.secret}</code>
            </p>
          </div>

          <div>
            <label htmlFor="verification-code" className="block text-sm font-medium text-gray-700 mb-2">
              Enter verification code from your app:
            </label>
            <Input
              id="verification-code"
              type="text"
              value={verificationCode}
              onChange={(e) => setVerificationCode(e.target.value)}
              placeholder="000000"
              maxLength={6}
              className="text-center text-lg tracking-wider"
            />
          </div>

          {/* Backup Codes */}
          <div className="bg-yellow-50 border border-yellow-200 p-4 rounded">
            <h4 className="font-medium text-yellow-800 mb-2">Backup Codes</h4>
            <p className="text-sm text-yellow-700 mb-3">
              Save these backup codes in a safe place. You can use them to access your account if you lose your authenticator device.
            </p>
            <div className="grid grid-cols-2 gap-2 text-sm font-mono bg-white p-3 rounded border">
              {mfaData.backupCodes.map((code, index) => (
                <div key={index} className="text-center py-1">
                  {code}
                </div>
              ))}
            </div>
            <Button 
              variant="outline" 
              size="sm"
              onClick={downloadBackupCodes}
              className="mt-3 w-full"
            >
              Download Backup Codes
            </Button>
          </div>

          <div className="flex space-x-3">
            <Button 
              onClick={handleVerifyMFA} 
              disabled={isLoading || !verificationCode}
              className="flex-1"
            >
              {isLoading ? 'Verifying...' : 'Verify & Enable MFA'}
            </Button>
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
          </div>
        </div>
      )}
    </Card>
  );
}