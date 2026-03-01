"use client";

import { useAuth } from "@clerk/nextjs";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { listNotifications } from "@/lib/api/notifications/listNotifications";
import { markAsRead } from "@/lib/api/notifications/markAsRead";
import { markAllAsRead } from "@/lib/api/notifications/markAllAsRead";
import type { NotificationItem } from "@/lib/validation/notification.api";
import { cn } from "@/lib/utils";

type NotificationDropdownProps = {
  isOpen: boolean;
  onClose: () => void;
};

function getNotificationMessage(item: NotificationItem): string {
  const actorName = item.actor?.displayName ?? "退会済みユーザー";

  switch (item.notificationType) {
    case "tag_movie_added":
      return `${actorName} がタグ「${item.tag?.title ?? ""}」に「${item.movieTitle ?? "映画"}」を追加しました`;
    case "tag_followed":
      return `${actorName} があなたのタグ「${item.tag?.title ?? ""}」をフォローしました`;
    case "user_followed":
      return `${actorName} があなたをフォローしました`;
    case "following_user_created_tag":
      return `${actorName} が新しいタグ「${item.tag?.title ?? ""}」を作成しました`;
    default:
      return "新しい通知があります";
  }
}

function getNotificationHref(item: NotificationItem): string {
  switch (item.notificationType) {
    case "tag_movie_added":
    case "tag_followed":
    case "following_user_created_tag":
      return item.tag?.id ? `/tags/${item.tag.id}` : "/";
    case "user_followed":
      return item.actor?.displayId ? `/${item.actor.displayId}` : "/";
    default:
      return "/";
  }
}

function formatRelativeTime(dateStr: string): string {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diff = now - date;

  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "たった今";
  if (minutes < 60) return `${minutes}分前`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}時間前`;

  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}日前`;

  const months = Math.floor(days / 30);
  return `${months}ヶ月前`;
}

export function NotificationDropdown({
  isOpen,
  onClose,
}: NotificationDropdownProps) {
  const { getToken } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["notifications", "list"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return listNotifications(token, 1, 20);
    },
    enabled: isOpen,
  });

  const markReadMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return markAsRead(notificationId, token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["notifications"],
      });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return markAllAsRead(token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["notifications"],
      });
    },
  });

  const handleNotificationClick = (item: NotificationItem) => {
    if (!item.isRead) {
      markReadMutation.mutate(item.id);
    }
    const href = getNotificationHref(item);
    onClose();
    router.push(href);
  };

  const handleMarkAllAsRead = () => {
    markAllReadMutation.mutate();
  };

  const notifications = data?.notifications ?? [];
  const hasUnread = notifications.some((n) => !n.isRead);

  return (
    <div className="absolute left-0 top-full mt-2 w-80 max-h-[480px] bg-white rounded-2xl shadow-lg border border-gray-200 overflow-hidden z-50">
      {/* ヘッダー */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100">
        <h3 className="text-sm font-bold text-gray-900">通知</h3>
        {hasUnread && (
          <button
            type="button"
            onClick={handleMarkAllAsRead}
            disabled={markAllReadMutation.isPending}
            className="text-xs font-medium text-blue-600 hover:text-blue-800 transition-colors disabled:opacity-50"
          >
            すべて既読にする
          </button>
        )}
      </div>

      {/* 通知リスト */}
      <div className="overflow-y-auto max-h-[420px]">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="w-5 h-5 border-2 border-gray-300 border-t-gray-600 rounded-full animate-spin" />
          </div>
        ) : notifications.length === 0 ? (
          <div className="flex items-center justify-center py-8 text-sm text-gray-400">
            通知はありません
          </div>
        ) : (
          notifications.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => handleNotificationClick(item)}
              className={cn(
                "w-full flex items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50",
                !item.isRead && "bg-blue-50/50",
              )}
            >
              {/* アバター */}
              <div className="flex-shrink-0 w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden">
                {item.actor?.avatarUrl ? (
                  <Image
                    src={item.actor.avatarUrl}
                    alt={item.actor.displayName}
                    width={32}
                    height={32}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <span className="text-xs font-bold text-gray-500">
                    {(item.actor?.displayName ?? "?").charAt(0)}
                  </span>
                )}
              </div>

              {/* コンテンツ */}
              <div className="flex-1 min-w-0">
                <p className="text-sm text-gray-700 leading-snug line-clamp-2">
                  {getNotificationMessage(item)}
                </p>
                <p className="text-xs text-gray-400 mt-1">
                  {formatRelativeTime(item.createdAt)}
                </p>
              </div>

              {/* 未読インジケーター */}
              {!item.isRead && (
                <div className="flex-shrink-0 mt-2">
                  <div className="w-2 h-2 rounded-full bg-blue-500" />
                </div>
              )}
            </button>
          ))
        )}
      </div>
    </div>
  );
}
