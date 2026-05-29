import { memo } from 'react'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import type { Components } from 'react-markdown'

interface StreamdownProps {
  content: string
}

const components: Components = {
  code({ className, children, ...props }) {
    const match = /language-(\w+)/.exec(className || '')
    const code = String(children).replace(/\n$/, '')
    if (match) {
      return (
        <SyntaxHighlighter style={oneLight} language={match[1]} PreTag="div">
          {code}
        </SyntaxHighlighter>
      )
    }
    return <code className="bg-gray-100 px-1 rounded text-sm" {...props}>{children}</code>
  },
  pre({ children }) {
    return <div className="my-2">{children}</div>
  },
}

function StreamdownInner({ content }: StreamdownProps) {
  return (
    <div className="prose prose-sm max-w-none prose-headings:text-gray-800 prose-p:text-gray-700 prose-a:text-emerald-600 prose-code:bg-gray-100 prose-code:px-1 prose-code:rounded prose-code:text-sm">
      <ReactMarkdown components={components}>
        {content || ''}
      </ReactMarkdown>
    </div>
  )
}

export const Streamdown = memo(StreamdownInner)
