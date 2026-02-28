import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, Tag, Users, Search } from "lucide-react";
import { MarqueeSection } from "./_components/MarqueeSection";

// --- ページ ---

export const metadata: Metadata = {
  title: "cinetag - 映画をタグでつながる、共有する",
  description:
    "cinetagは、映画に自由にタグを作成し、他のユーザーと共有できる新しい映画プラットフォームです。",
};

export default function LandingPage() {
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
              href="/sign-in"
              className="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
            >
              ログイン
            </Link>
            <Link
              href="/sign-up"
              className="inline-flex items-center px-5 py-2.5 bg-gray-900 text-white text-sm font-semibold rounded-full hover:bg-gray-800 transition-colors"
            >
              無料で始める
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

      {/* Tag Marquee Section (client-side fetch) */}
      <MarqueeSection />

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
