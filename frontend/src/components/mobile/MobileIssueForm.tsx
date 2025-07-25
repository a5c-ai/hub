'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import {
  PhotoIcon,
  PaperClipIcon,
  EyeIcon,
  PencilIcon,
  TagIcon,
  UserIcon,
} from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

interface IssueFormData {
  title: string;
  description: string;
  labels: string[];
  assignees: string[];
}

interface MobileIssueFormProps {
  onSubmit: (data: IssueFormData) => void;
  loading?: boolean;
  initialData?: Partial<IssueFormData>;
  availableLabels?: Array<{ id: string; name: string; color: string }>;
  availableAssignees?: Array<{ id: string; username: string; name: string }>;
}

export function MobileIssueForm({
  onSubmit,
  loading = false,
  initialData,
  availableLabels = [],
  availableAssignees = [],
}: MobileIssueFormProps) {
  const [isPreview, setIsPreview] = useState(false);
  const [selectedLabels, setSelectedLabels] = useState<string[]>(initialData?.labels || []);
  const [selectedAssignees, setSelectedAssignees] = useState<string[]>(initialData?.assignees || []);
  const [showLabelPicker, setShowLabelPicker] = useState(false);
  const [showAssigneePicker, setShowAssigneePicker] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<IssueFormData>({
    defaultValues: {
      title: initialData?.title || '',
      description: initialData?.description || '',
      labels: selectedLabels,
      assignees: selectedAssignees,
    },
  });

  const watchedDescription = watch('description');

  const onFormSubmit = (data: IssueFormData) => {
    onSubmit({
      ...data,
      labels: selectedLabels,
      assignees: selectedAssignees,
    });
  };

  const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      // Handle image upload logic here
      console.log('Image upload:', file);
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      // Handle file upload logic here
      console.log('File upload:', file);
    }
  };

  const toggleLabel = (labelId: string) => {
    const newLabels = selectedLabels.includes(labelId)
      ? selectedLabels.filter(id => id !== labelId)
      : [...selectedLabels, labelId];
    setSelectedLabels(newLabels);
    setValue('labels', newLabels);
  };

  const toggleAssignee = (assigneeId: string) => {
    const newAssignees = selectedAssignees.includes(assigneeId)
      ? selectedAssignees.filter(id => id !== assigneeId)
      : [...selectedAssignees, assigneeId];
    setSelectedAssignees(newAssignees);
    setValue('assignees', newAssignees);
  };

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="flex flex-col h-full">
      {/* Header */}
      <div className="sticky top-0 z-10 bg-background border-b border-border p-4">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-lg font-semibold">New Issue</h1>
          <div className="flex items-center space-x-2">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsPreview(!isPreview)}
              className={cn('flex items-center space-x-1', isPreview && 'bg-muted')}
            >
              {isPreview ? (
                <>
                  <PencilIcon className="h-4 w-4" />
                  <span>Edit</span>
                </>
              ) : (
                <>
                  <EyeIcon className="h-4 w-4" />
                  <span>Preview</span>
                </>
              )}
            </Button>
            <Button type="submit" disabled={loading} size="sm">
              {loading ? 'Creating...' : 'Create'}
            </Button>
          </div>
        </div>

        {/* Title input */}
        <Input
          {...register('title', { required: 'Title is required' })}
          placeholder="Issue title"
          className="w-full"
        />
        {errors.title && (
          <p className="text-destructive text-sm mt-1">{errors.title.message}</p>
        )}
      </div>

      {/* Content area */}
      <div className="flex-1 overflow-y-auto">
        <div className="p-4 space-y-4">
          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              Description
            </label>
            {isPreview ? (
              <Card className="p-4 min-h-[200px]">
                <div className="prose prose-sm max-w-none">
                  {watchedDescription ? (
                    <div className="whitespace-pre-wrap">{watchedDescription}</div>
                  ) : (
                    <p className="text-muted-foreground italic">Nothing to preview</p>
                  )}
                </div>
              </Card>
            ) : (
              <div>
                <textarea
                  {...register('description')}
                  placeholder="Describe the issue..."
                  className="w-full min-h-[200px] p-3 rounded-md border border-input bg-background text-sm resize-none focus:outline-none focus:ring-2 focus:ring-ring"
                />
                
                {/* Toolbar */}
                <div className="flex items-center justify-between mt-2 pt-2 border-t border-border">
                  <div className="flex items-center space-x-2">
                    <label htmlFor="image-upload">
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        asChild
                      >
                        <span>
                          <PhotoIcon className="h-4 w-4" />
                        </span>
                      </Button>
                    </label>
                    <input
                      id="image-upload"
                      type="file"
                      accept="image/*"
                      onChange={handleImageUpload}
                      className="hidden"
                    />
                    
                    <label htmlFor="file-upload">
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        asChild
                      >
                        <span>
                          <PaperClipIcon className="h-4 w-4" />
                        </span>
                      </Button>
                    </label>
                    <input
                      id="file-upload"
                      type="file"
                      onChange={handleFileUpload}
                      className="hidden"
                    />
                  </div>
                  
                  <div className="text-xs text-muted-foreground">
                    Markdown supported
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Labels */}
          <div>
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowLabelPicker(!showLabelPicker)}
              className="w-full justify-start mb-2"
            >
              <TagIcon className="h-4 w-4 mr-2" />
              Labels ({selectedLabels.length})
            </Button>
            
            {selectedLabels.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-2">
                {selectedLabels.map(labelId => {
                  const label = availableLabels.find(l => l.id === labelId);
                  return label ? (
                    <Badge key={label.id} variant="secondary">
                      {label.name}
                    </Badge>
                  ) : null;
                })}
              </div>
            )}
            
            {showLabelPicker && (
              <Card className="p-3">
                <div className="space-y-2">
                  {availableLabels.map(label => (
                    <button
                      key={label.id}
                      type="button"
                      onClick={() => toggleLabel(label.id)}
                      className={cn(
                        'w-full text-left p-2 rounded border transition-colors',
                        selectedLabels.includes(label.id)
                          ? 'bg-primary/10 border-primary'
                          : 'border-border hover:bg-muted'
                      )}
                    >
                      <Badge
                        variant="secondary"
                        style={{ backgroundColor: label.color }}
                      >
                        {label.name}
                      </Badge>
                    </button>
                  ))}
                </div>
              </Card>
            )}
          </div>

          {/* Assignees */}
          <div>
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowAssigneePicker(!showAssigneePicker)}
              className="w-full justify-start mb-2"
            >
              <UserIcon className="h-4 w-4 mr-2" />
              Assignees ({selectedAssignees.length})
            </Button>
            
            {selectedAssignees.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-2">
                {selectedAssignees.map(assigneeId => {
                  const assignee = availableAssignees.find(a => a.id === assigneeId);
                  return assignee ? (
                    <Badge key={assignee.id} variant="secondary">
                      {assignee.name || assignee.username}
                    </Badge>
                  ) : null;
                })}
              </div>
            )}
            
            {showAssigneePicker && (
              <Card className="p-3">
                <div className="space-y-2">
                  {availableAssignees.map(assignee => (
                    <button
                      key={assignee.id}
                      type="button"
                      onClick={() => toggleAssignee(assignee.id)}
                      className={cn(
                        'w-full text-left p-2 rounded border transition-colors flex items-center space-x-2',
                        selectedAssignees.includes(assignee.id)
                          ? 'bg-primary/10 border-primary'
                          : 'border-border hover:bg-muted'
                      )}
                    >
                      <div className="h-6 w-6 rounded-full bg-muted flex items-center justify-center text-xs">
                        {(assignee.name || assignee.username).charAt(0).toUpperCase()}
                      </div>
                      <span>{assignee.name || assignee.username}</span>
                    </button>
                  ))}
                </div>
              </Card>
            )}
          </div>
        </div>
      </div>
    </form>
  );
}