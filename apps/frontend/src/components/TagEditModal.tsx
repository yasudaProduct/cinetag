"use client";

import { X, Pencil } from "lucide-react";
import { useMemo, useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { useMutation } from "@tanstack/react-query";
import { updateTag } from "@/lib/api/tags/update";

type TagEditModalProps = {
  open: boolean;
  tagId: string;
  initialTitle: string;
  initialDescription: string;
  initialIsPublic?: boolean;
  onClose: () => void;
  onUpdated: () => void;
};

export const TagEditModal = ({
  open,
  tagId,
  initialTitle,
  initialDescription,
  initialIsPublic = true,
  onClose,
  onUpdated,
}: TagEditModalProps) => {
  const { getToken } = useAuth();

  const [title, setTitle] = useState(initialTitle);
  const [description, setDescription] = useState(initialDescription);
  const [isPublic, setIsPublic] = useState(initialIsPublic);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const canSubmit = useMemo(() => title.trim().length > 0, [title]);

  const updateMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) {
        throw new Error("認証情報の取得に失敗しました。再ログインしてください。");
      }

      return await updateTag({
        tagId,
        token,
        input: {
          title: title.trim(),
          description: description.trim().length > 0 ? description.trim() : null,
          is_public: isPublic,
        },
      });
    },
    onSuccess: () => {
      setErrorMessage(null);
      onUpdated();
      onClose();
    },
    onError: (err) => {
      setErrorMessage(err instanceof Error ? err.message : "更新に失敗しました。");
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

          <div className="flex items-center justify-between rounded-2xl border border-gray-200 bg-gray-50 px-4 py-3">
            <div>
              <div className="text-sm font-semibold text-gray-900">公開する</div>
              <div className="text-xs text-gray-600">オフにすると作成者のみ閲覧できます。</div>
            </div>
            <button
              type="button"
              onClick={() => setIsPublic((v) => !v)}
              className={[
                "h-9 w-16 rounded-full border transition-colors relative",
                isPublic ? "bg-blue-600 border-blue-600" : "bg-gray-200 border-gray-200",
              ].join(" ")}
              aria-pressed={isPublic}
            >
              <span
                className={[
                  "absolute top-1/2 -translate-y-1/2 h-7 w-7 rounded-full bg-white shadow transition-transform",
                  isPublic ? "translate-x-8" : "translate-x-1",
                ].join(" ")}
              />
            </button>
          </div>

          {errorMessage && <div className="text-sm text-red-600 font-medium">{errorMessage}</div>}

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


