"use client"

import * as React from "react"
import {
  IconApps,
  IconPlugConnected,
  IconSettings,
  IconLogout,
  IconBrandGoogle,
  IconBrandGithub,
  IconBrandInstagram,
  IconUser,
  IconRobot,
} from "@tabler/icons-react"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { signOut } from "@/app/auth/actions"

const navMain = [
  {
    title: "Agents",
    items: [
      {
        title: "All Agents",
        url: "/agents",
        icon: IconRobot,
        isActive: false as const,
      },
    ],
  },
  {
    title: "Integrations",
    items: [
      {
        title: "All Integrations",
        url: "/integrations",
        icon: IconPlugConnected,
        isActive: true as const,
      },
    ],
  },
  {
    title: "Providers",
    items: [
      { title: "Google", url: "/integrations#google", icon: IconBrandGoogle, isActive: false as const },
      { title: "GitHub", url: "/integrations#github", icon: IconBrandGithub, isActive: false as const },
      { title: "Instagram", url: "/integrations#instagram", icon: IconBrandInstagram, isActive: false as const },
    ],
  },
]

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  user?: { email?: string; name?: string }
}

export function AppSidebar({ user, ...props }: AppSidebarProps) {
  const initials = user?.name
    ? user.name.split(" ").map((n) => n[0]).join("").toUpperCase()
    : user?.email?.[0]?.toUpperCase() ?? "?"

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
                  <span className="font-semibold">Agent Marketplace</span>
                  <span className="text-xs text-muted-foreground">Integrations</span>
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
                    <SidebarMenuButton asChild isActive={item.isActive}>
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
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton size="lg">
                  <Avatar className="size-8">
                    <AvatarFallback>{initials}</AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col gap-0.5 leading-none">
                    <span className="font-medium">{user?.name ?? "User"}</span>
                    <span className="text-xs text-muted-foreground">
                      {user?.email ?? ""}
                    </span>
                  </div>
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width]"
                align="start"
              >
                <DropdownMenuItem asChild>
                  <a href="/settings">
                    <IconSettings />
                    Settings
                  </a>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => {
                    const form = document.createElement("form")
                    form.method = "POST"
                    form.action = "/auth/sign-out"
                    document.body.appendChild(form)
                    form.submit()
                  }}
                >
                  <IconLogout />
                  Sign out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
