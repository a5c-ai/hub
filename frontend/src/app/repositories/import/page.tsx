"use client"

import React from "react"

export default function ImportPage() {
  const [url, setUrl] = React.useState('')
  const [token, setToken] = React.useState('')
  const [jobId, setJobId] = React.useState<string | null>(null)
  const [status, setStatus] = React.useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setStatus('')
    setJobId(null)
    try {
      const res = await fetch('/api/v1/repositories/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, token }),
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
      const res = await fetch(`/api/v1/repositories/import/${jobId}`)
      if (!res.ok) throw new Error(res.statusText)
      const data = await res.json()
      setStatus(data.status)
    } catch (err: any) {
      setStatus(err.message)
    }
  }

  return (
    <div>
      <h1>Import Repository</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="url">Repository URL:</label>
          <input id="url" value={url} onChange={e => setUrl(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="token">Access Token (optional):</label>
          <input id="token" value={token} onChange={e => setToken(e.target.value)} />
        </div>
        <button type="submit">Start Import</button>
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
