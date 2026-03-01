import type { Metadata } from "next";
import { getUserByDisplayId } from "@/lib/api/users/getUser";
import { FollowingUsersList } from "../_components/tabs/FollowingUsersList";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ username: string }>;
}): Promise<Metadata> {
  const { username } = await params;

  try {
    const user = await getUserByDisplayId(username, { cache: "no-store" });
    const title = `${user.display_name}のフォロー | cinetag`;
    const description = `${user.display_name}がフォローしているユーザーの一覧です。`;

    return {
      title,
      description,
      alternates: {
        canonical: `/${username}/following`,
      },
      openGraph: { title, description },
      twitter: { card: "summary" as const, title, description },
    };
  } catch {
    return {
      title: "フォロー | cinetag",
    };
  }
}

export default async function FollowingPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;

  return (
    <div className="min-h-screen bg-gray-50/50">
      <main className="container mx-auto px-4 py-8 max-w-4xl">
        <div className="mb-6 flex items-center gap-4">
          <Link
            href={`/${username}`}
            className="p-2 hover:bg-gray-100 rounded-full transition-colors"
          >
            <ArrowLeft className="w-5 h-5 text-gray-600" />
          </Link>
          <h1 className="text-xl font-bold text-gray-900">フォロー中</h1>
        </div>
        <FollowingUsersList username={username} />
      </main>
    </div>
  );
}
