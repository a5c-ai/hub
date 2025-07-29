import { useEffect, useState } from 'react'

/**
 * useNotifications opens a WebSocket to receive real-time notifications.
 * Returns an array of incoming notifications in reverse chronological order.
 */
export function useNotifications<T = any>(): T[] {
  const [notifications, setNotifications] = useState<T[]>([])

  useEffect(() => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || ''
    const wsScheme = apiUrl.startsWith('https') ? 'wss' : 'ws'
    const wsUrl = apiUrl.replace(/^https?/, wsScheme) + '/notifications/subscribe'
    const ws = new WebSocket(wsUrl)

    ws.onmessage = (event) => {
      try {
        const notif = JSON.parse(event.data) as T
        setNotifications((prev) => [notif, ...prev])
      } catch {
        // ignore parse errors
      }
    }
    ws.onerror = () => {
      ws.close()
    }

    return () => {
      ws.close()
    }
  }, [])

  return notifications
}
