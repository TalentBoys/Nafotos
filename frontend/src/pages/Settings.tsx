import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { systemAPI } from '@/services/api'
import { RefreshCw } from 'lucide-react'

export default function Settings() {
  const { t, i18n } = useTranslation()
  const [scanning, setScanning] = useState(false)

  const handleScan = async () => {
    try {
      setScanning(true)
      await systemAPI.triggerScan()
      alert(t('settings.scanComplete'))
    } catch (error) {
      console.error('Failed to trigger scan:', error)
      alert('Failed to start scan')
    } finally {
      setScanning(false)
    }
  }

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang)
    localStorage.setItem('language', lang)
  }

  return (
    <div className="max-w-2xl space-y-8">
      <div>
        <h1 className="text-3xl font-bold mb-2">{t('settings.title')}</h1>
      </div>

      <div className="space-y-6">
        <div className="border rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-4">{t('settings.language')}</h2>
          <div className="flex gap-3">
            <button
              onClick={() => changeLanguage('en')}
              className={`px-4 py-2 rounded-lg ${
                i18n.language === 'en'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary hover:bg-secondary/80'
              }`}
            >
              English
            </button>
            <button
              onClick={() => changeLanguage('zh')}
              className={`px-4 py-2 rounded-lg ${
                i18n.language === 'zh'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-secondary hover:bg-secondary/80'
              }`}
            >
              中文
            </button>
          </div>
        </div>

        <div className="border rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-4">{t('settings.scan')}</h2>
          <p className="text-muted-foreground mb-4">
            Trigger a manual scan of all mounted directories to index new files.
          </p>
          <button
            onClick={handleScan}
            disabled={scanning}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
          >
            <RefreshCw size={20} className={scanning ? 'animate-spin' : ''} />
            {scanning ? t('settings.scanning') : t('settings.scan')}
          </button>
        </div>
      </div>
    </div>
  )
}
