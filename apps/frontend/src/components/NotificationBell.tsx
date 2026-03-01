"use client";

import { useState, useRef, useEffect } from "react";
import { Bell } from "lucide-react";
import { useAuth } from "@clerk/nextjs";
import { useQuery } from "@tanstack/react-query";
import { getUnreadCount } from "@/lib/api/notifications/getUnreadCount";
import dynamic from "next/dynamic";
import { cn } from "@/lib/utils";

const NotificationDropdown = dynamic(
  () =>
    import("@/components/NotificationDropdown").then(
      (mod) => mod.NotificationDropdown,
    ),
  { ssr: false },
);

export function NotificationBell() {
  const { getToken, isSignedIn, isLoaded } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // 外側クリックで閉じる
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  // 未読数ポーリング（60秒間隔）
  const { data: unreadCount = 0 } = useQuery({
    queryKey: ["notifications", "unread-count"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return getUnreadCount(token);
    },
    refetchInterval: 60_000,
    enabled: isLoaded && !!isSignedIn,
  });

  if (!isLoaded || !isSignedIn) return null;

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "relative flex items-center justify-center p-2 rounded-xl transition-all",
          isOpen ? "text-gray-900" : "text-gray-400 hover:text-gray-900",
        )}
        aria-label="通知"
      >
        <Bell className="w-5 h-5" />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 flex items-center justify-center min-w-[18px] h-[18px] px-1 bg-red-500 text-white text-[10px] font-bold rounded-full leading-none">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <NotificationDropdown
          isOpen={isOpen}
          onClose={() => setIsOpen(false)}
        />
      )}
    </div>
  );
}
