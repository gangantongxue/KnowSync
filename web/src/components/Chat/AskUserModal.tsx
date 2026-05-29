import { useState, useEffect } from 'react'

interface AskUserData {
  session_id: string
  question: string
  type: 'single' | 'multiple'
  options?: string[]
  has_other?: boolean
}

export default function AskUserModal() {
  const [data, setData] = useState<AskUserData | null>(null)

  useEffect(() => {
    const handler = (e: Event) => {
      setData((e as CustomEvent).detail as AskUserData)
    }
    window.addEventListener('ask-user', handler)
    return () => window.removeEventListener('ask-user', handler)
  }, [])

  if (!data) return null

  const handleSelect = (option: string) => {
    window.dispatchEvent(new CustomEvent('ask-user-response', { detail: { value: option } }))
    setData(null)
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 max-w-md w-full mx-4 shadow-xl">
        <h3 className="text-sm font-medium text-gray-800 mb-4">{data.question}</h3>
        <div className="space-y-2">
          {data.options?.map(opt => (
            <button
              key={opt}
              onClick={() => handleSelect(opt)}
              className="w-full px-3 py-2 text-sm text-left border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-emerald-300 transition-colors"
            >
              {opt}
            </button>
          ))}
        </div>
        <button onClick={() => setData(null)} className="mt-4 text-xs text-gray-400 hover:text-gray-600">
          取消
        </button>
      </div>
    </div>
  )
}
