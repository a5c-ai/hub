'use client'

import React, { useState } from 'react'
import { PullRequestFile } from '../../types'
import { Card } from '../ui/Card'
import { Badge } from '../ui/Badge'

interface PullRequestFilesProps {
  files: PullRequestFile[]
  owner: string
  repo: string
  number: number
}

export function PullRequestFiles({ files, owner, repo, number }: PullRequestFilesProps) {
  const [expandedFiles, setExpandedFiles] = useState<Set<string>>(new Set())

  const toggleFileExpansion = (fileId: string) => {
    const newExpanded = new Set(expandedFiles)
    if (newExpanded.has(fileId)) {
      newExpanded.delete(fileId)
    } else {
      newExpanded.add(fileId)
    }
    setExpandedFiles(newExpanded)
  }

  const getFileStatusColor = (status: string) => {
    switch (status) {
      case 'added':
        return 'bg-green-100 text-green-800'
      case 'deleted':
        return 'bg-red-100 text-red-800'
      case 'modified':
        return 'bg-blue-100 text-blue-800'
      case 'renamed':
        return 'bg-yellow-100 text-yellow-800'
      case 'copied':
        return 'bg-purple-100 text-purple-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getFileStatusText = (status: string) => {
    switch (status) {
      case 'added':
        return 'Added'
      case 'deleted':
        return 'Deleted'
      case 'modified':
        return 'Modified'
      case 'renamed':
        return 'Renamed'
      case 'copied':
        return 'Copied'
      default:
        return status
    }
  }

  const formatPatch = (patch: string) => {
    if (!patch) return null

    const lines = patch.split('\n')
    return lines.map((line, index) => {
      let className = 'font-mono text-sm px-4 py-1 '
      let prefix = ''

      if (line.startsWith('+')) {
        className += 'bg-green-50 text-green-900'
        prefix = '+'
      } else if (line.startsWith('-')) {
        className += 'bg-red-50 text-red-900'
        prefix = '-'
      } else if (line.startsWith('@@')) {
        className += 'bg-blue-50 text-blue-900 font-medium'
      } else {
        className += 'bg-gray-50 text-gray-700'
      }

      return (
        <div key={index} className={className}>
          <span className="select-none text-gray-400 mr-2 w-8 inline-block text-right">
            {index + 1}
          </span>
          {line}
        </div>
      )
    })
  }

  if (files.length === 0) {
    return (
      <Card className="p-8 text-center">
        <p className="text-gray-500">No files changed in this pull request</p>
      </Card>
    )
  }

  const totalAdditions = files.reduce((sum, file) => sum + file.additions, 0)
  const totalDeletions = files.reduce((sum, file) => sum + file.deletions, 0)

  return (
    <div className="space-y-4">
      {/* Summary */}
      <Card className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <span className="text-sm font-medium text-gray-900">
              {files.length} {files.length === 1 ? 'file' : 'files'} changed
            </span>
            <div className="flex items-center space-x-4 text-sm text-gray-600">
              <span className="flex items-center text-green-600">
                <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
                </svg>
                {totalAdditions} additions
              </span>
              <span className="flex items-center text-red-600">
                <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M5 10a1 1 0 011-1h8a1 1 0 110 2H6a1 1 0 01-1-1z" clipRule="evenodd" />
                </svg>
                {totalDeletions} deletions
              </span>
            </div>
          </div>
        </div>
      </Card>

      {/* Files List */}
      <div className="space-y-3">
        {files.map((file) => {
          const isExpanded = expandedFiles.has(file.id)
          const filename = file.filename
          const extension = filename.split('.').pop()?.toLowerCase()

          return (
            <Card key={file.id} className="overflow-hidden">
              {/* File Header */}
              <div 
                className="flex items-center justify-between p-4 bg-gray-50 border-b cursor-pointer hover:bg-gray-100 transition-colors"
                onClick={() => toggleFileExpansion(file.id)}
              >
                <div className="flex items-center space-x-3 min-w-0 flex-1">
                  <div className="flex items-center space-x-2">
                    <Badge className={getFileStatusColor(file.status)}>
                      {getFileStatusText(file.status)}
                    </Badge>
                    {file.status === 'renamed' && file.previous_filename && (
                      <span className="text-sm text-gray-500">
                        {file.previous_filename} → 
                      </span>
                    )}
                  </div>
                  
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {filename}
                    </p>
                    {file.status === 'renamed' && file.previous_filename && (
                      <p className="text-xs text-gray-500">
                        Previously: {file.previous_filename}
                      </p>
                    )}
                  </div>
                </div>

                <div className="flex items-center space-x-4 text-sm text-gray-600">
                  {file.additions > 0 && (
                    <span className="text-green-600 font-medium">
                      +{file.additions}
                    </span>
                  )}
                  {file.deletions > 0 && (
                    <span className="text-red-600 font-medium">
                      -{file.deletions}
                    </span>
                  )}
                  <div className="flex items-center">
                    <svg
                      className={`w-5 h-5 transform transition-transform ${
                        isExpanded ? 'rotate-180' : 'rotate-0'
                      }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                </div>
              </div>

              {/* File Content */}
              {isExpanded && (
                <div className="overflow-x-auto">
                  {file.patch ? (
                    <div className="border-l-4 border-gray-200">
                      {formatPatch(file.patch)}
                    </div>
                  ) : (
                    <div className="p-4 text-center text-gray-500">
                      {file.status === 'added' && 'File added'}
                      {file.status === 'deleted' && 'File deleted'}
                      {!['added', 'deleted'].includes(file.status) && 'No changes to display'}
                    </div>
                  )}
                </div>
              )}
            </Card>
          )
        })}
      </div>
    </div>
  )
}