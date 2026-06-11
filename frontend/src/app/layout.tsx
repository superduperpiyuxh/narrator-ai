import type { Metadata } from "next";
import { Space_Grotesk, Fira_Code } from "next/font/google";
import { Providers } from "@/components/Providers";
import { API_BASE } from "@/lib/api";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  variable: "--font-space-grotesk",
  subsets: ["latin"],
  display: "swap",
});

const firaCode = Fira_Code({
  variable: "--font-fira-code",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "Nexus — Security Incident Dashboard",
  description: "AI-powered security incident narrative generator for SOC analysts",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${spaceGrotesk.variable} ${firaCode.variable} dark h-full antialiased`}
    >
      <head>
        <link rel="preconnect" href={API_BASE} />
      </head>
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <a href="#main-content" className="skip-link">
          Skip to main content
        </a>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
