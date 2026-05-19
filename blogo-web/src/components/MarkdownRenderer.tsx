import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import { Typography } from 'antd'
import 'highlight.js/styles/github-dark.css'

const { Text, Title, Link } = Typography

interface Props {
  content: string
}

export default function MarkdownRenderer({ content }: Props) {
  return (
    <div className="markdown-body">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
        components={{
          h1: ({ children }) => <Title level={1} style={{ marginTop: 32, marginBottom: 16 }}>{children}</Title>,
          h2: ({ children }) => <Title level={2} style={{ marginTop: 28, marginBottom: 14 }}>{children}</Title>,
          h3: ({ children }) => <Title level={3} style={{ marginTop: 24, marginBottom: 12 }}>{children}</Title>,
          h4: ({ children }) => <Title level={4} style={{ marginTop: 20, marginBottom: 10 }}>{children}</Title>,
          p: ({ children }) => <Text style={{ fontSize: 16, lineHeight: 1.8, display: 'block', marginBottom: 16 }}>{children}</Text>,
          a: ({ href, children }) => <Link href={href} target="_blank" rel="noopener noreferrer">{children}</Link>,
          blockquote: ({ children }) => (
            <blockquote style={{ borderLeft: '4px solid #1890ff', paddingLeft: 16, margin: '16px 0', color: '#666', background: 'rgba(24,144,255,0.05)', padding: '12px 16px', borderRadius: 4 }}>
              {children}
            </blockquote>
          ),
          code: (props) => {
            const { className, children, ...rest } = props as { className?: string; children: React.ReactNode }
            const isInline = !className
            if (isInline) {
              return <code style={{ background: 'rgba(0,0,0,0.06)', padding: '2px 6px', borderRadius: 3, fontSize: '0.9em' }} {...rest}>{children}</code>
            }
            return <code className={className} {...rest}>{children}</code>
          },
          img: ({ src, alt }) => (
            <img
              src={src}
              alt={alt}
              loading="lazy"
              style={{ maxWidth: '100%', borderRadius: 8, margin: '16px 0' }}
            />
          ),
          table: ({ children }) => (
            <div style={{ overflowX: 'auto', marginBottom: 16 }}>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>{children}</table>
            </div>
          ),
          th: ({ children }) => <th style={{ border: '1px solid #e8e8e8', padding: '8px 12px', background: '#fafafa' }}>{children}</th>,
          td: ({ children }) => <td style={{ border: '1px solid #e8e8e8', padding: '8px 12px' }}>{children}</td>,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
