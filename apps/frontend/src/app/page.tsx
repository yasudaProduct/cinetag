"use client";

import { TagsGrid } from "@/components/TagsGrid";
import { Search, ChevronLeft, ChevronRight } from "lucide-react";
import { useState, useEffect } from "react";
import { TagModal, CreatedTagForList } from "@/components/TagModal";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listTags } from "@/lib/api/tags/list";
import type { TagListItem } from "@/lib/validation/tag.api";

const PAGE_SIZE = 10;

type SortOption = "popular" | "recent" | "movie_count";

export default function Home() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [sort, setSort] = useState<SortOption>("popular");
  const queryClient = useQueryClient();

  // デバウンス処理: 500ms後に検索クエリを更新
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1); // 検索時は1ページ目にリセット
    }, 500);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // ソート変更時にページを1にリセット
  const handleSortChange = (newSort: SortOption) => {
    setSort(newSort);
    setCurrentPage(1);
  };

  const tagsQuery = useQuery({
    queryKey: ["tags", debouncedSearchQuery, sort, currentPage],
    queryFn: () =>
      listTags({
        q: debouncedSearchQuery || undefined,
        sort: sort === "popular" ? undefined : sort,
        page: currentPage,
        pageSize: PAGE_SIZE,
      }),
  });

  const tags = tagsQuery.data?.items ?? [];
  const totalCount = tagsQuery.data?.totalCount ?? 0;
  const totalPages = Math.ceil(totalCount / PAGE_SIZE);

  // ページングボタン用のページ番号リストを生成
  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const maxVisible = 5; // 表示するページ番号の最大数

    if (totalPages <= maxVisible) {
      // 総ページ数が少ない場合は全て表示
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // 現在のページ周辺を表示
      if (currentPage <= 3) {
        // 最初の方
        for (let i = 1; i <= 4; i++) {
          pages.push(i);
        }
        pages.push("...");
        pages.push(totalPages);
      } else if (currentPage >= totalPages - 2) {
        // 最後の方
        pages.push(1);
        pages.push("...");
        for (let i = totalPages - 3; i <= totalPages; i++) {
          pages.push(i);
        }
      } else {
        // 中間
        pages.push(1);
        pages.push("...");
        for (let i = currentPage - 1; i <= currentPage + 1; i++) {
          pages.push(i);
        }
        pages.push("...");
        pages.push(totalPages);
      }
    }
    return pages;
  };

  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages) {
      setCurrentPage(page);
      window.scrollTo({ top: 0, behavior: "smooth" });
    }
  };

  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-12">
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
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="block w-full pl-12 pr-4 py-3.5 rounded-full border border-gray-900 bg-white text-gray-900 focus:ring-2 focus:ring-blue-500 focus:border-transparent shadow-sm"
            />
          </div>

          {/* Filter Tabs */}
          <div className="flex items-center bg-white rounded-full p-1 border border-gray-900">
            <button
              onClick={() => handleSortChange("popular")}
              className={`px-6 py-2 rounded-full text-sm font-medium transition-colors ${
                sort === "popular"
                  ? "bg-[#FFD75E] text-gray-900 font-bold"
                  : "text-gray-600 hover:bg-gray-100"
              }`}
            >
              人気
            </button>
            <button
              onClick={() => handleSortChange("recent")}
              className={`px-6 py-2 rounded-full text-sm font-medium transition-colors ${
                sort === "recent"
                  ? "bg-[#FFD75E] text-gray-900 font-bold"
                  : "text-gray-600 hover:bg-gray-100"
              }`}
            >
              新着
            </button>
            <button
              onClick={() => handleSortChange("movie_count")}
              className={`px-6 py-2 rounded-full text-sm font-medium transition-colors ${
                sort === "movie_count"
                  ? "bg-[#FFD75E] text-gray-900 font-bold"
                  : "text-gray-600 hover:bg-gray-100"
              }`}
            >
              映画数
            </button>
          </div>
        </div>

        {/* Category Grid */}
        {tagsQuery.isError ? (
          <div className="text-center py-12 text-red-600">
            タグ一覧の取得に失敗しました
          </div>
        ) : (
          <div className="mb-12">
            <TagsGrid
              tags={tags}
              isLoading={tagsQuery.isLoading}
              emptyMessage="タグがありません"
              columns="4"
            />
          </div>
        )}

        {/* Pagination */}
        {totalPages > 0 && (
          <div className="flex justify-center items-center gap-2">
            <button
              onClick={() => handlePageChange(currentPage - 1)}
              disabled={currentPage === 1}
              className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-300 bg-white hover:bg-gray-50 text-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronLeft className="w-5 h-5" />
            </button>
            {getPageNumbers().map((page, index) => {
              if (page === "...") {
                return (
                  <span
                    key={`ellipsis-${index}`}
                    className="text-gray-500 px-1"
                  >
                    ...
                  </span>
                );
              }
              const pageNum = page as number;
              const isActive = pageNum === currentPage;
              return (
                <button
                  key={pageNum}
                  onClick={() => handlePageChange(pageNum)}
                  className={`w-10 h-10 flex items-center justify-center rounded-full font-medium ${
                    isActive
                      ? "bg-blue-500 text-white font-bold"
                      : "border border-gray-900 bg-white hover:bg-gray-50 text-gray-900"
                  }`}
                >
                  {pageNum}
                </button>
              );
            })}
            <button
              onClick={() => handlePageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>
        )}
        {/* Create Tag Modal */}
        <TagModal
          open={isCreateModalOpen}
          onClose={() => setIsCreateModalOpen(false)}
          onCreated={(created: CreatedTagForList) => {
            // 検索クエリ、ソート、ページに応じてキャッシュを更新
            queryClient.setQueryData(
              ["tags", debouncedSearchQuery, sort, currentPage],
              (
                prev: { items: TagListItem[]; totalCount: number } | undefined
              ) => {
                if (!prev) return prev;
                return {
                  ...prev,
                  items: [created as unknown as TagListItem, ...prev.items],
                  totalCount: prev.totalCount + 1,
                };
              }
            );
          }}
        />
      </main>
    </div>
  );
}
