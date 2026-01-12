"use client";

import { Header } from "@/components/Header";
import { AvatarCircle } from "@/components/AvatarCircle";
import { CategoryCard } from "@/components/CategoryCard";
import { Search, Plus, UserPlus, UserMinus } from "lucide-react";
import { useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getUserByDisplayId, type UserProfile } from "@/lib/api/users/getUser";
import { getMe } from "@/lib/api/users/getMe";
import { listUserTags } from "@/lib/api/users/listUserTags";
import { getFollowStats } from "@/lib/api/users/getFollowStats";
import { followUser } from "@/lib/api/users/followUser";
import { unfollowUser } from "@/lib/api/users/unfollowUser";
import { listFollowing } from "@/lib/api/users/listFollowing";
import { listFollowers } from "@/lib/api/users/listFollowers";
import { listFollowingTags } from "@/lib/api/me/listFollowingTags";
import { notFound } from "next/navigation";
import Link from "next/link";

type TabType =
  | "created"
  | "registered"
  | "favorite"
  | "following"
  | "followers"
  | "followingTags";

export default function UserPageClient(props: {
  username: string;
  initialProfileUser: UserProfile;
}) {
  const username = props.username;
  const { isSignedIn, isLoaded, getToken } = useAuth();
  const [activeTab, setActiveTab] = useState<TabType>("created");
  const [searchQuery, setSearchQuery] = useState("");

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

  // ログインユーザー自身の情報を取得（自分のページかどうかの判定用）
  const { data: currentUser } = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return getMe(token);
    },
    enabled: isLoaded && isSignedIn,
  });

  // 自分のページかどうかを判定
  const isOwnPage =
    currentUser &&
    profileUser &&
    currentUser.display_id === profileUser.display_id;

  // ユーザーのタグ一覧を取得
  const { data: userTagsData, isLoading: isTagsLoading } = useQuery({
    queryKey: ["userTags", username],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" }).catch(
        () => null
      );
      return listUserTags({ displayId: username, token: token ?? undefined });
    },
    enabled: !!username && !!profileUser,
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

  // フォロー中ユーザー一覧を取得
  const { data: followingData, isLoading: isFollowingLoading } = useQuery({
    queryKey: ["following", username],
    queryFn: () => listFollowing(username),
    enabled: !!username && !!profileUser && activeTab === "following",
  });

  // フォロワー一覧を取得
  const { data: followersData, isLoading: isFollowersLoading } = useQuery({
    queryKey: ["followers", username],
    queryFn: () => listFollowers(username),
    enabled: !!username && !!profileUser && activeTab === "followers",
  });

  // フォロー中のタグ一覧を取得（自分のページのみ）
  const { data: followingTagsData, isLoading: isFollowingTagsLoading } =
    useQuery({
      queryKey: ["followingTags", username],
      queryFn: async () => {
        const token = await getToken({ template: "cinetag-backend" });
        if (!token) throw new Error("認証情報の取得に失敗しました");
        return listFollowingTags(token);
      },
      enabled:
        !!username &&
        !!profileUser &&
        isOwnPage &&
        activeTab === "followingTags",
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
      <div className="min-h-screen bg-[#FFF5F5]">
        <Header />
        <div className="flex items-center justify-center h-96">
          <p className="text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (isError || !profileUser) {
    notFound();
  }

  const displayName = profileUser.display_name;
  const userTags = userTagsData?.items ?? [];
  const createdCount = userTagsData?.totalCount ?? 0;
  const followingCount = followStats?.following_count ?? 0;
  const followersCount = followStats?.followers_count ?? 0;
  const isFollowing = followStats?.is_following ?? false;

  return (
    <div className="min-h-screen bg-[#FFF5F5]">
      <Header />

      <main className="container mx-auto px-4 md:px-6 py-12">
        <div className="flex flex-col lg:flex-row gap-8">
          {/* Left Sidebar - Profile */}
          <aside className="lg:w-80 flex-shrink-0">
            <div className="bg-white rounded-2xl border border-gray-200 p-8 shadow-sm sticky top-24">
              {/* Avatar */}
              <div className="flex justify-center mb-6">
                <AvatarCircle
                  name={displayName}
                  avatarUrl={profileUser.avatar_url ?? undefined}
                  className="w-32 h-32 text-4xl"
                  sizes="128px"
                />
              </div>

              {/* User Info */}
              <div className="text-center mb-6">
                <h1 className="text-2xl font-bold text-gray-900 mb-2">
                  {displayName}
                </h1>
                <p className="text-sm text-gray-500 mb-4">
                  {profileUser.bio ||
                    "映画のムードに合わせた最高のプレイリストを厳選"}
                </p>
              </div>

              {/* Follow Button - 他人のページのみ表示 */}
              {!isOwnPage && isLoaded && isSignedIn && (
                <button
                  onClick={handleFollowToggle}
                  disabled={
                    followMutation.isPending || unfollowMutation.isPending
                  }
                  className={`w-full mb-6 py-3 px-6 rounded-full flex items-center justify-center gap-2 font-medium transition-all ${
                    isFollowing
                      ? "bg-gray-200 text-gray-700 hover:bg-gray-300"
                      : "bg-pink-500 text-white hover:bg-pink-600"
                  } disabled:opacity-50 disabled:cursor-not-allowed`}
                >
                  {isFollowing ? (
                    <>
                      <UserMinus className="w-5 h-5" />
                      <span>フォロー中</span>
                    </>
                  ) : (
                    <>
                      <UserPlus className="w-5 h-5" />
                      <span>フォローする</span>
                    </>
                  )}
                </button>
              )}

              {/* Stats */}
              <div className="flex justify-center gap-8 mb-6">
                <div className="text-center">
                  <div className="text-3xl font-bold text-pink-500 mb-1">
                    {createdCount}
                  </div>
                  <div className="text-sm text-gray-600">作成カテゴリ</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-blue-500 mb-1">
                    {followingCount}
                  </div>
                  <div className="text-sm text-gray-600">フォロー中</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-purple-500 mb-1">
                    {followersCount}
                  </div>
                  <div className="text-sm text-gray-600">フォロワー</div>
                </div>
              </div>

              {/* Navigation - 自分のページのみ表示 */}
              {isOwnPage && (
                <>
                  <nav className="space-y-2">
                    <button className="w-full flex items-center gap-3 px-4 py-3 rounded-lg bg-pink-50 text-pink-600 font-medium">
                      <svg
                        className="w-5 h-5"
                        fill="currentColor"
                        viewBox="0 0 20 20"
                      >
                        <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z" />
                      </svg>
                      マイカテゴリ
                    </button>
                    <button className="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-gray-600 hover:bg-gray-50 font-medium">
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
                        />
                      </svg>
                      登録した映画
                    </button>
                    <button className="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-gray-600 hover:bg-gray-50 font-medium">
                      <svg
                        className="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
                        />
                      </svg>
                      お気に入り
                    </button>
                  </nav>

                  {/* New Category Button */}
                  <button className="w-full mt-6 bg-[#FFD75E] hover:bg-[#ffcf40] text-gray-900 font-bold py-3 px-6 rounded-full flex items-center justify-center gap-2 shadow-sm hover:shadow transition-all">
                    <Plus className="w-5 h-5" />
                    <span>新しいカテゴリを作成</span>
                  </button>
                </>
              )}
            </div>
          </aside>

          {/* Right Content - Tags */}
          <div className="flex-1">
            {/* Header */}
            <div className="mb-8">
              <h2 className="text-3xl font-bold text-gray-900 mb-2">
                {isOwnPage ? "マイカテゴリ" : `${displayName}のカテゴリ`}
              </h2>
              <p className="text-gray-600">
                {isOwnPage
                  ? "あなたが作成した映画カテゴリの一覧です"
                  : `${displayName}が作成した映画カテゴリの一覧です`}
              </p>
            </div>

            {/* Tabs */}
            <div className="flex items-center gap-2 mb-8 border-b border-gray-200 overflow-x-auto">
              <button
                onClick={() => setActiveTab("created")}
                className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
                  activeTab === "created"
                    ? "text-pink-600"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                作成したカテゴリ
                {activeTab === "created" && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
                )}
              </button>
              <button
                onClick={() => setActiveTab("registered")}
                className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
                  activeTab === "registered"
                    ? "text-pink-600"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                登録した映画
                {activeTab === "registered" && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
                )}
              </button>
              <button
                onClick={() => setActiveTab("favorite")}
                className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
                  activeTab === "favorite"
                    ? "text-pink-600"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                お気に入りカテゴリ
                {activeTab === "favorite" && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
                )}
              </button>
              <button
                onClick={() => setActiveTab("following")}
                className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
                  activeTab === "following"
                    ? "text-pink-600"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                フォロー中
                {activeTab === "following" && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
                )}
              </button>
              <button
                onClick={() => setActiveTab("followers")}
                className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
                  activeTab === "followers"
                    ? "text-pink-600"
                    : "text-gray-600 hover:text-gray-900"
                }`}
              >
                フォロワー
                {activeTab === "followers" && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
                )}
              </button>
              {isOwnPage && (
                <button
                  onClick={() => setActiveTab("followingTags")}
                  className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
                    activeTab === "followingTags"
                      ? "text-pink-600"
                      : "text-gray-600 hover:text-gray-900"
                  }`}
                >
                  フォロー中のタグ
                  {activeTab === "followingTags" && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
                  )}
                </button>
              )}
            </div>

            {/* Search Bar - カテゴリタブのみ表示 */}
            {(activeTab === "created" ||
              activeTab === "registered" ||
              activeTab === "favorite" ||
              activeTab === "followingTags") && (
              <div className="relative mb-8">
                <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                  <Search className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="text"
                  placeholder="カテゴリや映画を検索..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="block w-full pl-12 pr-4 py-3.5 rounded-full border border-gray-300 bg-white text-gray-900 focus:ring-2 focus:ring-pink-500 focus:border-transparent shadow-sm"
                />
              </div>
            )}

            {/* Cards Grid - カテゴリタブ */}
            {activeTab === "created" && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {isTagsLoading ? (
                  <p className="text-gray-600 col-span-full text-center py-8">
                    読み込み中...
                  </p>
                ) : userTags.length === 0 ? (
                  <p className="text-gray-600 col-span-full text-center py-8">
                    まだカテゴリがありません
                  </p>
                ) : (
                  userTags.map((tag) => (
                    <CategoryCard
                      key={tag.id}
                      title={tag.title}
                      description={tag.description ?? ""}
                      author={tag.author}
                      authorDisplayId={tag.authorDisplayId}
                      movieCount={tag.movieCount}
                      likes={tag.followerCount}
                      images={tag.images}
                      href={`/tags/${tag.id}`}
                    />
                  ))
                )}
              </div>
            )}

            {/* フォロー中ユーザー一覧 */}
            {activeTab === "following" && (
              <div className="space-y-4">
                {isFollowingLoading ? (
                  <p className="text-gray-600 text-center py-8">
                    読み込み中...
                  </p>
                ) : (followingData?.items ?? []).length === 0 ? (
                  <p className="text-gray-600 text-center py-8">
                    フォローしているユーザーはいません
                  </p>
                ) : (
                  (followingData?.items ?? []).map((user) => (
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
                  ))
                )}
              </div>
            )}

            {/* フォロワー一覧 */}
            {activeTab === "followers" && (
              <div className="space-y-4">
                {isFollowersLoading ? (
                  <p className="text-gray-600 text-center py-8">
                    読み込み中...
                  </p>
                ) : (followersData?.items ?? []).length === 0 ? (
                  <p className="text-gray-600 text-center py-8">
                    フォロワーはいません
                  </p>
                ) : (
                  (followersData?.items ?? []).map((user) => (
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
                  ))
                )}
              </div>
            )}

            {/* フォロー中のタグ一覧（自分のページのみ） */}
            {activeTab === "followingTags" && isOwnPage && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {isFollowingTagsLoading ? (
                  <p className="text-gray-600 col-span-full text-center py-8">
                    読み込み中...
                  </p>
                ) : (followingTagsData?.items ?? []).length === 0 ? (
                  <p className="text-gray-600 col-span-full text-center py-8">
                    フォローしているタグはありません
                  </p>
                ) : (
                  (followingTagsData?.items ?? []).map((tag) => (
                    <CategoryCard
                      key={tag.id}
                      title={tag.title}
                      description={tag.description ?? ""}
                      author={tag.author}
                      authorDisplayId={tag.authorDisplayId}
                      movieCount={tag.movieCount}
                      likes={tag.followerCount}
                      images={tag.images}
                      href={`/tags/${tag.id}`}
                    />
                  ))
                )}
              </div>
            )}

            {/* 登録した映画・お気に入り（未実装） */}
            {(activeTab === "registered" || activeTab === "favorite") && (
              <p className="text-gray-600 text-center py-8">
                この機能は準備中です
              </p>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
