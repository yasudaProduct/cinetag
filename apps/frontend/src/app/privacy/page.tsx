import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "プライバシーポリシー | cinetag",
  description:
    "cinetagのプライバシーポリシーです。個人情報の取り扱いについてご確認ください。",
};

export default function PrivacyPage() {
  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-10 max-w-4xl">
        <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-8 md:p-12">
          <h1 className="text-3xl font-bold text-gray-900 mb-8">
            プライバシーポリシー
          </h1>

          <div className="prose prose-gray max-w-none space-y-8">
            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                1. はじめに
              </h2>
              <p className="text-gray-700 leading-relaxed">
                cinetag（以下「本サービス」といいます）は、ユーザーの皆様のプライバシーを尊重し、個人情報の保護に努めています。本プライバシーポリシーは、本サービスにおける個人情報の取り扱いについて定めるものです。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                2. 収集する情報
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスでは、以下の情報を収集することがあります。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>
                  アカウント情報：メールアドレス、ユーザー名、プロフィール画像など
                </li>
                <li>
                  利用情報：作成したタグ、登録した映画、閲覧履歴などのサービス利用に関する情報
                </li>
                <li>
                  技術情報：IPアドレス、ブラウザの種類、デバイス情報、アクセス日時など
                </li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                3. 情報の利用目的
              </h2>
              <p className="text-gray-700 leading-relaxed">
                収集した情報は、以下の目的で利用します。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>本サービスの提供・運営・維持</li>
                <li>ユーザーサポートの提供</li>
                <li>サービスの改善・新機能の開発</li>
                <li>利用状況の分析・統計</li>
                <li>不正利用の防止・セキュリティの確保</li>
                <li>重要なお知らせの送信</li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                4. 情報の共有
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスでは、以下の場合を除き、ユーザーの個人情報を第三者に提供することはありません。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>ユーザーの同意がある場合</li>
                <li>法令に基づく場合</li>
                <li>
                  人の生命、身体または財産の保護のために必要がある場合であって、ユーザーの同意を得ることが困難である場合
                </li>
                <li>
                  サービス提供に必要な業務委託先に対して、必要な範囲で提供する場合
                </li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                5. 外部サービスの利用
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスでは、以下の外部サービスを利用しています。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>
                  <strong>Clerk</strong>
                  ：認証サービスとして利用しています。Clerkのプライバシーポリシーについては、Clerkの公式サイトをご確認ください。
                </li>
                <li>
                  <strong>The Movie Database (TMDB)</strong>
                  ：映画情報の取得に利用しています。
                </li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                6. Cookieの使用
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスでは、ユーザー体験の向上やセッション管理のためにCookieを使用しています。ブラウザの設定によりCookieを無効にすることができますが、一部の機能が正常に動作しなくなる場合があります。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                7. データの保護
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスは、ユーザーの個人情報を保護するために、適切な技術的・組織的セキュリティ対策を講じています。ただし、インターネット上のデータ送信は完全に安全であることを保証することはできません。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                8. ユーザーの権利
              </h2>
              <p className="text-gray-700 leading-relaxed">
                ユーザーは、自己の個人情報について、以下の権利を有します。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>個人情報の開示を請求する権利</li>
                <li>個人情報の訂正・削除を請求する権利</li>
                <li>個人情報の利用停止を請求する権利</li>
                <li>アカウントの削除を請求する権利</li>
              </ul>
              <p className="text-gray-700 leading-relaxed mt-2">
                これらの権利を行使する場合は、お問い合わせよりご連絡ください。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                9. 子どものプライバシー
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスは、13歳未満の方を対象としていません。13歳未満の方から意図せず個人情報を収集した場合は、速やかに削除いたします。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                10. プライバシーポリシーの変更
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本プライバシーポリシーは、必要に応じて変更されることがあります。重要な変更がある場合は、本サービス上でお知らせします。変更後も本サービスを継続して利用することにより、変更後のプライバシーポリシーに同意したものとみなされます。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                11. お問い合わせ
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本プライバシーポリシーに関するご質問やお問い合わせは、サービス内のお問い合わせフォームよりご連絡ください。
              </p>
            </section>

            <div className="pt-8 border-t border-gray-200">
              <p className="text-sm text-gray-500">
                制定日: 2025年1月1日
                <br />
                最終更新日: 2025年1月1日
              </p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
