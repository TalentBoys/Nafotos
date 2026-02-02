import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Link2, Copy, Check, Eye, Clock, Lock, Shield, Edit2, Trash2, Link } from 'lucide-react'
import { shareAPI, type Share, type UpdateShareRequest } from '@/services/shares'
import { format } from 'date-fns'

export default function ShareManagement() {
  const { t } = useTranslation()
  const [shares, setShares] = useState<Share[]>([])
  const [loading, setLoading] = useState(true)
  const [editingShare, setEditingShare] = useState<string | null>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  useEffect(() => {
    loadShares()
  }, [])

  const loadShares = async () => {
    try {
      setLoading(true)
      const response = await shareAPI.listShares()
      setShares(response.data.shares || [])
    } catch (error) {
      console.error('Failed to load shares:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCopyLink = (shareId: string) => {
    const url = `${window.location.origin}/s/${shareId}`
    navigator.clipboard.writeText(url)
    setCopiedId(shareId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const handleDelete = async (shareId: string) => {
    if (!confirm(t('share.confirmDelete'))) return

    try {
      await shareAPI.deleteShare(shareId)
      await loadShares()
    } catch (error) {
      console.error('Failed to delete share:', error)
      alert(t('share.deleteFailed'))
    }
  }

  const handleToggleEnabled = async (shareId: string, enabled: boolean) => {
    try {
      await shareAPI.updateShare(shareId, { enabled: !enabled })
      await loadShares()
    } catch (error) {
      console.error('Failed to update share:', error)
      alert(t('share.updateFailed'))
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link2 className="text-primary" size={32} />
          <div>
            <h1 className="text-3xl font-bold">{t('share.management')}</h1>
            <p className="text-gray-600 mt-1">{t('share.managementDescription')}</p>
          </div>
        </div>
      </div>

      {/* Shares List */}
      {shares.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <Link2 size={48} className="mx-auto text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {t('share.noShares')}
          </h3>
          <p className="text-gray-600">{t('share.noSharesDescription')}</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('share.link')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('share.settings')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('share.stats')}
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('share.status')}
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('common.actions')}
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {shares.map((share) => (
                <tr key={share.id} className={!share.enabled ? 'opacity-50' : ''}>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <Link size={16} className="text-gray-400" />
                      <code className="text-sm font-mono text-gray-900">
                        /s/{share.id}
                      </code>
                      <button
                        onClick={() => handleCopyLink(share.id)}
                        className="p-1 hover:bg-gray-100 rounded transition-colors"
                        title={t('share.copy')}
                      >
                        {copiedId === share.id ? (
                          <Check size={14} className="text-green-600" />
                        ) : (
                          <Copy size={14} className="text-gray-400" />
                        )}
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                      {t('share.created')}: {format(new Date(share.created_at), 'PPp')}
                    </p>
                  </td>

                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-2">
                      {share.requires_auth && (
                        <span className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                          <Shield size={12} />
                          {t('share.authRequired')}
                        </span>
                      )}
                      {share.has_password && (
                        <span className="inline-flex items-center gap-1 px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded">
                          <Lock size={12} />
                          {t('share.passwordProtected')}
                        </span>
                      )}
                      {share.expires_at && (
                        <span className="inline-flex items-center gap-1 px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded">
                          <Clock size={12} />
                          {format(new Date(share.expires_at), 'PP')}
                        </span>
                      )}
                    </div>
                  </td>

                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 text-sm">
                      <Eye size={14} className="text-gray-400" />
                      <span className="text-gray-900">{share.view_count}</span>
                      {share.max_views && (
                        <span className="text-gray-500">/ {share.max_views}</span>
                      )}
                    </div>
                  </td>

                  <td className="px-6 py-4">
                    {share.enabled ? (
                      <span className="inline-flex px-2 py-1 bg-green-100 text-green-800 text-xs rounded font-medium">
                        {t('share.active')}
                      </span>
                    ) : (
                      <span className="inline-flex px-2 py-1 bg-gray-100 text-gray-800 text-xs rounded font-medium">
                        {t('share.disabled')}
                      </span>
                    )}
                  </td>

                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => setEditingShare(share.id)}
                        className="p-2 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                        title={t('common.edit')}
                      >
                        <Edit2 size={16} />
                      </button>
                      <button
                        onClick={() => handleToggleEnabled(share.id, share.enabled)}
                        className="px-3 py-1 text-xs font-medium rounded transition-colors"
                        style={{
                          backgroundColor: share.enabled ? '#fee' : '#efe',
                          color: share.enabled ? '#c00' : '#080'
                        }}
                      >
                        {share.enabled ? t('share.disable') : t('share.enable')}
                      </button>
                      <button
                        onClick={() => handleDelete(share.id)}
                        className="p-2 text-red-600 hover:bg-red-50 rounded transition-colors"
                        title={t('common.delete')}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Edit Dialog */}
      {editingShare && (
        <EditShareDialog
          shareId={editingShare}
          onClose={() => setEditingShare(null)}
          onUpdate={loadShares}
        />
      )}
    </div>
  )
}

interface EditShareDialogProps {
  shareId: string
  onClose: () => void
  onUpdate: () => void
}

function EditShareDialog({ shareId, onClose, onUpdate }: EditShareDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [password, setPassword] = useState('')
  const [requiresAuth, setRequiresAuth] = useState(false)
  const [expiresIn, setExpiresIn] = useState<number | ''>('')
  const [removeExpiration, setRemoveExpiration] = useState(false)

  const handleSave = async () => {
    setLoading(true)
    try {
      const updates: UpdateShareRequest = {}

      if (password) {
        updates.password = password
      }

      updates.requires_auth = requiresAuth

      if (removeExpiration) {
        updates.expires_in = 0
      } else if (expiresIn && expiresIn > 0) {
        updates.expires_in = expiresIn
      }

      await shareAPI.updateShare(shareId, updates)
      onUpdate()
      onClose()
    } catch (error) {
      console.error('Failed to update share:', error)
      alert(t('share.updateFailed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-bold">{t('share.editShare')}</h2>
          <button onClick={onClose} className="p-1 hover:bg-gray-100 rounded">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-6 space-y-4">
          <div className="flex items-center gap-3 p-3 border rounded-lg">
            <Shield className="text-blue-600" size={20} />
            <div className="flex-1">
              <label className="font-medium">{t('share.requireAuth')}</label>
            </div>
            <input
              type="checkbox"
              checked={requiresAuth}
              onChange={(e) => setRequiresAuth(e.target.checked)}
              className="w-5 h-5"
            />
          </div>

          <div>
            <label className="block font-medium mb-2">{t('share.newPassword')}</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={t('share.leaveEmptyToKeep')}
              className="w-full px-4 py-2 border rounded-lg"
            />
          </div>

          <div>
            <label className="block font-medium mb-2">{t('share.expiration')}</label>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <input
                  type="number"
                  value={expiresIn}
                  onChange={(e) => {
                    setExpiresIn(e.target.value ? parseInt(e.target.value) : '')
                    setRemoveExpiration(false)
                  }}
                  placeholder="0"
                  min="0"
                  disabled={removeExpiration}
                  className="w-32 px-4 py-2 border rounded-lg disabled:opacity-50"
                />
                <span className="text-gray-600">{t('share.hoursFromNow')}</span>
              </div>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={removeExpiration}
                  onChange={(e) => {
                    setRemoveExpiration(e.target.checked)
                    if (e.target.checked) setExpiresIn('')
                  }}
                />
                <span className="text-sm">{t('share.removeExpiration')}</span>
              </label>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-end gap-3 p-6 border-t">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
          >
            {t('common.cancel')}
          </button>
          <button
            onClick={handleSave}
            disabled={loading}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? t('share.saving') : t('common.save')}
          </button>
        </div>
      </div>
    </div>
  )
}
