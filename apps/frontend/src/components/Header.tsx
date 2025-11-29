import Link from "next/link";
import { Search, Bell, User } from "lucide-react";

export const Header = () => {
  return (
    <header className="bg-white border-b border-gray-200 py-4 sticky top-0 z-50">
      <div className="container mx-auto px-4 md:px-6 flex items-center justify-between">
        {/* Logo & Navigation */}
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2">
            <div className="bg-blue-500 text-white p-1 rounded text-xs font-bold">
              🍿
            </div>
            <span className="text-xl font-bold text-gray-800">cinetag</span>
          </Link>
          <nav className="hidden md:flex items-center gap-6 text-sm font-medium text-gray-600">
            <Link href="/" className="text-gray-900">
              ホーム
            </Link>
            <Link
              href="/categories"
              className="hover:text-gray-900 text-pink-500"
            >
              カテゴリを探す
            </Link>
            <Link href="/mypage" className="hover:text-gray-900">
              マイページ
            </Link>
          </nav>
        </div>

        {/* User Actions */}
        <div className="flex items-center gap-4">
          <button className="p-2 rounded-full hover:bg-gray-100 border border-gray-200">
            <Bell className="w-5 h-5 text-gray-600" />
          </button>
          <button className="p-1 rounded-full border border-gray-200 hover:bg-gray-100">
            <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden">
              <User className="w-5 h-5 text-gray-500" />
            </div>
          </button>
        </div>
      </div>
    </header>
  );
};
