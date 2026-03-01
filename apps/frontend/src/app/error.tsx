"use client";

import Link from "next/link";

export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center max-w-md">
        <h1 className="text-6xl font-bold text-gray-300">500</h1>
        <p className="mt-4 text-xl font-bold text-gray-900">
          エラーが発生しました
        </p>
        <p className="mt-2 text-gray-500">
          一時的な問題が発生しています。しばらくしてからもう一度お試しください。
        </p>
        <div className="mt-8 flex flex-col sm:flex-row gap-3 justify-center">
          <button
            type="button"
            onClick={reset}
            className="px-6 py-3 bg-blue-500 text-white rounded-xl font-medium hover:bg-blue-600 transition-colors"
          >
            もう一度試す
          </button>
          <Link
            href="/"
            className="px-6 py-3 bg-gray-100 text-gray-700 rounded-xl font-medium hover:bg-gray-200 transition-colors"
          >
            トップページへ
          </Link>
        </div>
      </div>
    </div>
  );
}
