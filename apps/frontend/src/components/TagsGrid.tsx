"use client";

import { CategoryCard } from "@/components/CategoryCard";

type Tag = {
  id: string;
  title: string;
  description?: string | null;
  author: string;
  authorDisplayId?: string;
  movieCount: number;
  followerCount: number;
  images: string[];
};

type TagsGridProps = {
  tags: Tag[];
  isLoading?: boolean;
  emptyMessage?: string;
  columns?: "3" | "4";
};

export function TagsGrid({
  tags,
  isLoading = false,
  emptyMessage = "タグがありません",
  columns = "3",
}: TagsGridProps) {
  if (isLoading) {
    return (
      <p className="text-gray-600 col-span-full text-center py-8">
        読み込み中...
      </p>
    );
  }

  if (tags.length === 0) {
    return (
      <p className="text-gray-600 col-span-full text-center py-8">
        {emptyMessage}
      </p>
    );
  }

  const gridCols = columns === "4" ? "lg:grid-cols-4" : "lg:grid-cols-3";

  return (
    <div className={`grid grid-cols-1 sm:grid-cols-2 ${gridCols} gap-6`}>
      {tags.map((tag) => (
        <CategoryCard
          key={tag.id}
          title={tag.title}
          description={tag.description ?? ""}
          author={tag.author}
          authorDisplayId={tag.authorDisplayId ?? ""}
          movieCount={tag.movieCount}
          likes={tag.followerCount}
          images={tag.images}
          href={`/tags/${tag.id}`}
        />
      ))}
    </div>
  );
}
