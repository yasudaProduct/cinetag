import Image from "next/image";
import Link from "next/link";
import { Film, Flame, ArrowRight } from "lucide-react";
import type { TagListItem } from "@/lib/validation/tag.api";

const TMDB_IMAGE_BASE = "https://image.tmdb.org/t/p/w500";

function toFullImageUrl(src: string): string {
  return src.startsWith("http") ? src : `${TMDB_IMAGE_BASE}${src}`;
}

export function WeeklyTags({ items }: { items: TagListItem[] }) {
  if (items.length === 0) {
    // データが無い時はセクションごと非表示にして、レイアウトを崩さない。
    return null;
  }

  return (
    <section className="bg-white">
      <div className="max-w-6xl mx-auto px-6 py-12 md:py-16">
        <div className="flex items-end justify-between mb-6 md:mb-8">
          <div>
            <span className="inline-flex items-center gap-1.5 px-3 py-1 bg-orange-100 text-orange-700 text-xs font-bold rounded-full mb-3">
              <Flame className="w-3.5 h-3.5" />
              TRENDING
            </span>
            <h2 className="text-2xl md:text-3xl font-black tracking-tight text-gray-900">
              今週のタグ
            </h2>
            <p className="mt-2 text-sm md:text-base text-gray-500">
              直近1週間で利用が増えたタグ Top {items.length}
            </p>
          </div>
          <Link
            href="/tags"
            className="hidden sm:inline-flex items-center gap-1 text-sm font-semibold text-gray-700 hover:text-gray-900 transition-colors"
          >
            すべて見る
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>

        <ol className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {items.map((tag, index) => (
            <li key={tag.id}>
              <Link
                href={`/tags/${tag.id}`}
                className="group flex items-center gap-4 bg-white border border-gray-200 rounded-2xl px-4 py-3 hover:border-gray-400 hover:shadow-md transition-all"
              >
                <span className="flex-shrink-0 w-9 h-9 rounded-full bg-gradient-to-br from-amber-400 to-orange-500 flex items-center justify-center text-white text-base font-black shadow-sm">
                  {index + 1}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm md:text-base font-bold text-gray-900 truncate group-hover:text-orange-600 transition-colors">
                    {tag.title}
                  </p>
                  <p className="mt-0.5 flex items-center gap-1 text-xs text-gray-500">
                    <Film className="w-3.5 h-3.5" />
                    {tag.movieCount}本
                    <span className="mx-1.5">·</span>
                    by {tag.author}
                  </p>
                </div>
                <div className="hidden sm:flex -space-x-2 flex-shrink-0">
                  {tag.images.slice(0, 3).map((src, i) => (
                    <div
                      key={i}
                      className="relative w-8 h-12 rounded-md overflow-hidden shadow-sm border border-white"
                    >
                      <Image
                        src={toFullImageUrl(src)}
                        alt=""
                        fill
                        className="object-cover"
                        sizes="32px"
                      />
                    </div>
                  ))}
                </div>
                <ArrowRight className="w-4 h-4 text-gray-300 group-hover:text-orange-500 group-hover:translate-x-0.5 transition-all flex-shrink-0" />
              </Link>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
