import { useState } from 'react'
import Editor from '@monaco-editor/react'
import { Streamdown } from '../Chat/Streamdown'

interface MarkdownEditorProps {
  value: string
  onChange: (value: string) => void
  readOnly?: boolean
}

export default function MarkdownEditor({ value, onChange, readOnly }: MarkdownEditorProps) {
  const [preview, setPreview] = useState(false)

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 bg-gray-50">
        <span className="text-xs text-gray-400">Markdown</span>
        <button
          onClick={() => setPreview(!preview)}
          className="text-xs px-2 py-1 bg-white border border-gray-200 rounded hover:bg-gray-50"
        >
          {preview ? '编辑' : '预览'}
        </button>
      </div>
      <div className="flex-1">
        {preview ? (
          <div className="p-4 overflow-y-auto max-w-3xl mx-auto">
            <Streamdown content={value} />
          </div>
        ) : (
          <Editor
            defaultLanguage="markdown"
            value={value}
            onChange={(val) => onChange(val || '')}
            options={{
              readOnly,
              minimap: { enabled: false },
              fontSize: 14,
              lineNumbers: 'on',
              wordWrap: 'on',
              scrollBeyondLastLine: false,
            }}
          />
        )}
      </div>
    </div>
  )
}
