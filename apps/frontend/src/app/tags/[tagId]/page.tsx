"use client";

import { use, useState } from "react";
import dynamic from "next/dynamic";
import { MoviePosterCard } from "@/components/MoviePosterCard";
import { AvatarCircle } from "@/components/AvatarCircle";
import { getTagDetail } from "@/lib/api/tags/detail";
import { listTagMovies } from "@/lib/api/tags/movies";
import { deleteMovieFromTag } from "@/lib/api/tags/deleteMovie";
import { followTag } from "@/lib/api/tags/follow";
import { unfollowTag } from "@/lib/api/tags/unfollow";
import { getTagFollowStatus } from "@/lib/api/tags/getFollowStatus";
import { Search, Plus, Pencil, Heart, Film } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth, useUser } from "@clerk/nextjs";
import { Spinner } from "@/components/ui/spinner";

// 動的インポート: モーダルは初期表示に不要なため遅延ロード
const MovieAddModal = dynamic(
  () => import("@/components/MovieAddModal").then((mod) => mod.MovieAddModal),
  { ssr: false },
);
const TagModal = dynamic(
  () => import("@/components/TagModal").then((mod) => mod.TagModal),
  { ssr: false },
);
const TagFollowersModal = dynamic(
  () =>
    import("@/components/TagFollowersModal").then(
      (mod) => mod.TagFollowersModal,
    ),
  { ssr: false },
);

export default function TagDetailPage({
  params,
}: {
  params: Promise<{ tagId: string }>;
}) {
  const { tagId } = use(params);
  const [query, setQuery] = useState("");
  const [addOpen, setAddOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [followersOpen, setFollowersOpen] = useState(false);
  const { getToken } = useAuth();
  const { isSignedIn } = useUser();
  const queryClient = useQueryClient();

  const detailQuery = useQuery({
    queryKey: ["tagDetail", tagId],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" }).catch(
        () => null,
      );
      return await getTagDetail(tagId, { token: token ?? undefined });
    },
  });

  const moviesQuery = useQuery({
    queryKey: ["tagMovies", tagId],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" }).catch(
        () => null,
      );
      return await listTagMovies(tagId, { token: token ?? undefined });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (tagMovieId: string) => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証が必要です");
      return await deleteMovieFromTag({ tagId, tagMovieId, token });
    },
    onSuccess: () => {
      Promise.all([detailQuery.refetch(), moviesQuery.refetch()]);
    },
  });

  // フォロー状態取得
  const followStatusQuery = useQuery({
    queryKey: ["tagFollowStatus", tagId],
    queryFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) return { isFollowing: false };
      return await getTagFollowStatus(tagId, token);
    },
    enabled: isSignedIn === true,
  });

  // フォロー/アンフォローミューテーション
  const followMutation = useMutation({
    mutationFn: async () => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証が必要です");
      const isFollowing = followStatusQuery.data?.isFollowing ?? false;
      if (isFollowing) {
        await unfollowTag(tagId, token);
      } else {
        await followTag(tagId, token);
      }
    },
    onSuccess: () => {
      Promise.all([
        queryClient.invalidateQueries({ queryKey: ["tagFollowStatus", tagId] }),
        detailQuery.refetch(),
      ]);
    },
  });

  const detail = detailQuery.data ?? null;
  const isFollowing = followStatusQuery.data?.isFollowing ?? false;
  const movies = moviesQuery.data ?? [];
  const canEditTag = detail?.canEdit ?? false;
  const canAddMovie = detail?.canAddMovie ?? false;

  const filtered = (() => {
    const q = query.trim().toLowerCase();
    if (!q) return movies;
    return movies.filter((m) =>
      `${m.title} ${m.year}`.toLowerCase().includes(q),
    );
  })();

  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-10">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
          {/* Left: Tag info card */}
          <aside className="lg:col-span-4">
            <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-7">
              <h1 className="text-2xl md:text-3xl font-extrabold text-gray-900 tracking-tight">
                {detail?.title ?? "読み込み中..."}
              </h1>
              <p className="mt-3 text-sm md:text-base text-gray-600 leading-relaxed">
                {detail?.description}
              </p>

              {/* Stats: フォロー数 & 映画数 */}
              <div className="mt-5 flex items-center gap-4">
                <div className="flex items-center gap-1.5 text-sm text-gray-600">
                  <Heart className="w-4 h-4 text-pink-500" />
                  <span className="font-bold text-gray-900">{detail?.followerCount ?? 0}</span>
                  <span>フォロー</span>
                </div>
                <div className="flex items-center gap-1.5 text-sm text-gray-600">
                  <Film className="w-4 h-4 text-blue-500" />
                  <span className="font-bold text-gray-900">{detail?.movieCount ?? 0}</span>
                  <span>本の映画</span>
                </div>
              </div>

              {/* Author */}
              <div className="mt-6 flex items-center gap-3">
                <AvatarCircle
                  name={detail?.author?.name ?? "author"}
                  avatarUrl={detail?.owner?.avatarUrl}
                  displayId={detail?.owner?.displayId}
                  className="h-10 w-10"
                />
                <div>
                  <div className="text-xs text-gray-500 font-medium">
                    作成者
                  </div>
                  <div className="text-sm font-bold text-gray-900">
                    {detail?.owner?.displayId ? (
                      <a
                        href={`/${detail.owner.displayId}`}
                        className="hover:underline"
                      >
                        {detail?.author?.name ?? "-"}
                      </a>
                    ) : (
                      (detail?.author?.name ?? "-")
                    )}
                  </div>
                </div>
              </div>

              {/* 参加者 */}
              <div className="mt-6">
                <button
                  type="button"
                  onClick={() => setFollowersOpen(true)}
                  className="w-full text-left hover:bg-gray-50 rounded-xl p-2 -m-2 transition-colors"
                >
                  <div className="text-xs text-gray-500 font-semibold">
                    {detail?.participantCount ?? 0}人の参加者
                  </div>
                  <div className="mt-3 flex items-center">
                    {(detail?.participants ?? []).slice(0, 4).map((p, idx) => (
                      <div
                        key={`${p.name}-${idx}`}
                        className={idx === 0 ? "" : "-ml-2"}
                      >
                        <AvatarCircle
                          name={p.name}
                          avatarUrl={p.avatarUrl}
                          className="h-9 w-9"
                        />
                      </div>
                    ))}
                    {detail && detail.participantCount > 4 && (
                      <div className="-ml-2">
                        <div className="h-9 w-9 rounded-full bg-pink-100 border border-pink-200 flex items-center justify-center text-xs font-bold text-pink-600">
                          +{detail.participantCount - 4}
                        </div>
                      </div>
                    )}
                  </div>
                </button>
              </div>

              {/* Actions */}
              <div className="mt-7 space-y-3">
                {/* フォローボタン（ログインユーザーのみ表示、自分が作成者でない場合） */}
                {isSignedIn && !canEditTag && (
                  <button
                    type="button"
                    disabled={followMutation.isPending}
                    className={`w-full font-bold py-3 rounded-full flex items-center justify-center gap-2 shadow-sm hover:shadow transition-all ${
                      isFollowing
                        ? "bg-pink-100 text-pink-600 border border-pink-300 hover:bg-pink-200"
                        : "bg-white text-gray-700 border border-gray-300 hover:bg-gray-50"
                    }`}
                    onClick={() => followMutation.mutate()}
                  >
                    <Heart
                      className={`w-5 h-5 ${isFollowing ? "fill-current" : ""}`}
                    />
                    {followMutation.isPending ? (
                      <span className="flex items-center gap-2">
                        <Spinner size="sm" />
                        処理中
                      </span>
                    ) : isFollowing ? (
                      "フォロー中"
                    ) : (
                      "フォローする"
                    )}
                  </button>
                )}
                {canAddMovie ? (
                  <button
                    type="button"
                    className="w-full bg-[#FF5C5C] hover:bg-[#ff4a4a] text-white font-bold py-3 rounded-full flex items-center justify-center gap-2 shadow-sm hover:shadow transition-all"
                    onClick={() => setAddOpen(true)}
                  >
                    <Plus className="w-5 h-5" />
                    映画を追加する
                  </button>
                ) : null}
                {canEditTag ? (
                  <button
                    type="button"
                    className="w-full bg-gray-900 hover:bg-black text-white font-bold py-3 rounded-full flex items-center justify-center gap-2 border border-gray-200 shadow-sm hover:shadow transition-all"
                    onClick={() => setEditOpen(true)}
                  >
                    <Pencil className="w-4 h-4" />
                    タグを編集
                  </button>
                ) : null}
              </div>
            </div>
          </aside>

          {/* Right: Movies */}
          <section className="lg:col-span-8">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
              <div className="text-xl md:text-2xl font-extrabold text-gray-900">
                {movies.length}本の映画
              </div>
              <div className="relative w-full sm:max-w-md">
                <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                  <Search className="h-5 w-5 text-gray-400" />
                </div>
                <input
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="このタグ内の映画を検索..."
                  className="block w-full pl-12 pr-4 py-3.5 rounded-full border border-gray-900 bg-white text-gray-900 focus:ring-2 focus:ring-blue-500 focus:border-transparent shadow-sm"
                />
              </div>
            </div>

            {(detailQuery.isError || moviesQuery.isError) && (
              <div className="mb-6 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                {(detailQuery.error as Error | null)?.message ??
                  (moviesQuery.error as Error | null)?.message ??
                  "読み込みに失敗しました"}
              </div>
            )}

            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
              {filtered.map((m) => (
                <MoviePosterCard
                  key={m.id}
                  title={m.title}
                  year={m.year}
                  posterUrl={m.posterUrl}
                  onDelete={
                    m.canDelete
                      ? () => {
                          if (
                            confirm(
                              `「${m.title}」をこのタグから削除しますか？`,
                            )
                          ) {
                            deleteMutation.mutate(m.id);
                          }
                        }
                      : undefined
                  }
                  isDeleting={deleteMutation.isPending}
                />
              ))}
            </div>

            {(detailQuery.isLoading || moviesQuery.isLoading) && (
              <div className="mt-10 flex justify-center">
                <Spinner size="md" className="text-gray-600" />
              </div>
            )}

            {!moviesQuery.isLoading && filtered.length === 0 && (
              <div className="mt-10 text-center text-gray-600">
                該当する映画がありません
              </div>
            )}
          </section>
        </div>
      </main>

      <MovieAddModal
        open={addOpen}
        tagId={tagId}
        onClose={() => setAddOpen(false)}
        onAdded={() => {
          Promise.all([detailQuery.refetch(), moviesQuery.refetch()]);
        }}
      />

      {editOpen && canEditTag ? (
        <TagModal
          key={tagId}
          open={true}
          tag={{
            id: tagId,
            title: detail?.title ?? "",
            description: detail?.description ?? "",
            is_public: true,
            add_movie_policy: detail?.addMoviePolicy ?? "everyone",
          }}
          onClose={() => setEditOpen(false)}
          onUpdated={() => {
            detailQuery.refetch();
          }}
        />
      ) : null}

      <TagFollowersModal
        open={followersOpen}
        tagId={tagId}
        tagTitle={detail?.title ?? ""}
        onClose={() => setFollowersOpen(false)}
      />
    </div>
  );
}
