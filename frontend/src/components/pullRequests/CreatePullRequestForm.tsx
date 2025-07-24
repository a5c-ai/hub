'use client'

import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { pullRequestApi } from '../../lib/pullRequestApi'
import { CreatePullRequestRequest } from '../../types'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Card } from '../ui/Card'

interface CreatePullRequestFormProps {
  owner: string
  repo: string
  defaultHead?: string
  defaultBase?: string
}

export function CreatePullRequestForm({ 
  owner, 
  repo, 
  defaultHead = '', 
  defaultBase = 'main' 
}: CreatePullRequestFormProps) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [branches, setBranches] = useState<string[]>([])
  
  const [formData, setFormData] = useState<CreatePullRequestRequest>({
    title: '',
    body: '',
    head: defaultHead,
    base: defaultBase,
    draft: false,
    maintainer_can_modify: true
  })

  useEffect(() => {
    // In a real implementation, we would fetch available branches
    setBranches(['main', 'develop', 'feature/example', 'hotfix/urgent'])
  }, [owner, repo])

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
      
      const pullRequest = await pullRequestApi.createPullRequest(owner, repo, formData)
      
      // Redirect to the created pull request
      router.push(`/${owner}/${repo}/pull/${pullRequest.issue.number}`)
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

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Create a new pull request</h1>
        <p className="text-gray-600 mt-2">
          Create a pull request to propose changes to {owner}/{repo}
        </p>
      </div>

      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Branch Selection */}
        <Card className="p-6">
          <h3 className="text-lg font-medium mb-4">Comparing changes</h3>
          
          <div className="flex items-center space-x-4">
            <div className="flex-1">
              <label htmlFor="base" className="block text-sm font-medium text-gray-700 mb-2">
                Base branch
              </label>
              <select
                id="base"
                value={formData.base}
                onChange={(e) => handleInputChange('base', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={loading}
              >
                {branches.map(branch => (
                  <option key={branch} value={branch}>{branch}</option>
                ))}
              </select>
            </div>
            
            <div className="flex items-center text-gray-500">
              ←
            </div>
            
            <div className="flex-1">
              <label htmlFor="head" className="block text-sm font-medium text-gray-700 mb-2">
                Compare branch
              </label>
              <select
                id="head"
                value={formData.head}
                onChange={(e) => handleInputChange('head', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={loading}
              >
                <option value="">Select a branch</option>
                {branches.map(branch => (
                  <option key={branch} value={branch}>{branch}</option>
                ))}
              </select>
            </div>
          </div>
          
          <p className="text-sm text-gray-600 mt-2">
            Choose two branches to see what&apos;s changed or to start a new pull request.
          </p>
        </Card>

        {/* Pull Request Details */}
        <Card className="p-6">
          <div className="space-y-4">
            <div>
              <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
                Title <span className="text-red-500">*</span>
              </label>
              <Input
                id="title"
                type="text"
                value={formData.title}
                onChange={(e) => handleInputChange('title', e.target.value)}
                placeholder="Enter pull request title"
                disabled={loading}
                className="w-full"
              />
            </div>

            <div>
              <label htmlFor="body" className="block text-sm font-medium text-gray-700 mb-2">
                Description
              </label>
              <textarea
                id="body"
                value={formData.body}
                onChange={(e) => handleInputChange('body', e.target.value)}
                placeholder="Describe your changes (optional)"
                rows={8}
                disabled={loading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-vertical"
              />
              <p className="text-sm text-gray-500 mt-2">
                You can use Markdown to format your description.
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
                className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
              />
              <span className="ml-2 text-sm text-gray-700">
                Create as draft pull request
              </span>
            </label>
            <p className="text-sm text-gray-500 ml-6">
              Draft pull requests cannot be merged until they are marked as ready for review.
            </p>

            <label className="flex items-center">
              <input
                type="checkbox"
                checked={formData.maintainer_can_modify}
                onChange={(e) => handleInputChange('maintainer_can_modify', e.target.checked)}
                disabled={loading}
                className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
              />
              <span className="ml-2 text-sm text-gray-700">
                Allow edits by maintainers
              </span>
            </label>
            <p className="text-sm text-gray-500 ml-6">
              Maintainers will be able to edit your pull request and push commits to your branch.
            </p>
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
    </div>
  )
}