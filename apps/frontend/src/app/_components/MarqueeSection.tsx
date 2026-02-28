"use client";

import { useQuery } from "@tanstack/react-query";
import { listTags } from "@/lib/api/tags/list";
import { TagCard, type MarqueeTag } from "./TagCard";

const TMDB_IMAGE_BASE = "https://image.tmdb.org/t/p/w500";

const TAG_COLORS = [
  "bg-blue-400",
  "bg-amber-400",
  "bg-green-400",
  "bg-pink-400",
  "bg-indigo-400",
  "bg-teal-400",
  "bg-purple-400",
  "bg-rose-400",
  "bg-emerald-400",
  "bg-orange-400",
  "bg-cyan-400",
  "bg-red-400",
  "bg-yellow-400",
  "bg-slate-400",
  "bg-violet-400",
  "bg-lime-400",
];

const FALLBACK_ROW1: MarqueeTag[] = [
  {
    title: "雨の日に観たい映画",
    movieCount: 24,
    color: "bg-blue-400",
    images: [],
    fallbackPosters: [
      { title: "雨に唄えば", bg: "from-blue-700 to-blue-950" },
      { title: "天気の子", bg: "from-sky-600 to-sky-900" },
      { title: "ブレードランナー", bg: "from-indigo-700 to-indigo-950" },
    ],
  },
  {
    title: "人生で一度は観るべき名作",
    movieCount: 89,
    color: "bg-amber-400",
    images: [],
    fallbackPosters: [
      { title: "ショーシャンクの空に", bg: "from-amber-700 to-amber-950" },
      { title: "ゴッドファーザー", bg: "from-stone-700 to-stone-950" },
      { title: "フォレスト・ガンプ", bg: "from-emerald-700 to-emerald-950" },
    ],
  },
  {
    title: "笑えるコメディ映画",
    movieCount: 156,
    color: "bg-green-400",
    images: [],
    fallbackPosters: [
      { title: "ホーム・アローン", bg: "from-green-600 to-green-900" },
      { title: "テルマエ・ロマエ", bg: "from-yellow-700 to-yellow-950" },
      { title: "翔んで埼玉", bg: "from-lime-700 to-lime-950" },
    ],
  },
  {
    title: "泣ける恋愛映画",
    movieCount: 67,
    color: "bg-pink-400",
    images: [],
    fallbackPosters: [
      { title: "タイタニック", bg: "from-pink-700 to-pink-950" },
      { title: "君の名は。", bg: "from-sky-600 to-indigo-900" },
      { title: "ノッティングヒルの恋人", bg: "from-rose-600 to-rose-900" },
    ],
  },
  {
    title: "夜更かしにぴったり",
    movieCount: 31,
    color: "bg-indigo-400",
    images: [],
    fallbackPosters: [
      { title: "ファイト・クラブ", bg: "from-indigo-800 to-gray-950" },
      { title: "マトリックス", bg: "from-green-800 to-gray-950" },
      { title: "インセプション", bg: "from-slate-700 to-slate-950" },
    ],
  },
  {
    title: "友達と観たい映画",
    movieCount: 43,
    color: "bg-teal-400",
    images: [],
    fallbackPosters: [
      { title: "ジュラシック・パーク", bg: "from-teal-700 to-teal-950" },
      {
        title: "バック・トゥ・ザ・フューチャー",
        bg: "from-blue-600 to-blue-900",
      },
      { title: "ミッション:インポッシブル", bg: "from-red-800 to-red-950" },
    ],
  },
  {
    title: "頭を使うサスペンス",
    movieCount: 71,
    color: "bg-purple-400",
    images: [],
    fallbackPosters: [
      { title: "シャッター アイランド", bg: "from-purple-800 to-purple-950" },
      { title: "メメント", bg: "from-gray-700 to-gray-950" },
      { title: "ゴーン・ガール", bg: "from-zinc-700 to-zinc-950" },
    ],
  },
  {
    title: "音楽が最高な映画",
    movieCount: 33,
    color: "bg-rose-400",
    images: [],
    fallbackPosters: [
      { title: "ラ・ラ・ランド", bg: "from-violet-700 to-indigo-950" },
      { title: "ボヘミアン・ラプソディ", bg: "from-rose-700 to-rose-950" },
      { title: "SING", bg: "from-fuchsia-600 to-fuchsia-900" },
    ],
  },
];

const FALLBACK_ROW2: MarqueeTag[] = [
  {
    title: "旅に出たくなる映画",
    movieCount: 45,
    color: "bg-emerald-400",
    images: [],
    fallbackPosters: [
      { title: "LIFE!", bg: "from-emerald-600 to-emerald-900" },
      {
        title: "食べて、祈って、恋をして",
        bg: "from-orange-600 to-orange-900",
      },
      { title: "イントゥ・ザ・ワイルド", bg: "from-green-700 to-green-950" },
    ],
  },
  {
    title: "家族で観たい映画",
    movieCount: 62,
    color: "bg-orange-400",
    images: [],
    fallbackPosters: [
      { title: "となりのトトロ", bg: "from-green-600 to-green-900" },
      { title: "トイ・ストーリー", bg: "from-sky-500 to-sky-800" },
      { title: "サマーウォーズ", bg: "from-blue-500 to-blue-800" },
    ],
  },
  {
    title: "夏に観たい爽快映画",
    movieCount: 52,
    color: "bg-cyan-400",
    images: [],
    fallbackPosters: [
      {
        title: "サマータイムマシン・ブルース",
        bg: "from-cyan-600 to-cyan-900",
      },
      { title: "菊次郎の夏", bg: "from-sky-600 to-sky-900" },
      { title: "ウォーターボーイズ", bg: "from-blue-500 to-blue-800" },
    ],
  },
  {
    title: "クリスマスに観たい映画",
    movieCount: 38,
    color: "bg-red-400",
    images: [],
    fallbackPosters: [
      { title: "ホーム・アローン", bg: "from-red-600 to-red-900" },
      { title: "ラブ・アクチュアリー", bg: "from-rose-700 to-rose-950" },
      { title: "素晴らしき哉、人生!", bg: "from-amber-700 to-amber-950" },
    ],
  },
  {
    title: "元気が出る映画",
    movieCount: 94,
    color: "bg-yellow-400",
    images: [],
    fallbackPosters: [
      {
        title: "グレイテスト・ショーマン",
        bg: "from-yellow-600 to-yellow-900",
      },
      {
        title: "リトル・ミス・サンシャイン",
        bg: "from-amber-600 to-amber-900",
      },
      { title: "ROOKIES", bg: "from-orange-600 to-orange-900" },
    ],
  },
  {
    title: "一人で静かに観たい映画",
    movieCount: 57,
    color: "bg-slate-400",
    images: [],
    fallbackPosters: [
      { title: "ドライブ・マイ・カー", bg: "from-slate-700 to-slate-950" },
      { title: "パターソン", bg: "from-gray-600 to-gray-900" },
      { title: "万引き家族", bg: "from-stone-700 to-stone-950" },
    ],
  },
  {
    title: "SF好きにおすすめ",
    movieCount: 83,
    color: "bg-violet-400",
    images: [],
    fallbackPosters: [
      { title: "インターステラー", bg: "from-violet-800 to-gray-950" },
      { title: "ブレードランナー 2049", bg: "from-orange-800 to-gray-950" },
      { title: "メッセージ", bg: "from-slate-600 to-slate-900" },
    ],
  },
  {
    title: "実話ベースの感動作",
    movieCount: 46,
    color: "bg-lime-400",
    images: [],
    fallbackPosters: [
      { title: "シンドラーのリスト", bg: "from-gray-700 to-gray-950" },
      { title: "ビューティフル・マインド", bg: "from-blue-700 to-blue-950" },
      { title: "ソーシャル・ネットワーク", bg: "from-sky-800 to-sky-950" },
    ],
  },
];

function toFullImageUrl(src: string): string {
  return src.startsWith("http") ? src : `${TMDB_IMAGE_BASE}${src}`;
}

function buildRows(result: { items: Array<{ title: string; movieCount: number; images: string[] }> }): {
  row1: MarqueeTag[];
  row2: MarqueeTag[];
} {
  if (result.items.length === 0) {
    return { row1: FALLBACK_ROW1, row2: FALLBACK_ROW2 };
  }

  const tags: MarqueeTag[] = result.items.map((item, i) => ({
    title: item.title,
    movieCount: item.movieCount,
    color: TAG_COLORS[i % TAG_COLORS.length],
    images: item.images.slice(0, 3).map(toFullImageUrl),
  }));

  const row1 =
    tags.length >= 8
      ? tags.slice(0, 8)
      : [
          ...tags.slice(0, Math.ceil(tags.length / 2)),
          ...FALLBACK_ROW1,
        ].slice(0, 8);
  const row2 =
    tags.length >= 16
      ? tags.slice(8, 16)
      : tags.length > 8
        ? [...tags.slice(8), ...FALLBACK_ROW2].slice(0, 8)
        : [...tags.slice(Math.ceil(tags.length / 2)), ...FALLBACK_ROW2].slice(
            0,
            8,
          );

  return { row1, row2 };
}

function MarqueeSkeleton() {
  return (
    <div className="space-y-4">
      {[0, 1].map((row) => (
        <div key={row} className="flex gap-4 pl-4 overflow-hidden">
          {Array.from({ length: 8 }).map((_, i) => (
            <div
              key={i}
              className="flex items-center gap-3 bg-gray-100 rounded-2xl pl-5 pr-3 py-3 shrink-0 animate-pulse"
            >
              <div className="w-3 h-3 rounded-full bg-gray-200" />
              <div className="w-28 h-4 bg-gray-200 rounded" />
              <div className="w-10 h-3 bg-gray-200 rounded" />
              <div className="flex -space-x-2 ml-1">
                {[0, 1, 2].map((j) => (
                  <div key={j} className="w-8 h-12 rounded-md bg-gray-200" />
                ))}
              </div>
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

export function MarqueeSection() {
  const { data, isLoading } = useQuery({
    queryKey: ["tags", "marquee"],
    queryFn: () => listTags({ sort: "popular", pageSize: 16 }),
    staleTime: 60 * 60 * 1000, // 1時間
  });

  const { row1, row2 } = data
    ? buildRows(data)
    : { row1: FALLBACK_ROW1, row2: FALLBACK_ROW2 };

  if (isLoading) {
    return (
      <section className="py-16 md:py-20 bg-white overflow-hidden">
        <div className="text-center mb-10">
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-gray-900">
            みんなが作ったタグ
          </h2>
          <p className="mt-4 text-base md:text-lg text-gray-500">
            ユーザーが自由に作成したタグの一部をご紹介
          </p>
        </div>
        <MarqueeSkeleton />
      </section>
    );
  }

  return (
    <section className="py-16 md:py-20 bg-white overflow-hidden">
      <div className="text-center mb-10">
        <h2 className="text-3xl md:text-4xl font-black tracking-tight text-gray-900">
          みんなが作ったタグ
        </h2>
        <p className="mt-4 text-base md:text-lg text-gray-500">
          ユーザーが自由に作成したタグの一部をご紹介
        </p>
      </div>

      {/* Row 1 - scrolls left */}
      <div className="marquee-track mb-4">
        <div className="animate-marquee-left flex w-max gap-4 pl-4">
          {[...row1, ...row1].map((tag, i) => (
            <TagCard key={`r1-${i}`} tag={tag} />
          ))}
        </div>
      </div>

      {/* Row 2 - scrolls right */}
      <div className="marquee-track">
        <div className="animate-marquee-right flex w-max gap-4 pl-4">
          {[...row2, ...row2].map((tag, i) => (
            <TagCard key={`r2-${i}`} tag={tag} />
          ))}
        </div>
      </div>
    </section>
  );
}
