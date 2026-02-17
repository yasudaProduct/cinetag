"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useUser, useClerk, useAuth } from "@clerk/nextjs";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Camera, Trash2, AlertTriangle, LogOut } from "lucide-react";
import Link from "next/link";
import Image from "next/image";
import { Spinner } from "@/components/ui/spinner";
import { ImageUploadModal } from "@/components/ImageUploadModal";
import { getMe } from "@/lib/api/users/getMe";
import { updateMe } from "@/lib/api/users/updateMe";

export function SettingsClient() {
  const router = useRouter();
  const { user, isLoaded } = useUser();
  const { signOut } = useClerk();
  const { getToken } = useAuth();
  const queryClient = useQueryClient();

  const [displayName, setDisplayName] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);
  const [isImageModalOpen, setIsImageModalOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleteConfirmText, setDeleteConfirmText] = useState("");
  const [message, setMessage] = useState<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // バックエンドからユーザー情報を取得
  const { data: backendUser } = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return getMe(token);
    },
    enabled: isLoaded && !!user,
  });

  // バックエンドからユーザー情報が取得できたら表示名を設定
  useEffect(() => {
    if (backendUser) {
      setDisplayName(backendUser.display_name);
    }
  }, [backendUser]);

  // 未ログインの場合はリダイレクト
  useEffect(() => {
    if (isLoaded && !user) {
      router.push("/");
    }
  }, [isLoaded, user, router]);

  if (!isLoaded) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner size="lg" />
      </div>
    );
  }

  if (!user) {
    return null;
  }

  const handleUpdateDisplayName = async () => {
    if (!displayName.trim()) {
      setMessage({ type: "error", text: "表示名を入力してください" });
      return;
    }

    setIsUpdating(true);
    setMessage(null);

    try {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) {
        setMessage({ type: "error", text: "認証情報の取得に失敗しました" });
        return;
      }

      await updateMe(token, { display_name: displayName.trim() });

      // キャッシュを更新
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });

      setMessage({ type: "success", text: "表示名を更新しました" });
    } catch (err) {
      setMessage({
        type: "error",
        text: err instanceof Error ? err.message : "表示名の更新に失敗しました",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  const handleImageUpload = async (file: File) => {
    await user.setProfileImage({ file });
    // Clerk Webhook でバックエンドが更新されるまで少し待機してからキャッシュを無効化
    setTimeout(() => {
      queryClient.invalidateQueries({ queryKey: ["users", "me"] });
    }, 1000);
    setMessage({ type: "success", text: "プロフィール画像を更新しました" });
  };

  const handleDeleteAccount = async () => {
    if (deleteConfirmText !== "削除する") {
      setMessage({
        type: "error",
        text: "確認テキストを正しく入力してください",
      });
      return;
    }

    setIsDeleting(true);
    setMessage(null);

    try {
      await user.delete();
      await signOut();
      router.push("/");
    } catch {
      setMessage({ type: "error", text: "アカウントの削除に失敗しました" });
      setIsDeleting(false);
    }
  };

  // 連携アカウント情報
  const externalAccounts = user.externalAccounts || [];

  return (
    <div className="min-h-screen">
      <div className="max-w-2xl mx-auto px-4 py-8">
        {/* ヘッダー */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            href="/"
            className="p-2 rounded-full hover:bg-gray-100 transition-colors"
          >
            <ArrowLeft className="w-6 h-6 text-gray-600" />
          </Link>
          <h1 className="text-2xl font-bold text-gray-900">設定</h1>
        </div>

        {/* メッセージ表示 */}
        {message && (
          <div
            className={`mb-6 p-4 rounded-xl ${
              message.type === "success"
                ? "bg-green-50 text-green-800 border border-green-200"
                : "bg-red-50 text-red-800 border border-red-200"
            }`}
          >
            {message.text}
          </div>
        )}

        {/* プロフィール確認・更新セクション */}
        <section className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 mb-6">
          <h2 className="text-lg font-bold text-gray-900 mb-6">プロフィール</h2>

          {/* プロフィール画像 */}
          <div className="flex items-center gap-6 mb-8">
            <div className="relative">
              <div className="w-24 h-24 rounded-full overflow-hidden bg-gray-100 border-4 border-white shadow-md">
                {user.imageUrl ? (
                  <Image
                    src={user.imageUrl}
                    alt="プロフィール画像"
                    width={96}
                    height={96}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400 text-3xl font-bold">
                    {displayName.charAt(0).toUpperCase() || "?"}
                  </div>
                )}
              </div>
              <button
                type="button"
                onClick={() => setIsImageModalOpen(true)}
                className="absolute bottom-0 right-0 p-2 bg-[#FFD75E] rounded-full shadow-md hover:bg-[#ffcf40] transition-colors"
              >
                <Camera className="w-4 h-4 text-gray-900" />
              </button>
            </div>
            <div>
              <p className="text-sm text-gray-500 mb-1">プロフィール画像</p>
              <p className="text-xs text-gray-400">
                クリックして画像を変更（5MB以下）
              </p>
            </div>
          </div>

          {/* 表示名 */}
          <div className="mb-6">
            <label
              htmlFor="displayName"
              className="block text-sm font-medium text-gray-700 mb-2"
            >
              表示名
            </label>
            <div className="flex gap-3">
              <input
                id="displayName"
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="表示名を入力"
                className="flex-1 px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#FFD75E] focus:border-transparent"
              />
              <button
                type="button"
                onClick={handleUpdateDisplayName}
                disabled={isUpdating}
                className="px-6 py-3 bg-[#FFD75E] text-gray-900 font-bold rounded-xl hover:bg-[#ffcf40] transition-colors disabled:opacity-50"
              >
                {isUpdating ? <Spinner size="sm" /> : "更新"}
              </button>
            </div>
          </div>

          {/* メールアドレス（読み取り専用） */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              メールアドレス
            </label>
            <p className="px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-gray-600">
              {user.primaryEmailAddress?.emailAddress || "未設定"}
            </p>
          </div>

          {/* 連携アカウント */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              連携アカウント
            </label>
            {externalAccounts.length > 0 ? (
              <div className="space-y-2">
                {externalAccounts.map((account) => (
                  <div
                    key={account.id}
                    className="flex items-center gap-3 px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
                  >
                    <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center">
                      {account.provider.includes("google") ? (
                        <span className="text-sm">G</span>
                      ) : account.provider.includes("github") ? (
                        <span className="text-sm">GH</span>
                      ) : (
                        <span className="text-sm">🔗</span>
                      )}
                    </div>
                    <div>
                      <p className="text-sm font-medium text-gray-900 capitalize">
                        {account.provider.replace(/^oauth_/, "")}
                      </p>
                      <p className="text-xs text-gray-500">
                        {account.emailAddress || "連携済み"}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-gray-500">
                連携アカウントはありません
              </p>
            )}
          </div>
        </section>

        {/* サインアウトセクション */}
        <section className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100 mb-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">サインアウト</h2>
          <p className="text-sm text-gray-600 mb-4">
            このデバイスからサインアウトします。
          </p>
          <button
            type="button"
            onClick={() => signOut({ redirectUrl: "/" })}
            className="flex items-center gap-2 px-4 py-3 bg-gray-100 text-gray-700 font-bold rounded-xl hover:bg-gray-200 transition-colors"
          >
            <LogOut className="w-4 h-4" />
            サインアウト
          </button>
        </section>

        {/* 退会セクション */}
        <section className="bg-white rounded-2xl p-6 shadow-sm border border-red-100">
          <h2 className="text-lg font-bold text-red-600 mb-4 flex items-center gap-2">
            <AlertTriangle className="w-5 h-5" />
            アカウントの削除
          </h2>
          <p className="text-sm text-gray-600 mb-4">
            アカウントを削除すると、すべてのデータが完全に削除されます。この操作は取り消せません。
          </p>

          {!showDeleteConfirm ? (
            <button
              type="button"
              onClick={() => setShowDeleteConfirm(true)}
              className="flex items-center gap-2 px-4 py-3 bg-red-50 text-red-600 font-bold rounded-xl hover:bg-red-100 transition-colors"
            >
              <Trash2 className="w-4 h-4" />
              アカウントを削除
            </button>
          ) : (
            <div className="space-y-4 p-4 bg-red-50 rounded-xl border border-red-200">
              <p className="text-sm font-medium text-red-800">
                本当に削除しますか？確認のため「削除する」と入力してください。
              </p>
              <input
                type="text"
                value={deleteConfirmText}
                onChange={(e) => setDeleteConfirmText(e.target.value)}
                placeholder="削除する"
                className="w-full px-4 py-3 border border-red-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300"
              />
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => {
                    setShowDeleteConfirm(false);
                    setDeleteConfirmText("");
                  }}
                  className="flex-1 px-4 py-3 bg-gray-100 text-gray-700 font-bold rounded-xl hover:bg-gray-200 transition-colors"
                >
                  キャンセル
                </button>
                <button
                  type="button"
                  onClick={handleDeleteAccount}
                  disabled={isDeleting || deleteConfirmText !== "削除する"}
                  className="flex-1 px-4 py-3 bg-red-600 text-white font-bold rounded-xl hover:bg-red-700 transition-colors disabled:opacity-50"
                >
                  {isDeleting ? <Spinner size="sm" /> : "削除を実行"}
                </button>
              </div>
            </div>
          )}
        </section>
      </div>

      {/* 画像アップロードモーダル */}
      <ImageUploadModal
        open={isImageModalOpen}
        onClose={() => setIsImageModalOpen(false)}
        onUpload={handleImageUpload}
        currentImageUrl={user.imageUrl}
      />
    </div>
  );
}
