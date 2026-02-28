import type { Metadata } from "next";
import { notFound } from "next/navigation";
import UserPageClient from "./UserPageClient";
import { getUserByDisplayId } from "@/lib/api/users/getUser";

// ISR: 10分ごとに再生成
export const revalidate = 600;

export async function generateMetadata({
  params,
}: {
  params: Promise<{ username: string }>;
}): Promise<Metadata> {
  const { username } = await params;

  try {
    const user = await getUserByDisplayId(username);
    const title = `${user.display_name} | cinetag`;
    const description =
      user.bio || `${user.display_name}のプロフィール - cinetag`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        type: "profile",
        ...(user.avatar_url && { images: [user.avatar_url] }),
      },
      twitter: {
        card: "summary",
        title,
        description,
      },
    };
  } catch {
    return {
      title: "ユーザー | cinetag",
    };
  }
}

export default async function UserPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;

  let user;
  try {
    user = await getUserByDisplayId(username);
  } catch (error) {
    if (error instanceof Error && error.message.includes("404")) {
      notFound();
    }
    throw error;
  }

  return <UserPageClient username={username} initialProfileUser={user} />;
}
