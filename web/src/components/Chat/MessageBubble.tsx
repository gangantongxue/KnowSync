import { Streamdown } from './Streamdown'

interface MessageBubbleProps {
  role: 'user' | 'assistant'
  content: string
  thinking?: string
  isStreaming?: boolean
  userAvatar?: string
}

export default function MessageBubble({ role, content, thinking, isStreaming, userAvatar }: MessageBubbleProps) {
  const isUser = role === 'user'

  return (
    <div className={`flex gap-3 mb-4 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm shrink-0 overflow-hidden ${isUser ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-200 text-gray-600'}`}>
        {isUser ? (
          userAvatar ? (
            <img src={userAvatar} alt="" className="w-full h-full object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }} />
          ) : (
            'U'
          )
        ) : (
          <img src="/img/KK.jpg" alt="" className="w-full h-full object-cover" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }} />
        )}
      </div>

      <div className={`max-w-[70%] ${isUser ? 'items-end' : 'items-start'}`}>
        {thinking && (
          <details className="mb-2 text-sm">
            <summary className="text-gray-400 cursor-pointer hover:text-gray-600">思考过程</summary>
            <div className="mt-1 p-2 bg-gray-50 rounded text-gray-500 text-xs whitespace-pre-wrap">
              {thinking}
            </div>
          </details>
        )}
        <div className={`px-4 py-2.5 rounded-2xl text-sm leading-relaxed ${isUser ? 'bg-emerald-500 text-white' : 'bg-gray-100 text-gray-800'}`}>
          {isUser ? content : <Streamdown content={content} />}
        </div>
        {isStreaming && (
          <span className="inline-block w-2 h-4 bg-emerald-500 animate-pulse ml-1" />
        )}
      </div>
    </div>
  )
}
