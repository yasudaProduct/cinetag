"use client";

import { AvatarCircle } from "@/components/AvatarCircle";
import { UserPlus, UserMinus, Settings } from "lucide-react";
import type { UserProfile as UserProfileType } from "@/lib/api/users/getUser";

type UserProfileProps = {
  profileUser: UserProfileType;
  createdCount: number;
  followingCount: number;
  followersCount: number;
  isOwnPage: boolean;
  isFollowing: boolean;
  isSignedIn: boolean;
  isLoaded: boolean;
  isFollowPending: boolean;
  onFollowToggle: () => void;
};

export function UserProfile({
  profileUser,
  createdCount,
  followingCount,
  followersCount,
  isOwnPage,
  isFollowing,
  isSignedIn,
  isLoaded,
  isFollowPending,
  onFollowToggle,
}: UserProfileProps) {
  const displayName = profileUser.display_name;

  return (
    <div className="w-full">
      <div className="bg-white rounded-2xl border border-gray-200 p-8 shadow-sm">
        <div className="flex flex-col md:flex-row items-center md:items-start gap-8">
          {/* Avatar */}
          <div className="flex-shrink-0">
            <AvatarCircle
              name={displayName}
              avatarUrl={profileUser.avatar_url ?? undefined}
              className="w-32 h-32 text-4xl"
              sizes="128px"
            />
          </div>

          {/* User Info & Stats */}
          <div className="flex-1 text-center md:text-left">
            <div className="mb-4">
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                {displayName}
              </h1>
              <p className="text-sm text-gray-500 mb-4 max-w-2xl">
                {profileUser.bio || ""}
              </p>
            </div>

            {/* Stats */}
            <div className="flex justify-center md:justify-start gap-8 mb-6">
              <div className="text-center md:text-left">
                <div className="text-2xl font-bold text-pink-500 mb-1">
                  {createdCount}
                </div>
                <div className="text-xs text-gray-600">作成タグ</div>
              </div>
              <div className="text-center md:text-left">
                <div className="text-2xl font-bold text-blue-500 mb-1">
                  {followingCount}
                </div>
                <div className="text-xs text-gray-600">フォロー中</div>
              </div>
              <div className="text-center md:text-left">
                <div className="text-2xl font-bold text-purple-500 mb-1">
                  {followersCount}
                </div>
                <div className="text-xs text-gray-600">フォロワー</div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex justify-center md:justify-start gap-4">
              {/* Follow Button - 他人のページのみ表示 */}
              {!isOwnPage && isLoaded && isSignedIn && (
                <button
                  onClick={onFollowToggle}
                  disabled={isFollowPending}
                  className={`py-2 px-6 rounded-full flex items-center justify-center gap-2 font-medium transition-all ${
                    isFollowing
                      ? "bg-gray-200 text-gray-700 hover:bg-gray-300"
                      : "bg-pink-500 text-white hover:bg-pink-600"
                  } disabled:opacity-50 disabled:cursor-not-allowed`}
                >
                  {isFollowing ? (
                    <>
                      <UserMinus className="w-4 h-4" />
                      <span>フォロー中</span>
                    </>
                  ) : (
                    <>
                      <UserPlus className="w-4 h-4" />
                      <span>フォローする</span>
                    </>
                  )}
                </button>
              )}

              {/* Navigation - 自分のページのみ表示 */}
              {isOwnPage && (
                <div className="flex gap-2">
                  <button className="flex items-center gap-2 px-4 py-2 rounded-lg bg-pink-50 text-pink-600 font-medium text-sm">
                    <svg
                      className="w-4 h-4"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z" />
                    </svg>
                    マイカテゴリ
                  </button>
                  <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-gray-600 hover:bg-gray-50 font-medium text-sm border border-gray-200">
                    <Settings className="w-4 h-4" />
                    設定
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
