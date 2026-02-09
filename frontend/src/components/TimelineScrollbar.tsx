import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import type { YearInfo } from '@/types'

interface TimelineScrollbarProps {
  years: YearInfo[]
  currentYear: number | null
  currentMonth?: number | null
  onYearClick: (year: number, month?: number) => void
}

// 生成时间轴上的所有点（年份 + 月份）
interface TimelinePoint {
  type: 'year' | 'month'
  year: number
  month?: number
  label?: string
}

export default function TimelineScrollbar({
  years,
  currentYear,
  currentMonth,
  onYearClick,
}: TimelineScrollbarProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [hoverPoint, setHoverPoint] = useState<TimelinePoint | null>(null)
  const [dragPoint, setDragPoint] = useState<TimelinePoint | null>(null)
  const trackRef = useRef<HTMLDivElement>(null)

  // 生成所有时间点（年份作为标签，月份作为点）
  const timelinePoints = useMemo(() => {
    const points: TimelinePoint[] = []

    // 按年份降序排列
    const sortedYears = [...years].sort((a, b) => b.year - a.year)

    sortedYears.forEach((yearInfo, yearIndex) => {
      // 添加年份标记
      points.push({
        type: 'year',
        year: yearInfo.year,
        label: yearInfo.year.toString(),
      })

      // 在年份之间添加月份点（除了最后一个年份）
      if (yearIndex < sortedYears.length - 1) {
        // 从12月到1月（降序）
        for (let month = 12; month >= 1; month--) {
          points.push({
            type: 'month',
            year: yearInfo.year,
            month,
          })
        }
      }
    })

    return points
  }, [years])

  // 根据 Y 位置计算对应的时间点
  const getPointFromPosition = useCallback(
    (clientY: number): TimelinePoint | null => {
      if (!trackRef.current || timelinePoints.length === 0) return null

      const rect = trackRef.current.getBoundingClientRect()
      const relativeY = clientY - rect.top
      const progress = Math.max(0, Math.min(1, relativeY / rect.height))
      const index = Math.round(progress * (timelinePoints.length - 1))
      return timelinePoints[index] ?? null
    },
    [timelinePoints]
  )

  // 计算当前位置的进度（0-1）
  const getCurrentProgress = useCallback(() => {
    if (!currentYear || timelinePoints.length === 0) return 0

    // 找到当前年份/月份对应的索引
    let targetIndex = timelinePoints.findIndex(
      (p) => p.type === 'year' && p.year === currentYear
    )

    // 如果有月份信息，找更精确的位置
    if (currentMonth && targetIndex !== -1) {
      const monthIndex = timelinePoints.findIndex(
        (p) => p.type === 'month' && p.year === currentYear && p.month === currentMonth
      )
      if (monthIndex !== -1) {
        targetIndex = monthIndex
      }
    }

    if (targetIndex === -1) return 0
    return targetIndex / Math.max(1, timelinePoints.length - 1)
  }, [currentYear, currentMonth, timelinePoints])

  // 处理鼠标按下
  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      setIsDragging(true)
      const point = getPointFromPosition(e.clientY)
      if (point) {
        setDragPoint(point)
        onYearClick(point.year, point.month)
      }
    },
    [getPointFromPosition, onYearClick]
  )

  // 处理鼠标移动
  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (!isDragging) {
        const point = getPointFromPosition(e.clientY)
        setHoverPoint(point)
      }
    },
    [isDragging, getPointFromPosition]
  )

  // 拖动时的全局事件
  useEffect(() => {
    if (!isDragging) return

    const handleGlobalMouseMove = (e: MouseEvent) => {
      const point = getPointFromPosition(e.clientY)
      if (point && (point.year !== dragPoint?.year || point.month !== dragPoint?.month)) {
        setDragPoint(point)
        onYearClick(point.year, point.month)
      }
    }

    const handleMouseUp = () => {
      setIsDragging(false)
      setDragPoint(null)
    }

    document.addEventListener('mousemove', handleGlobalMouseMove)
    document.addEventListener('mouseup', handleMouseUp)

    return () => {
      document.removeEventListener('mousemove', handleGlobalMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, dragPoint, getPointFromPosition, onYearClick])

  if (years.length === 0) return null

  const currentProgress = getCurrentProgress()
  const displayPoint = dragPoint || hoverPoint

  // 格式化月份显示
  const formatMonth = (month: number) => {
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    return months[month - 1]
  }

  return (
    <div className="fixed right-3 top-1/2 -translate-y-1/2 z-40 flex items-center gap-2">
      {/* Tooltip */}
      {displayPoint && (
        <div className="px-2 py-1 bg-foreground text-background rounded text-xs font-medium whitespace-nowrap">
          {displayPoint.type === 'year'
            ? displayPoint.year
            : `${formatMonth(displayPoint.month!)} ${displayPoint.year}`
          }
        </div>
      )}

      {/* 滚动轨道 */}
      <div
        ref={trackRef}
        className="relative flex flex-col items-center py-2 cursor-pointer select-none"
        style={{ height: 'min(70vh, 500px)' }}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseLeave={() => !isDragging && setHoverPoint(null)}
      >
        {/* 轨道背景线 */}
        <div className="absolute inset-y-2 w-[2px] bg-border rounded-full" />

        {/* 时间点 */}
        {timelinePoints.map((point, index) => {
          const isYearPoint = point.type === 'year'
          const progress = index / Math.max(1, timelinePoints.length - 1)

          return (
            <div
              key={`${point.year}-${point.month || 'year'}`}
              className="absolute flex items-center"
              style={{
                top: `calc(${progress * 100}% + 8px)`,
                transform: 'translateY(-50%)',
              }}
            >
              {isYearPoint ? (
                // 年份标签
                <div className="flex items-center">
                  <span className="text-[10px] font-medium text-muted-foreground mr-1 w-8 text-right">
                    {point.label}
                  </span>
                  <div className="w-2 h-[2px] bg-muted-foreground rounded-full" />
                </div>
              ) : (
                // 月份小点
                <div className="ml-[38px] w-1 h-1 rounded-full bg-border" />
              )}
            </div>
          )
        })}

        {/* 当前位置指示器（小横线） */}
        <div
          className="absolute left-[30px] w-4 h-[3px] bg-primary rounded-full transition-all duration-150 shadow-sm"
          style={{
            top: `calc(${currentProgress * 100}% + 8px)`,
            transform: 'translateY(-50%)',
          }}
        />
      </div>
    </div>
  )
}
