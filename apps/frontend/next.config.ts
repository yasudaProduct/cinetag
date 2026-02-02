import type { NextConfig } from "next";

const isDev = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "placehold.co",
      },
      {
        protocol: "https",
        hostname: "image.tmdb.org",
      },
      // Clerk user avatar URLs
      {
        protocol: "https",
        hostname: "img.clerk.com",
      },
      {
        protocol: "https",
        hostname: "images.clerk.dev",
      },
    ],
  },

  async headers() {
    const backendApiBase = process.env.NEXT_PUBLIC_BACKEND_API_BASE;
    const backendOrigin = (() => {
      if (!backendApiBase) return undefined;
      try {
        return new URL(backendApiBase).origin;
      } catch {
        return undefined;
      }
    })();

    const clerkPublishableKey = process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;
    const clerkAccountsHosts = (() => {
      if (clerkPublishableKey?.startsWith("pk_live_")) {
        return ["https://*.clerk.accounts.com"];
      }
      if (clerkPublishableKey?.startsWith("pk_test_")) {
        return ["https://*.clerk.accounts.dev"];
      }

      return isDev
        ? ["https://*.clerk.accounts.dev"]
        : ["https://*.clerk.accounts.com", "https://*.clerk.accounts.dev"];
    })();
    const clerkAccountsSrc = clerkAccountsHosts.join(" ");

    const scriptSrc = isDev
      ? `'self' 'unsafe-inline' 'unsafe-eval' ${clerkAccountsSrc} https://clerk.com`
      : `'self' 'unsafe-inline' 'unsafe-eval' ${clerkAccountsSrc} https://clerk.com`; // TODO: 将来的にNonceベースに移行

    const connectSrc = [
      "'self'",
      "https://clerk.com",
      ...clerkAccountsHosts,
      ...(isDev ? ["http://localhost:8080"] : []),
      ...(backendOrigin ? [backendOrigin] : []),
    ].join(" ");

    const workerSrc = ["'self'", "blob:"].join(" ");

    return [
      {
        source: "/:path*",
        headers: [
          // Content Security Policy
          {
            key: "Content-Security-Policy",
            value: [
              // デフォルト: 同一オリジンのみ
              "default-src 'self'",

              // スクリプト: 自サイト + インラインスクリプト（Next.jsとReactに必要）
              `script-src ${scriptSrc}`,

              // スタイル: 自サイト + インラインスタイル + Google Fonts
              "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",

              // フォント: 自サイト + Google Fonts
              "font-src 'self' https://fonts.gstatic.com data:",

              // 画像: 自サイト + data URIs + blob URIs + 外部画像サービス
              "img-src 'self' data: blob: https://placehold.co https://image.tmdb.org https://img.clerk.com https://images.clerk.dev",

              // 接続: 自サイト + Clerk + バックエンドAPI
              `connect-src ${connectSrc}`,

              // Worker: 自サイト + blob URL（Web Worker生成用）
              `worker-src ${workerSrc}`,

              // フレーム: 自サイト + Clerk（認証モーダル用）
              `frame-src 'self' https://clerk.com ${clerkAccountsSrc}`,

              // オブジェクト: 禁止（Flash等のプラグイン対策）
              "object-src 'none'",

              // Base URI: 自サイトのみ（相対URLハイジャック対策）
              "base-uri 'self'",

              // フォーム送信先: 自サイトのみ
              "form-action 'self'",

              // 他サイトでのiframe埋め込み: 禁止（クリックジャッキング対策）
              "frame-ancestors 'none'",

              // HTTPをHTTPSに自動アップグレード（本番環境のみ）
              ...(isDev ? [] : ["upgrade-insecure-requests"]),
            ].join("; "),
          },

          // X-Content-Type-Options: MIMEタイプスニッフィング防止
          {
            key: "X-Content-Type-Options",
            value: "nosniff",
          },

          // X-Frame-Options: クリックジャッキング対策（CSPと重複だが念のため）
          {
            key: "X-Frame-Options",
            value: "DENY",
          },

          // X-XSS-Protection: 旧ブラウザ向けXSS対策
          {
            key: "X-XSS-Protection",
            value: "1; mode=block",
          },

          // Referrer-Policy: リファラー情報の制御
          {
            key: "Referrer-Policy",
            value: "strict-origin-when-cross-origin",
          },

          // Permissions-Policy: 不要な機能の無効化
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
