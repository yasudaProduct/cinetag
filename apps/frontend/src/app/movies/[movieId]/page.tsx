"use client";

import { use } from "react";
import Image from "next/image";
import Link from "next/link";
import {
  Star,
  Calendar,
  Clock,
  Globe,
  Heart,
  Plus,
  Tag,
  Film,
} from "lucide-react";
import {
  getMockMovieDetail,
  getMockRelatedTags,
} from "@/lib/mock/movieDetail";

export default function MovieDetailPage({
  params,
}: {
  params: Promise<{ movieId: string }>;
}) {
  const { movieId } = use(params);
  const movie = getMockMovieDetail(movieId);
  const relatedTags = getMockRelatedTags(movieId);

  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-10">
        {/* メインカード */}
        <div className="bg-white rounded-[20px] border border-gray-200 shadow-sm overflow-hidden p-6 md:p-8">
          <div className="flex flex-col md:flex-row gap-8">
            {/* 左: ポスター画像 */}
            <div className="shrink-0 mx-auto md:mx-0">
              <div className="relative w-[220px] md:w-[300px] aspect-[2/3] rounded-2xl overflow-hidden border border-gray-200 shadow-lg">
                <Image
                  src={movie.posterUrl}
                  alt={`${movie.title} poster`}
                  fill
                  className="object-cover"
                  sizes="(max-width: 768px) 220px, 300px"
                />
              </div>
            </div>

            {/* 右: 映画情報 */}
            <div className="flex-1 min-w-0">
              {/* タイトル */}
              <h1 className="text-2xl md:text-4xl font-bold text-gray-900 tracking-tight">
                {movie.title}
              </h1>
              <p className="mt-1 text-base md:text-lg text-gray-500">
                {movie.originalTitle}
              </p>

              {/* 評価 */}
              <div className="mt-4 flex items-center gap-3 flex-wrap">
                <div className="flex items-center gap-2 bg-[#FFD75E] rounded-xl px-4 py-2">
                  <Star className="w-5 h-5 text-gray-900 fill-current" />
                  <span className="text-xl font-bold text-gray-900">
                    {movie.rating}
                  </span>
                </div>
                <div className="flex items-center gap-0.5">
                  {[1, 2, 3, 4, 5].map((i) => (
                    <Star
                      key={i}
                      className="w-6 h-6 text-gray-300 stroke-gray-300"
                    />
                  ))}
                </div>
              </div>

              {/* メタ情報 */}
              <div className="mt-5 grid grid-cols-1 sm:grid-cols-2 gap-x-12 gap-y-2">
                <div className="flex items-center gap-2 text-sm text-gray-500">
                  <Calendar className="w-4 h-4 shrink-0" />
                  <span>{movie.releaseYear}年</span>
                </div>
                <div className="flex items-center gap-2 text-sm text-gray-500">
                  <Clock className="w-4 h-4 shrink-0" />
                  <span>{movie.runtime}分</span>
                </div>
                <div className="flex items-center gap-2 text-sm text-gray-500">
                  <Globe className="w-4 h-4 shrink-0" />
                  <span>{movie.country}</span>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <span className="font-bold text-gray-900">監督:</span>
                  <span className="text-gray-500">{movie.director}</span>
                </div>
              </div>

              {/* ジャンル */}
              <div className="mt-4 flex flex-wrap gap-2">
                {movie.genres.map((genre) => (
                  <span
                    key={genre}
                    className="px-3 py-1 text-sm font-medium text-gray-600 bg-gray-100 rounded-full"
                  >
                    {genre}
                  </span>
                ))}
              </div>

              {/* あらすじ */}
              <p className="mt-5 text-base text-gray-600 leading-relaxed">
                {movie.overview}
              </p>

              {/* キャスト */}
              <div className="mt-5">
                <h2 className="text-base font-bold text-gray-900">キャスト</h2>
                <div className="mt-2 flex flex-wrap gap-2">
                  {movie.cast.map((name) => (
                    <span
                      key={name}
                      className="px-3 py-1.5 text-sm text-gray-600 bg-white border border-gray-200 rounded-lg"
                    >
                      {name}
                    </span>
                  ))}
                </div>
              </div>

              {/* アクションボタン */}
              <div className="mt-6 flex flex-col sm:flex-row gap-3">
                <button
                  type="button"
                  className="flex items-center justify-center gap-2 px-6 py-3 text-sm font-bold text-gray-900 bg-white border border-gray-900 rounded-2xl hover:bg-gray-50 transition-colors"
                >
                  <Heart className="w-5 h-5" />
                  お気に入りに追加
                </button>
                <button
                  type="button"
                  className="flex items-center justify-center gap-2 px-6 py-3 text-sm font-bold text-white bg-[#2b7fff] rounded-2xl hover:bg-blue-600 transition-colors"
                >
                  <Plus className="w-5 h-5" />
                  タグに追加
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* この映画が含まれるタグ */}
        <section className="mt-10">
          <h2 className="text-xl font-bold text-gray-900 mb-4 px-6">
            この映画が含まれるタグ
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 px-6">
            {relatedTags.map((tag) => (
              <Link
                key={tag.id}
                href={`/tags/${tag.id}`}
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
        </section>
      </main>
    </div>
  );
}
