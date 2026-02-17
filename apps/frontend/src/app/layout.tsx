import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { ClerkProvider } from "@clerk/nextjs";
import { Providers } from "../components/providers/query-client-provider";
import { AppShell } from "../components/AppShell";
import { jaJP } from "@clerk/localizations";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "cinetag",
  description: "cinetag",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <ClerkProvider
      localization={jaJP}
      appearance={{
        variables: {
          colorPrimary: "#3b82f6",
          borderRadius: "0.75rem",
        },
        elements: {
          card: "shadow-lg border border-gray-200",
          headerTitle: "text-gray-900",
          headerSubtitle: "text-gray-600",
          formButtonPrimary:
            "bg-blue-500 hover:bg-blue-600 text-white rounded-lg",
          formFieldInput:
            "rounded-lg border border-gray-300 focus:ring-2 focus:ring-blue-500",
          footerActionLink: "text-blue-600 hover:text-blue-700",
        },
      }}
    >
      <html lang="ja" suppressHydrationWarning>
        <body
          suppressHydrationWarning
          className={`${geistSans.variable} ${geistMono.variable} antialiased bg-[#FFF9F3] text-gray-900`}
        >
          <Providers>
            <AppShell>{children}</AppShell>
          </Providers>
        </body>
      </html>
    </ClerkProvider>
  );
}
