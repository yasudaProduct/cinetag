"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@clerk/nextjs";
import { listUserTags } from "@/lib/api/users/listUserTags";
import { TagsGrid } from "@/components/TagsGrid";

type CreatedTagsListProps = {
  username: string;
};

export function CreatedTagsList({ username }: CreatedTagsListProps) {
  const { getToken } = useAuth();

  const { data: userTagsData, isLoading } = useQuery({
    queryKey: ["userTags", username],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" }).catch(
        () => null
      );
      return listUserTags({ displayId: username, token: token ?? undefined });
    },
    enabled: !!username,
  });

  return (
    <TagsGrid
      tags={userTagsData?.items ?? []}
      isLoading={isLoading}
      emptyMessage="まだカテゴリがありません"
      columns="3"
    />
  );
}
