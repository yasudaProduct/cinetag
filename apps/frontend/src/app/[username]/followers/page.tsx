import type { Metadata } from "next";
import { getUserByDisplayId } from "@/lib/api/users/getUser";
import { FollowersList } from "../_components/tabs/FollowersList";
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
    const title = `${user.display_name}のフォロワー | cinetag`;

    return {
      title,
      openGraph: { title },
    };
  } catch {
    return {
      title: "フォロワー | cinetag",
    };
  }
}

export default async function FollowersPage({
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
          <h1 className="text-xl font-bold text-gray-900">フォロワー</h1>
        </div>
        <FollowersList username={username} />
      </main>
    </div>
  );
}
