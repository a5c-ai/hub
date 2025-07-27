'use client';

import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button, Input, Card, CardHeader, CardTitle, CardContent } from '@/components/ui';
import { useOrganizationStore } from '@/store/organization';

interface CreateOrganizationFormData {
  login: string;
  name: string;
  description?: string;
}

export default function NewOrganizationPage() {
  const router = useRouter();
  const { createOrganization, isLoading, error, clearError } = useOrganizationStore();
  
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<CreateOrganizationFormData>();

  const organizationLogin = watch('login');

  const onSubmit = async (data: CreateOrganizationFormData) => {
    try {
      clearError();
      console.log('Creating organization:', data);
      const organization = await createOrganization({
        login: data.login,
        name: data.name,
        description: data.description,
      });
      console.log('Organization created successfully:', organization);
      
      // Navigate to the organization page
      router.push(`/organizations/${organization.login}`);
    } catch (err) {
      console.error('Failed to create organization:', err);
      // Error is handled by the store
    }
  };

  return (
    <AppLayout>
      <div className="container mx-auto px-4 py-8 max-w-2xl">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">
            Create a new organization
          </h1>
          <p className="text-muted-foreground">
            Organizations are shared accounts where teams can collaborate across many projects at once.
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Organization Details</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {error && (
                <div className="rounded-md bg-destructive/10 p-3">
                  <div className="text-sm text-destructive">{error}</div>
                </div>
              )}

              <div>
                <Input
                  label="Organization username"
                  placeholder="my-awesome-org"
                  error={errors.login?.message}
                  {...register('login', {
                    required: 'Organization username is required',
                    pattern: {
                      value: /^[a-zA-Z0-9._-]+$/,
                      message: 'Username can only contain letters, numbers, dots, hyphens, and underscores',
                    },
                    minLength: {
                      value: 1,
                      message: 'Username must be at least 1 character',
                    },
                    maxLength: {
                      value: 39,
                      message: 'Username must be less than 40 characters',
                    },
                  })}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  This will be your organization&apos;s username on the platform.
                </p>
              </div>

              <div>
                <Input
                  label="Organization name"
                  placeholder="My Awesome Organization"
                  error={errors.name?.message}
                  {...register('name', {
                    required: 'Organization name is required',
                    minLength: {
                      value: 1,
                      message: 'Organization name must be at least 1 character',
                    },
                    maxLength: {
                      value: 100,
                      message: 'Organization name must be less than 100 characters',
                    },
                  })}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  This is the display name for your organization.
                </p>
              </div>

              <div>
                <Input
                  label="Description (optional)"
                  placeholder="A short description of your organization"
                  error={errors.description?.message}
                  {...register('description', {
                    maxLength: {
                      value: 500,
                      message: 'Description must be less than 500 characters',
                    },
                  })}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Help others understand what your organization is about.
                </p>
              </div>

              <div className="border-t pt-6">
                <h3 className="text-lg font-medium mb-3">Organization features</h3>
                <div className="space-y-3 text-sm text-muted-foreground">
                  <div className="flex items-center space-x-2">
                    <svg className="w-4 h-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    <span>Unlimited public repositories</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <svg className="w-4 h-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    <span>Team and member management</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <svg className="w-4 h-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                    <span>Advanced permissions and access controls</span>
                  </div>
                </div>
              </div>

              <div className="flex items-center space-x-3 pt-4">
                <Button
                  type="submit"
                  loading={isLoading}
                  disabled={isLoading || !organizationLogin}
                  className="flex-1"
                >
                  Create organization
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => router.back()}
                  disabled={isLoading}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}