import { memo } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
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
  table({ children }) {
    return <div className="overflow-x-auto my-3"><table className="min-w-full border-collapse border border-gray-300 text-sm">{children}</table></div>
  },
  th({ children }) {
    return <th className="border border-gray-300 px-3 py-2 bg-gray-100 font-medium text-left">{children}</th>
  },
  td({ children }) {
    return <td className="border border-gray-300 px-3 py-2">{children}</td>
  },
  img({ src, alt }) {
    return (
      <img
        src={src}
        alt={alt}
        className="max-w-full h-auto rounded-lg my-2"
        loading="lazy"
      />
    )
  },
  a({ href, children }) {
    if (!href) return <>{children}</>
    const isImageOnly = (() => {
      if (!children) return false
      if (typeof children === 'object' && 'type' in children) {
        const child = children as any
        if (child.type === 'img') return true
      }
      return false
    })()
    if (isImageOnly) {
      return <>{children}</>
    }
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" className="text-emerald-600 hover:text-emerald-700 underline">
        {children}
      </a>
    )
  },
}

function StreamdownInner({ content }: StreamdownProps) {
  return (
    <div className="prose prose-sm max-w-none prose-headings:text-gray-800 prose-p:text-gray-700 prose-a:text-emerald-600 prose-code:bg-gray-100 prose-code:px-1 prose-code:rounded prose-code:text-sm">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {content || ''}
      </ReactMarkdown>
    </div>
  )
}

export const Streamdown = memo(StreamdownInner)
