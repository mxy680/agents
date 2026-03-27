"use client"

import * as React from "react"
import {
  IconApps,
  IconPlugConnected,
  IconRobot,
  IconCalendarEvent,
  IconUsers,
  IconExternalLink,
} from "@tabler/icons-react"

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"

const navMain = [
  {
    title: "Dashboard",
    items: [
      {
        title: "Integrations",
        url: "/integrations",
        icon: IconPlugConnected,
      },
      {
        title: "Agents",
        url: "/agents",
        icon: IconRobot,
      },
      {
        title: "Jobs",
        url: "/jobs",
        icon: IconCalendarEvent,
      },
    ],
  },
  {
    title: "Admin",
    items: [
      {
        title: "Clients",
        url: "/admin/clients",
        icon: IconUsers,
      },
      {
        title: "Client Portal",
        url: "/client",
        icon: IconExternalLink,
      },
    ],
  },
]

export function AppSidebar(props: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="/integrations">
                <div className="flex size-8 items-center justify-center bg-primary text-primary-foreground">
                  <IconApps className="size-4" />
                </div>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="font-semibold">Engagent</span>
                  <span className="text-xs text-muted-foreground">Engineer-as-a-Service</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {navMain.map((section) => (
          <SidebarGroup key={section.title}>
            <SidebarGroupLabel>{section.title}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {section.items.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild>
                      <a href={item.url}>
                        <item.icon />
                        {item.title}
                      </a>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
