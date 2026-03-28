import { Geist, Geist_Mono, Space_Grotesk, Lora, Figtree } from "next/font/google"

import "./globals.css"
import { cn } from "@/lib/utils";

const loraHeading = Lora({subsets:['latin'],variable:'--font-heading'});

const figtree = Figtree({subsets:['latin'],variable:'--font-sans'})
const geistMono = Geist_Mono({ subsets: ["latin"], variable: "--font-mono" })

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      className={cn("dark antialiased", geistMono.variable, "font-sans", figtree.variable, loraHeading.variable)}
    >
      <body>{children}</body>
    </html>
  )
}
