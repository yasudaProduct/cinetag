import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  robots: { index: false, follow: false },
};

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-[#FFF9F3] px-4 py-12">
      <Link
        href="/"
        className="mb-8 text-3xl font-black tracking-tight text-gray-900"
      >
        cinetag
      </Link>

      <div className="w-full max-w-md">{children}</div>

      <p className="mt-8 text-xs text-gray-400">&copy; 2026 cinetag</p>
    </div>
  );
}
