"""Email Assistant Agent — Claude Agent SDK entry point."""

import os
import subprocess

import anthropic


def run_tool(command: str) -> str:
    """Execute an integrations CLI command and return the output."""
    result = subprocess.run(
        command, shell=True, capture_output=True, text=True, timeout=30
    )
    if result.returncode != 0:
        return f"Error: {result.stderr.strip()}"
    return result.stdout.strip()


def main():
    client = anthropic.Anthropic()

    # Read role and instructions
    role_path = os.environ.get("AGENT_ROLE_PATH", "/agent/workspace/role.md")
    claude_path = os.environ.get("AGENT_CLAUDE_PATH", "/agent/workspace/CLAUDE.md")

    system_prompt = ""
    for path in [role_path, claude_path]:
        if os.path.exists(path):
            with open(path) as f:
                system_prompt += f.read() + "\n\n"

    if not system_prompt:
        system_prompt = "You are a helpful email assistant."

    print(f"Email Assistant started (model: {os.environ.get('AGENT_MODEL', 'claude-sonnet-4-6')})")
    print("Ready to process tasks.")

    # In production, this would connect to a task queue or WebSocket
    # For now, just verify the agent can initialize
    response = client.messages.create(
        model=os.environ.get("AGENT_MODEL", "claude-sonnet-4-6"),
        max_tokens=1024,
        system=system_prompt,
        messages=[{"role": "user", "content": "Briefly confirm you are online and ready."}],
    )
    print(f"Agent response: {response.content[0].text}")


if __name__ == "__main__":
    main()
