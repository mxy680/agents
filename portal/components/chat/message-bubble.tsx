"use client"

import * as React from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { IconRobot, IconUser, IconFile, IconDownload, IconPhoto } from "@tabler/icons-react"
import { ToolCard } from "./tool-card"
import { cn } from "@/lib/utils"

interface ContentBlock {
  type: "text" | "tool" | "file"
  content?: string
  id?: string
  name?: string
  finalInput?: string
  result?: string
  url?: string
  size?: number
  fileType?: string
}

interface MessageBubbleProps {
  role: "user" | "assistant"
  blocks: ContentBlock[]
  isStreaming: boolean
}

function InlineMarkdown({ content }: { content: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
        strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
        em: ({ children }) => <em className="italic">{children}</em>,
        del: ({ children }) => <del className="line-through text-muted-foreground">{children}</del>,
        ul: ({ children }) => <ul className="mb-2 ml-4 list-disc last:mb-0">{children}</ul>,
        ol: ({ children }) => <ol className="mb-2 ml-4 list-decimal last:mb-0">{children}</ol>,
        li: ({ children }) => <li className="mb-0.5">{children}</li>,
        code: ({ children, className }) => {
          const isBlock = className?.includes("language-")
          if (isBlock) {
            return (
              <pre className="my-2 overflow-auto rounded bg-background/80 p-2 text-[11px] font-mono">
                <code>{children}</code>
              </pre>
            )
          }
          return (
            <code className="rounded bg-background/60 px-1 py-0.5 text-[11px] font-mono">
              {children}
            </code>
          )
        },
        pre: ({ children }) => <>{children}</>,
        h1: ({ children }) => <h1 className="mb-2 text-base font-bold">{children}</h1>,
        h2: ({ children }) => <h2 className="mb-2 text-sm font-bold">{children}</h2>,
        h3: ({ children }) => <h3 className="mb-1 text-xs font-bold">{children}</h3>,
        blockquote: ({ children }) => (
          <blockquote className="my-2 border-l-2 border-border pl-3 text-muted-foreground italic">
            {children}
          </blockquote>
        ),
        hr: () => <hr className="my-3 border-border/60" />,
        a: ({ href, children }) => (
          <a href={href} target="_blank" rel="noopener noreferrer" className="text-primary underline underline-offset-2 hover:text-primary/80">
            {children}
          </a>
        ),
        table: ({ children }) => (
          <div className="my-2 overflow-x-auto">
            <table className="w-full border-collapse text-[11px]">{children}</table>
          </div>
        ),
        thead: ({ children }) => <thead className="bg-background/60">{children}</thead>,
        tr: ({ children }) => <tr className="border-b border-border/40">{children}</tr>,
        th: ({ children }) => <th className="px-2 py-1.5 text-left font-semibold border border-border/40">{children}</th>,
        td: ({ children }) => <td className="px-2 py-1.5 border border-border/40">{children}</td>,
        input: ({ checked, ...props }) => (
          <input type="checkbox" checked={checked} readOnly className="mr-1.5 align-middle" {...props} />
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  )
}

export function MessageBubble({ role, blocks, isStreaming }: MessageBubbleProps) {
  const isUser = role === "user"
  const hasContent = blocks.length > 0
  const isThinking = isStreaming && !hasContent
  const lastBlock = blocks[blocks.length - 1]
  const isStreamingText = isStreaming && lastBlock?.type === "text"

  return (
    <div className={cn("flex gap-2.5 mb-4", isUser && "flex-row-reverse")}>
      <div
        className={cn(
          "flex size-7 shrink-0 items-center justify-center mt-0.5",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}
      >
        {isUser ? <IconUser className="size-3.5" /> : <IconRobot className="size-3.5" />}
      </div>
      <div className={cn("flex flex-col gap-1.5 max-w-[80%]", isUser && "items-end")}>
        {isThinking && (
          <div className="bg-muted/60 border border-border/40 px-3 py-2">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <div className="flex gap-1">
                <span className="size-1.5 rounded-full bg-muted-foreground/60 animate-bounce" style={{ animationDelay: "0ms" }} />
                <span className="size-1.5 rounded-full bg-muted-foreground/60 animate-bounce" style={{ animationDelay: "150ms" }} />
                <span className="size-1.5 rounded-full bg-muted-foreground/60 animate-bounce" style={{ animationDelay: "300ms" }} />
              </div>
              <span className="italic">Working…</span>
            </div>
          </div>
        )}

        {blocks.map((block, i) => {
          if (block.type === "text") {
            const isLast = i === blocks.length - 1
            return (
              <div
                key={i}
                className={cn(
                  "px-3 py-2 text-xs/relaxed",
                  isUser
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted/60 text-foreground border border-border/40"
                )}
              >
                {isUser ? (
                  block.content
                ) : (
                  <InlineMarkdown content={block.content || ""} />
                )}
                {isLast && isStreamingText && (
                  <span className="inline-block w-1.5 h-3 ml-0.5 bg-current animate-pulse align-text-bottom" />
                )}
              </div>
            )
          }

          if (block.type === "tool" && block.id && block.name) {
            return (
              <ToolCard
                key={block.id}
                name={block.name}
                id={block.id}
                input={block.finalInput || ""}
                result={block.result}
                isStreaming={isStreaming}
              />
            )
          }

          if (block.type === "file" && block.url && block.name) {
            const isImage = block.fileType?.startsWith("image/")
            const sizeStr = block.size
              ? block.size > 1024 * 1024
                ? `${(block.size / (1024 * 1024)).toFixed(1)} MB`
                : `${Math.round(block.size / 1024)} KB`
              : ""
            return (
              <div key={i} className="border border-border rounded bg-muted/30 overflow-hidden max-w-xs">
                {isImage && (
                  <img src={block.url} alt={block.name} className="w-full max-h-48 object-cover" />
                )}
                <a
                  href={block.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  download={block.name}
                  className="flex items-center gap-2 px-3 py-2 hover:bg-muted/60 transition-colors"
                >
                  {isImage ? (
                    <IconPhoto className="size-4 text-blue-400 shrink-0" />
                  ) : (
                    <IconFile className="size-4 text-muted-foreground shrink-0" />
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium truncate">{block.name}</p>
                    {sizeStr && <p className="text-[10px] text-muted-foreground">{sizeStr}</p>}
                  </div>
                  <IconDownload className="size-3.5 text-muted-foreground shrink-0" />
                </a>
              </div>
            )
          }

          return null
        })}
      </div>
    </div>
  )
}
