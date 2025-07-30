'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { pullRequestApi } from '../../lib/pullRequestApi'
import { CreatePullRequestRequest, Repository } from '../../types'
import { apiClient, searchApi } from '../../lib/api'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Card } from '../ui/Card'

interface CreatePullRequestFormProps {
  repositoryOwner: string
  repositoryName: string
  defaultHead?: string
  defaultBase?: string
}

export function CreatePullRequestForm({ 
  repositoryOwner, 
  repositoryName, 
  defaultHead = '', 
  defaultBase = 'main' 
}: CreatePullRequestFormProps) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [branches, setBranches] = useState<string[]>([])
  const [repos, setRepos] = useState<Repository[]>([])
  const [headRepoName, setHeadRepoName] = useState<string>(repositoryName)
  
  const [formData, setFormData] = useState<CreatePullRequestRequest>({
    title: '',
    body: '',
    head: defaultHead,
    base: defaultBase,
    head_repository_id: undefined,
    draft: false,
    maintainer_can_modify: true
  })

  // load repositories in this owner/org for head repo selection
  useEffect(() => {
    (async () => {
      try {
        const resp = await searchApi.searchRepositories('', { user: repositoryOwner, per_page: 100 })
        const items = (resp.data as any).items as Repository[]
        setRepos(items)
      } catch {
        setRepos([])
      }
    })()
  }, [repositoryOwner])

  // update active head repo name when selection changes
  useEffect(() => {
    if (formData.head_repository_id) {
      const sel = repos.find(r => r.id === formData.head_repository_id)
      setHeadRepoName(sel ? sel.name : repositoryName)
    } else {
      setHeadRepoName(repositoryName)
    }
  }, [formData.head_repository_id, repos, repositoryName])

  // fetch branches for selected head repository
  useEffect(() => {
    let mounted = true
    ;(async () => {
      try {
        const resp = await apiClient.get<string[]>(`/repositories/${repositoryOwner}/${headRepoName}/branches`)
        if (mounted) setBranches(resp.data)
      } catch {
        if (mounted) setBranches([])
      }
    })()
    return () => { mounted = false }
  }, [repositoryOwner, headRepoName])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!formData.title.trim()) {
      setError('Title is required')
      return
    }
    
    if (!formData.head) {
      setError('Head branch is required')
      return
    }
    
    if (!formData.base) {
      setError('Base branch is required')
      return
    }
    
    if (formData.head === formData.base) {
      setError('Head and base branches cannot be the same')
      return
    }

    try {
      setLoading(true)
      setError(null)
      
      const pullRequest = await pullRequestApi.createPullRequest(repositoryOwner, repositoryName, formData)
      
      // Redirect to the created pull request
      router.push(`/${repositoryOwner}/${repositoryName}/pull/${pullRequest.issue.number}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create pull request')
    } finally {
      setLoading(false)
    }
  }

  const handleInputChange = (field: keyof CreatePullRequestRequest, value: string | boolean) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    if (error) setError(null)
  }

  return <div className="max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-foreground">Create a new pull request</h1>
        <p className="text-muted-foreground mt-2">
          Create a pull request to propose and collaborate on changes to a repository.
        </p>
      </div>

      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Repository Selection */}
        <Card className="p-6">
          <label htmlFor="head_repository_id" className="block text-sm font-medium text-foreground mb-2">
            Compare repository
          </label>
          <select
            id="head_repository_id"
            value={formData.head_repository_id ?? repositoryName}
            onChange={(e) => handleInputChange('head_repository_id', e.target.value)}
            disabled={loading}
            className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground"
          >
            <option value={repositoryName}>{repositoryName}</option>
            {repos.map(repo => (
              <option key={repo.id} value={repo.id}>{repo.name}</option>
            ))}
          </select>
          <p className="text-sm text-muted-foreground mt-2">
            Select the repository containing your changes
          </p>
        </Card>
        {/* Branch Selection */}
        <Card className="p-6">
          <h3 className="text-lg font-medium mb-4">Comparing changes</h3>

          <div className="flex items-center space-x-4">
            <div className="flex-1">
              <label htmlFor="base" className="block text-sm font-medium text-foreground mb-2">
                Base branch
              </label>
              <select
                id="base"
                value={formData.base}
                onChange={(e) => handleInputChange('base', e.target.value)}
                className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground"
                disabled={loading}
              >
                {branches.map(branch => <option key={branch} value={branch}>{branch}</option>)}
              </select>
            </div>
            
            <div className="flex items-center text-muted-foreground">
              <svg className="w-4 h-4 mx-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M12.293 5.293a1 1 0 011.414 0l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-2.293-2.293a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
            </div>
            
            <div className="flex-1">
              <label htmlFor="head" className="block text-sm font-medium text-foreground mb-2">
                Compare branch
              </label>
              <select
                id="head"
                value={formData.head}
                onChange={(e) => handleInputChange('head', e.target.value)}
                className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent bg-background text-foreground"
                disabled={loading}
              >
                <option value="">Select a branch</option>
                {branches.map(branch => <option key={branch} value={branch}>{branch}</option>)}
              </select>
            </div>
          </div>
          
          <p className="text-sm text-muted-foreground mt-2">
            Choose a branch that contains your changes
          </p>
        </Card>

        {/* Pull Request Details */}
        <Card className="p-6">
          <div className="space-y-4">
            <div>
              <label htmlFor="title" className="block text-sm font-medium text-foreground mb-2">
                Title <span className="text-red-500">*</span>
              </label>
              <Input
                id="title"
                type="text"
                value={formData.title}
                onChange={(e) => handleInputChange('title', e.target.value)}
                placeholder="Brief description of your changes"
                disabled={loading}
                className="w-full"
              />
            </div>

            <div>
              <label htmlFor="body" className="block text-sm font-medium text-foreground mb-2 mt-4">
                Description
              </label>
              <textarea
                id="body"
                value={formData.body}
                onChange={(e) => handleInputChange('body', e.target.value)}
                placeholder="Provide a detailed description of your changes..."
                rows={6}
                disabled={loading}
                className="w-full px-3 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent resize-vertical bg-background text-foreground placeholder:text-muted-foreground"
              />
              <p className="text-sm text-muted-foreground mt-2">
                You can use markdown to format your description
              </p>
            </div>
          </div>
        </Card>

        {/* Options */}
        <Card className="p-6">
          <h3 className="text-lg font-medium mb-4">Options</h3>
          
          <div className="space-y-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={formData.draft}
                onChange={(e) => handleInputChange('draft', e.target.checked)}
                disabled={loading}
                className="rounded border-border text-primary shadow-sm focus:border-ring focus:ring focus:ring-ring/20"
              />
              <span className="ml-2 text-sm text-foreground">
                Create as draft
              </span>
              <p className="text-sm text-muted-foreground ml-6">
                Draft pull requests cannot be merged and do not request reviews
              </p>
            </label>

            <label className="flex items-center">
              <input
                type="checkbox"
                checked={formData.maintainer_can_modify}
                onChange={(e) => handleInputChange('maintainer_can_modify', e.target.checked)}
                disabled={loading}
                className="rounded border-border text-primary shadow-sm focus:border-ring focus:ring focus:ring-ring/20"
              />
              <span className="ml-2 text-sm text-foreground">
                Allow edits by maintainers
              </span>
              <p className="text-sm text-muted-foreground ml-6">
                Maintainers can edit and update your pull request
              </p>
            </label>
          </div>
        </Card>

        {/* Actions */}
        <div className="flex justify-end space-x-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => router.back()}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={loading || !formData.title.trim() || !formData.head || !formData.base}
          >
            {loading ? 'Creating...' : 'Create pull request'}
          </Button>
        </div>
      </form>
    </div>;
}
