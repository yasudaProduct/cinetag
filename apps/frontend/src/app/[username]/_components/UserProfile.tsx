"use client";

import Link from "next/link";
import { AvatarCircle } from "@/components/AvatarCircle";
import { UserPlus, UserMinus } from "lucide-react";
import type { UserProfile as UserProfileType } from "@/lib/api/users/getUser";

type UserProfileProps = {
  profileUser: UserProfileType;
  createdCount: number;
  followingCount: number;
  followersCount: number;
  showFollowButton: boolean;
  isFollowing: boolean;
  isFollowPending: boolean;
  onFollowToggle: () => void;
};

export function UserProfile({
  profileUser,
  createdCount,
  followingCount,
  followersCount,
  showFollowButton,
  isFollowing,
  isFollowPending,
  onFollowToggle,
}: UserProfileProps) {
  const displayName = profileUser.display_name;

  return (
    <div className="w-full">
      <div className="bg-white rounded-2xl border border-gray-200 p-4 md:p-8 shadow-sm">
        <div className="flex flex-col md:flex-row items-center md:items-start gap-4 md:gap-8">
          {/* Avatar */}
          <div className="flex-shrink-0">
            <AvatarCircle
              name={displayName}
              avatarUrl={profileUser.avatar_url ?? undefined}
              className="w-24 h-24 md:w-32 md:h-32 text-2xl md:text-4xl"
              sizes="128px"
            />
          </div>

          {/* User Info & Stats */}
          <div className="flex-1 text-center md:text-left">
            <div className="mb-4">
              <h1 className="text-xl md:text-2xl font-bold text-gray-900 mb-2">
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
              <Link
                href={`/${profileUser.display_id}/following`}
                className="text-center md:text-left hover:opacity-70 transition-opacity"
              >
                <div className="text-2xl font-bold text-blue-500 mb-1">
                  {followingCount}
                </div>
                <div className="text-xs text-gray-600">フォロー中</div>
              </Link>
              <Link
                href={`/${profileUser.display_id}/followers`}
                className="text-center md:text-left hover:opacity-70 transition-opacity"
              >
                <div className="text-2xl font-bold text-purple-500 mb-1">
                  {followersCount}
                </div>
                <div className="text-xs text-gray-600">フォロワー</div>
              </Link>
            </div>

            {/* Actions */}
            <div className="flex justify-center md:justify-start gap-4">
              {showFollowButton && (
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
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
