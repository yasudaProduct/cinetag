"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import {
  Search,
  Plus,
  Tag,
  User,
  Settings,
  Bell,
  FileText,
  ShieldCheck,
} from "lucide-react";
import {
  SignedIn,
  SignedOut,
  SignInButton,
  UserButton,
  useAuth,
  useClerk,
} from "@clerk/nextjs";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getMe } from "@/lib/api/users/getMe";
import { cn } from "@/lib/utils";
import { TagModal } from "@/components/TagModal";

export const Sidebar = () => {
  const pathname = usePathname();
  const { isSignedIn, isLoaded, getToken } = useAuth();
  const { openSignIn } = useClerk();
  const queryClient = useQueryClient();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

  const { data: currentUser } = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return getMe(token);
    },
    enabled: isLoaded && isSignedIn,
  });

  const myPageHref = currentUser?.display_id
    ? `/${currentUser.display_id}`
    : "/mypage";

  const menuItems = [
    { icon: Search, label: "タグを検索", href: "/tags" },
    ...(isLoaded && isSignedIn
      ? [
          { icon: Tag, label: "フォローしたタグ", href: "/tags/following" },
          { icon: User, label: "マイページ", href: myPageHref },
          { icon: Settings, label: "設定", href: "/#settings" },
          { icon: Bell, label: "通知", href: "/#notifications" },
        ]
      : []),
  ];

  const bottomMenuItems = [
    { icon: Bell, label: "お知らせ", href: "/#news" },
    { icon: FileText, label: "利用規約", href: "/terms" },
    { icon: ShieldCheck, label: "プライバシーポリシー", href: "/privacy" },
  ];

  return (
    <>
      {/* Desktop Sidebar */}
      <aside className="hidden md:flex fixed left-0 top-0 h-screen w-64 bg-[#f7e6e6] border-gray-200 flex-col z-50">
        {/* Brand Logo */}
        <div className="p-8">
          {/* <Link href="/" className="flex items-center gap-3">
            <div className="bg-blue-500 text-white p-1.5 rounded-lg text-sm font-bold shadow-sm">
              🍿
            </div>
            <span className="text-2xl font-black text-gray-900 tracking-tighter">
              cinetag
            </span>
          </Link> */}
        </div>

        {/* Main Menu */}
        <nav className="flex-1 px-4 py-2 space-y-1.5">
          {menuItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <div className="relative group" key={item.href}>
                <div
                  className={cn(
                    "absolute left-0 top-0 bottom-0 w-1.5 bg-[#FFD75E] transition-opacity",
                    isActive
                      ? "opacity-100"
                      : "opacity-0 group-hover:opacity-100"
                  )}
                />
                <Link
                  href={item.href}
                  className={cn(
                    "relative flex items-center gap-3 px-4 py-3 rounded-r-2xl text-sm font-bold transition-all overflow-hidden",
                    isActive
                      ? "text-gray-900"
                      : "text-gray-600 group-hover:text-gray-900"
                  )}
                >
                  <item.icon
                    className={cn(
                      "w-5 h-5 transition-colors",
                      isActive
                        ? "text-gray-900"
                        : "text-gray-400 group-hover:text-gray-900"
                    )}
                  />
                  {item.label}
                </Link>
              </div>
            );
          })}

          <button
            type="button"
            disabled={!isLoaded}
            onClick={() => {
              if (isSignedIn) {
                setIsCreateModalOpen(true);
                return;
              }
              openSignIn({}); // 未ログイン時はサインインを促す
            }}
            className={cn(
              "w-full flex items-center justify-center gap-2 mt-4 px-4 py-3 bg-[#FFD75E] text-gray-900 text-sm font-bold rounded-2xl transition-all shadow-sm hover:shadow active:scale-[0.98]",
              isLoaded
                ? "hover:bg-[#ffcf40]"
                : "opacity-60 cursor-not-allowed hover:bg-[#FFD75E]"
            )}
          >
            <Plus className="w-5 h-5" />
            新しいタグを作成
          </button>
        </nav>

        {/* Bottom Menu */}
        <div className="px-4 py-6 space-y-2">
          <div className="flex items-center gap-3 px-4 py-2 text-sm font-medium text-gray-300 cursor-not-allowed">
            {/* <Circle className="w-10 h-10" /> */}
            {/* <span>アイコン(未作成)</span> */}
            <div className="bg-blue-500 text-white p-1.5 rounded-lg text-sm font-bold shadow-sm">
              🍿
            </div>
          </div>

          {bottomMenuItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 px-4 py-2 rounded-xl text-xs font-bold transition-colors",
                  isActive
                    ? "text-blue-600"
                    : "text-gray-400 hover:text-gray-600"
                )}
              >
                <item.icon className="w-4 h-4" />
                {item.label}
              </Link>
            );
          })}

          <div className="px-4 py-2 text-[10px] font-bold text-gray-300 uppercase tracking-widest">
            © 2026 cinetag
          </div>

          {/* User Status */}
          <div className="mt-4 pt-4 border-t border-gray-100 px-2">
            <SignedIn>
              <div className="flex items-center justify-between bg-gray-50 p-2 rounded-2xl border border-gray-100">
                <div className="flex items-center gap-2 pl-1">
                  <span className="text-xs font-bold text-gray-900">
                    Account
                  </span>
                </div>
                <UserButton
                  appearance={{
                    elements: {
                      avatarBox: "w-8 h-8 border-2 border-white shadow-sm",
                    },
                  }}
                />
              </div>
            </SignedIn>
            <SignedOut>
              <SignInButton mode="modal">
                <button className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-[#FFD75E] text-gray-900 text-sm font-bold rounded-2xl hover:bg-[#ffcf40] transition-all shadow-sm hover:shadow active:scale-[0.98]">
                  <User className="w-4 h-4" />
                  ログイン
                </button>
              </SignInButton>
            </SignedOut>
          </div>
        </div>
      </aside>

      {/* Mobile Bottom Navigation */}
      <nav className="fixed bottom-0 left-0 right-0 z-50 flex items-center justify-around bg-white/90 backdrop-blur-md border-t border-gray-200 px-2 py-3 md:hidden safe-area-bottom shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)]">
        {menuItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center p-2 rounded-xl transition-all",
                isActive ? "text-[#FFD75E]" : "text-gray-400"
              )}
            >
              <item.icon
                className={cn(
                  "w-6 h-6",
                  isActive ? "fill-current" : "stroke-current"
                )}
              />
            </Link>
          );
        })}

        {/* Mobile Create Button */}
        <button
          type="button"
          disabled={!isLoaded}
          onClick={() => {
            if (isSignedIn) {
              setIsCreateModalOpen(true);
              return;
            }
            openSignIn({});
          }}
          className={cn(
            "flex items-center justify-center p-3 rounded-full bg-[#FFD75E] text-gray-900 shadow-lg active:scale-95 transition-transform",
            isLoaded ? "opacity-100" : "opacity-50"
          )}
        >
          <Plus className="w-6 h-6" />
        </button>
      </nav>

      <SignedIn>
        <TagModal
          open={isCreateModalOpen}
          onClose={() => setIsCreateModalOpen(false)}
          onCreated={() => {
            queryClient.invalidateQueries({
              predicate: (query) => query.queryKey[0] === "tags",
            });
          }}
        />
      </SignedIn>
    </>
  );
};
