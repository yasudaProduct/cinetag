import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "利用規約 | cinetag",
  description: "cinetagの利用規約です。サービスの利用条件をご確認ください。",
};

export default function TermsPage() {
  return (
    <div className="min-h-screen">
      <main className="container mx-auto px-4 md:px-6 py-10 max-w-4xl">
        <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-8 md:p-12">
          <h1 className="text-3xl font-bold text-gray-900 mb-8">利用規約</h1>

          <div className="prose prose-gray max-w-none space-y-8">
            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第1条（適用）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本規約は、cinetag（以下「本サービス」といいます）の利用に関する条件を定めるものです。ユーザーの皆様には、本規約に同意いただいた上で、本サービスをご利用いただきます。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第2条（定義）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本規約において使用する用語の定義は、以下のとおりとします。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>
                  「ユーザー」とは、本サービスを利用する全ての方を指します。
                </li>
                <li>
                  「タグ」とは、ユーザーが作成する映画のプレイリストを指します。
                </li>
                <li>
                  「コンテンツ」とは、ユーザーが本サービス上で作成・投稿した情報を指します。
                </li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第3条（アカウント登録）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本サービスの一部機能を利用するには、アカウント登録が必要です。ユーザーは、正確かつ最新の情報を提供し、登録情報に変更があった場合は速やかに更新するものとします。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第4条（禁止事項）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                ユーザーは、本サービスの利用にあたり、以下の行為を行ってはなりません。
              </p>
              <ul className="list-disc list-inside text-gray-700 mt-2 space-y-1">
                <li>法令または公序良俗に違反する行為</li>
                <li>犯罪行為に関連する行為</li>
                <li>他のユーザーまたは第三者の権利を侵害する行為</li>
                <li>
                  本サービスの運営を妨害する行為、またはそのおそれのある行為
                </li>
                <li>不正アクセス、またはこれを試みる行為</li>
                <li>他のユーザーに成りすます行為</li>
                <li>スパム行為、または本サービスを悪用した宣伝行為</li>
                <li>その他、運営者が不適切と判断する行為</li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第5条（コンテンツの権利）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                ユーザーが作成したコンテンツの著作権は、当該ユーザーに帰属します。ただし、ユーザーは、本サービスの提供・改善のために必要な範囲で、運営者がコンテンツを利用することを許諾するものとします。
              </p>
              <p className="text-gray-700 leading-relaxed mt-2">
                本サービスで表示される映画情報は、The Movie Database
                (TMDB)より提供されています。映画のポスター画像、タイトル、その他のメタデータの著作権は、それぞれの権利者に帰属します。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第6条（サービスの変更・中断・終了）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                運営者は、ユーザーへの事前通知なく、本サービスの内容を変更、中断、または終了することができます。これによりユーザーに生じた損害について、運営者は責任を負いません。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第7条（免責事項）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                運営者は、本サービスの利用によりユーザーに生じた損害について、故意または重大な過失がある場合を除き、責任を負いません。また、本サービスで提供される情報の正確性、完全性、有用性等について、いかなる保証も行いません。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第8条（規約の変更）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                運営者は、必要と判断した場合には、ユーザーへの事前通知なく本規約を変更することができます。変更後の利用規約は、本サービス上に掲載した時点で効力を生じるものとします。
              </p>
            </section>

            <section>
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                第9条（準拠法・裁判管轄）
              </h2>
              <p className="text-gray-700 leading-relaxed">
                本規約の解釈にあたっては、日本法を準拠法とします。本サービスに関して紛争が生じた場合には、運営者の所在地を管轄する裁判所を専属的合意管轄とします。
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
