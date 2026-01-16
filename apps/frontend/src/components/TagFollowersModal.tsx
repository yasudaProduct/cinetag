"use client";

import { useQuery } from "@tanstack/react-query";
import { Modal } from "@/components/Modal";
import { AvatarCircle } from "@/components/AvatarCircle";
import { listTagFollowers } from "@/lib/api/tags/listFollowers";
import { X, Users } from "lucide-react";

type TagFollowersModalProps = {
  open: boolean;
  tagId: string;
  tagTitle: string;
  onClose: () => void;
};

export function TagFollowersModal({
  open,
  tagId,
  tagTitle,
  onClose,
}: TagFollowersModalProps) {
  const followersQuery = useQuery({
    queryKey: ["tagFollowers", tagId],
    queryFn: () => listTagFollowers(tagId, 1, 50),
    enabled: open,
  });

  const followers = followersQuery.data?.items ?? [];
  const totalCount = followersQuery.data?.totalCount ?? 0;

  return (
    <Modal open={open} onClose={onClose}>
      <div className="bg-white rounded-3xl shadow-xl w-full max-w-md mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-gray-600" />
            <h2 className="text-lg font-bold text-gray-900">フォロワー</h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-2 rounded-full hover:bg-gray-100 transition-colors"
            aria-label="閉じる"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Tag title */}
        <div className="px-6 py-3 bg-gray-50 border-b border-gray-100">
          <p className="text-sm text-gray-600 truncate">
            タグ: <span className="font-semibold text-gray-900">{tagTitle}</span>
          </p>
          <p className="text-xs text-gray-500 mt-1">{totalCount}人のフォロワー</p>
        </div>

        {/* Content */}
        <div className="px-6 py-4 max-h-80 overflow-y-auto">
          {followersQuery.isLoading ? (
            <div className="text-center py-8 text-gray-500">読み込み中...</div>
          ) : followersQuery.isError ? (
            <div className="text-center py-8 text-red-500">
              読み込みに失敗しました
            </div>
          ) : followers.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              まだフォロワーがいません
            </div>
          ) : (
            <ul className="space-y-3">
              {followers.map((follower) => (
                <li key={follower.id}>
                  <a
                    href={`/${follower.displayId}`}
                    className="flex items-center gap-3 p-2 rounded-xl hover:bg-gray-50 transition-colors"
                  >
                    <AvatarCircle
                      name={follower.displayName}
                      avatarUrl={follower.avatarUrl}
                      className="h-10 w-10"
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-semibold text-gray-900 truncate">
                        {follower.displayName}
                      </p>
                      <p className="text-xs text-gray-500 truncate">
                        @{follower.displayId}
                      </p>
                    </div>
                  </a>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-gray-100">
          <button
            type="button"
            onClick={onClose}
            className="w-full py-2.5 bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold rounded-full transition-colors"
          >
            閉じる
          </button>
        </div>
      </div>
    </Modal>
  );
}
