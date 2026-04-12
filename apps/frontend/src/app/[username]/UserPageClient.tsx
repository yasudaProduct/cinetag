"use client";

import { useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getUserByDisplayId, type UserProfile } from "@/lib/api/users/getUser";
import { getFollowStats } from "@/lib/api/users/getFollowStats";
import { listUserTags } from "@/lib/api/users/listUserTags";
import { followUser } from "@/lib/api/users/followUser";
import { unfollowUser } from "@/lib/api/users/unfollowUser";
import { notFound } from "next/navigation";
import { Spinner } from "@/components/ui/spinner";
import {
  UserProfile as UserProfileComponent,
  UserPageTabs,
  CreatedTagsList,
  FollowingTagsList,
  LikedTagsList,
  type TabType,
} from "./_components";

export default function UserPageClient(props: {
  username: string;
  initialProfileUser: UserProfile;
  isOwnPage: boolean;
}) {
  const username = props.username;
  const isOwnPage = props.isOwnPage;
  const { isSignedIn, isLoaded, getToken } = useAuth();
  const [activeTab, setActiveTab] = useState<TabType>("created");

  // プロフィールユーザー（ページの対象ユーザー）を取得
  const {
    data: profileUser,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["user", username],
    queryFn: () => getUserByDisplayId(username),
    enabled: !!username,
    initialData: props.initialProfileUser,
  });

  // ユーザーが作成したタグ一覧を取得（totalCount をプロフィールに表示）
  const { data: userTagsData } = useQuery({
    queryKey: ["userTags", username],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" }).catch(
        () => null
      );
      return listUserTags({ displayId: username, token: token ?? undefined });
    },
    enabled: !!username,
  });

  // フォロー統計を取得
  const { data: followStats } = useQuery({
    queryKey: ["followStats", username],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" }).catch(
        () => null
      );
      return getFollowStats(username, token ?? undefined);
    },
    enabled: !!username && !!profileUser,
  });

  const queryClient = useQueryClient();

  // フォローミューテーション
  const followMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return followUser(username, token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["followStats", username] });
    },
  });

  // アンフォローミューテーション
  const unfollowMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return unfollowUser(username, token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["followStats", username] });
    },
  });

  const handleFollowToggle = () => {
    if (followStats?.is_following) {
      unfollowMutation.mutate();
    } else {
      followMutation.mutate();
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen">
        <div className="flex items-center justify-center h-96">
          <Spinner size="lg" className="text-gray-600" />
        </div>
      </div>
    );
  }

  if (isError || !profileUser) {
    notFound();
  }

  const createdCount = userTagsData?.totalCount ?? 0;
  const followingCount = followStats?.following_count ?? 0;
  const followersCount = followStats?.followers_count ?? 0;
  const isFollowing = followStats?.is_following ?? false;

  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-8 md:py-12">
        <div className="flex flex-col gap-6 md:gap-8">
          {/* User Profile */}
          <UserProfileComponent
            profileUser={profileUser}
            createdCount={createdCount}
            followingCount={followingCount}
            followersCount={followersCount}
            showFollowButton={!isOwnPage && isLoaded && (isSignedIn ?? false)}
            isFollowing={isFollowing}
            isFollowPending={
              followMutation.isPending || unfollowMutation.isPending
            }
            onFollowToggle={handleFollowToggle}
          />

          {/* Tabs & Content */}
          <div className="w-full">
            <UserPageTabs
              activeTab={activeTab}
              onTabChange={setActiveTab}
            />

            {/* Tab Content */}
            {activeTab === "created" && <CreatedTagsList username={username} />}
            {activeTab === "registered" && (
              <FollowingTagsList
                username={username}
                isOwnPage={isOwnPage}
              />
            )}
            {activeTab === "favorite" && (
              <LikedTagsList isOwnPage={isOwnPage} />
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
