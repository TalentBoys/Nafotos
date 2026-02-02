import { Link } from 'react-router-dom'
import { useState, useEffect } from 'react'
import type { File } from '@/types'
import { fileAPI } from '@/services/api'
import { Play, Check } from 'lucide-react'
import { useContainerWidth } from '@/hooks/useContainerWidth'
import { useResponsiveValue } from '@/hooks/useResponsiveValue'
import { useJustifiedLayout } from '@/hooks/useJustifiedLayout'

// Utility function to merge class names
const cn = (...classes: (string | boolean | undefined)[]) => {
  return classes.filter(Boolean).join(' ')
}

interface FileGridProps {
  files: File[]
  selectedFileIds?: number[]
  onSelectionChange?: (selectedIds: number[]) => void
  selectionMode?: boolean
}

export default function FileGrid({
  files,
  selectedFileIds = [],
  onSelectionChange,
  selectionMode = false
}: FileGridProps) {
  const [hoveredFileId, setHoveredFileId] = useState<number | null>(null)
  const [lastSelectedId, setLastSelectedId] = useState<number | null>(null)
  const [isShiftPressed, setIsShiftPressed] = useState(false)

  // Responsive layout hooks
  const [containerRef, containerWidth] = useContainerWidth<HTMLDivElement>()
  const targetHeight = useResponsiveValue({
    mobile: 150,
    tablet: 200,
    desktop: 250
  })
  const rows = useJustifiedLayout(files, containerWidth, targetHeight, 4)

  // Monitor Shift key state
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Shift') {
        setIsShiftPressed(true)
      }
    }

    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.key === 'Shift') {
        setIsShiftPressed(false)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('keyup', handleKeyUp)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('keyup', handleKeyUp)
    }
  }, [])

  const handleFileClick = (e: React.MouseEvent, fileId: number) => {
    if (selectionMode || selectedFileIds.length > 0) {
      e.preventDefault()
      toggleSelection(fileId, e.shiftKey)
    }
  }

  const handleCheckboxClick = (e: React.MouseEvent, fileId: number) => {
    e.preventDefault()
    e.stopPropagation()
    toggleSelection(fileId, e.shiftKey)
  }

  const toggleSelection = (fileId: number, shiftKey: boolean = false) => {
    if (!onSelectionChange) return

    // Shift + Click for range selection
    if (shiftKey && lastSelectedId !== null) {
      const fileIds = files.map(f => f.id)
      const lastIndex = fileIds.indexOf(lastSelectedId)
      const currentIndex = fileIds.indexOf(fileId)

      if (lastIndex !== -1 && currentIndex !== -1) {
        const start = Math.min(lastIndex, currentIndex)
        const end = Math.max(lastIndex, currentIndex)
        const rangeIds = fileIds.slice(start, end + 1)

        // Add range to selection (union)
        const newSelection = Array.from(new Set([...selectedFileIds, ...rangeIds]))
        onSelectionChange(newSelection)
        return
      }
    }

    // Normal toggle
    const newSelection = selectedFileIds.includes(fileId)
      ? selectedFileIds.filter(id => id !== fileId)
      : [...selectedFileIds, fileId]

    setLastSelectedId(fileId)
    onSelectionChange(newSelection)
  }

  const isSelected = (fileId: number) => selectedFileIds.includes(fileId)

  // Calculate range preview when Shift is pressed
  const getRangePreview = (hoveredId: number): number[] => {
    if (!isShiftPressed || lastSelectedId === null || selectedFileIds.length === 0) {
      return []
    }

    const fileIds = files.map(f => f.id)
    const lastIndex = fileIds.indexOf(lastSelectedId)
    const hoveredIndex = fileIds.indexOf(hoveredId)

    if (lastIndex === -1 || hoveredIndex === -1) {
      return []
    }

    const start = Math.min(lastIndex, hoveredIndex)
    const end = Math.max(lastIndex, hoveredIndex)
    return fileIds.slice(start, end + 1)
  }

  const isInPreviewRange = (fileId: number): boolean => {
    if (!hoveredFileId) return false
    const previewRange = getRangePreview(hoveredFileId)
    return previewRange.includes(fileId) && !isSelected(fileId)
  }

  return (
    <div ref={containerRef} className="w-full">
      {rows.map((row, rowIndex) => (
        <div key={rowIndex} className="flex gap-1 mb-1">
          {row.photos.map(({ file, width, height }) => (
            <Link
              key={file.id}
              to={`/file/${file.id}`}
              className={cn(
                "group relative rounded-lg overflow-hidden transition-all flex-shrink-0",
                isSelected(file.id)
                  ? "bg-gray-300 dark:bg-gray-700"
                  : "bg-secondary"
              )}
              style={{ width: `${width}px`, height: `${height}px` }}
              onClick={(e) => handleFileClick(e, file.id)}
              onMouseEnter={() => setHoveredFileId(file.id)}
              onMouseLeave={() => setHoveredFileId(null)}
            >
              <div
                className={cn(
                  "absolute inset-0 transition-transform duration-200 rounded-lg overflow-hidden",
                  isSelected(file.id) ? "scale-[0.85]" : "scale-100"
                )}
              >
                <img
                  src={fileAPI.getThumbnailUrl(file.id)}
                  alt={file.filename}
                  className="w-full h-full object-cover"
                  loading="lazy"
                />

                {file.file_type === 'video' && (
                  <div className="absolute inset-0 flex items-center justify-center bg-black/20">
                    <Play className="text-white" size={32} />
                  </div>
                )}

                {/* Range preview overlay - warm yellow for Shift selection preview */}
                {isInPreviewRange(file.id) && (
                  <div className="absolute inset-0 bg-amber-400/40 pointer-events-none transition-opacity" />
                )}

                {/* Top gradient shadow on hover */}
                {hoveredFileId === file.id && !isInPreviewRange(file.id) && (
                  <div className="absolute top-0 left-0 right-0 h-20 bg-gradient-to-b from-black/25 to-transparent transition-opacity" />
                )}

                {/* Bottom gradient with filename */}
                <div className="absolute bottom-0 left-0 right-0 h-20 bg-gradient-to-t from-black/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
                  <div className="absolute bottom-0 left-0 right-0 p-2">
                    <p className="text-white text-xs truncate">{file.filename}</p>
                  </div>
                </div>
              </div>

              {/* Checkbox - positioned relative to parent Link, not the scaled div */}
              {(hoveredFileId === file.id || selectedFileIds.length > 0 || selectionMode) && (
                <button
                  type="button"
                  className={cn(
                    "absolute top-2 left-2 p-1 rounded-full flex items-center justify-center cursor-pointer transition-all z-10",
                    isSelected(file.id)
                      ? "bg-gray-200 dark:bg-gray-800"
                      : "bg-white/90 hover:bg-white"
                  )}
                  onClick={(e) => handleCheckboxClick(e, file.id)}
                >
                  {isSelected(file.id) ? (
                    <svg
                      width="24"
                      height="24"
                      viewBox="0 0 24 24"
                      className="text-blue-600"
                      fill="currentColor"
                    >
                      <path d="M12 2C6.5 2 2 6.5 2 12S6.5 22 12 22 22 17.5 22 12 17.5 2 12 2M10 17L5 12L6.41 10.59L10 14.17L17.59 6.58L19 8L10 17Z" />
                    </svg>
                  ) : (
                    <div className="w-6 h-6 rounded-full border-2 border-white" />
                  )}
                </button>
              )}
            </Link>
          ))}
        </div>
      ))}
    </div>
  )
}
