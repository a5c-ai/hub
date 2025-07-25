'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { Button, Input, Card, CardHeader, CardTitle, CardContent } from '@/components/ui';
import { useRepositoryStore } from '@/store/repository';

interface CreateRepositoryFormData {
  name: string;
  description?: string;
  private: boolean;
  auto_init?: boolean;
}

export default function NewRepositoryPage() {
  const router = useRouter();
  const { createRepository, isLoading, error, clearError } = useRepositoryStore();
  
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
  } = useForm<CreateRepositoryFormData>({
    defaultValues: {
      private: false,
      auto_init: true,
    },
  });

  const repositoryName = watch('name');

  const onSubmit = async (data: CreateRepositoryFormData) => {
    try {
      clearError();
      console.log('Creating repository:', data);
      const repository = await createRepository({
        name: data.name,
        description: data.description,
        private: data.private,
        auto_init: data.auto_init,
      });
      console.log('Repository created successfully:', repository);
      
      // Build the repository URL - use owner info if available
      const repoUrl = repository.full_name || `admin/${repository.name}`;
      console.log('Navigating to:', `/repositories/${repoUrl}`);
      router.push(`/repositories/${repoUrl}`);
    } catch (err) {
      console.error('Failed to create repository:', err);
      // Error is handled by the store
    }
  };

  return (
    <AppLayout>
      <div className="container mx-auto px-4 py-8 max-w-2xl">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">
            Create a new repository
          </h1>
          <p className="text-muted-foreground">
            A repository contains all project files, including the revision history.
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Repository Details</CardTitle>
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
                  label="Repository name"
                  placeholder="my-awesome-project"
                  error={errors.name?.message}
                  {...register('name', {
                    required: 'Repository name is required',
                    pattern: {
                      value: /^[a-zA-Z0-9._-]+$/,
                      message: 'Repository name can only contain letters, numbers, dots, hyphens, and underscores',
                    },
                    minLength: {
                      value: 1,
                      message: 'Repository name must be at least 1 character',
                    },
                    maxLength: {
                      value: 100,
                      message: 'Repository name must be less than 100 characters',
                    },
                  })}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Great repository names are short and memorable.
                </p>
              </div>

              <div>
                <Input
                  label="Description (optional)"
                  placeholder="A short description of your repository"
                  error={errors.description?.message}
                  {...register('description', {
                    maxLength: {
                      value: 500,
                      message: 'Description must be less than 500 characters',
                    },
                  })}
                />
              </div>

              <div className="space-y-4">
                <h3 className="text-lg font-medium">Repository settings</h3>
                
                <div className="space-y-3">
                  <div className="flex items-start space-x-3">
                    <input
                      id="visibility-public"
                      type="radio"
                      value="false"
                      className="mt-1"
                      {...register('private')}
                    />
                    <div className="flex-1">
                      <label htmlFor="visibility-public" className="block text-sm font-medium text-foreground">
                        Public
                      </label>
                      <p className="text-xs text-muted-foreground">
                        Anyone on the internet can see this repository. You choose who can commit.
                      </p>
                    </div>
                  </div>

                  <div className="flex items-start space-x-3">
                    <input
                      id="visibility-private"
                      type="radio"
                      value="true"
                      className="mt-1"
                      {...register('private')}
                    />
                    <div className="flex-1">
                      <label htmlFor="visibility-private" className="block text-sm font-medium text-foreground">
                        Private
                      </label>
                      <p className="text-xs text-muted-foreground">
                        You choose who can see and commit to this repository.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <div className="border-t pt-4">
                <div className="flex items-center space-x-3">
                  <input
                    id="auto-init"
                    type="checkbox"
                    className="rounded border-border"
                    {...register('auto_init')}
                  />
                  <div className="flex-1">
                    <label htmlFor="auto-init" className="block text-sm font-medium text-foreground">
                      Initialize this repository with a README
                    </label>
                    <p className="text-xs text-muted-foreground">
                      This will create an initial commit with a README file.
                    </p>
                  </div>
                </div>
              </div>

              <div className="flex items-center space-x-3 pt-4">
                <Button
                  type="submit"
                  loading={isLoading}
                  disabled={isLoading || !repositoryName}
                  className="flex-1"
                >
                  Create repository
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