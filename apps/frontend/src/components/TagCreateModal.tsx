"use client";

import { X, Film } from "lucide-react";
import { useState, FormEvent } from "react";
import { useAuth, useUser } from "@clerk/nextjs";
import { useMutation } from "@tanstack/react-query";
import {
  TagCreateInputSchema,
  getFirstZodErrorMessage,
  type AddMoviePolicyForm,
} from "@/lib/validation/tag.form";
import { createTag } from "@/lib/api/tags/create";
import { getBackendTokenOrThrow } from "@/lib/api/_shared/auth";

interface TagCreateModalProps {
  open: boolean;
  onClose: () => void;
  onCreated: (tag: CreatedTagForList) => void;
}

export interface CreatedTagForList extends Record<string, unknown> {
  id: string;
  title: string;
  description?: string | null;
  author: string;
  movie_count: number;
  follower_count: number;
  images: string[];
  created_at?: string;
}

export const TagCreateModal = ({
  open,
  onClose,
  onCreated,
}: TagCreateModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [addMoviePolicy, setAddMoviePolicy] = useState<AddMoviePolicyForm>("everyone");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const { getToken } = useAuth();
  const { user } = useUser();

  const createMutation = useMutation({
    mutationFn: async (input: { title: string; description?: string; add_movie_policy: AddMoviePolicyForm }) => {
      const token = await getBackendTokenOrThrow(getToken);
      return await createTag({
        token,
        input: {
          title: input.title,
          description: input.description,
          is_public: true,
          add_movie_policy: input.add_movie_policy,
        },
      });
    },
    onSuccess: (created) => {
      const author = user?.username ?? user?.fullName ?? "me";
      onCreated({
        id: created.id,
        title: created.title,
        description: created.description ?? null,
        author,
        movie_count: created.movie_count ?? 0,
        follower_count: created.follower_count ?? 0,
        images: [],
        created_at: created.created_at,
      });

      setName("");
      setDescription("");
      setAddMoviePolicy("everyone");
      onClose();
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "作成に失敗しました。"
      );
    },
  });

  if (!open) return null;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);

    const parsedInput = TagCreateInputSchema.safeParse({
      title: name,
      description: description.length > 0 ? description : undefined,
      add_movie_policy: addMoviePolicy,
    });
    if (!parsedInput.success) {
      setErrorMessage(getFirstZodErrorMessage(parsedInput.error));
      return;
    }

    const { title, description: desc, add_movie_policy } = parsedInput.data;
    await createMutation.mutateAsync({
      title,
      description: desc && desc.length > 0 ? desc : undefined,
      add_movie_policy,
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      {/* Card */}
      <div className="w-full max-w-xl mx-4 rounded-3xl bg-[#FFF9F3] shadow-xl border border-[#F3E1D6]">
        {/* Header */}
        <div className="flex items-start justify-between px-8 pt-8">
          <div>
            <h2 className="text-2xl md:text-3xl font-extrabold text-[#1F1A2B] tracking-tight">
              新しいタグを作成
            </h2>
            <p className="mt-2 text-sm md:text-base text-[#7C7288]">
              あなたの映画コレクションを世界とシェアしましょう。
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="閉じる"
            className="ml-4 inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#E4D3C7] bg-white text-[#7C7288] hover:bg-[#FDF1E7] hover:text-[#1F1A2B] transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="px-8 pb-8 pt-6 space-y-6">
          {/* Tag Name */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              名前
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Mind-Bending Sci-Fi"
              className="w-full rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] px-4 py-3 text-sm text-[#1F1A2B] shadow-[0_1px_0_rgba(0,0,0,0.03)] focus:outline-none focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8]"
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              説明
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="このタグについての簡単な説明を書いてください。"
              rows={4}
              className="w-full rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] px-4 py-3 text-sm text-[#1F1A2B] shadow-[0_1px_0_rgba(0,0,0,0.03)] focus:outline-none focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8] resize-none"
            />
          </div>

          {/* Add Movie Policy */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
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
                  className="w-4 h-4 text-[#FF5C5C] border-[#E4D3C7] focus:ring-[#FF8C75]"
                />
                <span className="text-sm text-[#1F1A2B]">誰でも追加可能</span>
              </label>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="radio"
                  name="add_movie_policy"
                  value="owner_only"
                  checked={addMoviePolicy === "owner_only"}
                  onChange={() => setAddMoviePolicy("owner_only")}
                  className="w-4 h-4 text-[#FF5C5C] border-[#E4D3C7] focus:ring-[#FF8C75]"
                />
                <span className="text-sm text-[#1F1A2B]">作成者のみ</span>
              </label>
            </div>
          </div>

          {/* Cover Image (dummy uploader) */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              カバー画像
            </label>
            <div className="rounded-2xl border-2 border-dashed border-[#E4D3C7] bg-[#FFFDF8] px-6 py-8 flex flex-col items-center justify-center text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-[#FFF2E0] mb-3">
                <Film className="w-7 h-7 text-[#FF8C75]" />
              </div>
              <p className="text-sm">
                <span className="font-semibold text-[#FF5C5C]">
                  Upload a file
                </span>{" "}
                <span className="text-[#7C7288]">or drag and drop</span>
              </p>
              <p className="mt-1 text-xs text-[#B09EA0]">
                PNG, JPG, GIF up to 10MB
              </p>
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end pt-2">
            {errorMessage && (
              <div className="mr-auto text-sm text-red-600 font-medium">
                {errorMessage}
              </div>
            )}
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="inline-flex items-center justify-center rounded-full bg-[#FF5C5C] px-8 py-3 text-sm font-semibold text-white shadow-[0_8px_0_#D44242] hover:translate-y-0.5 hover:shadow-[0_6px_0_#D44242] active:translate-y-1 active:shadow-[0_3px_0_#D44242] transition-transform"
            >
              {createMutation.isPending ? "作成中..." : "タグを作成"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
