import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Lock, AlertCircle, LogIn } from 'lucide-react'
import { type Share } from '@/services/shares'
import { fileAPI } from '@/services/api'
import { useAuth } from '@/contexts/AuthContext'
import axios from 'axios'
import FileViewer from '@/components/FileViewer'

export default function PublicShare() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const { user } = useAuth()

  const [share, setShare] = useState<Share | null>(null)
  const [accessToken, setAccessToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [requiresPassword, setRequiresPassword] = useState(false)
  const [password, setPassword] = useState('')
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const loadingRef = React.useRef(false)

  useEffect(() => {
    if (!id) return

    // Prevent duplicate calls during React StrictMode double-invocation
    if (loadingRef.current) return
    loadingRef.current = true

    // Try to load from session storage first to avoid duplicate API calls
    const sessionKey = `share_data_${id}`
    const cachedData = sessionStorage.getItem(sessionKey)

    if (cachedData) {
      try {
        const { share, access_token } = JSON.parse(cachedData)
        setShare(share)
        setAccessToken(access_token)
        setLoading(false)
        return
      } catch (e) {
        // Invalid cache, load from API
        sessionStorage.removeItem(sessionKey)
      }
    }

    // Load from API if not cached
    loadShare()
  }, [id])

  const loadShare = async () => {
    if (!id) return

    setLoading(true)
    setError(null)
    setPasswordError(null)

    try {
      // Create a separate axios instance without the auth redirect interceptor
      const publicAxios = axios.create({
        baseURL: '/api',
        timeout: 10000,
        withCredentials: true,
      })

      const response = await publicAxios.get<{ share: Share; access_token: string }>(`/s/${id}`, {
        params: password ? { password } : undefined
      })
      setShare(response.data.share)
      setAccessToken(response.data.access_token)

      // Cache the data in session storage to prevent duplicate API calls (and duplicate counting)
      const sessionKey = `share_data_${id}`
      sessionStorage.setItem(sessionKey, JSON.stringify({
        share: response.data.share,
        access_token: response.data.access_token
      }))
    } catch (err: any) {
      const errorData = err.response?.data

      if (err.response?.status === 401 && errorData?.requires_password) {
        setRequiresPassword(true)
        setPasswordError(t('share.passwordRequired'))
      } else if (err.response?.status === 403) {
        if (errorData?.requires_auth || errorData?.error?.includes('login')) {
          // Requires authentication
          setError(t('share.authRequiredMessage'))
        } else if (errorData?.error?.includes('disabled')) {
          // Share is disabled
          setError(t('share.shareDisabled'))
        } else if (errorData?.error?.includes('Maximum views')) {
          // Max views reached
          setError(t('share.maxViewsReached'))
        } else {
          setError(errorData?.error || t('share.accessDenied'))
        }
      } else if (err.response?.status === 404) {
        setError(t('share.notFound'))
      } else if (err.response?.status === 410) {
        setError(t('share.expired'))
      } else {
        setError(errorData?.error || t('share.loadFailed'))
      }
    } finally {
      setLoading(false)
    }
  }

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!password.trim()) {
      setPasswordError(t('share.passwordRequired'))
      return
    }
    await loadShare()
  }

  const handleLogin = () => {
    // Navigate to login with redirect parameter
    navigate(`/login?redirect=${encodeURIComponent(`/s/${id}`)}`)
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full text-center">
          <AlertCircle size={48} className="mx-auto text-red-500 mb-4" />
          <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('share.error')}</h2>
          <p className="text-gray-600 mb-6">{error}</p>

          {!user ? (
            <button
              onClick={handleLogin}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2 mx-auto"
            >
              <LogIn size={18} />
              {t('auth.login')}
            </button>
          ) : (
            <button
              onClick={() => navigate('/')}
              className="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700"
            >
              {t('common.backToHome')}
            </button>
          )}
        </div>
      </div>
    )
  }

  if (requiresPassword && !share) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full">
          <div className="text-center mb-6">
            <Lock size={48} className="mx-auto text-blue-600 mb-4" />
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              {t('share.passwordRequired')}
            </h2>
            <p className="text-gray-600">{t('share.enterPassword')}</p>
          </div>

          <form onSubmit={handlePasswordSubmit} className="space-y-4">
            <div>
              <input
                type="password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value)
                  setPasswordError(null)
                }}
                placeholder={t('share.password')}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                autoFocus
              />
              {passwordError && (
                <p className="text-red-600 text-sm mt-1">{passwordError}</p>
              )}
            </div>

            <button
              type="submit"
              className="w-full px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
            >
              {t('common.submit')}
            </button>
          </form>
        </div>
      </div>
    )
  }

  if (!share) {
    return null
  }

  // Display the shared content
  return (
    <>
      {share.share_type === 'file' && accessToken && (
        <SharedFileDisplay fileId={share.resource_id} accessToken={accessToken} />
      )}
    </>
  )
}

interface SharedFileDisplayProps {
  fileId: number
  accessToken: string
}

function SharedFileDisplay({ fileId, accessToken }: SharedFileDisplayProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [file, setFile] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadFile()
  }, [fileId, accessToken])

  const loadFile = async () => {
    try {
      // Create a separate axios instance without the auth redirect interceptor
      const publicAxios = axios.create({
        baseURL: '/api',
        timeout: 10000,
        withCredentials: true,
      })

      const response = await publicAxios.get(`/public/files/${fileId}`, {
        params: { token: accessToken }
      })
      setFile(response.data)
    } catch (error) {
      console.error('Failed to load file:', error)
    } finally {
      setLoading(false)
    }
  }

  const getFileUrl = (fileId: number) => {
    return `/api/public/files/${fileId}/download?token=${encodeURIComponent(accessToken)}`
  }

  if (loading) {
    return (
      <div className="fixed inset-0 bg-black flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white"></div>
      </div>
    )
  }

  if (!file) {
    return (
      <div className="fixed inset-0 bg-black flex items-center justify-center">
        <div className="text-white text-center">
          <p>{t('file.notFound')}</p>
        </div>
      </div>
    )
  }

  return (
    <FileViewer
      file={file}
      mode="share"
      onClose={() => navigate('/')}
      getImageUrl={getFileUrl}
      getDownloadUrl={getFileUrl}
    />
  )
}
