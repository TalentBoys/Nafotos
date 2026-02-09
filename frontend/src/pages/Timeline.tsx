import { useState, useEffect, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { fileAPI } from '@/services/api'
import type { File, YearInfo } from '@/types'
import FileGrid from '@/components/FileGrid'
import SelectionBar from '@/components/SelectionBar'
import TimelineScrollbar from '@/components/TimelineScrollbar'
import { format } from 'date-fns'
import { Upload } from 'lucide-react'

export default function Timeline() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [selectedFileIds, setSelectedFileIds] = useState<number[]>([])
  const [years, setYears] = useState<YearInfo[]>([])
  const [currentYear, setCurrentYear] = useState<number | null>(null)
  const [currentMonth, setCurrentMonth] = useState<number | null>(null)
  const yearRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const monthRefs = useRef<Record<string, HTMLDivElement | null>>({})

  // Load years for scrollbar
  useEffect(() => {
    fileAPI.getTimelineYears().then((res) => {
      setYears(res.data.years || [])
    })
  }, [])

  useEffect(() => {
    loadFiles()
  }, [page])

  // Track current visible year/month using IntersectionObserver
  useEffect(() => {
    if (files.length === 0) return

    const observerOptions = {
      root: null,
      rootMargin: '-100px 0px -50% 0px',
      threshold: 0,
    }

    // 观察月份级别的元素
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          const year = entry.target.getAttribute('data-year')
          const month = entry.target.getAttribute('data-month')
          if (year) {
            setCurrentYear(parseInt(year))
            if (month) {
              setCurrentMonth(parseInt(month))
            }
          }
        }
      })
    }, observerOptions)

    // 观察所有月份区块
    Object.values(monthRefs.current).forEach((el) => {
      if (el) observer.observe(el)
    })

    return () => observer.disconnect()
  }, [files])

  const loadFiles = async () => {
    try {
      setLoading(true)
      const response = await fileAPI.getTimeline(page, 50)
      const newFiles = response.data.files || []

      if (page === 1) {
        setFiles(newFiles)
      } else {
        setFiles((prev) => [...prev, ...newFiles])
      }

      setHasMore(newFiles.length === 50)
    } catch (error) {
      console.error('Failed to load files:', error)
    } finally {
      setLoading(false)
    }
  }

  // Group files by year, then by month, then by date
  const groupFilesByYearMonth = (files: File[]) => {
    const structure: Record<string, Record<string, Record<string, File[]>>> = {}

    files.forEach((file) => {
      const date = new Date(file.taken_at)
      const year = format(date, 'yyyy')
      const month = format(date, 'MM')
      const day = format(date, 'yyyy-MM-dd')

      if (!structure[year]) structure[year] = {}
      if (!structure[year][month]) structure[year][month] = {}
      if (!structure[year][month][day]) structure[year][month][day] = []

      structure[year][month][day].push(file)
    })

    return structure
  }

  const fileStructure = groupFilesByYearMonth(files)

  const handleShare = () => {
    navigate(`/share?fileIds=${selectedFileIds.join(',')}`)
  }

  const handleClearSelection = () => {
    setSelectedFileIds([])
  }

  const handleTimelineClick = useCallback((year: number, month?: number) => {
    // 优先滚动到月份
    if (month) {
      const monthKey = `${year}-${month.toString().padStart(2, '0')}`
      const element = monthRefs.current[monthKey]
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'start' })
        return
      }
    }

    // 回退到年份
    const element = yearRefs.current[year.toString()]
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }, [])

  // 格式化月份显示
  const formatMonthName = (month: string) => {
    const date = new Date(2000, parseInt(month) - 1, 1)
    return format(date, 'MMMM')
  }

  return (
    <div className="space-y-8">
      {selectedFileIds.length > 0 && (
        <SelectionBar
          selectedCount={selectedFileIds.length}
          onClear={handleClearSelection}
          onShare={handleShare}
        />
      )}

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('timeline.title')}</h1>
          <p className="text-muted-foreground">{t('app.title')}</p>
        </div>
        <button
          onClick={() => navigate('/upload')}
          className="flex items-center gap-2 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors"
        >
          <Upload className="h-5 w-5" />
          Upload Images
        </button>
      </div>

      {loading && page === 1 ? (
        <div className="text-center py-12">{t('timeline.loading')}</div>
      ) : files.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-muted-foreground mb-4">
            {t('timeline.empty')}
          </div>
          <button
            onClick={() => navigate('/upload')}
            className="inline-flex items-center gap-2 px-6 py-3 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors"
          >
            <Upload className="h-5 w-5" />
            Upload Your First Images
          </button>
        </div>
      ) : (
        <div className="space-y-12 pr-12">
          {Object.entries(fileStructure)
            .sort(([a], [b]) => b.localeCompare(a))
            .map(([year, months]) => (
              <div
                key={year}
                ref={(el) => {
                  yearRefs.current[year] = el
                }}
              >
                {/* Year header */}
                <h2 className="text-3xl font-bold mb-8 sticky top-16 bg-background/95 backdrop-blur-sm py-2 z-10">
                  {year}
                </h2>

                {/* Months within this year */}
                <div className="space-y-10">
                  {Object.entries(months)
                    .sort(([a], [b]) => b.localeCompare(a))
                    .map(([month, days]) => {
                      const monthKey = `${year}-${month}`

                      return (
                        <div
                          key={monthKey}
                          data-year={year}
                          data-month={month}
                          ref={(el) => {
                            monthRefs.current[monthKey] = el
                          }}
                        >
                          {/* Month header */}
                          <h3 className="text-xl font-semibold mb-4 text-muted-foreground">
                            {formatMonthName(month)}
                          </h3>

                          {/* Days within this month */}
                          <div className="space-y-6">
                            {Object.entries(days)
                              .sort(([a], [b]) => b.localeCompare(a))
                              .map(([date, dateFiles]) => (
                                <div key={date}>
                                  <h4 className="text-sm font-medium mb-2 text-muted-foreground/70">
                                    {format(new Date(date), 'd')}
                                  </h4>
                                  <FileGrid
                                    files={dateFiles}
                                    allFiles={files}
                                    selectedFileIds={selectedFileIds}
                                    onSelectionChange={setSelectedFileIds}
                                  />
                                </div>
                              ))}
                          </div>
                        </div>
                      )
                    })}
                </div>
              </div>
            ))}

          {hasMore && (
            <div className="text-center">
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={loading}
                className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
              >
                {loading ? t('timeline.loading') : t('timeline.loadMore')}
              </button>
            </div>
          )}
        </div>
      )}

      {/* Timeline Scrollbar */}
      {years.length > 0 && files.length > 0 && (
        <TimelineScrollbar
          years={years}
          currentYear={currentYear}
          currentMonth={currentMonth}
          onYearClick={handleTimelineClick}
        />
      )}
    </div>
  )
}
