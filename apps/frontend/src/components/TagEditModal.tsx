"use client";

import { X, Pencil } from "lucide-react";
import { useMemo, useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { useMutation } from "@tanstack/react-query";
import { updateTag } from "@/lib/api/tags/update";
import { Switch } from "@/components/ui/switch";
import { getBackendTokenOrThrow } from "@/lib/api/_shared/auth";
import type { AddMoviePolicy } from "@/lib/validation/tag.api";

type TagEditModalProps = {
  open: boolean;
  tag: {
    id: string;
    title: string;
    description: string;
    is_public: boolean;
    add_movie_policy?: AddMoviePolicy;
  };
  onClose: () => void;
  onUpdated: () => void;
};

export const TagEditModal = ({
  open,
  tag,
  onClose,
  onUpdated,
}: TagEditModalProps) => {
  const { getToken } = useAuth();

  const [title, setTitle] = useState(tag.title);
  const [description, setDescription] = useState(tag.description);
  const [isPublic, setIsPublic] = useState(tag.is_public);
  const [addMoviePolicy, setAddMoviePolicy] = useState<AddMoviePolicy>(
    tag.add_movie_policy ?? "everyone"
  );
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const canSubmit = useMemo(() => title.trim().length > 0, [title]);

  const updateMutation = useMutation({
    mutationFn: async () => {
      const token = await getBackendTokenOrThrow(getToken);

      return await updateTag({
        tagId: tag.id,
        token,
        input: {
          title: title.trim(),
          description:
            description.trim().length > 0 ? description.trim() : null,
          is_public: isPublic,
          add_movie_policy: addMoviePolicy,
        },
      });
    },
    onSuccess: () => {
      setErrorMessage(null);
      onUpdated();
      onClose();
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "更新に失敗しました。"
      );
    },
  });

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-xl mx-4 rounded-3xl bg-white shadow-xl border border-gray-200">
        <div className="flex items-start justify-between px-7 pt-7">
          <div>
            <h2 className="text-xl md:text-2xl font-extrabold text-gray-900 tracking-tight">
              タグを編集
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              タイトル・説明・公開設定を更新できます。
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="閉じる"
            className="ml-4 inline-flex h-9 w-9 items-center justify-center rounded-full border border-gray-200 bg-white text-gray-600 hover:bg-gray-50 hover:text-gray-900 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="px-7 pb-7 pt-5 space-y-5">
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-gray-600">
              タイトル
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-gray-600">
              説明
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              className="w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
            />
          </div>

          {/* Add Movie Policy */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-gray-600">
              映画の追加権限
            </label>
            <div className="flex flex-col gap-2">
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="radio"
                  name="add_movie_policy"
                  value="everyone"
                  checked={addMoviePolicy === "everyone"}
                  onChange={() => setAddMoviePolicy("everyone")}
                  className="w-4 h-4 text-blue-600 border-gray-300 focus:ring-blue-500"
                />
                <span className="text-sm text-gray-900">誰でも追加可能</span>
              </label>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="radio"
                  name="add_movie_policy"
                  value="owner_only"
                  checked={addMoviePolicy === "owner_only"}
                  onChange={() => setAddMoviePolicy("owner_only")}
                  className="w-4 h-4 text-blue-600 border-gray-300 focus:ring-blue-500"
                />
                <span className="text-sm text-gray-900">作成者のみ</span>
              </label>
            </div>
          </div>

          {/* TODO: 現状タグは全て公開設定とする。タグのオプションを整理後実装 */}
          {/* <div className="flex items-center justify-between rounded-2xl border border-gray-200 bg-gray-50 px-4 py-3">
            <div>
              <div className="text-sm font-semibold text-gray-900">
                公開する
              </div>
              <div className="text-xs text-gray-600">
                オフにすると作成者のみ閲覧できます。
              </div>
            </div>
            <Switch
              checked={isPublic}
              onCheckedChange={setIsPublic}
              aria-label="公開する"
              className={[
                "h-9 w-16 border border-transparent",
                "data-[state=checked]:bg-blue-600 data-[state=unchecked]:bg-gray-200",
                "[&_[data-slot=switch-thumb]]:size-7",
                "[&_[data-slot=switch-thumb]]:bg-white",
                "[&_[data-slot=switch-thumb]]:shadow",
                "[&_[data-slot=switch-thumb]]:data-[state=checked]:translate-x-8",
                "[&_[data-slot=switch-thumb]]:data-[state=unchecked]:translate-x-1",
              ].join(" ")}
            />
          </div> */}

          {errorMessage && (
            <div className="text-sm text-red-600 font-medium">
              {errorMessage}
            </div>
          )}

          <div className="flex justify-end gap-3 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="rounded-full border border-gray-200 bg-white px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
            >
              キャンセル
            </button>
            <button
              type="button"
              disabled={updateMutation.isPending || !canSubmit}
              onClick={() => updateMutation.mutate()}
              className="inline-flex items-center justify-center rounded-full bg-gray-900 px-7 py-3 text-sm font-semibold text-white shadow-sm hover:bg-black disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Pencil className="w-4 h-4 mr-2" />
              {updateMutation.isPending ? "保存中..." : "保存する"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
