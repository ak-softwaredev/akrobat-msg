import { createContext, useContext, useEffect, useState } from 'react'
import { api, setToken } from '../api/client'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      setLoading(false)
      return
    }
    api
      .getMe()
      .then(setUser)
      .catch(() => setToken(null))
      .finally(() => setLoading(false))
  }, [])

  async function login(username, password) {
    const res = await api.login(username, password)
    setToken(res.auth_token)
    setUser({ id: res.id, username: res.username })
    return res
  }

  async function register(username, password) {
    const res = await api.register(username, password)
    setToken(res.auth_token)
    setUser({ id: res.id, username: res.username })
    return res
  }

  async function logout() {
    try {
      await api.logout()
    } catch {
      // ignore network errors on logout
    }
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, setUser, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}