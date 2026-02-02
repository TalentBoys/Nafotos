import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { X, Copy, Check, Lock, Clock, Eye, Shield } from 'lucide-react'
import { shareAPI, type CreateShareRequest } from '@/services/shares'
import axios from 'axios'

interface ShareDialogProps {
  fileId: number
  fileName: string
  onClose: () => void
}

export default function ShareDialog({ fileId, fileName, onClose }: ShareDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [shareUrl, setShareUrl] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Form state
  const [password, setPassword] = useState('')
  const [requiresAuth, setRequiresAuth] = useState(false)
  const [expiresIn, setExpiresIn] = useState<number | ''>('')
  const [maxViews, setMaxViews] = useState<number | ''>('')

  // Check if domain is configured
  useEffect(() => {
    checkDomainConfig()
  }, [])

  const checkDomainConfig = async () => {
    try {
      const response = await axios.get('/api/domain-config')
      if (!response.data.domain) {
        setError(t('share.noDomainConfigured'))
      }
    } catch (err) {
      setError(t('share.noDomainConfigured'))
    }
  }

  const handleCreate = async () => {
    if (error && error === t('share.noDomainConfigured')) {
      return
    }

    setLoading(true)
    setError(null)

    try {
      const request: CreateShareRequest = {
        share_type: 'file',
        resource_id: fileId,
        requires_auth: requiresAuth,
      }

      if (password) {
        request.password = password
      }

      if (expiresIn && expiresIn > 0) {
        request.expires_in = expiresIn
      }

      if (maxViews && maxViews > 0) {
        request.max_views = maxViews
      }

      const response = await shareAPI.createShare(request)
      setShareUrl(response.data.url)
    } catch (err: any) {
      setError(err.response?.data?.error || t('share.createFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = () => {
    if (shareUrl) {
      navigator.clipboard.writeText(shareUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-bold">{t('share.title')}</h2>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6">
          {/* File name */}
          <div>
            <p className="text-sm text-gray-600">{t('share.sharingFile')}</p>
            <p className="font-medium text-gray-900 mt-1 truncate">{fileName}</p>
          </div>

          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-red-800 text-sm">
              {error}
            </div>
          )}

          {/* Share URL (if created) */}
          {shareUrl ? (
            <div className="space-y-4">
              <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                <p className="text-sm text-green-800 font-medium mb-2">
                  {t('share.linkCreated')}
                </p>
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={shareUrl}
                    readOnly
                    className="flex-1 px-3 py-2 bg-white border border-green-300 rounded text-sm font-mono"
                  />
                  <button
                    onClick={handleCopy}
                    className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 transition-colors flex items-center gap-2"
                  >
                    {copied ? <Check size={16} /> : <Copy size={16} />}
                    {copied ? t('share.copied') : t('share.copy')}
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Requires Auth */}
              <div className="flex items-center gap-3 p-3 border border-gray-200 rounded-lg">
                <Shield className="text-blue-600" size={20} />
                <div className="flex-1">
                  <label className="font-medium text-gray-900 block">
                    {t('share.requireAuth')}
                  </label>
                  <p className="text-sm text-gray-600 mt-1">
                    {t('share.requireAuthHelp')}
                  </p>
                </div>
                <input
                  type="checkbox"
                  checked={requiresAuth}
                  onChange={(e) => setRequiresAuth(e.target.checked)}
                  className="w-5 h-5"
                />
              </div>

              {/* Password Protection */}
              <div className="space-y-2">
                <label className="flex items-center gap-2 font-medium text-gray-900">
                  <Lock size={18} />
                  {t('share.password')}
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder={t('share.passwordPlaceholder')}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
                <p className="text-sm text-gray-600">{t('share.passwordHelp')}</p>
              </div>

              {/* Expiration */}
              <div className="space-y-2">
                <label className="flex items-center gap-2 font-medium text-gray-900">
                  <Clock size={18} />
                  {t('share.expiration')}
                </label>
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    value={expiresIn}
                    onChange={(e) => setExpiresIn(e.target.value ? parseInt(e.target.value) : '')}
                    placeholder="0"
                    min="0"
                    className="w-32 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                  <span className="text-gray-600">{t('share.hours')}</span>
                </div>
                <p className="text-sm text-gray-600">{t('share.expirationHelp')}</p>
              </div>

              {/* Max Views */}
              <div className="space-y-2">
                <label className="flex items-center gap-2 font-medium text-gray-900">
                  <Eye size={18} />
                  {t('share.maxViews')}
                </label>
                <input
                  type="number"
                  value={maxViews}
                  onChange={(e) => setMaxViews(e.target.value ? parseInt(e.target.value) : '')}
                  placeholder={t('share.unlimited')}
                  min="0"
                  className="w-32 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
                <p className="text-sm text-gray-600">{t('share.maxViewsHelp')}</p>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-gray-200">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          >
            {shareUrl ? t('common.close') : t('common.cancel')}
          </button>
          {!shareUrl && (
            <button
              onClick={handleCreate}
              disabled={loading || (error === t('share.noDomainConfigured'))}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? t('share.creating') : t('share.createLink')}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
