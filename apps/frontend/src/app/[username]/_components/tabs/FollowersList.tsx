"use client";

import { useQuery } from "@tanstack/react-query";
import { listFollowers } from "@/lib/api/users/listFollowers";
import { AvatarCircle } from "@/components/AvatarCircle";
import { Spinner } from "@/components/ui/spinner";
import Link from "next/link";

type FollowersListProps = {
  username: string;
};

export function FollowersList({ username }: FollowersListProps) {
  const { data: followersData, isLoading } = useQuery({
    queryKey: ["followers", username],
    queryFn: () => listFollowers(username),
    enabled: !!username,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <Spinner size="md" className="text-gray-600" />
      </div>
    );
  }

  if ((followersData?.items ?? []).length === 0) {
    return (
      <p className="text-gray-600 text-center py-8">フォロワーはいません</p>
    );
  }

  return (
    <div className="space-y-4 w-[40%] mx-auto">
      {(followersData?.items ?? []).map((user) => (
        <Link
          key={user.id}
          href={`/${user.display_id}`}
          className="flex items-center gap-4 p-4 bg-white rounded-2xl border border-gray-200 hover:border-pink-300 hover:shadow-sm transition-all"
        >
          <AvatarCircle
            name={user.display_name}
            avatarUrl={user.avatar_url ?? undefined}
            className="w-12 h-12"
          />
          <div className="flex-1 min-w-0">
            <div className="font-bold text-gray-900 truncate">
              {user.display_name}
            </div>
            <div className="text-sm text-gray-500 truncate">
              @{user.display_id}
            </div>
            {user.bio && (
              <div className="text-sm text-gray-600 mt-1 line-clamp-1">
                {user.bio}
              </div>
            )}
          </div>
        </Link>
      ))}
    </div>
  );
}
