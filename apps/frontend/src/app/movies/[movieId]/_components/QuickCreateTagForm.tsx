"use client";

import { useState, type FormEvent } from "react";
import { Plus, Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import { useMutation } from "@tanstack/react-query";
import { createTag } from "@/lib/api/tags/create";
import { addMoviesToTag } from "@/lib/api/tags/addMovie";
import { getBackendTokenOrThrow } from "@/lib/api/_shared/auth";
import {
  TagCreateInputSchema,
  getFirstZodErrorMessage,
} from "@/lib/validation/tag.form";

type QuickCreateTagFormProps = {
  tmdbMovieId: number;
  onCreatedAndAdded: (tagId: string) => void;
};

export function QuickCreateTagForm({
  tmdbMovieId,
  onCreatedAndAdded,
}: QuickCreateTagFormProps) {
  const { getToken } = useAuth();
  const router = useRouter();
  const [title, setTitle] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async (tagTitle: string) => {
      const token = await getBackendTokenOrThrow(getToken);

      const created = await createTag({
        token,
        input: {
          title: tagTitle,
          is_public: false,
          add_movie_policy: "everyone",
        },
      });

      await addMoviesToTag({
        tagId: created.id,
        token,
        movies: [{ tmdb_movie_id: tmdbMovieId }],
      });

      return created;
    },
    onSuccess: (created) => {
      setTitle("");
      setErrorMessage(null);
      onCreatedAndAdded(created.id);
      router.push(`/tags/${created.id}`);
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "作成に失敗しました",
      );
    },
  });

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setErrorMessage(null);

    const parsed = TagCreateInputSchema.safeParse({
      title: title,
    });
    if (!parsed.success) {
      setErrorMessage(getFirstZodErrorMessage(parsed.error));
      return;
    }

    mutation.mutate(parsed.data.title);
  };

  return (
    <div className="flex flex-col gap-2">
      <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
        新しいタグを作成して追加
      </label>
      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="タグ名を入力..."
          className="flex-1 rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] px-4 py-2.5 text-sm text-[#1F1A2B] focus:outline-none focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8]"
        />
        <button
          type="submit"
          disabled={title.trim().length === 0 || mutation.isPending}
          className="shrink-0 inline-flex items-center gap-1.5 rounded-xl bg-[#1F1A2B] px-4 py-2.5 text-sm font-semibold text-white hover:bg-[#2D2640] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {mutation.isPending ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <Plus className="w-4 h-4" />
          )}
          作成して追加
        </button>
      </form>
      <p className="text-xs text-[#C2B5A8]">
        詳細設定はタグページから変更できます
      </p>
      {errorMessage && (
        <p className="text-sm text-red-600 font-medium">{errorMessage}</p>
      )}
    </div>
  );
}
