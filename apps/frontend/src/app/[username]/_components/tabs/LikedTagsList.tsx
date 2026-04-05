"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/nextjs";
import { listLikedTags } from "@/lib/api/me/listLikedTags";
import { TagsGrid } from "@/components/TagsGrid";

type LikedTagsListProps = {
  isOwnPage: boolean;
};

export function LikedTagsList({ isOwnPage }: LikedTagsListProps) {
  const { getToken, isSignedIn, isLoaded } = useAuth();

  const { data: likedTagsData, isLoading } = useQuery({
    queryKey: ["likedTags", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return listLikedTags(token);
    },
    enabled: isLoaded && isSignedIn && isOwnPage,
  });

  if (!isOwnPage) {
    return (
      <p className="text-gray-600 text-center py-8">
        いいねしたタグは本人のみ表示されます
      </p>
    );
  }

  return (
    <TagsGrid
      tags={likedTagsData?.items ?? []}
      isLoading={isLoading}
      emptyMessage="いいねしたタグはまだありません"
      columns="3"
    />
  );
}
