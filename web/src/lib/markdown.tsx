import type { Components } from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

/** 解析相对路径，处理 ./ 和 ../ */
export function resolveRelativePath(baseDir: string, relative: string): string {
  let path = relative.replace(/^\.\/+/, '')
  const parts = baseDir ? baseDir.split('/') : []
  while (path.startsWith('../')) {
    if (parts.length > 0) parts.pop()
    path = path.slice(3)
  }
  if (parts.length > 0) {
    return parts.join('/') + '/' + path
  }
  return path
}

/** 将 markdown 中的相对路径图片转为绝对文件 URL */
export function resolveImageUrls(md: string, fileDir: string, baseUrl: string): string {
  return md.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_match, alt, src) => {
    if (/^https?:\/\//.test(src) || src.startsWith('/')) return _match
    return `![${alt}](${baseUrl}/${resolveRelativePath(fileDir, src)})`
  })
}

/** 文章渲染用的 ReactMarkdown components，完整排版样式 */
export const markdownComponents: Components = {
  h1({ children }) {
    return <h1 className="text-2xl font-bold text-gray-900 mt-8 mb-4 pb-2 border-b border-gray-200">{children}</h1>
  },
  h2({ children }) {
    return <h2 className="text-xl font-semibold text-gray-900 mt-7 mb-3 pb-1.5 border-b border-gray-100">{children}</h2>
  },
  h3({ children }) {
    return <h3 className="text-lg font-semibold text-gray-900 mt-6 mb-2">{children}</h3>
  },
  h4({ children }) {
    return <h4 className="text-base font-semibold text-gray-900 mt-5 mb-2">{children}</h4>
  },
  h5({ children }) {
    return <h5 className="text-sm font-semibold text-gray-900 mt-4 mb-2">{children}</h5>
  },
  h6({ children }) {
    return <h6 className="text-sm font-medium text-gray-500 mt-4 mb-2">{children}</h6>
  },
  p({ children }) {
    return <p className="text-sm text-gray-700 leading-7 my-3">{children}</p>
  },
  blockquote({ children }) {
    return <blockquote className="border-l-4 border-emerald-300 bg-emerald-50/50 pl-4 pr-2 py-2 my-4 text-sm text-gray-600 rounded-r">{children}</blockquote>
  },
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
    return <code className="bg-gray-100 text-rose-600 text-sm px-1.5 py-0.5 rounded font-mono" {...props}>{children}</code>
  },
  pre({ children }) {
    return (
      <div className="my-4 rounded-lg border border-gray-200 overflow-hidden shadow-sm">
        {children}
      </div>
    )
  },
  table({ children }) {
    return (
      <div className="overflow-x-auto my-4 rounded-lg border border-gray-200">
        <table className="min-w-full border-collapse text-sm">{children}</table>
      </div>
    )
  },
  thead({ children }) {
    return <thead className="bg-gray-50 border-b border-gray-200">{children}</thead>
  },
  th({ children }) {
    return <th className="border border-gray-200 px-4 py-2.5 font-medium text-gray-700 text-left">{children}</th>
  },
  td({ children }) {
    return <td className="border border-gray-200 px-4 py-2.5 text-gray-700">{children}</td>
  },
  ul({ children }) {
    return <ul className="list-disc list-inside my-3 space-y-1 text-sm text-gray-700">{children}</ul>
  },
  ol({ children }) {
    return <ol className="list-decimal list-inside my-3 space-y-1 text-sm text-gray-700">{children}</ol>
  },
  li({ children }) {
    return <li className="my-1">{children}</li>
  },
  hr() {
    return <hr className="my-6 border-gray-200" />
  },
  img({ src, alt }) {
    return (
      <img
        src={src}
        alt={alt}
        className="max-w-full h-auto rounded-lg my-4"
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
