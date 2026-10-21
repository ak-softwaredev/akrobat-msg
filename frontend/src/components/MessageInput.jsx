import { useState } from 'react'

export default function MessageInput({ onSend }) {
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    const content = text.trim()
    if (!content || busy) return
    setBusy(true)
    setText('')
    try {
      await onSend(content)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="border-t bg-white px-4 py-3 flex gap-2"
    >
      <input
        className="flex-1 border rounded-full px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        placeholder="Type a message…"
        value={text}
        onChange={(e) => setText(e.target.value)}
        disabled={busy}
      />
      <button
        type="submit"
        disabled={!text.trim() || busy}
        className="px-5 py-2 rounded-full bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-40"
      >
        Send
      </button>
    </form>
  )
}