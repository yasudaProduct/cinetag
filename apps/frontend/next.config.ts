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
    // 開発環境では緩い設定、本番環境では厳格な設定
    const scriptSrc = isDev
      ? "'self' 'unsafe-inline' 'unsafe-eval'"
      : "'self' 'unsafe-inline' 'unsafe-eval'"; // TODO: 将来的にNonceベースに移行

    const connectSrc = isDev
      ? "'self' https://clerk.com https://*.clerk.accounts.dev http://localhost:8080"
      : "'self' https://clerk.com https://*.clerk.accounts.dev";

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

              // 画像: 自サイト + data URIs + 外部画像サービス
              "img-src 'self' data: https://placehold.co https://image.tmdb.org https://img.clerk.com https://images.clerk.dev",

              // 接続: 自サイト + Clerk + バックエンドAPI
              `connect-src ${connectSrc}`,

              // フレーム: 自サイト + Clerk（認証モーダル用）
              "frame-src 'self' https://clerk.com https://*.clerk.accounts.dev",

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
            value: "camera=(), microphone=(), geolocation=(), interest-cohort=()",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
