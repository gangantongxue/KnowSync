import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useChat } from '../store/chat-context'
import ChatWindow from '../components/Chat/ChatWindow'
import SessionList from '../components/Chat/SessionList'

export default function Chat() {
  const { sessionId } = useParams()
  const { loadSessions, loadMessages, currentSessionId, setCurrentSession } = useChat()

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  useEffect(() => {
    if (sessionId && sessionId !== currentSessionId) {
      loadMessages(sessionId)
    }
  }, [sessionId, currentSessionId, loadMessages])

  useEffect(() => {
    if (!currentSessionId && !sessionId) {
      setCurrentSession(null)
    }
  }, [currentSessionId, sessionId, setCurrentSession])

  return (
    <div className="h-full flex">
      <div className="flex-1 flex flex-col min-w-0">
        <ChatWindow />
      </div>

      <div className="w-60 border-l border-gray-200 bg-gray-50 shrink-0 overflow-y-auto">
        <SessionList />
      </div>
    </div>
  )
}
