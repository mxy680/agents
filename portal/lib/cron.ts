import { CronExpressionParser } from "cron-parser"

/**
 * Get the next run time for a cron expression.
 */
export function getNextRun(cronExpression: string, from?: Date): Date {
  const interval = CronExpressionParser.parse(cronExpression, {
    currentDate: from ?? new Date(),
  })
  return interval.next().toDate()
}

/**
 * Check if a cron expression is due within a time window.
 * Used by the scheduler to determine if a job should run.
 */
export function isDue(cronExpression: string, now: Date, windowMs: number = 60000): boolean {
  const windowStart = new Date(now.getTime() - windowMs)
  const interval = CronExpressionParser.parse(cronExpression, {
    currentDate: windowStart,
  })
  const nextRun = interval.next().toDate()
  return nextRun.getTime() <= now.getTime()
}

/**
 * Convert a cron expression to a human-readable string.
 */
export function toHumanReadable(cronExpression: string): string {
  const parts = cronExpression.trim().split(/\s+/)
  if (parts.length !== 5) return cronExpression

  const [minute, hour, dayOfMonth, month, dayOfWeek] = parts

  // Every minute
  if (minute === "*" && hour === "*" && dayOfMonth === "*" && month === "*" && dayOfWeek === "*") {
    return "Every minute"
  }

  // Every N minutes
  if (minute.startsWith("*/") && hour === "*" && dayOfMonth === "*" && month === "*" && dayOfWeek === "*") {
    return `Every ${minute.slice(2)} minutes`
  }

  // Hourly
  if (minute !== "*" && hour === "*" && dayOfMonth === "*" && month === "*" && dayOfWeek === "*") {
    return `Hourly at :${minute.padStart(2, "0")}`
  }

  // Every N hours
  if (minute !== "*" && hour.startsWith("*/") && dayOfMonth === "*" && month === "*" && dayOfWeek === "*") {
    return `Every ${hour.slice(2)} hours at :${minute.padStart(2, "0")}`
  }

  // Daily
  if (minute !== "*" && hour !== "*" && !hour.includes("/") && !hour.includes(",") && dayOfMonth === "*" && month === "*" && dayOfWeek === "*") {
    return `Daily at ${hour.padStart(2, "0")}:${minute.padStart(2, "0")} UTC`
  }

  // Weekly
  const dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"]
  if (minute !== "*" && hour !== "*" && dayOfMonth === "*" && month === "*" && dayOfWeek !== "*") {
    const dayIdx = parseInt(dayOfWeek)
    const dayName = dayNames[dayIdx] ?? dayOfWeek
    return `Weekly on ${dayName} at ${hour.padStart(2, "0")}:${minute.padStart(2, "0")} UTC`
  }

  // Monthly
  if (minute !== "*" && hour !== "*" && dayOfMonth !== "*" && month === "*" && dayOfWeek === "*") {
    const suffix = dayOfMonth === "1" ? "st" : dayOfMonth === "2" ? "nd" : dayOfMonth === "3" ? "rd" : "th"
    return `Monthly on the ${dayOfMonth}${suffix} at ${hour.padStart(2, "0")}:${minute.padStart(2, "0")} UTC`
  }

  return cronExpression
}
