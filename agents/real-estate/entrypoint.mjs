#!/usr/bin/env node

import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";

// Resolve Agent SDK from global npm modules (ESM ignores NODE_PATH)
const __dirname = dirname(fileURLToPath(import.meta.url));

// Try multiple global npm locations (macOS Homebrew vs Linux)
let query;
for (const prefix of [
  "/opt/homebrew/lib/node_modules",
  "/usr/local/lib/node_modules",
  "/usr/lib/node_modules",
]) {
  try {
    const mod = await import(join(prefix, "@anthropic-ai", "claude-agent-sdk", "sdk.mjs"));
    query = mod.query;
    break;
  } catch {}
}
if (!query) {
  // Fallback: try bare import (works if NODE_PATH is set correctly)
  const mod = await import("@anthropic-ai/claude-agent-sdk");
  query = mod.query;
}

const sessionPath = process.argv[2];
if (!sessionPath) {
  console.error("Usage: node entrypoint.mjs <session.json>");
  process.exit(1);
}

const session = JSON.parse(readFileSync(sessionPath, "utf-8"));
const systemPrompt = session.systemPrompt || "You are a helpful assistant.";

// When stdout is a TTY (terminal), use colored output.
// When piped (portal local-runner), output raw NDJSON for parsing.
const isTTY = process.stdout.isTTY;

const queryOptions = {
  cwd: process.cwd(),
  permissionMode: "bypassPermissions",
  allowDangerouslySkipPermissions: true,
  systemPrompt,
  maxTurns: session.maxTurns || 500,
  model: session.model || "claude-opus-4-6",
  includePartialMessages: !isTTY,
};

// Resume existing conversation if sessionId is provided
if (session.sessionId) {
  queryOptions.resume = session.sessionId;
}

const conversation = query({
  prompt: session.prompt,
  options: queryOptions,
});

// #13: Wrap all agent loops in try/catch so errors are reported cleanly
try {
  if (isTTY) {
    // ── Terminal mode: colored human-readable output ──
    const DIM = "\x1b[2m";
    const RESET = "\x1b[0m";
    const BOLD = "\x1b[1m";
    const CYAN = "\x1b[36m";
    const GREEN = "\x1b[32m";
    const YELLOW = "\x1b[33m";
    const RED = "\x1b[31m";
    const GRAY = "\x1b[90m";
    const write = (s) => process.stdout.write(s);
    let turnCount = 0;

    for await (const event of conversation) {
      if (event.type === "assistant" && event.message?.content) {
        for (const block of event.message.content) {
          if (block.type === "text") {
            write(`\n${BOLD}${block.text}${RESET}\n`);
          } else if (block.type === "tool_use") {
            turnCount++;
            const name = block.name || "tool";
            const input = block.input || {};
            if (name === "Bash") {
              const cmd = (input.command || "").split("\n")[0].slice(0, 120);
              const desc = input.description || "";
              write(`\n${CYAN}[${turnCount}] ${desc}${RESET}\n`);
              write(`${DIM}  $ ${cmd}${cmd.length >= 120 ? "..." : ""}${RESET}\n`);
            } else if (name === "Write") {
              write(`\n${YELLOW}[${turnCount}] Writing ${input.file_path || "file"}${RESET}\n`);
            } else if (name === "Read") {
              write(`${GRAY}[${turnCount}] Reading ${input.file_path || "file"}${RESET}\n`);
            } else if (name === "Agent") {
              write(`\n${RED}[${turnCount}] Spawning agent: ${input.description || ""}${RESET}\n`);
            } else {
              const summary = JSON.stringify(input).slice(0, 100);
              write(`\n${CYAN}[${turnCount}] ${name}: ${summary}${RESET}\n`);
            }
          }
        }
      } else if (event.type === "tool_result") {
        const raw = typeof event.content === "string" ? event.content : JSON.stringify(event.content || "");
        const lines = raw.slice(0, 200).split("\n").slice(0, 3).join("\n    ");
        if (raw.toLowerCase().includes("error")) {
          write(`  ${RED}✗ ${lines}${RESET}\n`);
        } else {
          write(`  ${GREEN}✓${RESET} ${DIM}${lines}${raw.length > 200 ? "..." : ""}${RESET}\n`);
        }
      } else if (event.type === "result") {
        write(`\n${GREEN}${BOLD}━━━ Agent finished (${turnCount} tool calls) ━━━${RESET}\n`);
      }
    }
  } else {
    // ── Piped mode: raw NDJSON for portal local-runner ──
    for await (const event of conversation) {
      process.stdout.write(JSON.stringify(event) + "\n");
    }
  }
} catch (err) {
  const message = err instanceof Error ? err.message : String(err);
  const stack = err instanceof Error ? err.stack : "";
  process.stderr.write(`Agent error: ${message}\n`);
  if (stack) process.stderr.write(`Stack: ${stack}\n`);
  // Also write as NDJSON so the portal can see it
  if (!isTTY) {
    process.stdout.write(JSON.stringify({ type: "result", subtype: "error", is_error: true, result: message }) + "\n");
  }
  process.exit(1);
}
