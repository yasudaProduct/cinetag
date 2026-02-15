"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/nextjs";
import { listFollowingTags } from "@/lib/api/me/listFollowingTags";
import { TagsGrid } from "@/components/TagsGrid";

type FollowingTagsListProps = {
  username?: string; // オプショナル：互換性のために残す
  isOwnPage?: boolean; // オプショナル：互換性のために残す
};

export function FollowingTagsList({
  isOwnPage,
}: FollowingTagsListProps) {
  const { getToken, isSignedIn, isLoaded } = useAuth();

  const { data: followingTagsData, isLoading } = useQuery({
    queryKey: ["followingTags", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return listFollowingTags(token);
    },
    enabled: isLoaded && isSignedIn && (isOwnPage === undefined || isOwnPage),
  });

  return (
    <TagsGrid
      tags={followingTagsData?.items ?? []}
      isLoading={isLoading}
      emptyMessage="フォローしているタグはありません"
      columns="3"
    />
  );
}
