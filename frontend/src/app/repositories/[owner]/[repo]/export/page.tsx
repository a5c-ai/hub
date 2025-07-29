"use client"

import React from "react"
import { useParams } from "next/navigation"

export default function ExportPage() {
  const { owner, repo } = useParams()
  return (
    <div>
      <h1>Export Repository: {owner}/{repo}</h1>
      <p>Export functionality is coming soon.</p>
    </div>
  )
}
