"use client";

import { Pencil, Film, X } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { useMemo, useState, type FormEvent } from "react";
import { useAuth, useUser } from "@clerk/nextjs";
import { useMutation } from "@tanstack/react-query";
import {
  TagCreateInputSchema,
  getFirstZodErrorMessage,
  type AddMoviePolicyForm,
} from "@/lib/validation/tag.form";
import type { AddMoviePolicy } from "@/lib/validation/tag.api";
import { createTag } from "@/lib/api/tags/create";
import { updateTag } from "@/lib/api/tags/update";
import { getBackendTokenOrThrow } from "@/lib/api/_shared/auth";
import { Modal } from "@/components/Modal";

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

type TagModalProps = {
  open: boolean;
  onClose: () => void;
  tag?: {
    id: string;
    title: string;
    description: string;
    is_public: boolean;
    add_movie_policy?: AddMoviePolicy;
  };
  onCreated?: (tag: CreatedTagForList) => void;
  onUpdated?: () => void;
};

export const TagModal = (props: TagModalProps) => {
  const { open, onClose, tag } = props;
  const { getToken } = useAuth();
  const { user } = useUser();

  const isEditMode = tag !== undefined;

  // 初期値の設定
  const getInitialFormValues = () => {
    if (isEditMode && tag) {
      return {
        title: tag.title,
        description: tag.description,
        addMoviePolicy: (tag.add_movie_policy ?? "everyone") as AddMoviePolicy,
      };
    } else {
      return {
        title: "",
        description: "",
        addMoviePolicy: "everyone" as AddMoviePolicyForm,
      };
    }
  };

  const [formValues, setFormValues] = useState(getInitialFormValues);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isPublic, setIsPublic] = useState(
    isEditMode && tag ? tag.is_public : true
  ); // TODO デフォルトfalseに変更する

  const canSubmit = useMemo(
    () => formValues.title.trim().length > 0,
    [formValues.title]
  );

  // Create mutation
  const createMutation = useMutation({
    mutationFn: async (input: {
      title: string;
      description?: string;
      add_movie_policy: AddMoviePolicyForm;
    }) => {
      const token = await getBackendTokenOrThrow(getToken);
      return await createTag({
        token,
        input: {
          title: input.title,
          description: input.description,
          is_public: isPublic,
          add_movie_policy: input.add_movie_policy,
        },
      });
    },
    onSuccess: (created) => {
      if (!isEditMode && props.onCreated) {
        const author = user?.username ?? user?.fullName ?? "me";
        props.onCreated({
          id: created.id,
          title: created.title,
          description: created.description ?? null,
          author,
          movie_count: created.movie_count ?? 0,
          follower_count: created.follower_count ?? 0,
          images: [],
          created_at: created.created_at,
        });
      }

      setFormValues({
        title: "",
        description: "",
        addMoviePolicy: "everyone",
      });
      onClose();
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "作成に失敗しました。"
      );
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: async () => {
      if (!isEditMode || !tag) return;
      const token = await getBackendTokenOrThrow(getToken);

      return await updateTag({
        tagId: tag.id,
        token,
        input: {
          title: formValues.title.trim(),
          description:
            formValues.description.trim().length > 0
              ? formValues.description.trim()
              : null,
          is_public: isPublic,
          add_movie_policy: formValues.addMoviePolicy,
        },
      });
    },
    onSuccess: () => {
      if (isEditMode && props.onUpdated) {
        setErrorMessage(null);
        props.onUpdated();
        onClose();
      }
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "更新に失敗しました。"
      );
    },
  });

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!isEditMode) {
      setErrorMessage(null);

      const parsedInput = TagCreateInputSchema.safeParse({
        title: formValues.title,
        description:
          formValues.description.length > 0
            ? formValues.description
            : undefined,
        add_movie_policy: formValues.addMoviePolicy as AddMoviePolicyForm,
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
    } else {
      if (!canSubmit || updateMutation.isPending) return;
      updateMutation.mutate();
    }
  };

  const handleTitleChange = (value: string) => {
    setFormValues({ ...formValues, title: value });
  };

  const handleDescriptionChange = (value: string) => {
    setFormValues({ ...formValues, description: value });
  };

  const handleAddMoviePolicyChange = (
    value: AddMoviePolicyForm | AddMoviePolicy
  ) => {
    setFormValues({ ...formValues, addMoviePolicy: value });
  };

  const title = isEditMode ? "タグを編集" : "新しいタグを作成";
  const subtitle = isEditMode
    ? "タイトル・説明・公開設定を更新できます。"
    : "あなたの映画コレクションを世界とシェアしましょう。";
  const submitLabel = isEditMode ? "保存する" : "タグを作成";
  const isSubmitting = isEditMode
    ? updateMutation.isPending
    : createMutation.isPending;

  return (
    <Modal open={open} onClose={onClose}>
      <div className="w-full max-w-xl mx-4 rounded-3xl bg-[#FFF9F3] shadow-xl border border-[#F3E1D6]">
        {/* Header */}
        <div className="flex items-start justify-between px-8 pt-8">
          <div>
            <h2 className="text-2xl md:text-3xl font-extrabold text-[#1F1A2B] tracking-tight">
              {title}
            </h2>
            <p className="mt-2 text-sm md:text-base text-[#7C7288]">
              {subtitle}
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
        <div className="px-8 pb-8 pt-6 space-y-6">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Tag Name */}
            <div className="space-y-2">
              <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
                名前
              </label>
              <input
                type="text"
                value={formValues.title}
                onChange={(e) => handleTitleChange(e.target.value)}
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
                value={formValues.description}
                onChange={(e) => handleDescriptionChange(e.target.value)}
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
                    checked={formValues.addMoviePolicy === "everyone"}
                    onChange={() => handleAddMoviePolicyChange("everyone")}
                    className="w-4 h-4 text-[#FF5C5C] border-[#E4D3C7] focus:ring-[#FF8C75]"
                  />
                  <span className="text-sm text-[#1F1A2B]">誰でも追加可能</span>
                </label>
                <label className="flex items-center gap-3 cursor-pointer">
                  <input
                    type="radio"
                    name="add_movie_policy"
                    value="owner_only"
                    checked={formValues.addMoviePolicy === "owner_only"}
                    onChange={() => handleAddMoviePolicyChange("owner_only")}
                    className="w-4 h-4 text-[#FF5C5C] border-[#E4D3C7] focus:ring-[#FF8C75]"
                  />
                  <span className="text-sm text-[#1F1A2B]">作成者のみ</span>
                </label>
              </div>
            </div>

            {/* Is Public */}
            <div className="flex items-center justify-between rounded-2xl border border-[#E4D3C7] bg-[#FFFDF8] px-4 py-3">
              <div>
                <div className="text-sm font-semibold text-[#1F1A2B]">
                  公開する
                </div>
                <div className="text-xs text-[#7C7288]">
                  オフにすると作成者のみ閲覧できます。
                </div>
              </div>
              <Switch
                checked={isPublic}
                onCheckedChange={setIsPublic}
                aria-label="公開する"
                className="h-9 w-16 border border-transparent data-[state=checked]:bg-[#FF5C5C] data-[state=unchecked]:bg-[#E4D3C7] [&_[data-slot=switch-thumb]]:size-7 [&_[data-slot=switch-thumb]]:bg-white [&_[data-slot=switch-thumb]]:shadow [&_[data-slot=switch-thumb]]:data-[state=checked]:translate-x-8 [&_[data-slot=switch-thumb]]:data-[state=unchecked]:translate-x-1"
              />
            </div>

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
              {isEditMode && (
                <button
                  type="button"
                  onClick={onClose}
                  className="rounded-full border border-gray-200 bg-white px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50 mr-3"
                >
                  キャンセル
                </button>
              )}
              <button
                type="submit"
                disabled={isSubmitting || (isEditMode && !canSubmit)}
                className="inline-flex items-center justify-center rounded-full bg-[#FF5C5C] px-8 py-3 text-sm font-semibold text-white shadow-[0_8px_0_#D44242] hover:translate-y-0.5 hover:shadow-[0_6px_0_#D44242] active:translate-y-1 active:shadow-[0_3px_0_#D44242] transition-transform disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-[0_8px_0_#D44242]"
              >
                {isEditMode && <Pencil className="w-4 h-4 mr-2" />}
                {isSubmitting ? "処理中..." : submitLabel}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Modal>
  );
};
