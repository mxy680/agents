"use client"

import * as React from "react"
import { getScenariosForAgent } from "@/lib/demo-scenarios"
import { Card, CardContent } from "@/components/ui/card"
import {
  IconRobot,
  IconMailOpened,
  IconUrgent,
  IconSearch,
  IconGitPullRequest,
  IconChecklist,
  IconBook,
  IconChartBar,
  IconBell,
  IconUsers,
  IconHeartbeat,
  IconDashboard,
} from "@tabler/icons-react"

const SCENARIO_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  IconMailOpened,
  IconUrgent,
  IconSearch,
  IconGitPullRequest,
  IconChecklist,
  IconBook,
  IconChartBar,
  IconBell,
  IconUsers,
  IconHeartbeat,
  IconDashboard,
}

interface PromptPaletteProps {
  agentName: string
  agentDisplayName: string
  onSelect: (prompt: string) => void
}

export function PromptPalette({ agentName, agentDisplayName, onSelect }: PromptPaletteProps) {
  const scenarios = getScenariosForAgent(agentName)

  return (
    <div className="flex flex-col items-center gap-6 py-12 px-4 max-w-2xl mx-auto">
      <div className="flex flex-col items-center gap-2">
        <div className="flex size-12 items-center justify-center bg-muted rounded-lg">
          <IconRobot className="size-6" />
        </div>
        <h2 className="text-lg font-semibold">{agentDisplayName}</h2>
        <p className="text-sm text-muted-foreground text-center">
          Choose a task below or type your own message.
        </p>
      </div>

      {scenarios.length > 0 && (
        <div className="grid gap-3 w-full sm:grid-cols-2 lg:grid-cols-3">
          {scenarios.map((scenario) => {
            const ScenarioIcon = SCENARIO_ICONS[scenario.icon] || IconRobot
            return (
              <Card
                key={scenario.id}
                className="cursor-pointer transition-colors hover:bg-muted/50"
                onClick={() => onSelect(scenario.prompt)}
              >
                <CardContent className="flex flex-col gap-2 p-4">
                  <div className="flex items-center gap-2">
                    <ScenarioIcon className="size-4 text-muted-foreground" />
                    <span className="text-sm font-medium">{scenario.label}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {scenario.description}
                  </p>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
