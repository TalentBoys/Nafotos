import { X, Share2 } from 'lucide-react'

interface SelectionBarProps {
  selectedCount: number
  onClear: () => void
  onShare?: () => void
  onDelete?: () => void
}

export default function SelectionBar({
  selectedCount,
  onClear,
  onShare,
  onDelete
}: SelectionBarProps) {
  return (
    <div className="fixed top-0 left-0 right-0 z-50 bg-blue-600 text-white shadow-lg">
      <div className="container mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={onClear}
            className="p-2 hover:bg-white/20 rounded-lg transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
          <div className="flex items-center gap-3">
            <span className="font-medium">
              {selectedCount} 项目已选择
            </span>
            <span className="text-sm text-blue-100 hidden sm:inline">
              • 按住 Shift 点击可连续选择
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {onShare && (
            <button
              onClick={onShare}
              className="flex items-center gap-2 px-4 py-2 bg-white text-blue-600 rounded-lg hover:bg-blue-50 transition-colors font-medium"
            >
              <Share2 className="h-4 w-4" />
              分享
            </button>
          )}
          {onDelete && (
            <button
              onClick={onDelete}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium"
            >
              删除
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
