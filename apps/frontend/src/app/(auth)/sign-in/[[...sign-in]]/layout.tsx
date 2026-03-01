import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "ログイン | cinetag",
  description: "cinetagにログインして、映画タグの作成・共有を始めましょう。",
};

export default function SignInLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
