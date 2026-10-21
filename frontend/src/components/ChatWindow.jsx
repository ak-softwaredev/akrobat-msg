import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '../api/client'
import { useWebSocket } from '../hooks/useWebSocket'
import MessageList from './MessageList'
import MessageInput from './MessageInput'

export default function ChatWindow({ chatId, currentUserId }) {
  const [chat, setChat] = useState(null)
  const [messages, setMessages] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const token = localStorage.getItem('token')

  // Load chat info + initial messages
  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError('')

    Promise.all([api.getChat(chatId), api.getMessages(chatId)])
      .then(([chatData, messagesData]) => {
        if (cancelled) return
        setChat(chatData)
        // Backend returns newest-first; reverse for chronological display
        setMessages([...(messagesData ?? [])].reverse())
      })
      .catch((err) => {
        if (!cancelled) setError(err.message)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [chatId])

  // WebSocket event handler
  const handleWsEvent = useCallback(
    (event) => {
      const { type, data } = event
      if (type === 'message.created' && data) {
        setMessages((prev) => {
          if (prev.some((m) => m.id === data.id)) return prev
          return [...prev, data]
        })
      }
      if (type === 'message.updated' && data) {
        setMessages((prev) => prev.map((m) => (m.id === data.id ? data : m)))
      }
      if (type === 'message.deleted' && data) {
        const id = typeof data === 'object' ? data.id : data
        setMessages((prev) => prev.filter((m) => m.id !== id))
      }
    },
    [],
  )

  useWebSocket(chatId, token, handleWsEvent)

  async function handleSend(content) {
    try {
      const msg = await api.postMessage(chatId, content)
      // Optimistically add; WS will dedupe
      setMessages((prev) =>
        prev.some((m) => m.id === msg.id) ? prev : [...prev, msg],
      )
    } catch (err) {
      alert(err.message)
    }
  }

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center text-gray-400">
        Loading chat…
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center text-red-500">
        {error}
      </div>
    )
  }

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <header className="px-5 py-3 border-b bg-white">
        <h2 className="font-semibold text-lg">{chat?.name}</h2>
        {chat?.description && (
          <p className="text-sm text-gray-500">{chat.description}</p>
        )}
        {chat?.members && (
          <p className="text-xs text-gray-400 mt-1">
            {chat.members.length} member{chat.members.length === 1 ? '' : 's'}
          </p>
        )}
      </header>

      <MessageList messages={messages} currentUserId={currentUserId} />

      <MessageInput onSend={handleSend} />
    </div>
  )
}