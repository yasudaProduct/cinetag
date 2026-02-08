import { NextRequest, NextResponse } from "next/server";

/**
 * CSP違反レポートを受け取るエンドポイント
 *
 * ブラウザがContent Security Policyに違反したリソースを検出した際に、
 * このエンドポイントにレポートを送信します。
 *
 * 本番環境では、これらのレポートをログ収集サービス（Sentry、DataDog等）に
 * 転送することを推奨します。
 */
export async function POST(request: NextRequest) {
  try {
    const report = await request.json();

    // 開発環境ではコンソールに出力
    if (process.env.NODE_ENV === "development") {
      console.warn("🚨 CSP Violation Report:", JSON.stringify(report, null, 2));
    }

    // 本番環境では、ここでログ収集サービスに送信
    // 例: Sentryへの送信
    // if (process.env.NODE_ENV === 'production') {
    //   Sentry.captureMessage('CSP Violation', {
    //     level: 'warning',
    //     extra: report,
    //     tags: {
    //       type: 'csp_violation',
    //     },
    //   });
    // }

    // TODO: データベースやログサービスへの保存を実装
    // 例:
    // await prisma.cspViolation.create({
    //   data: {
    //     report: JSON.stringify(report),
    //     userAgent: request.headers.get('user-agent') || 'unknown',
    //     createdAt: new Date(),
    //   },
    // });

    return NextResponse.json(
      { received: true, message: "CSP report received" },
      { status: 200 }
    );
  } catch (error) {
    console.error("Error processing CSP report:", error);
    return NextResponse.json(
      { received: false, error: "Failed to process report" },
      { status: 500 }
    );
  }
}

// GETリクエストは許可しない
export async function GET() {
  return NextResponse.json(
    { error: "Method not allowed" },
    { status: 405 }
  );
}
