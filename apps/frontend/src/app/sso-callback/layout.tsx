import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "認証処理中 | cinetag",
  robots: { index: false, follow: false },
};

export default function SSOCallbackLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
