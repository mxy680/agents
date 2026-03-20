"use client"

import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

export function MarkdownContent({ content, className }: { content: string; className?: string }) {
  return (
    <div className={className}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          p: ({ children }) => <p className="mb-3 last:mb-0 leading-relaxed">{children}</p>,
          strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
          em: ({ children }) => <em className="italic">{children}</em>,
          del: ({ children }) => <del className="line-through text-muted-foreground">{children}</del>,
          ul: ({ children }) => <ul className="mb-3 ml-5 list-disc last:mb-0 space-y-1">{children}</ul>,
          ol: ({ children }) => <ol className="mb-3 ml-5 list-decimal last:mb-0 space-y-1">{children}</ol>,
          li: ({ children }) => <li className="leading-relaxed">{children}</li>,
          code: ({ children, className: codeClassName }) => {
            const isBlock = codeClassName?.includes("language-")
            if (isBlock) {
              return (
                <pre className="my-3 overflow-auto bg-background/80 border border-border/40 p-3 text-xs font-mono">
                  <code>{children}</code>
                </pre>
              )
            }
            return <code className="bg-background/60 border border-border/30 px-1.5 py-0.5 text-xs font-mono">{children}</code>
          },
          pre: ({ children }) => <>{children}</>,
          h1: ({ children }) => <h1 className="mt-6 mb-3 text-xl font-bold first:mt-0">{children}</h1>,
          h2: ({ children }) => <h2 className="mt-5 mb-2 text-lg font-bold first:mt-0">{children}</h2>,
          h3: ({ children }) => <h3 className="mt-4 mb-2 text-base font-semibold first:mt-0">{children}</h3>,
          h4: ({ children }) => <h4 className="mt-3 mb-1 text-sm font-semibold first:mt-0">{children}</h4>,
          blockquote: ({ children }) => (
            <blockquote className="my-3 border-l-2 border-border pl-4 text-muted-foreground italic">{children}</blockquote>
          ),
          hr: () => <hr className="my-4 border-border/60" />,
          a: ({ href, children }) => (
            <a href={href} target="_blank" rel="noopener noreferrer" className="text-primary underline underline-offset-2 hover:text-primary/80">
              {children}
            </a>
          ),
          table: ({ children }) => (
            <div className="my-3 overflow-x-auto border border-border/40">
              <table className="w-full border-collapse text-sm">{children}</table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-muted/40">{children}</thead>
          ),
          tbody: ({ children }) => <tbody>{children}</tbody>,
          tr: ({ children }) => (
            <tr className="border-b border-border/40">{children}</tr>
          ),
          th: ({ children }) => (
            <th className="px-3 py-2 text-left font-semibold text-xs border-r border-border/40 last:border-r-0">{children}</th>
          ),
          td: ({ children }) => (
            <td className="px-3 py-2 text-xs border-r border-border/40 last:border-r-0">{children}</td>
          ),
          input: ({ checked, ...props }) => (
            <input
              type="checkbox"
              checked={checked}
              readOnly
              className="mr-1.5 align-middle"
              {...props}
            />
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
