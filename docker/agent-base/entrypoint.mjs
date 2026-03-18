#!/usr/bin/env node

/**
 * Docker container entrypoint for agent pods.
 *
 * Reads session.json from /agent/workspace/, invokes Claude via the Agent SDK,
 * and streams events as NDJSON to stdout.
 */

import { readFileSync, existsSync } from "fs";
import { query } from "@anthropic-ai/claude-agent-sdk";

const SESSION_PATH = "/agent/workspace/session.json";

function loadSystemPrompt() {
  let prompt = "";
  const rolePath = "/agent/workspace/role.md";
  const claudePath = "/agent/workspace/CLAUDE.md";

  if (existsSync(rolePath)) {
    prompt += readFileSync(rolePath, "utf-8");
  }
  if (existsSync(claudePath)) {
    if (prompt) prompt += "\n\n";
    prompt += readFileSync(claudePath, "utf-8");
  }
  return prompt || "You are a helpful assistant.";
}

function buildOptions(prompt, sessionId, model) {
  return {
    prompt,
    options: {
      cwd: "/agent/workspace",
      permissionMode: "bypassPermissions",
      allowDangerouslySkipPermissions: true,
      systemPrompt: loadSystemPrompt(),
      includePartialMessages: true,
      maxTurns: 20,
      model: model || "claude-sonnet-4-6",
      ...(sessionId ? { resume: sessionId } : {}),
    },
  };
}

async function run(prompt, sessionId, model) {
  const conversation = query(buildOptions(prompt, sessionId, model));
  for await (const event of conversation) {
    process.stdout.write(JSON.stringify(event) + "\n");
  }
}

async function main() {
  let session;
  try {
    session = JSON.parse(readFileSync(SESSION_PATH, "utf-8"));
  } catch {
    // No session.json — fall back to env vars or default prompt
    session = {
      prompt:
        process.env.AGENT_PROMPT ||
        "List my recent unread emails using: integrations gmail messages list --query=is:unread --since=24h --json. Then summarize them.",
    };
  }

  const { prompt, sessionId, model } = session;

  if (!prompt) {
    process.stderr.write("session.json must contain a 'prompt' field\n");
    process.exit(1);
  }

  process.stderr.write(`Agent starting (model: ${model || "claude-sonnet-4-6"})\n`);

  try {
    await run(prompt, sessionId, model);
  } catch (err) {
    // If resume fails (stale session), retry without sessionId
    if (sessionId && err.message?.includes("No conversation found")) {
      process.stderr.write(`Session ${sessionId} not found, starting fresh\n`);
      try {
        await run(prompt, null, model);
        return;
      } catch (retryErr) {
        process.stderr.write(`Agent error: ${retryErr.message}\n`);
        process.exit(1);
      }
    }
    process.stderr.write(`Agent error: ${err.message}\n${err.stack || ""}\n`);
    process.exit(1);
  }
}

main();
