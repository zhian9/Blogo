import MDEditor from '@uiw/react-md-editor'
import '@uiw/react-md-editor/markdown-editor.css'

interface Props {
  value?: string
  onChange?: (value?: string) => void
  height?: number
}

export default function MarkdownEditor({ value = '', onChange, height = 400 }: Props) {
  return (
    <div className="markdown-editor-wrapper">
      <MDEditor
        value={value}
        onChange={(v) => onChange?.(v || '')}
        height={height}
        preview="live"
        hideToolbar={false}
      />
    </div>
  )
}
