import { NextResponse } from "next/server"
import { join } from "path"
import archiver from "archiver"
import { PassThrough } from "stream"

export async function GET() {
  const extensionDir = join(process.cwd(), "extension")

  const archive = archiver("zip", { zlib: { level: 9 } })
  const passthrough = new PassThrough()
  archive.pipe(passthrough)

  // Add all files from extension/ recursively, excluding README files
  archive.glob("**/*", {
    cwd: extensionDir,
    ignore: ["**/README.md", "**/.DS_Store"],
  })

  await archive.finalize()

  // Collect stream into a buffer
  const chunks: Buffer[] = []
  for await (const chunk of passthrough) {
    chunks.push(chunk as Buffer)
  }
  const buffer = Buffer.concat(chunks)

  return new NextResponse(buffer, {
    headers: {
      "Content-Type": "application/zip",
      "Content-Disposition": 'attachment; filename="emdash-extension.zip"',
      "Content-Length": buffer.length.toString(),
    },
  })
}
