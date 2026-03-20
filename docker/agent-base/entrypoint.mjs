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
  } catch (err) {
    if (err.code === "ENOENT") {
      // session.json not present — fall back to AGENT_PROMPT env var.
      const prompt = process.env.AGENT_PROMPT;
      if (!prompt) {
        process.stderr.write(
          `ERROR: ${SESSION_PATH} not found and AGENT_PROMPT is not set. Cannot determine what to do.\n`
        );
        process.exit(1);
      }
      session = { prompt };
    } else {
      // File exists but JSON is malformed — hard error, do not silently swallow.
      process.stderr.write(
        `ERROR: Failed to parse ${SESSION_PATH}: ${err.message}\n`
      );
      process.exit(1);
    }
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
    // If resume fails (stale session), retry without sessionId.
    // WARNING: retrying without a session may cause duplicate side effects if the
    // previous run partially completed.
    if (sessionId && err.message?.includes("No conversation found")) {
      process.stderr.write(
        `WARNING: Session ${sessionId} not found. Retrying without session — this may cause duplicate actions if the previous run partially completed.\n`
      );
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
