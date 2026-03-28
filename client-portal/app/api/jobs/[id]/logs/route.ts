import { NextRequest } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params

  const admin = createAdminClient()

  // Verify the run exists
  const { data: run } = await admin
    .from("local_job_runs")
    .select("id, log, status, deliverables")
    .eq("id", id)
    .single()

  if (!run) {
    return new Response("Not found", { status: 404 })
  }

  const encoder = new TextEncoder()

  const stream = new ReadableStream({
    async start(controller) {
      let lastLength = 0

      // If already done, send full log + done event and close
      if (run.status === "completed" || run.status === "failed") {
        if (run.log) {
          controller.enqueue(
            encoder.encode(
              `data: ${JSON.stringify({ type: "log", content: run.log })}\n\n`
            )
          )
        }
        controller.enqueue(
          encoder.encode(
            `data: ${JSON.stringify({ type: "done", status: run.status, deliverables: run.deliverables ?? {} })}\n\n`
          )
        )
        controller.close()
        return
      }

      // Stream live logs by polling every 500ms
      const interval = setInterval(async () => {
        try {
          const { data } = await admin
            .from("local_job_runs")
            .select("log, status, deliverables")
            .eq("id", id)
            .single()

          if (!data) return

          const currentLog = data.log ?? ""
          if (currentLog.length > lastLength) {
            const newContent = currentLog.slice(lastLength)
            controller.enqueue(
              encoder.encode(
                `data: ${JSON.stringify({ type: "log", content: newContent })}\n\n`
              )
            )
            lastLength = currentLog.length
          }

          if (data.status === "completed" || data.status === "failed") {
            controller.enqueue(
              encoder.encode(
                `data: ${JSON.stringify({ type: "done", status: data.status, deliverables: data.deliverables ?? {} })}\n\n`
              )
            )
            clearInterval(interval)
            controller.close()
          }
        } catch (err) {
          console.error("[jobs/logs] Poll error:", err)
          clearInterval(interval)
          controller.close()
        }
      }, 500)

      // Clean up if client disconnects
      request.signal.addEventListener("abort", () => {
        clearInterval(interval)
        try { controller.close() } catch { /* already closed */ }
      })
    },
  })

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    },
  })
}
