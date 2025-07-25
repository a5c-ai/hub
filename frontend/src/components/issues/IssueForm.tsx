'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { CreateIssueRequest } from '@/types';
import { useIssueStore } from '@/store/issues';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

interface IssueFormProps {
  repositoryOwner: string;
  repositoryName: string;
  mode?: 'create' | 'edit';
  initialData?: Partial<CreateIssueRequest>;
}

export function IssueForm({ 
  repositoryOwner, 
  repositoryName, 
  mode = 'create',
  initialData = {}
}: IssueFormProps) {
  const router = useRouter();
  const { createIssue, isCreating, operationError } = useIssueStore();
  
  const [formData, setFormData] = useState<CreateIssueRequest>({
    title: initialData.title || '',
    body: initialData.body || '',
    assignee_id: initialData.assignee_id,
    milestone_id: initialData.milestone_id,
    label_ids: initialData.label_ids || [],
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = () => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.title.trim()) {
      newErrors.title = 'Title is required';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      const issue = await createIssue(repositoryOwner, repositoryName, formData);
      router.push(`/repositories/${repositoryOwner}/${repositoryName}/issues/${issue.number}`);
    } catch (error) {
      // Error is handled by the store
      console.error('Failed to create issue:', error);
    }
  };

  const handleCancel = () => {
    router.push(`/repositories/${repositoryOwner}/${repositoryName}/issues`);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Title */}
      <div>
        <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
          Title *
        </label>
        <Input
          id="title"
          value={formData.title}
          onChange={(e) => setFormData({ ...formData, title: e.target.value })}
          placeholder="Brief description of the issue"
          className={errors.title ? 'border-red-300' : ''}
        />
        {errors.title && (
          <p className="mt-1 text-sm text-red-600">{errors.title}</p>
        )}
      </div>

      {/* Body */}
      <div>
        <label htmlFor="body" className="block text-sm font-medium text-gray-700 mb-2">
          Description
        </label>
        <textarea
          id="body"
          rows={10}
          value={formData.body}
          onChange={(e) => setFormData({ ...formData, body: e.target.value })}
          placeholder="Detailed description of the issue..."
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      {/* Form actions */}
              <div className="flex items-center justify-between pt-6 border-t border-border">
        <div>
          {operationError && (
            <p className="text-sm text-red-600">{operationError}</p>
          )}
        </div>
        <div className="flex items-center space-x-3">
          <Button
            type="button"
            variant="outline"
            onClick={handleCancel}
            disabled={isCreating}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={isCreating || !formData.title.trim()}
          >
            {isCreating ? (
              <div className="flex items-center">
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                Creating...
              </div>
            ) : (
              `${mode === 'create' ? 'Create' : 'Update'} Issue`
            )}
          </Button>
        </div>
      </div>
    </form>
  );
}