"use client";

import { useState, useMemo, useCallback } from "react";
import { X, Search, Check, Film, Loader2 } from "lucide-react";
import { useAuth } from "@clerk/nextjs";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Modal } from "@/components/Modal";
import { Spinner } from "@/components/ui/spinner";
import { getMe } from "@/lib/api/users/getMe";
import { listUserTags } from "@/lib/api/users/listUserTags";
import { addMoviesToTag } from "@/lib/api/tags/addMovie";
import { getBackendTokenOrThrow } from "@/lib/api/_shared/auth";
import { QuickCreateTagForm } from "@/app/movies/[movieId]/_components/QuickCreateTagForm";

type AddToTagModalProps = {
  open: boolean;
  onClose: () => void;
  tmdbMovieId: number;
  movieTitle: string;
  relatedTagIds: string[];
};

export function AddToTagModal({
  open,
  onClose,
  tmdbMovieId,
  movieTitle,
  relatedTagIds,
}: AddToTagModalProps) {
  const { getToken, isSignedIn, isLoaded } = useAuth();
  const queryClient = useQueryClient();

  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTagIds, setSelectedTagIds] = useState<Set<string>>(new Set());
  const [addedTagIds, setAddedTagIds] = useState<Set<string>>(new Set());
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const meQuery = useQuery({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証情報の取得に失敗しました");
      return getMe(token);
    },
    enabled: isLoaded && isSignedIn,
  });

  const displayId = meQuery.data?.display_id ?? "";

  const userTagsQuery = useQuery({
    queryKey: ["userTags", displayId],
    queryFn: async () => {
      const token = await getBackendTokenOrThrow(getToken);
      return listUserTags({ displayId, token, pageSize: 100 });
    },
    enabled: open && displayId.length > 0,
  });

  const userTags = useMemo(
    () => userTagsQuery.data?.items ?? [],
    [userTagsQuery.data],
  );

  const alreadyAddedIds = useMemo(() => {
    const ids = new Set(relatedTagIds);
    for (const id of addedTagIds) ids.add(id);
    return ids;
  }, [relatedTagIds, addedTagIds]);

  const filteredTags = useMemo(() => {
    if (searchQuery.trim().length === 0) return userTags;
    const q = searchQuery.trim().toLowerCase();
    return userTags.filter((tag) => tag.title.toLowerCase().includes(q));
  }, [userTags, searchQuery]);

  const selectableCount = useMemo(
    () => [...selectedTagIds].filter((id) => !alreadyAddedIds.has(id)).length,
    [selectedTagIds, alreadyAddedIds],
  );

  const toggleTag = useCallback(
    (tagId: string) => {
      if (alreadyAddedIds.has(tagId)) return;
      setSelectedTagIds((prev) => {
        const next = new Set(prev);
        if (next.has(tagId)) next.delete(tagId);
        else next.add(tagId);
        return next;
      });
    },
    [alreadyAddedIds],
  );

  const addMutation = useMutation({
    mutationFn: async (tagIds: string[]) => {
      const token = await getBackendTokenOrThrow(getToken);
      const results = await Promise.allSettled(
        tagIds.map((tagId) =>
          addMoviesToTag({
            tagId,
            token,
            movies: [{ tmdb_movie_id: tmdbMovieId }],
          }),
        ),
      );
      const failed = results.filter((r) => r.status === "rejected");
      if (failed.length > 0 && failed.length === tagIds.length) {
        throw new Error("タグへの追加に失敗しました");
      }
      return { total: tagIds.length, failed: failed.length };
    },
    onSuccess: (_result, tagIds) => {
      setAddedTagIds((prev) => {
        const next = new Set(prev);
        for (const id of tagIds) next.add(id);
        return next;
      });
      setSelectedTagIds(new Set());
      setErrorMessage(null);
      queryClient.invalidateQueries({
        queryKey: ["movieRelatedTags", tmdbMovieId],
      });
      onClose();
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "タグへの追加に失敗しました",
      );
    },
  });

  const handleSubmit = () => {
    const tagIds = [...selectedTagIds].filter((id) => !alreadyAddedIds.has(id));
    if (tagIds.length === 0) return;
    addMutation.mutate(tagIds);
  };

  const handleTagCreatedAndAdded = (tagId: string) => {
    setAddedTagIds((prev) => new Set(prev).add(tagId));
    queryClient.invalidateQueries({ queryKey: ["userTags", displayId] });
    queryClient.invalidateQueries({
      queryKey: ["movieRelatedTags", tmdbMovieId],
    });
  };

  const handleClose = () => {
    setSearchQuery("");
    setSelectedTagIds(new Set());
    setErrorMessage(null);
    onClose();
  };

  const submitLabel =
    selectableCount > 0 ? `${selectableCount}件のタグに追加` : "タグに追加";

  return (
    <Modal open={open} onClose={handleClose}>
      <div className="w-full max-w-lg mx-auto rounded-3xl bg-[#FFF9F3] shadow-xl border border-[#F3E1D6] max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-start justify-between px-6 pt-6 pb-2 flex-shrink-0">
          <div className="min-w-0">
            <h2 className="text-xl font-extrabold text-[#1F1A2B] tracking-tight">
              タグに追加
            </h2>
            <p className="mt-1 text-sm text-[#7C7288] truncate">
              「{movieTitle}」
            </p>
          </div>
          <button
            type="button"
            onClick={handleClose}
            aria-label="閉じる"
            className="ml-4 inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#E4D3C7] bg-white text-[#7C7288] hover:bg-[#FDF1E7] hover:text-[#1F1A2B] transition-colors shrink-0"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Body */}
        <div className="px-6 pb-6 pt-2 flex-1 overflow-hidden flex flex-col gap-5">
          {/* 既存タグ選択セクション */}
          <div className="flex flex-col gap-3">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              あなたのタグから選択
            </label>

            {/* 検索 */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#C2B5A8]" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="タグを検索..."
                className="w-full rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] pl-9 pr-4 py-2.5 text-sm text-[#1F1A2B] focus:outline-none focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8]"
              />
            </div>

            {/* タグ一覧 */}
            <div className="max-h-[240px] overflow-y-auto rounded-xl border border-[#E4D3C7] bg-[#FFFDF8]">
              {open && isSignedIn && meQuery.isLoading && (
                <div className="flex justify-center py-8">
                  <Spinner size="sm" className="text-[#7C7288]" />
                </div>
              )}

              {open && isSignedIn && meQuery.isError && (
                <div className="px-4 py-6 text-center">
                  <p className="text-sm text-red-600">
                    プロフィール情報の取得に失敗しました
                  </p>
                  <button
                    type="button"
                    onClick={() => meQuery.refetch()}
                    className="mt-2 text-sm font-medium text-[#FF5C5C] hover:underline"
                  >
                    再試行
                  </button>
                </div>
              )}

              {open &&
                isSignedIn &&
                displayId.length > 0 &&
                userTagsQuery.isLoading && (
                  <div className="flex justify-center py-8">
                    <Spinner size="sm" className="text-[#7C7288]" />
                  </div>
                )}

              {open &&
                isSignedIn &&
                displayId.length > 0 &&
                userTagsQuery.isError && (
                  <div className="px-4 py-6 text-center">
                    <p className="text-sm text-red-600">
                      タグの取得に失敗しました
                    </p>
                    <button
                      type="button"
                      onClick={() => userTagsQuery.refetch()}
                      className="mt-2 text-sm font-medium text-[#FF5C5C] hover:underline"
                    >
                      再試行
                    </button>
                  </div>
                )}

              {open &&
                isSignedIn &&
                displayId.length > 0 &&
                !meQuery.isLoading &&
                !userTagsQuery.isLoading &&
                !userTagsQuery.isError &&
                userTags.length === 0 && (
                  <p className="px-4 py-6 text-center text-sm text-[#7C7288]">
                    まだタグがありません。下のフォームから作成しましょう
                  </p>
                )}

              {open &&
                isSignedIn &&
                displayId.length > 0 &&
                !userTagsQuery.isLoading &&
                !userTagsQuery.isError &&
                userTags.length > 0 &&
                filteredTags.length === 0 && (
                  <p className="px-4 py-6 text-center text-sm text-[#7C7288]">
                    一致するタグが見つかりません
                  </p>
                )}

              {filteredTags.map((tag) => {
                const isAlreadyAdded = alreadyAddedIds.has(tag.id);
                const isSelected = selectedTagIds.has(tag.id);

                return (
                  <button
                    key={tag.id}
                    type="button"
                    onClick={() => toggleTag(tag.id)}
                    disabled={isAlreadyAdded}
                    className="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-[#FDF1E7] transition-colors border-b border-[#F3E1D6] last:border-b-0 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {/* チェックボックス */}
                    <div
                      className={`shrink-0 w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                        isAlreadyAdded
                          ? "border-[#C2B5A8] bg-[#E4D3C7]"
                          : isSelected
                            ? "border-[#FF5C5C] bg-[#FF5C5C]"
                            : "border-[#E4D3C7] bg-white"
                      }`}
                    >
                      {(isAlreadyAdded || isSelected) && (
                        <Check className="w-3 h-3 text-white" />
                      )}
                    </div>

                    {/* タグ情報 */}
                    <div className="flex-1 min-w-0">
                      <span className="text-sm font-medium text-[#1F1A2B] truncate block">
                        {tag.title}
                      </span>
                    </div>

                    {/* メタ */}
                    <div className="shrink-0 flex items-center gap-1 text-xs text-[#7C7288]">
                      <Film className="w-3.5 h-3.5" />
                      <span>{tag.movieCount}作品</span>
                    </div>

                    {isAlreadyAdded && (
                      <span className="shrink-0 text-xs font-medium text-[#C2B5A8] bg-[#F3E1D6] rounded-full px-2 py-0.5">
                        追加済み
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>

          {/* 区切り */}
          <div className="flex items-center gap-3">
            <div className="flex-1 h-px bg-[#E4D3C7]" />
            <span className="text-xs text-[#C2B5A8] font-medium">または</span>
            <div className="flex-1 h-px bg-[#E4D3C7]" />
          </div>

          {/* 新規タグ作成 */}
          <QuickCreateTagForm
            tmdbMovieId={tmdbMovieId}
            onCreatedAndAdded={handleTagCreatedAndAdded}
          />

          {/* エラーメッセージ */}
          {errorMessage && (
            <p className="text-sm text-red-600 font-medium">{errorMessage}</p>
          )}

          {/* 確定ボタン */}
          <button
            type="button"
            onClick={handleSubmit}
            disabled={selectableCount === 0 || addMutation.isPending}
            className="w-full inline-flex items-center justify-center rounded-full bg-[#FF5C5C] px-6 py-3 text-sm font-semibold text-white shadow-[0_8px_0_#D44242] hover:translate-y-0.5 hover:shadow-[0_6px_0_#D44242] active:translate-y-1 active:shadow-[0_3px_0_#D44242] transition-transform disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-[0_8px_0_#D44242]"
          >
            {addMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                追加中...
              </>
            ) : (
              submitLabel
            )}
          </button>
        </div>
      </div>
    </Modal>
  );
}
