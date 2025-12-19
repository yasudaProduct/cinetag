"use client";

import { Header } from "@/components/Header";
import { CategoryCard } from "@/components/CategoryCard";
import { Search, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { useState } from "react";
import { TagCreateModal, CreatedTagForList } from "@/components/TagCreateModal";
import { SignedIn, SignedOut, SignInButton } from "@clerk/nextjs";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listTags } from "@/lib/api/tags/list";
import type { TagListItem } from "@/lib/validation/tag.api";

export default function Home() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const queryClient = useQueryClient();
  const tagsQuery = useQuery({
    queryKey: ["tags"],
    queryFn: listTags,
  });
  const tags = tagsQuery.data ?? [];

  return (
    <div className="min-h-screen bg-[#FFF5F5]">
      <Header />

      <main className="container mx-auto px-4 md:px-6 py-12">
        {/* Hero Section */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-10">
          <div>
            <h1 className="text-3xl md:text-4xl font-bold text-gray-900 mb-2">
              タグを探そう！
            </h1>
            <p className="text-gray-600">
              お気に入りの映画リストを見つけたり、自分だけのタグを作ってみよう。
            </p>
          </div>
          <SignedIn>
            <button
              type="button"
              className="bg-[#FFD75E] hover:bg-[#ffcf40] text-gray-900 font-bold py-3 px-6 rounded-full flex items-center gap-2 shadow-sm hover:shadow transition-all"
              onClick={() => setIsCreateModalOpen(true)}
            >
              <Plus className="w-5 h-5" />
              <span>新しいタグを作成</span>
            </button>
          </SignedIn>
          <SignedOut>
            <SignInButton mode="modal">
              <button
                type="button"
                className="bg-[#FFD75E] hover:bg-[#ffcf40] text-gray-900 font-bold py-3 px-6 rounded-full flex items-center gap-2 shadow-sm hover:shadow transition-all"
              >
                <Plus className="w-5 h-5" />
                <span>新しいタグを作成</span>
              </button>
            </SignInButton>
          </SignedOut>
        </div>

        {/* Search & Filter */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-10">
          {/* Search Bar */}
          <div className="relative w-full md:max-w-2xl">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Search className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="text"
              placeholder="「90年代SF」などで検索..."
              className="block w-full pl-12 pr-4 py-3.5 rounded-full border border-gray-900 bg-white text-gray-900 focus:ring-2 focus:ring-blue-500 focus:border-transparent shadow-sm"
            />
          </div>

          {/* Filter Tabs */}
          <div className="flex items-center bg-white rounded-full p-1 border border-gray-900">
            <button className="px-6 py-2 rounded-full bg-[#FFD75E] text-gray-900 font-bold text-sm">
              人気
            </button>
            <button className="px-6 py-2 rounded-full text-gray-600 hover:bg-gray-100 font-medium text-sm">
              新着
            </button>
            <button className="px-6 py-2 rounded-full text-gray-600 hover:bg-gray-100 font-medium text-sm">
              映画数
            </button>
          </div>
        </div>

        {/* Category Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
          {/* {MOCK_CATEGORIES.map((category) => ( */}
          {tagsQuery.isLoading ? (
            <div>読み込み中...</div>
          ) : tagsQuery.isError ? (
            <div>タグ一覧の取得に失敗しました</div>
          ) : tags.length > 0 ? (
            tags.map((tag) => (
              <CategoryCard
                key={tag.id}
                title={tag.title}
                description={tag.description ?? ""}
                author={tag.author}
                movieCount={tag.movie_count}
                likes={tag.follower_count}
                images={tag.images || []}
                href={`/tags/${tag.id}`}
              />
            ))
          ) : (
            <div>タグがありません</div>
          )}
          {/* ))} */}
        </div>

        {/* Pagination */}
        <div className="flex justify-center items-center gap-2">
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-300 bg-white hover:bg-gray-50 text-gray-600 disabled:opacity-50">
            <ChevronLeft className="w-5 h-5" />
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full bg-blue-500 text-white font-bold">
            1
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            2
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            3
          </button>
          <span className="text-gray-500 px-1">...</span>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            12
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            <ChevronRight className="w-5 h-5" />
          </button>
        </div>
        {/* Create Tag Modal */}
        <TagCreateModal
          open={isCreateModalOpen}
          onClose={() => setIsCreateModalOpen(false)}
          onCreated={(created: CreatedTagForList) => {
            queryClient.setQueryData<TagListItem[]>(["tags"], (prev) => [
              created as unknown as TagListItem,
              ...(prev ?? []),
            ]);
          }}
        />
      </main>
    </div>
  );
}
