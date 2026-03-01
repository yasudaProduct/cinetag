import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "新規登録 | cinetag",
  description:
    "cinetagに無料で登録して、映画タグの作成・共有を始めましょう。",
};

export default function SignUpLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
