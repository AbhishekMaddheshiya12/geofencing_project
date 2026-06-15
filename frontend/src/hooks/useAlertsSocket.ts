import { useEffect, useRef, useState } from 'react'
import type { AlertEnvelope } from '../api/types'

function getWsURL() {
  const apiBase =
    (import.meta as any).env?.VITE_API_BASE || 'http://localhost:8080'
  const wsProtocol = apiBase.startsWith('https') ? 'wss:' : 'ws:'
  const apiHost = new URL(apiBase).host
  return `${wsProtocol}//${apiHost}/ws/alerts`
}

const MAX_FEED = 100

export function useAlertsSocket(onAlert?: (a: AlertEnvelope) => void) {
  const [connected, setConnected] = useState(false)
  const [feed, setFeed] = useState<AlertEnvelope[]>([])
  const wsRef = useRef<WebSocket | null>(null)
  const backoffRef = useRef(1000)
  const onAlertRef = useRef(onAlert)
  onAlertRef.current = onAlert

  useEffect(() => {
    let stopped = false
    function connect() {
      const ws = new WebSocket(getWsURL())
      wsRef.current = ws
      ws.onopen = () => {
        setConnected(true)
        backoffRef.current = 1000
      }
      ws.onmessage = (e) => {
        try {
          const alert = JSON.parse(e.data) as AlertEnvelope
          setFeed((prev) => [alert, ...prev].slice(0, MAX_FEED))
          onAlertRef.current?.(alert)
        } catch {
          // ignore malformed frames
        }
      }
      ws.onclose = () => {
        setConnected(false)
        if (stopped) return
        const delay = Math.min(backoffRef.current, 30000)
        backoffRef.current = Math.min(backoffRef.current * 2, 30000)
        setTimeout(connect, delay)
      }
      ws.onerror = () => ws.close()
    }
    connect()
    return () => {
      stopped = true
      wsRef.current?.close()
    }
  }, [])

  return { connected, feed }
}
