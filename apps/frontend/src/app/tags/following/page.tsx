import { auth } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";
import { FollowingTagsPageClient } from "./FollowingTagsPageClient";

export default async function FollowingTagsPage() {
  const { userId } = await auth();

  // 未ログインの場合はサインインページへ
  if (!userId) {
    redirect("/sign-in");
  }

  return <FollowingTagsPageClient />;
}
