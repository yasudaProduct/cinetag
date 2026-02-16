"use client";

import { FollowingTagsList } from "@/app/(app)/[username]/_components/tabs/FollowingTagsList";

export function FollowingTagsPageClient() {
  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-12">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            フォローしているタグ
          </h1>
          <p className="text-gray-600">
            あなたがフォローしているタグの一覧です
          </p>
        </div>

        <FollowingTagsList />
      </main>
    </div>
  );
}
