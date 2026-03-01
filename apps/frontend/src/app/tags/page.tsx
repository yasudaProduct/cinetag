import type { Metadata } from "next";
import HomeClient from "./HomeClient";

export const metadata: Metadata = {
  title: "タグ一覧 | cinetag",
  description:
    "みんなが作ったタグを探そう。映画をテーマごとにまとめたタグを閲覧・フォローできます。",
  alternates: {
    canonical: "/tags",
  },
  openGraph: {
    title: "タグ一覧 | cinetag",
    description:
      "みんなが作ったタグを探そう。映画をテーマごとにまとめたタグを閲覧・フォローできます。",
  },
  twitter: {
    card: "summary",
    title: "タグ一覧 | cinetag",
    description:
      "みんなが作ったタグを探そう。映画をテーマごとにまとめたタグを閲覧・フォローできます。",
  },
};

export default function TagsPage() {
  return <HomeClient />;
}
