import { useEffect, useRef } from 'react'

/**
 * Connects to the chat WebSocket and calls onEvent for each message.
 * The backend must accept `?token=` for WS auth (see note in the README).
 */
export function useWebSocket(chatId, token, onEvent) {
  const wsRef = useRef(null)

  useEffect(() => {
    if (!chatId || !token) return

    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    // Use the same host; Vite proxy forwards to the backend
    const url = `${proto}//${window.location.host}/api/v1/chats/${chatId}/ws?token=${encodeURIComponent(
      token,
    )}`

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onmessage = (msg) => {
      try {
        onEvent(JSON.parse(msg.data))
      } catch {
        // ignore non-JSON frames
      }
    }

    ws.onerror = (err) => console.warn('WS error', err)

    return () => {
      ws.close()
      wsRef.current = null
    }
  }, [chatId, token, onEvent])

  return wsRef
}