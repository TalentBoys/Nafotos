import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { fileAPI } from '@/services/api'
import type { File } from '@/types'
import FileViewer from '@/components/FileViewer'

export default function FileDetail() {
  const { id } = useParams<{ id: string }>()
  const { t } = useTranslation()
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (id) {
      loadFile(parseInt(id))
    }
  }, [id])

  const loadFile = async (fileId: number) => {
    try {
      const response = await fileAPI.getFileById(fileId)
      setFile(response.data)
    } catch (error) {
      console.error('Failed to load file:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="text-center py-12">{t('timeline.loading')}</div>
  }

  if (!file) {
    return <div className="text-center py-12">File not found</div>
  }

  return <FileViewer file={file} mode="detail" />
}
