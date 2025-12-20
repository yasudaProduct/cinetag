"use client";

import { useMemo, useState } from "react";
import { X, Search, Plus } from "lucide-react";
import { useAuth } from "@clerk/nextjs";
import { useMutation, useQuery } from "@tanstack/react-query";
import { searchMovies } from "@/lib/api/movies/search";
import { addMovieToTag } from "@/lib/api/tags/addMovie";

type MovieAddModalProps = {
  open: boolean;
  tagId: string;
  onClose: () => void;
  onAdded: () => void;
};

export const MovieAddModal = ({
  open,
  tagId,
  onClose,
  onAdded,
}: MovieAddModalProps) => {
  const { getToken } = useAuth();
  const [q, setQ] = useState("");
  const [selected, setSelected] = useState<null | {
    tmdb_movie_id: number;
    title: string;
  }>(null);
  const [note, setNote] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const trimmedQ = useMemo(() => q.trim(), [q]);

  const searchQuery = useQuery({
    queryKey: ["movieSearch", trimmedQ],
    queryFn: () => searchMovies({ q: trimmedQ, page: 1 }),
    enabled: open && trimmedQ.length >= 2,
  });

  const addMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) {
        throw new Error(
          "認証情報の取得に失敗しました。再ログインしてください。"
        );
      }
      if (!selected) {
        throw new Error("追加する映画を選択してください。");
      }
      return await addMovieToTag({
        tagId,
        token,
        input: {
          tmdb_movie_id: selected.tmdb_movie_id,
          note: note.trim().length > 0 ? note.trim() : undefined,
          position: 0,
        },
      });
    },
    onSuccess: () => {
      setErrorMessage(null);
      setSelected(null);
      setNote("");
      setQ("");
      onAdded();
      onClose();
    },
    onError: (err) => {
      setErrorMessage(
        err instanceof Error ? err.message : "追加に失敗しました。"
      );
    },
  });

  if (!open) return null;

  const items = searchQuery.data?.items ?? [];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="w-full max-w-2xl mx-4 rounded-3xl bg-white shadow-xl border border-gray-200">
        <div className="flex items-start justify-between px-7 pt-7">
          <div>
            <h2 className="text-xl md:text-2xl font-extrabold text-gray-900 tracking-tight">
              映画を追加
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              タイトルで検索して、追加したい映画を選択してください。
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

        <div className="px-7 pb-2 space-y-5">
          <div className="mt-2 text-xs text-gray-500">
            2文字以上で検索します。
          </div>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center justify-center pointer-events-none">
              <Search className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="text"
              value={q}
              onChange={(e) => {
                setQ(e.target.value);
                setSelected(null);
              }}
              placeholder="例: Interstellar"
              className="block w-full pl-12 pr-4 py-3.5 rounded-full border border-gray-300 bg-white text-gray-900 focus:ring-2 focus:ring-blue-500 focus:border-transparent shadow-sm"
            />
          </div>

          {/* {trimmedQ.length >= 2 && ( */}
          <div className="min-h-64 rounded-2xl border border-gray-200 bg-gray-50 p-3">
            {searchQuery.isLoading && (
              <div className="text-sm text-gray-600">検索中...</div>
            )}
            {searchQuery.isError && (
              <div className="text-sm text-red-700">
                {(searchQuery.error as Error | null)?.message ??
                  "検索に失敗しました"}
              </div>
            )}
            {!searchQuery.isLoading &&
              trimmedQ.length >= 2 &&
              items.length === 0 && (
                <div className="text-sm text-gray-600">候補がありません</div>
              )}

            {items.length > 0 ? (
              <div className="max-h-64 overflow-auto divide-y divide-gray-200 bg-white rounded-xl border border-gray-200">
                {items.slice(0, 20).map((it) => {
                  const isSelected =
                    selected?.tmdb_movie_id === it.tmdb_movie_id;
                  const year = it.release_date
                    ? it.release_date.slice(0, 4)
                    : "";
                  return (
                    <button
                      key={it.tmdb_movie_id}
                      type="button"
                      onClick={() =>
                        setSelected({
                          tmdb_movie_id: it.tmdb_movie_id,
                          title: it.title,
                        })
                      }
                      className={[
                        "w-full text-left px-4 py-3 hover:bg-gray-50",
                        isSelected ? "bg-blue-50" : "bg-white",
                      ].join(" ")}
                    >
                      <div className="flex items-center justify-between gap-4">
                        <div className="min-w-0">
                          <div className="font-semibold text-gray-900 truncate">
                            {it.title}{" "}
                            {year ? (
                              <span className="text-gray-500 font-medium">
                                ({year})
                              </span>
                            ) : null}
                          </div>
                          {it.original_title &&
                            it.original_title !== it.title && (
                              <div className="text-xs text-gray-500 truncate">
                                {it.original_title}
                              </div>
                            )}
                        </div>
                        <div className="text-xs text-gray-500">
                          TMDB: {it.tmdb_movie_id}
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            ) : (
              <div className="text-sm text-gray-600 text-center py-4">
                候補がありません
              </div>
            )}
          </div>
          {/* )} */}

          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-gray-600">
              メモ（任意）
            </label>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              rows={3}
              placeholder="この映画を追加した理由など"
              className="w-full rounded-xl border border-gray-300 bg-white px-4 py-3 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
            />
          </div>

          {errorMessage && (
            <div className="text-sm text-red-600 font-medium">
              {errorMessage}
            </div>
          )}

          <div className="flex justify-end gap-3 pt-1 pb-7">
            <button
              type="button"
              onClick={onClose}
              className="rounded-full border border-gray-200 bg-white px-6 py-3 text-sm font-semibold text-gray-700 hover:bg-gray-50"
            >
              キャンセル
            </button>
            <button
              type="button"
              disabled={addMutation.isPending || !selected}
              onClick={() => addMutation.mutate()}
              className="inline-flex items-center justify-center rounded-full bg-[#FF5C5C] px-7 py-3 text-sm font-semibold text-white shadow-sm hover:bg-[#ff4a4a] disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus className="w-4 h-4 mr-2" />
              {addMutation.isPending ? "追加中..." : "追加する"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
