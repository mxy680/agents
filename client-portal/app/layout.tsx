import { Geist, Geist_Mono, Space_Grotesk, Lora } from "next/font/google"

import "./globals.css"
import { cn } from "@/lib/utils";

const loraHeading = Lora({subsets:['latin'],variable:'--font-heading'});

const spaceGrotesk = Space_Grotesk({subsets:['latin'],variable:'--font-sans'})
const geistMono = Geist_Mono({ subsets: ["latin"], variable: "--font-mono" })

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      className={cn("dark antialiased", geistMono.variable, "font-sans", spaceGrotesk.variable, loraHeading.variable)}
    >
      <body>{children}</body>
    </html>
  )
}
