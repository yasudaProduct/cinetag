"use client";

import { CategoryListItem } from "@/components/CategoryListItem";
import { Spinner } from "@/components/ui/spinner";

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

type TagsListProps = {
  tags: Tag[];
  isLoading?: boolean;
  emptyMessage?: string;
};

export function TagsList({
  tags,
  isLoading = false,
  emptyMessage = "タグがありません",
}: TagsListProps) {
  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <Spinner size="md" className="text-gray-600" />
      </div>
    );
  }

  if (tags.length === 0) {
    return (
      <p className="text-gray-600 text-center py-8">
        {emptyMessage}
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {tags.map((tag) => (
        <CategoryListItem
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
