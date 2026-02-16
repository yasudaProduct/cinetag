"use client";

import { use } from "react";
import Image from "next/image";
import Link from "next/link";
import {
  Star,
  Calendar,
  Clock,
  Globe,
  Plus,
  Tag,
  Film,
  Heart,
} from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { getMovieDetail } from "@/lib/api/movies/detail";
import { getMovieRelatedTags } from "@/lib/api/movies/tags";
import { Spinner } from "@/components/ui/spinner";

export default function MovieDetailPage({
  params,
}: {
  params: Promise<{ movieId: string }>;
}) {
  const { movieId } = use(params);
  const tmdbMovieId = Number(movieId);

  const movieQuery = useQuery({
    queryKey: ["movieDetail", tmdbMovieId],
    queryFn: () => getMovieDetail(tmdbMovieId),
    enabled: !Number.isNaN(tmdbMovieId) && tmdbMovieId > 0,
  });

  const tagsQuery = useQuery({
    queryKey: ["movieRelatedTags", tmdbMovieId],
    queryFn: () => getMovieRelatedTags(tmdbMovieId),
    enabled: !Number.isNaN(tmdbMovieId) && tmdbMovieId > 0,
  });

  const movie = movieQuery.data;
  const relatedTags = tagsQuery.data ?? [];

  if (movieQuery.isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Spinner size="md" className="text-gray-600" />
      </div>
    );
  }

  if (movieQuery.isError || !movie) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 text-lg font-bold">
            {(movieQuery.error as Error | null)?.message ??
              "映画情報の取得に失敗しました"}
          </p>
        </div>
      </div>
    );
  }

  const posterUrl = movie.posterPath
    ? `https://image.tmdb.org/t/p/w500${movie.posterPath}`
    : undefined;

  const releaseYear = movie.releaseDate
    ? new Date(movie.releaseDate).getFullYear()
    : undefined;

  const country = movie.productionCountries[0]?.name;
  const directorText = movie.directors.join(", ");

  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-10">
        {/* メインカード */}
        <div className="bg-white rounded-[20px] border border-gray-200 shadow-sm overflow-hidden p-6 md:p-8">
          <div className="flex flex-col md:flex-row gap-8">
            {/* 左: ポスター画像 */}
            <div className="shrink-0 mx-auto md:mx-0">
              <div className="relative w-[220px] md:w-[300px] aspect-[2/3] rounded-2xl overflow-hidden border border-gray-200 shadow-lg bg-gray-100">
                {posterUrl ? (
                  <Image
                    src={posterUrl}
                    alt={`${movie.title} poster`}
                    fill
                    className="object-cover"
                    sizes="(max-width: 768px) 220px, 300px"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400">
                    <Film className="w-16 h-16" />
                  </div>
                )}
              </div>
            </div>

            {/* 右: 映画情報 */}
            <div className="flex-1 min-w-0">
              {/* タイトル */}
              <h1 className="text-2xl md:text-4xl font-bold text-gray-900 tracking-tight">
                {movie.title}
              </h1>
              {movie.originalTitle && (
                <p className="mt-1 text-base md:text-lg text-gray-500">
                  {movie.originalTitle}
                </p>
              )}

              {/* 評価バッジ */}
              {movie.voteAverage != null && (
                <div className="mt-4 flex items-center gap-3 flex-wrap">
                  <div className="flex items-center gap-2 bg-[#FFD75E] rounded-xl px-4 py-2">
                    <Star className="w-5 h-5 text-gray-900 fill-current" />
                    <span className="text-xl font-bold text-gray-900">
                      {movie.voteAverage.toFixed(1)}
                    </span>
                  </div>
                </div>
              )}

              {/* メタ情報 */}
              <div className="mt-5 grid grid-cols-1 sm:grid-cols-2 gap-x-12 gap-y-2">
                {releaseYear && (
                  <div className="flex items-center gap-2 text-sm text-gray-500">
                    <Calendar className="w-4 h-4 shrink-0" />
                    <span>{releaseYear}年</span>
                  </div>
                )}
                {movie.runtime != null && (
                  <div className="flex items-center gap-2 text-sm text-gray-500">
                    <Clock className="w-4 h-4 shrink-0" />
                    <span>{movie.runtime}分</span>
                  </div>
                )}
                {country && (
                  <div className="flex items-center gap-2 text-sm text-gray-500">
                    <Globe className="w-4 h-4 shrink-0" />
                    <span>{country}</span>
                  </div>
                )}
                {directorText && (
                  <div className="flex items-center gap-2 text-sm">
                    <span className="font-bold text-gray-900">監督:</span>
                    <span className="text-gray-500">{directorText}</span>
                  </div>
                )}
              </div>

              {/* ジャンル */}
              {movie.genres.length > 0 && (
                <div className="mt-4 flex flex-wrap gap-2">
                  {movie.genres.map((genre) => (
                    <span
                      key={genre.id}
                      className="px-3 py-1 text-sm font-medium text-gray-600 bg-gray-100 rounded-full"
                    >
                      {genre.name}
                    </span>
                  ))}
                </div>
              )}

              {/* あらすじ */}
              {movie.overview && (
                <p className="mt-5 text-base text-gray-600 leading-relaxed">
                  {movie.overview}
                </p>
              )}

              {/* キャスト */}
              {movie.cast.length > 0 && (
                <div className="mt-5">
                  <h2 className="text-base font-bold text-gray-900">
                    キャスト
                  </h2>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {movie.cast.map((member) => (
                      <span
                        key={member.name}
                        className="px-3 py-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg"
                      >
                        {member.name}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* この映画が含まれるタグ */}
        <section className="mt-10">
          <h2 className="text-xl font-bold text-gray-900 mb-4 px-6">
            この映画が含まれるタグ
          </h2>

          {tagsQuery.isLoading && (
            <div className="flex justify-center py-8">
              <Spinner size="sm" className="text-gray-600" />
            </div>
          )}

          {relatedTags.length > 0 && (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 px-6">
              {relatedTags.map((tag) => (
                <Link
                  key={tag.tagId}
                  href={`/tags/${tag.tagId}`}
                  className="bg-white border border-gray-200 rounded-xl p-5 hover:shadow-md transition-shadow"
                >
                  {/* タグアイコン + タイトル */}
                  <div className="flex items-start gap-3">
                    <div className="shrink-0 w-9 h-9 bg-[rgba(255,215,94,0.2)] rounded-lg flex items-center justify-center">
                      <Tag className="w-5 h-5 text-[#FFD75E]" />
                    </div>
                    <h3 className="text-[15px] font-bold text-gray-900 line-clamp-1 pt-1">
                      {tag.title}
                    </h3>
                  </div>

                  {/* 統計 */}
                  <div className="mt-4 flex items-center gap-4 text-xs text-gray-500">
                    <div className="flex items-center gap-1">
                      <Heart className="w-4 h-4" />
                      <span>{tag.followerCount}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Film className="w-4 h-4" />
                      <span>{tag.movieCount}作品</span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          )}

          {!tagsQuery.isLoading && relatedTags.length === 0 && (
            <div className="text-center text-gray-500 py-8 px-6">
              この映画が含まれるタグはまだありません
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
