import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Globe, Save, AlertCircle, CheckCircle } from 'lucide-react'
import axios from 'axios'

interface DomainConfig {
  id?: number
  protocol: string
  domain: string
  port: string
  updated_by?: number
  updated_at?: string
}

export default function DomainConfig() {
  const { t } = useTranslation()
  const [config, setConfig] = useState<DomainConfig>({
    protocol: 'http',
    domain: 'localhost',
    port: '8080'
  })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)

  // Load current configuration
  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    try {
      setLoading(true)
      const response = await axios.get('/api/domain-config')
      setConfig(response.data)
    } catch (error) {
      console.error('Failed to load domain configuration:', error)
      setMessage({ type: 'error', text: t('domainConfig.loadFailed') })
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    // Validation
    if (!config.protocol) {
      setMessage({ type: 'error', text: t('domainConfig.protocolRequired') })
      return
    }
    if (!config.domain) {
      setMessage({ type: 'error', text: t('domainConfig.domainRequired') })
      return
    }
    if (!config.port) {
      setMessage({ type: 'error', text: t('domainConfig.portRequired') })
      return
    }

    try {
      setSaving(true)
      setMessage(null)
      await axios.post('/api/domain-config', {
        protocol: config.protocol,
        domain: config.domain,
        port: config.port
      })
      setMessage({ type: 'success', text: t('domainConfig.saveSuccess') })
      // Reload to get updated data
      setTimeout(() => loadConfig(), 1000)
    } catch (error: any) {
      console.error('Failed to save domain configuration:', error)
      setMessage({
        type: 'error',
        text: error.response?.data?.error || t('domainConfig.saveFailed')
      })
    } finally {
      setSaving(false)
    }
  }

  const getFullURL = () => {
    const url = `${config.protocol}://${config.domain}`
    if ((config.protocol === 'http' && config.port !== '80') ||
        (config.protocol === 'https' && config.port !== '443')) {
      return `${url}:${config.port}`
    }
    return url
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
          <Globe className="text-primary" size={32} />
          <div>
            <h1 className="text-3xl font-bold">{t('domainConfig.title')}</h1>
            <p className="text-gray-600 mt-1">
              {t('domainConfig.description')}
            </p>
          </div>
        </div>
      </div>

      {/* Message */}
      {message && (
        <div className={`p-4 rounded-lg flex items-center gap-3 ${
          message.type === 'success'
            ? 'bg-green-50 text-green-800 border border-green-200'
            : 'bg-red-50 text-red-800 border border-red-200'
        }`}>
          {message.type === 'success' ? (
            <CheckCircle size={20} />
          ) : (
            <AlertCircle size={20} />
          )}
          <span>{message.text}</span>
        </div>
      )}

      {/* Configuration Form */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="space-y-6">
          {/* Protocol */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {t('domainConfig.protocol')}
            </label>
            <select
              value={config.protocol}
              onChange={(e) => setConfig({ ...config, protocol: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="http">HTTP</option>
              <option value="https">HTTPS</option>
            </select>
            <p className="text-sm text-gray-500 mt-1">
              {t('domainConfig.protocolHelp')}
            </p>
          </div>

          {/* Domain */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {t('domainConfig.domain')}
            </label>
            <input
              type="text"
              value={config.domain}
              onChange={(e) => setConfig({ ...config, domain: e.target.value })}
              placeholder={t('domainConfig.domainPlaceholder')}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
            <p className="text-sm text-gray-500 mt-1">
              {t('domainConfig.domainHelp')}
            </p>
          </div>

          {/* Port */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {t('domainConfig.port')}
            </label>
            <input
              type="text"
              value={config.port}
              onChange={(e) => setConfig({ ...config, port: e.target.value })}
              placeholder={t('domainConfig.portPlaceholder')}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
            <p className="text-sm text-gray-500 mt-1">
              {t('domainConfig.portHelp')}
            </p>
          </div>

          {/* Preview */}
          <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
            <h3 className="text-sm font-medium text-gray-700 mb-2">{t('domainConfig.preview')}</h3>
            <p className="text-lg font-mono text-primary">
              {getFullURL()}
            </p>
            <p className="text-sm text-gray-500 mt-1">
              {t('domainConfig.previewHelp')}
            </p>
          </div>

          {/* Save Button */}
          <div className="flex justify-end">
            <button
              onClick={handleSave}
              disabled={saving}
              className="flex items-center gap-2 px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Save size={18} />
              {saving ? t('domainConfig.saving') : t('domainConfig.save')}
            </button>
          </div>
        </div>
      </div>

      {/* Last Updated Info */}
      {config.updated_at && (
        <div className="text-sm text-gray-500">
          {t('domainConfig.lastUpdated')}: {new Date(config.updated_at).toLocaleString()}
        </div>
      )}
    </div>
  )
}
