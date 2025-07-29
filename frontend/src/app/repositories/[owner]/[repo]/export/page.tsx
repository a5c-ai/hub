"use client"

import React from "react"
import { useParams } from "next/navigation"

export default function ExportPage() {
  const { owner, repo } = useParams()
  const [remoteURL, setRemoteURL] = React.useState('')
  const [token, setToken] = React.useState('')
  const [jobId, setJobId] = React.useState<string | null>(null)
  const [status, setStatus] = React.useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setStatus('')
    setJobId(null)
    try {
      const res = await fetch(`/api/v1/repositories/${owner}/${repo}/export`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ remote_url: remoteURL, token }),
      })
      if (!res.ok) throw new Error((await res.json()).error || res.statusText)
      const data = await res.json()
      setJobId(data.job_id)
      setStatus('pending')
    } catch (err: any) {
      setStatus(err.message)
    }
  }

  const checkStatus = async () => {
    if (!jobId) return
    try {
      const res = await fetch(`/api/v1/repositories/${owner}/${repo}/export/${jobId}`)
      if (!res.ok) throw new Error(res.statusText)
      const data = await res.json()
      setStatus(data.status)
    } catch (err: any) {
      setStatus(err.message)
    }
  }

  return (
    <div>
      <h1>Export Repository: {owner}/{repo}</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="remoteURL">Destination URL:</label>
          <input id="remoteURL" value={remoteURL} onChange={e => setRemoteURL(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="token">Access Token (optional):</label>
          <input id="token" value={token} onChange={e => setToken(e.target.value)} />
        </div>
        <button type="submit">Start Export</button>
      </form>
      {jobId && (
        <div>
          <p>Job ID: {jobId}</p>
          <button onClick={checkStatus}>Check Status</button>
          <p>Status: {status}</p>
        </div>
      )}
    </div>
  )
}
