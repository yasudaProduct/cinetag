"use client";

import { useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { listNotifications } from "@/lib/api/notifications/listNotifications";
import { markAsRead } from "@/lib/api/notifications/markAsRead";
import { markAllAsRead } from "@/lib/api/notifications/markAllAsRead";
import {
  getNotificationMessage,
  getNotificationHref,
  formatRelativeTime,
} from "@/components/NotificationDropdown";
import type { NotificationItem } from "@/lib/validation/notification.api";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 20;

export function NotificationsClient() {
  const { getToken, isSignedIn, isLoaded } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ["notifications", "all", page],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return listNotifications(token, page, PAGE_SIZE);
    },
    enabled: isLoaded && isSignedIn,
  });

  const markReadMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return markAsRead(notificationId, token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return markAllAsRead(token);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const handleNotificationClick = (item: NotificationItem) => {
    if (!item.isRead) {
      markReadMutation.mutate(item.id);
    }
    router.push(getNotificationHref(item));
  };

  const notifications = data?.notifications ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);
  const hasUnread = notifications.some((n) => !n.isRead);

  if (!isLoaded) {
    return (
      <div className="min-h-screen">
        <div className="flex items-center justify-center h-96">
          <Spinner size="lg" className="text-gray-600" />
        </div>
      </div>
    );
  }

  if (isLoaded && !isSignedIn) {
    router.replace("/sign-in");
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-2xl mx-auto px-4 py-8">
        {/* ヘッダー */}
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-xl font-bold text-gray-900">通知</h1>
          {hasUnread && (
            <button
              type="button"
              onClick={() => markAllReadMutation.mutate()}
              disabled={markAllReadMutation.isPending}
              className="text-sm font-medium text-blue-600 hover:text-blue-800 transition-colors disabled:opacity-50"
            >
              すべて既読にする
            </button>
          )}
        </div>

        {/* 通知リスト */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Spinner size="lg" className="text-gray-600" />
            </div>
          ) : notifications.length === 0 ? (
            <div className="flex items-center justify-center py-16 text-sm text-gray-400">
              通知はありません
            </div>
          ) : (
            notifications.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => handleNotificationClick(item)}
                className={cn(
                  "w-full flex items-start gap-3 px-4 py-4 text-left transition-colors hover:bg-gray-50 border-b border-gray-100 last:border-b-0",
                  !item.isRead && "bg-blue-50/50",
                )}
              >
                {/* アバター */}
                <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden">
                  {item.actor?.avatarUrl ? (
                    <Image
                      src={item.actor.avatarUrl}
                      alt={item.actor.displayName}
                      width={40}
                      height={40}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <span className="text-sm font-bold text-gray-500">
                      {(item.actor?.displayName ?? "?").charAt(0)}
                    </span>
                  )}
                </div>

                {/* コンテンツ */}
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-gray-700 leading-snug">
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

        {/* ページネーション */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-6">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              前へ
            </button>
            <span className="text-sm text-gray-500">
              {page} / {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              次へ
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
