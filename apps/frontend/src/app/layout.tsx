import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { ClerkProvider } from "@clerk/nextjs";
import { Providers } from "../components/providers/query-client-provider";
import { jaJP } from "@clerk/localizations";
import { Sidebar } from "../components/Sidebar";

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
      <html lang="ja">
        <body
          className={`${geistSans.variable} ${geistMono.variable} antialiased flex min-h-screen bg-[#FFF5F5] text-gray-900`}
        >
          <Providers>
            <Sidebar />
            <main className="flex-1 md:ml-64 min-h-screen pb-20 md:pb-0">
              {children}
            </main>
          </Providers>
        </body>
      </html>
    </ClerkProvider>
  );
}
