"use client";

import { useEffect, useMemo, useState } from "react";
import { Header } from "@/components/Header";
import { MoviePosterCard } from "@/components/MoviePosterCard";
import { fetchTagDetail, fetchTagMovies } from "@/lib/api/tag";
import type { TagDetail, TagMovie } from "@/lib/mock/tagDetail";
import { Search, Plus, Pencil } from "lucide-react";

function AvatarCircle({ name, className }: { name: string; className?: string }) {
  const initial = (name?.trim()?.[0] ?? "?").toUpperCase();
  return (
    <div
      className={[
        "flex items-center justify-center rounded-full bg-white border border-gray-200 text-gray-700 font-bold",
        className ?? "",
      ].join(" ")}
      aria-label={name}
      title={name}
    >
      <span className="text-xs">{initial}</span>
    </div>
  );
}

export default function TagDetailPage({ params }: { params: { tagId: string } }) {
  const tagId = params.tagId;
  const [detail, setDetail] = useState<TagDetail | null>(null);
  const [movies, setMovies] = useState<TagMovie[]>([]);
  const [query, setQuery] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      setError(null);

      const [d, m] = await Promise.all([fetchTagDetail(tagId), fetchTagMovies(tagId)]);
      if (cancelled) return;

      if (d.ok) setDetail(d.data);
      else setError(d.error);

      if (m.ok) setMovies(m.data);
      else setError((prev) => prev ?? m.error);
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [tagId]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return movies;
    return movies.filter((m) => `${m.title} ${m.year}`.toLowerCase().includes(q));
  }, [movies, query]);

  return (
    <div className="min-h-screen bg-[#FFF5F5]">
      <Header />

      <main className="container mx-auto px-4 md:px-6 py-10">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
          {/* Left: Tag info card */}
          <aside className="lg:col-span-4">
            <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-7">
              <h1 className="text-2xl md:text-3xl font-extrabold text-gray-900 tracking-tight">
                {detail?.title ?? "タグ"}
              </h1>
              <p className="mt-3 text-sm md:text-base text-gray-600 leading-relaxed">
                {detail?.description ?? "読み込み中..."}
              </p>

              {/* Author */}
              <div className="mt-6 flex items-center gap-3">
                <AvatarCircle name={detail?.author?.name ?? "author"} className="h-10 w-10" />
                <div>
                  <div className="text-xs text-gray-500 font-medium">作成者</div>
                  <div className="text-sm font-bold text-gray-900">
                    {detail?.author?.name ?? "-"}
                  </div>
                </div>
              </div>

              {/* Participants */}
              <div className="mt-6">
                <div className="text-xs text-gray-500 font-semibold">
                  {detail?.participantCount ?? 0}人の参加者
                </div>
                <div className="mt-3 flex items-center">
                  {(detail?.participants ?? []).slice(0, 4).map((p, idx) => (
                    <div key={`${p.name}-${idx}`} className={idx === 0 ? "" : "-ml-2"}>
                      <AvatarCircle name={p.name} className="h-9 w-9" />
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
              </div>

              {/* Actions */}
              <div className="mt-7 space-y-3">
                <button
                  type="button"
                  className="w-full bg-[#FF5C5C] hover:bg-[#ff4a4a] text-white font-bold py-3 rounded-full flex items-center justify-center gap-2 shadow-sm hover:shadow transition-all"
                  onClick={() => alert("映画追加は未実装です（モック画面）。")}
                >
                  <Plus className="w-5 h-5" />
                  映画を追加する
                </button>
                <button
                  type="button"
                  className="w-full bg-gray-100 text-gray-500 font-bold py-3 rounded-full flex items-center justify-center gap-2 border border-gray-200 cursor-not-allowed"
                  disabled
                >
                  <Pencil className="w-4 h-4" />
                  タグを編集
                </button>
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

            {error && (
              <div className="mb-6 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                {error}
              </div>
            )}

            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
              {filtered.map((m) => (
                <MoviePosterCard key={m.id} title={m.title} year={m.year} posterUrl={m.posterUrl} />
              ))}
            </div>

            {filtered.length === 0 && (
              <div className="mt-10 text-center text-gray-600">該当する映画がありません</div>
            )}
          </section>
        </div>
      </main>
    </div>
  );
}


