import { useChat } from '../../store/chat-context'
import LoadingSpinner from './LoadingSpinner'

export default function SessionList() {
  const { sessions, virtualSession, createNewSession, isLoadingSessions, currentSessionId, loadMessages, deleteSession } = useChat()

  const displaySessions = virtualSession ? [virtualSession, ...sessions] : sessions

  return (
    <div className="p-2">
      <div className="text-xs text-gray-400 font-medium px-2 py-1">会话列表</div>
      {isLoadingSessions && displaySessions.length === 0 ? (
        <LoadingSpinner />
      ) : displaySessions.length === 0 ? (
        <div className="py-4">
          <button
            onClick={createNewSession}
            className="w-full px-3 py-2 text-sm text-emerald-600 border border-emerald-200 rounded-lg hover:bg-emerald-50 transition-colors"
          >
            开始新对话
          </button>
        </div>
      ) : (
        displaySessions.map(session => {
          const isVirtual = virtualSession && session.id === virtualSession.id
          return (
            <div key={session.id} className="group relative">
              <button
                onClick={() => loadMessages(session.id)}
                className={`w-full text-left px-2 py-2 rounded-lg text-sm mb-0.5 transition-colors ${currentSessionId === session.id ? 'bg-emerald-50 text-emerald-700' : 'text-gray-600 hover:bg-gray-100'} ${isVirtual ? 'italic' : ''}`}
              >
                <div className="truncate">{session.title || '新对话'}</div>
                <div className="text-xs text-gray-400 mt-0.5">
                  {new Date(session.updated_at * 1000).toLocaleDateString('zh-CN')}
                </div>
              </button>
              {!isVirtual && (
                <button
                  onClick={() => deleteSession(session.id)}
                  className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 w-5 h-5 flex items-center justify-center text-gray-400 hover:text-red-500 text-xs rounded hover:bg-gray-200"
                  title="删除会话"
                >
                  ✕
                </button>
              )}
            </div>
          )
        })
      )}
    </div>
  )
}
