import Link from "next/link";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <h1 className="text-6xl font-bold text-gray-300">404</h1>
        <p className="mt-4 text-xl font-bold text-gray-900">
          ページが見つかりません
        </p>
        <p className="mt-2 text-gray-500">
          お探しのページは存在しないか、移動した可能性があります。
        </p>
        <div className="mt-8 flex flex-col sm:flex-row gap-3 justify-center">
          <Link
            href="/"
            className="px-6 py-3 bg-blue-500 text-white rounded-xl font-medium hover:bg-blue-600 transition-colors"
          >
            トップページへ
          </Link>
          <Link
            href="/tags"
            className="px-6 py-3 bg-gray-100 text-gray-700 rounded-xl font-medium hover:bg-gray-200 transition-colors"
          >
            タグ一覧を見る
          </Link>
        </div>
      </div>
    </div>
  );
}
