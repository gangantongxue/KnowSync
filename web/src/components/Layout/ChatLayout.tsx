import { type ReactNode } from 'react'

interface ChatLayoutProps {
  sidebar?: ReactNode
  main: ReactNode
  rightPanel?: ReactNode
}

export default function ChatLayout({ sidebar, main, rightPanel }: ChatLayoutProps) {
  return (
    <div className="h-full flex">
      {sidebar && (
        <div className="w-56 border-r border-gray-200 bg-gray-50 shrink-0 overflow-y-auto">
          {sidebar}
        </div>
      )}
      <div className="flex-1 flex flex-col min-w-0">
        {main}
      </div>
      {rightPanel && (
        <div className="w-60 border-l border-gray-200 bg-gray-50 shrink-0 overflow-y-auto">
          {rightPanel}
        </div>
      )}
    </div>
  )
}
