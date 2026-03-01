import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "マイページ | cinetag",
  robots: { index: false, follow: false },
};

export default function MyPageLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
