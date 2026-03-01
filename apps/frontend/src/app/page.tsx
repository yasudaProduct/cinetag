import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, Tag, Users, Search } from "lucide-react";
import { listTags, type ListTagsResult } from "@/lib/api/tags/list";
import { TagCard, type MarqueeTag } from "./_components/TagCard";

// --- カラーパレット（実データ用の循環色） ---

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

// --- フォールバック用サンプルデータ ---

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

// --- データ取得 ---

const TMDB_IMAGE_BASE = "https://image.tmdb.org/t/p/w500";

function toFullImageUrl(src: string): string {
  return src.startsWith("http") ? src : `${TMDB_IMAGE_BASE}${src}`;
}

async function fetchMarqueeTags(): Promise<{
  row1: MarqueeTag[];
  row2: MarqueeTag[];
}> {
  try {
    const result: ListTagsResult = await listTags({
      sort: "popular",
      pageSize: 16,
    });

    if (result.items.length === 0) {
      return { row1: FALLBACK_ROW1, row2: FALLBACK_ROW2 };
    }

    const tags: MarqueeTag[] = result.items.map((item, i) => ({
      title: item.title,
      movieCount: item.movieCount,
      color: TAG_COLORS[i % TAG_COLORS.length],
      images: item.images.slice(0, 3).map(toFullImageUrl),
    }));

    // 8件ずつに分割、足りない行はフォールバックで補完
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
  } catch {
    return { row1: FALLBACK_ROW1, row2: FALLBACK_ROW2 };
  }
}

// --- ページ ---

export const metadata: Metadata = {
  title: "cinetag - 映画をタグでつながる、共有する",
  description:
    "cinetagは、映画に自由にタグを作成し、他のユーザーと共有できる新しい映画プラットフォームです。",
};

export default async function LandingPage() {
  const { row1, row2 } = await fetchMarqueeTags();

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-gray-100">
        <div className="max-w-6xl mx-auto px-6 h-[78px] flex items-center justify-between">
          <Link href="/" className="text-2xl font-black tracking-tight">
            cinetag
          </Link>
          <div className="flex items-center gap-4">
            <Link
              href="/tags"
              className="inline-flex items-center px-5 py-2.5 bg-gray-900 text-white text-sm font-semibold rounded-full hover:bg-gray-800 transition-colors"
            >
              タグを探す
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="bg-[#FFF5F5]">
        <div className="max-w-6xl mx-auto px-6 py-16 md:py-24">
          <div className="flex flex-col md:flex-row items-center gap-12 md:gap-16">
            {/* Left content */}
            <div className="flex-1 space-y-6">
              <span className="inline-block px-4 py-2 bg-pink-100 text-pink-700 text-xs font-bold rounded-full">
                映画の新しい楽しみ方
              </span>
              <h1 className="text-4xl md:text-5xl font-black leading-tight tracking-tight text-gray-900">
                映画をタグで
                <br />
                つながる、共有する
              </h1>
              <p className="text-base md:text-lg text-gray-600 leading-relaxed max-w-xl">
                cinetagは、映画に自由にタグを作成し、他のユーザーと共有できる新しい映画プラットフォームです。あなただけのプレイリストを作って、映画の楽しみ方を広げましょう。
              </p>
              <div className="flex flex-wrap gap-4 pt-2">
                <Link
                  href="/tags"
                  className="inline-flex items-center gap-2 px-7 py-3.5 bg-gray-900 text-white font-semibold rounded-full hover:bg-gray-800 transition-colors"
                >
                  タグを探す
                  <ArrowRight className="w-5 h-5" />
                </Link>
                <a
                  href="#how-to-use"
                  className="inline-flex items-center px-7 py-3.5 border-2 border-gray-300 text-gray-700 font-semibold rounded-full hover:border-gray-400 hover:text-gray-900 transition-colors"
                >
                  使い方を見る
                </a>
              </div>
            </div>

            {/* Right image */}
            <div className="flex-1 relative">
              <div className="relative w-full aspect-square max-w-[480px] mx-auto">
                {/* Hero image placeholder - replace with actual cinema image */}
                <div className="w-full h-full rounded-3xl overflow-hidden shadow-2xl bg-gradient-to-br from-indigo-950 via-blue-900 to-indigo-950">
                  <div className="w-full h-full flex items-center justify-center relative">
                    {/* Cinema seats pattern */}
                    <div className="absolute inset-0 opacity-30">
                      {Array.from({ length: 5 }).map((_, row) => (
                        <div
                          key={row}
                          className="flex justify-center gap-2 mt-4"
                          style={{
                            transform: `perspective(500px) rotateX(${10 + row * 5}deg)`,
                          }}
                        >
                          {Array.from({ length: 7 }).map((_, seat) => (
                            <div
                              key={seat}
                              className="w-8 h-10 md:w-12 md:h-14 bg-blue-600 rounded-t-lg"
                            />
                          ))}
                        </div>
                      ))}
                    </div>
                    {/* Screen glow */}
                    <div className="absolute top-4 left-1/2 -translate-x-1/2 w-3/4 h-16 bg-white/10 rounded-lg blur-sm" />
                  </div>
                </div>

                {/* Stats badge */}
                <div className="absolute -bottom-4 -left-4 md:bottom-4 md:left-[-24px] bg-white rounded-2xl shadow-xl px-5 py-4 flex items-center gap-3">
                  <div className="w-10 h-10 bg-pink-100 rounded-xl flex items-center justify-center">
                    <Tag className="w-5 h-5 text-pink-600" />
                  </div>
                  <div>
                    <p className="text-xl font-black text-gray-900">10,000+</p>
                    <p className="text-xs text-gray-500 font-medium">
                      作成されたタグ
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Tag Marquee Section */}
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

      {/* Features Section */}
      <section className="bg-gray-50/80">
        <div className="max-w-6xl mx-auto px-6 py-16 md:py-24">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-black tracking-tight text-gray-900">
              cinetagの特徴
            </h2>
            <p className="mt-4 text-base md:text-lg text-gray-500">
              映画をもっと自由に、もっと楽しく
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Feature 1 */}
            <div className="bg-white rounded-2xl border border-gray-200 p-8 hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-pink-50 rounded-2xl flex items-center justify-center mb-6">
                <Tag className="w-8 h-8 text-pink-500" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-3">
                自由なタグ作成
              </h3>
              <p className="text-gray-500 leading-relaxed">
                「雨の日に観たい映画」「元気が出る映画」など、あなただけのテーマでタグを作成。映画を自由に分類できます。
              </p>
            </div>

            {/* Feature 2 */}
            <div className="bg-white rounded-2xl border border-gray-200 p-8 hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-blue-50 rounded-2xl flex items-center justify-center mb-6">
                <Users className="w-8 h-8 text-blue-500" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-3">
                みんなで共有
              </h3>
              <p className="text-gray-500 leading-relaxed">
                作成したタグは他のユーザーも閲覧・編集可能。協力してプレイリストを充実させられます。
              </p>
            </div>

            {/* Feature 3 */}
            <div className="bg-white rounded-2xl border border-gray-200 p-8 hover:shadow-lg transition-shadow">
              <div className="w-16 h-16 bg-amber-50 rounded-2xl flex items-center justify-center mb-6">
                <Search className="w-8 h-8 text-amber-500" />
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-3">簡単検索</h3>
              <p className="text-gray-500 leading-relaxed">
                興味のあるテーマのタグを検索して、新しい映画との出会いを楽しめます。
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How to Use Section */}
      <section id="how-to-use" className="bg-white scroll-mt-20">
        <div className="max-w-6xl mx-auto px-6 py-16 md:py-24">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-black tracking-tight text-gray-900">
              使い方はとても簡単
            </h2>
            <p className="mt-4 text-base md:text-lg text-gray-500">
              3ステップで始められます
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-12 md:gap-8">
            {[
              {
                step: "1",
                title: "タグを作成",
                description:
                  "好きなテーマでタグを作成し、名前と説明を設定します",
              },
              {
                step: "2",
                title: "映画を追加",
                description: "タグに映画を追加して、プレイリストを充実させます",
              },
              {
                step: "3",
                title: "共有・発見",
                description:
                  "他のユーザーとタグを共有し、新しい映画を発見します",
              },
            ].map((item) => (
              <div key={item.step} className="text-center">
                <div className="w-20 h-20 mx-auto mb-6 bg-gradient-to-br from-amber-400 to-orange-400 rounded-full flex items-center justify-center shadow-lg shadow-amber-200/50">
                  <span className="text-3xl font-black text-white">
                    {item.step}
                  </span>
                </div>
                <h3 className="text-lg font-bold text-gray-900 mb-3">
                  {item.title}
                </h3>
                <p className="text-gray-500 leading-relaxed max-w-xs mx-auto">
                  {item.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="px-6 py-16 md:py-24">
        <div className="max-w-4xl mx-auto bg-[#FFD75E] rounded-3xl px-8 py-12 md:px-16 md:py-16 text-center">
          <h2 className="text-3xl md:text-4xl font-black tracking-tight text-gray-900">
            今すぐcinetagを始めよう
          </h2>
          <p className="mt-4 text-base md:text-lg text-gray-700">
            無料でアカウントを作成して、映画の新しい楽しみ方を体験してください
          </p>
          <Link
            href="/tags"
            className="inline-flex items-center gap-2 mt-8 px-8 py-4 bg-gray-900 text-white font-semibold rounded-full hover:bg-gray-800 transition-colors text-lg"
          >
            無料で始める
            <ArrowRight className="w-5 h-5" />
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-50 border-t border-gray-200">
        <div className="max-w-6xl mx-auto px-6 py-12">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            {/* Brand */}
            <div className="col-span-2 md:col-span-1">
              <p className="text-xl font-black tracking-tight text-gray-900">
                cinetag
              </p>
              <p className="mt-3 text-sm text-gray-500 leading-relaxed">
                映画をタグでつながる、
                <br />
                共有する新しいプラットフォーム
              </p>
            </div>

            {/* Support */}
            <div>
              <h4 className="text-sm font-bold text-gray-900 mb-4">サポート</h4>
              <ul className="space-y-3">
                <li>
                  <span className="text-sm text-gray-500">お問い合わせ</span>
                </li>
              </ul>
            </div>

            {/* Legal */}
            <div>
              <h4 className="text-sm font-bold text-gray-900 mb-4">法的情報</h4>
              <ul className="space-y-3">
                <li>
                  <Link
                    href="/terms"
                    className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
                  >
                    利用規約
                  </Link>
                </li>
                <li>
                  <Link
                    href="/privacy"
                    className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
                  >
                    プライバシーポリシー
                  </Link>
                </li>
              </ul>
            </div>
          </div>

          {/* Copyright */}
          <div className="mt-12 pt-6 border-t border-gray-200 text-center">
            <p className="text-sm text-gray-400">
              &copy; 2026 cinetag. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}
