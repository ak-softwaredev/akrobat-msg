import { useEffect, useRef } from 'react'

export default function MessageList({ messages, currentUserId }) {
  const bottomRef = useRef(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  if (!messages.length) {
    return (
      <div className="flex-1 flex items-center justify-center text-gray-400 text-sm">
        No messages yet. Say hello!
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto px-5 py-4 space-y-3 bg-gray-50">
      {messages.map((msg) => {
        const isOwn = msg.sender_id === currentUserId
        return (
          <div
            key={msg.id}
            className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[70%] rounded-2xl px-4 py-2 text-sm whitespace-pre-wrap break-words ${
                isOwn
                  ? 'bg-indigo-600 text-white rounded-br-sm'
                  : 'bg-white border rounded-bl-sm'
              }`}
            >
              {msg.content}
            </div>
          </div>
        )
      })}
      <div ref={bottomRef} />
    </div>
  )
}