"use client";

import { useCallback, useMemo, useState } from "react";
import { X, Search, Plus, Check, Film } from "lucide-react";
import Image from "next/image";
import { useAuth } from "@clerk/nextjs";
import { useMutation, useQuery } from "@tanstack/react-query";
import { searchMovies, type MovieSearchItem } from "@/lib/api/movies/search";
import { getMovieDetail } from "@/lib/api/movies/detail";
import { addMovieToTag } from "@/lib/api/tags/addMovie";
import { getBackendTokenOrThrow } from "@/lib/api/_shared/auth";
import { Modal } from "@/components/Modal";
import { Spinner } from "@/components/ui/spinner";

type SelectedMovie = {
  tmdb_movie_id: number;
  title: string;
  poster_path?: string | null;
};

type MovieAddModalProps = {
  open: boolean;
  tagId: string;
  onClose: () => void;
  onAdded: () => void;
};

/** 検索結果の各行に監督名を遅延取得して表示する */
function MovieDirectorLabel({ tmdbMovieId }: { tmdbMovieId: number }) {
  const { data, isLoading } = useQuery({
    queryKey: ["movieDetail", tmdbMovieId],
    queryFn: () => getMovieDetail(tmdbMovieId),
    staleTime: 1000 * 60 * 60,
  });

  if (isLoading) {
    return <span className="text-xs text-[#B09EA0]">...</span>;
  }
  if (!data?.directors?.length) return null;

  return (
    <span className="text-xs text-[#7C7288] truncate">
      {data.directors.join(", ")}
    </span>
  );
}

function PosterThumbnail({
  posterPath,
  title,
  size = "md",
}: {
  posterPath?: string | null;
  title: string;
  size?: "sm" | "md";
}) {
  const dimensions =
    size === "sm"
      ? { w: 28, h: 42, cls: "w-7 h-[42px]" }
      : { w: 48, h: 72, cls: "w-12 h-[72px]" };

  if (!posterPath) {
    return (
      <div
        className={`${dimensions.cls} rounded-lg bg-[#F3E1D6] flex items-center justify-center flex-shrink-0`}
      >
        <Film className="w-4 h-4 text-[#B09EA0]" />
      </div>
    );
  }

  return (
    <Image
      src={`https://image.tmdb.org/t/p/w200${posterPath}`}
      alt={title}
      width={dimensions.w}
      height={dimensions.h}
      className={`${dimensions.cls} rounded-lg object-cover flex-shrink-0`}
    />
  );
}

export const MovieAddModal = ({
  open,
  tagId,
  onClose,
  onAdded,
}: MovieAddModalProps) => {
  const { getToken } = useAuth();
  const [q, setQ] = useState("");
  const [selectedMovies, setSelectedMovies] = useState<SelectedMovie[]>([]);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [addProgress, setAddProgress] = useState<{
    done: number;
    total: number;
  } | null>(null);

  const trimmedQ = useMemo(() => q.trim(), [q]);

  const searchQuery = useQuery({
    queryKey: ["movieSearch", trimmedQ],
    queryFn: () => searchMovies({ q: trimmedQ, page: 1 }),
    enabled: open && trimmedQ.length >= 2,
  });

  const isSelected = useCallback(
    (tmdbMovieId: number) =>
      selectedMovies.some((m) => m.tmdb_movie_id === tmdbMovieId),
    [selectedMovies],
  );

  const toggleMovie = useCallback((movie: MovieSearchItem) => {
    setSelectedMovies((prev) => {
      const exists = prev.some(
        (m) => m.tmdb_movie_id === movie.tmdb_movie_id,
      );
      if (exists) {
        return prev.filter((m) => m.tmdb_movie_id !== movie.tmdb_movie_id);
      }
      return [
        ...prev,
        {
          tmdb_movie_id: movie.tmdb_movie_id,
          title: movie.title,
          poster_path: movie.poster_path,
        },
      ];
    });
  }, []);

  const removeMovie = useCallback((tmdbMovieId: number) => {
    setSelectedMovies((prev) =>
      prev.filter((m) => m.tmdb_movie_id !== tmdbMovieId),
    );
  }, []);

  const addMutation = useMutation({
    mutationFn: async () => {
      const token = await getBackendTokenOrThrow(getToken);
      if (selectedMovies.length === 0) {
        throw new Error("追加する映画を選択してください。");
      }

      const errors: string[] = [];
      setAddProgress({ done: 0, total: selectedMovies.length });

      for (let i = 0; i < selectedMovies.length; i++) {
        try {
          await addMovieToTag({
            tagId,
            token,
            input: {
              tmdb_movie_id: selectedMovies[i].tmdb_movie_id,
              position: 0,
            },
          });
        } catch (err) {
          const msg = err instanceof Error ? err.message : "不明なエラー";
          errors.push(`${selectedMovies[i].title}: ${msg}`);
        }
        setAddProgress({ done: i + 1, total: selectedMovies.length });
      }

      if (errors.length > 0) {
        throw new Error(errors.join("\n"));
      }
    },
    onSuccess: () => {
      setErrorMessage(null);
      setSelectedMovies([]);
      setQ("");
      setAddProgress(null);
      onAdded();
      onClose();
    },
    onError: (err) => {
      setAddProgress(null);
      setErrorMessage(
        err instanceof Error ? err.message : "追加に失敗しました。",
      );
    },
  });

  const items = searchQuery.data?.items ?? [];

  return (
    <Modal open={open} onClose={onClose}>
      <div className="w-full max-w-2xl mx-auto rounded-3xl bg-[#FFF9F3] shadow-xl border border-[#F3E1D6] max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-start justify-between px-7 pt-7 flex-shrink-0">
          <div>
            <h2 className="text-xl md:text-2xl font-extrabold text-[#1F1A2B] tracking-tight">
              映画を追加
            </h2>
            <p className="mt-1 text-sm text-[#7C7288]">
              タイトルで検索して、追加したい映画を選択してください。
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
        <div className="px-7 pb-7 pt-4 space-y-4 overflow-y-auto flex-1 min-h-0">
          {/* Search input */}
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Search className="h-5 w-5 text-[#B09EA0]" />
            </div>
            <input
              type="text"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="映画タイトルで検索（2文字以上）"
              className="block w-full pl-12 pr-4 py-3 rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] text-[#1F1A2B] text-sm focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8] shadow-[0_1px_0_rgba(0,0,0,0.03)]"
            />
          </div>

          {/* Selected movies chips */}
          {selectedMovies.length > 0 && (
            <div className="space-y-2">
              <div className="text-xs font-semibold text-[#7C7288] tracking-wide">
                選択中（{selectedMovies.length}件）
              </div>
              <div className="flex flex-wrap gap-2">
                {selectedMovies.map((movie) => (
                  <div
                    key={movie.tmdb_movie_id}
                    className="inline-flex items-center gap-2 rounded-full bg-[#FF5C5C]/10 border border-[#FF5C5C]/20 pl-1.5 pr-2 py-1"
                  >
                    <PosterThumbnail
                      posterPath={movie.poster_path}
                      title={movie.title}
                      size="sm"
                    />
                    <span className="text-xs font-medium text-[#1F1A2B] max-w-[120px] truncate">
                      {movie.title}
                    </span>
                    <button
                      type="button"
                      onClick={() => removeMovie(movie.tmdb_movie_id)}
                      className="inline-flex h-5 w-5 items-center justify-center rounded-full text-[#FF5C5C] hover:bg-[#FF5C5C]/20 transition-colors"
                      aria-label={`${movie.title}の選択を解除`}
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Search results */}
          <div className="min-h-48 rounded-2xl border border-[#E4D3C7] bg-[#FFFDF8] p-2">
            {searchQuery.isLoading && (
              <div className="flex items-center justify-center gap-2 py-8 text-sm text-[#7C7288]">
                <Spinner size="sm" />
                検索中...
              </div>
            )}
            {searchQuery.isError && (
              <div className="text-sm text-red-600 px-3 py-4">
                {(searchQuery.error as Error | null)?.message ??
                  "検索に失敗しました"}
              </div>
            )}
            {!searchQuery.isLoading &&
              trimmedQ.length >= 2 &&
              items.length === 0 && (
                <div className="text-sm text-[#7C7288] text-center py-8">
                  候補がありません
                </div>
              )}
            {!searchQuery.isLoading &&
              trimmedQ.length < 2 &&
              items.length === 0 && (
                <div className="flex flex-col items-center justify-center py-8 text-[#B09EA0]">
                  <Search className="w-8 h-8 mb-2" />
                  <span className="text-sm">
                    映画タイトルを入力してください
                  </span>
                </div>
              )}

            {items.length > 0 && (
              <div className="max-h-80 overflow-auto space-y-1">
                {items.slice(0, 20).map((it) => {
                  const checked = isSelected(it.tmdb_movie_id);
                  const year = it.release_date
                    ? it.release_date.slice(0, 4)
                    : "";
                  return (
                    <button
                      key={it.tmdb_movie_id}
                      type="button"
                      onClick={() => toggleMovie(it)}
                      className={[
                        "w-full text-left px-3 py-2.5 rounded-xl flex items-center gap-3 transition-colors",
                        checked
                          ? "bg-[#FF5C5C]/8 ring-1 ring-[#FF5C5C]/30"
                          : "hover:bg-[#FDF1E7]",
                      ].join(" ")}
                    >
                      {/* Poster with check badge */}
                      <div className="relative flex-shrink-0">
                        <PosterThumbnail
                          posterPath={it.poster_path}
                          title={it.title}
                        />
                        {checked && (
                          <div className="absolute -top-1.5 -right-1.5 w-5 h-5 rounded-full bg-[#FF5C5C] flex items-center justify-center shadow-sm">
                            <Check className="w-3 h-3 text-white" />
                          </div>
                        )}
                      </div>

                      {/* Movie info */}
                      <div className="min-w-0 flex-1">
                        <div className="font-semibold text-sm text-[#1F1A2B] truncate">
                          {it.title}
                        </div>
                        {it.original_title &&
                          it.original_title !== it.title && (
                            <div className="text-xs text-[#B09EA0] truncate">
                              {it.original_title}
                            </div>
                          )}
                        <div className="flex items-center gap-3 mt-0.5">
                          {year && (
                            <span className="text-xs text-[#7C7288]">
                              {year}年
                            </span>
                          )}
                          <MovieDirectorLabel
                            tmdbMovieId={it.tmdb_movie_id}
                          />
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          {/* Error */}
          {errorMessage && (
            <div className="text-sm text-red-600 font-medium whitespace-pre-line">
              {errorMessage}
            </div>
          )}

          {/* Progress */}
          {addProgress && (
            <div className="flex items-center gap-2 text-sm text-[#7C7288]">
              <Spinner size="sm" />
              追加中... ({addProgress.done}/{addProgress.total})
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="rounded-full border border-[#E4D3C7] bg-white px-6 py-3 text-sm font-semibold text-[#7C7288] hover:bg-[#FDF1E7] hover:text-[#1F1A2B] transition-colors"
            >
              キャンセル
            </button>
            <button
              type="button"
              disabled={addMutation.isPending || selectedMovies.length === 0}
              onClick={() => addMutation.mutate()}
              className="inline-flex items-center justify-center rounded-full bg-[#FF5C5C] px-7 py-3 text-sm font-semibold text-white shadow-[0_8px_0_#D44242] hover:translate-y-0.5 hover:shadow-[0_6px_0_#D44242] active:translate-y-1 active:shadow-[0_3px_0_#D44242] transition-transform disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-[0_8px_0_#D44242]"
            >
              {addMutation.isPending ? (
                <Spinner size="sm" className="mr-2" />
              ) : (
                <Plus className="w-4 h-4 mr-2" />
              )}
              {addMutation.isPending
                ? "追加中..."
                : selectedMovies.length > 0
                  ? `追加する（${selectedMovies.length}件）`
                  : "追加する"}
            </button>
          </div>
        </div>
      </div>
    </Modal>
  );
};
