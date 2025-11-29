import Link from "next/link";
import { Bell, User } from "lucide-react";
import { SignedIn, SignedOut, SignInButton, UserButton } from "@clerk/nextjs";

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
          <SignedIn>
            <button className="p-2 rounded-full hover:bg-gray-100 border border-gray-200">
              <Bell className="w-5 h-5 text-gray-600" />
            </button>
            <UserButton
              appearance={{
                elements: {
                  avatarBox: "w-9 h-9",
                },
              }}
            />
          </SignedIn>
          <SignedOut>
            <SignInButton mode="modal">
              <button className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white text-sm font-medium rounded-lg hover:bg-blue-600 transition-colors">
                <User className="w-4 h-4" />
                ログイン
              </button>
            </SignInButton>
          </SignedOut>
        </div>
      </div>
    </header>
  );
};
