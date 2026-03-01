import type { Metadata } from "next";
import { auth } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";
import { FollowingTagsPageClient } from "./FollowingTagsPageClient";

export const metadata: Metadata = {
  title: "フォロー中のタグ | cinetag",
  description: "あなたがフォローしているタグの一覧です。",
  robots: { index: false, follow: false },
};

export default async function FollowingTagsPage() {
  const { userId } = await auth();

  // 未ログインの場合はサインインページへ
  if (!userId) {
    redirect("/sign-in");
  }

  return <FollowingTagsPageClient />;
}
