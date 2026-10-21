import { useEffect, useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../api/client'

export default function ProfilePage() {
  const { user, setUser, logout } = useAuth()
  const [bio, setBio] = useState('')
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [status, setStatus] = useState('')

  useEffect(() => {
    api
      .getMe()
      .then((me) => {
        setBio(me.bio ?? '')
        setUsername(me.username ?? '')
        setUser(me)
      })
      .catch((err) => setError(err.message))
  }, [setUser])

  async function save(field, value) {
    setError('')
    setStatus('')
    try {
      const updated = await api.updateMe(field, value)
      setUser(updated)
      setStatus('Saved')
      setTimeout(() => setStatus(''), 2000)
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleDelete() {
    if (!confirm('Delete your account permanently?')) return
    try {
      await api.deleteMe()
      await logout()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div className="max-w-lg mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Profile</h1>

      {error && <p className="text-sm text-red-600">{error}</p>}
      {status && <p className="text-sm text-green-600">{status}</p>}

      <section className="bg-white rounded-xl shadow p-5 space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1">Username</label>
          <div className="flex gap-2">
            <input
              className="flex-1 border rounded px-3 py-2"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <button
              onClick={() => save('username', username)}
              className="px-4 py-2 bg-indigo-600 text-white rounded text-sm"
            >
              Save
            </button>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">Bio</label>
          <div className="flex gap-2">
            <textarea
              className="flex-1 border rounded px-3 py-2"
              rows={3}
              value={bio}
              onChange={(e) => setBio(e.target.value)}
            />
            <button
              onClick={() => save('bio', bio)}
              className="px-4 py-2 bg-indigo-600 text-white rounded text-sm h-fit"
            >
              Save
            </button>
          </div>
        </div>
      </section>

      <section className="bg-white rounded-xl shadow p-5">
        <h2 className="font-semibold text-red-600 mb-2">Danger zone</h2>
        <p className="text-sm text-gray-600 mb-3">
          Deleting your account is permanent and cannot be undone.
        </p>
        <button
          onClick={handleDelete}
          className="px-4 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700"
        >
          Delete my account
        </button>
      </section>
    </div>
  )
}