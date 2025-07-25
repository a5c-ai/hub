'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Button, Input, Card, CardHeader, CardTitle, CardContent } from '@/components/ui';
import { useAuthStore } from '@/store/auth';
import { OAuthButtons } from '@/components/auth/OAuthButtons';

interface LoginFormData {
  email: string;
  password: string;
  mfaCode?: string;
  remember?: boolean;
}

export function LoginForm() {
  const router = useRouter();
  const { login, isLoading, error, clearError } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);
  const [needsMFA, setNeedsMFA] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>();

  const onSubmit = async (data: LoginFormData) => {
    try {
      clearError();
      console.log('LoginForm: Starting login process...', { email: data.email });
      await login(data.email, data.password, data.mfaCode);
      console.log('LoginForm: Login successful, redirecting to dashboard...');
      
      // Add small delay to ensure tokens are stored before navigation
      setTimeout(() => {
        console.log('LoginForm: Checking localStorage before redirect:', {
          authToken: !!localStorage.getItem('auth_token'),
          refreshToken: !!localStorage.getItem('refresh_token')
        });
        router.push('/dashboard');
      }, 100);
    } catch (err: unknown) {
      console.error('LoginForm: Login failed:', err);
      if (
        err && 
        typeof err === 'object' && 
        'response' in err && 
        err.response &&
        typeof err.response === 'object' &&
        'data' in err.response &&
        err.response.data &&
        typeof err.response.data === 'object' &&
        'error' in err.response.data &&
        err.response.data.error === 'MFA code required'
      ) {
        setNeedsMFA(true);
      }
    }
  };

  const handleOAuthLogin = (provider: string) => {
    // Generate state parameter for security
    const state = Math.random().toString(36).substring(2, 15);
    sessionStorage.setItem('oauth_state', state);
    
    // Redirect to OAuth URL - use the API base URL
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
    window.location.href = `${baseURL}/auth/oauth/${provider}?state=${state}`;
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4 py-12 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-bold tracking-tight text-foreground">
            Sign in to Hub
          </h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Or{' '}
            <Link
              href="/register"
              className="font-medium text-primary hover:text-primary/80"
            >
              create a new account
            </Link>
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Sign In</CardTitle>
          </CardHeader>
          <CardContent>
            <form className="space-y-6" onSubmit={handleSubmit(onSubmit)}>
              {error && (
                <div className="rounded-md bg-destructive/10 p-3" data-testid="error-message">
                  <div className="text-sm text-destructive">{error}</div>
                </div>
              )}

              <Input
                label="Email address"
                type="email"
                autoComplete="email"
                data-testid="email-input"
                error={errors.email?.message}
                {...register('email', {
                  required: 'Email is required',
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: 'Invalid email address',
                  },

                })}
              />

              <div className="relative">
                <Input
                  label="Password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  data-testid="password-input"
                  error={errors.password?.message}
                  {...register('password', {
                    required: 'Password is required',
                    minLength: {
                      value: 12,
                      message: 'Password must be at least 12 characters',
                    },
                  })}
                />
                <button
                  type="button"
                  className="absolute right-3 top-8 text-sm text-muted-foreground hover:text-foreground"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? 'Hide' : 'Show'}
                </button>
              </div>

              {needsMFA && (
                <Input
                  label="Two-Factor Authentication Code"
                  type="text"
                  placeholder="000000"
                  maxLength={6}
                  error={errors.mfaCode?.message}
                  {...register('mfaCode', {
                    required: needsMFA ? 'MFA code is required' : false,
                    pattern: {
                      value: /^\d{6}$/,
                      message: 'MFA code must be 6 digits',
                    },
                  })}
                />
              )}

              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <input
                    id="remember-me"
                    type="checkbox"
                    className="h-4 w-4 rounded border-border text-primary focus:ring-primary"
                    {...register('remember')}
                  />
                  <label htmlFor="remember-me" className="ml-2 block text-sm text-muted-foreground">
                    Remember me
                  </label>
                </div>

                <Link
                  href="/forgot-password"
                  className="text-sm font-medium text-primary hover:text-primary/80"
                >
                  Forgot your password?
                </Link>
              </div>

              <Button
                type="submit"
                className="w-full"
                loading={isLoading}
                disabled={isLoading}
                data-testid="login-button"
              >
                Sign in
              </Button>
            </form>

            <div className="mt-6">
              <OAuthButtons onOAuthLogin={handleOAuthLogin} isLoading={isLoading} />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}