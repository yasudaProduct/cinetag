"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import { useQuery } from "@tanstack/react-query";
import { getMe } from "@/lib/api/users/getMe";
import { Header } from "@/components/Header";

export default function MyPage() {
  const router = useRouter();
  const { isSignedIn, isLoaded, getToken } = useAuth();

  const {
    data: user,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return getMe(token);
    },
    enabled: isLoaded && isSignedIn,
  });

  useEffect(() => {
    if (isLoaded && !isSignedIn) {
      router.replace("/sign-in");
      return;
    }

    if (user?.display_id) {
      router.replace(`/${user.display_id}`);
    }
  }, [isLoaded, isSignedIn, user, router]);

  if (!isLoaded || isLoading) {
    return (
      <div className="min-h-screen bg-[#FFF5F5]">
        <Header />
        <div className="flex items-center justify-center h-96">
          <p className="text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="min-h-screen bg-[#FFF5F5]">
        <Header />
        <div className="flex items-center justify-center h-96">
          <p className="text-gray-600">エラーが発生しました</p>
        </div>
      </div>
    );
  }

  // リダイレクト中
  return (
    <div className="min-h-screen bg-[#FFF5F5]">
      <Header />
      <div className="flex items-center justify-center h-96">
        <p className="text-gray-600">リダイレクト中...</p>
      </div>
    </div>
  );
}
