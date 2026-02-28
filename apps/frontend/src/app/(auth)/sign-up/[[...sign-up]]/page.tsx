"use client";

import { useState } from "react";
import { useSignUp } from "@clerk/nextjs";
import { isClerkAPIResponseError } from "@clerk/nextjs/errors";
import Link from "next/link";
import { Spinner } from "@/components/ui/spinner";
import { GoogleIcon } from "@/components/ui/google-icon";

export default function SignUpPage() {
  const { isLoaded, signUp } = useSignUp();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleGoogleSignUp = async () => {
    if (!isLoaded || !signUp) return;

    setErrorMessage(null);
    setIsSubmitting(true);
    try {
      await signUp.authenticateWithRedirect({
        strategy: "oauth_google",
        redirectUrl: "/sso-callback",
        redirectUrlComplete: "/tags",
      });
    } catch (err) {
      if (isClerkAPIResponseError(err)) {
        const clerkMessage =
          err.errors[0]?.longMessage || err.errors[0]?.message;
        setErrorMessage(clerkMessage || "アカウントの作成に失敗しました。");
      } else {
        setErrorMessage(
          "アカウントの作成に失敗しました。もう一度お試しください。"
        );
      }
      setIsSubmitting(false);
    }
  };

  if (!isLoaded) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    );
  }

  return (
    <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-bold text-gray-900">新規登録</h1>
        <p className="mt-1 text-sm text-gray-500">
          アカウントを作成してcinetagを始めましょう
        </p>
      </div>

      {errorMessage && (
        <div
          role="alert"
          className="mb-4 rounded-xl bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700"
        >
          {errorMessage}
        </div>
      )}

      <button
        type="button"
        onClick={handleGoogleSignUp}
        disabled={!isLoaded || isSubmitting}
        className="w-full rounded-xl border border-gray-200 bg-white px-6 py-3 text-sm font-bold text-gray-700 hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
      >
        {isSubmitting ? (
          <>
            <Spinner size="sm" />
            リダイレクト中...
          </>
        ) : (
          <>
            <GoogleIcon className="w-5 h-5" />
            Googleで新規登録
          </>
        )}
      </button>

      <p className="mt-6 text-center text-sm text-gray-600">
        既にアカウントをお持ちの方は{" "}
        <Link
          href="/sign-in"
          className="text-blue-600 hover:text-blue-700 font-semibold"
        >
          ログイン
        </Link>
      </p>
    </div>
  );
}

/*
 * === メール/パスワード認証フォーム（将来用） ===
 *
 * 【実装時の留意事項】
 * メール/パスワード認証を有効化する際は、以下の対応が必要:
 *
 * 1. Clerk Dashboard で Bot Protection を有効にする
 *    - Configure → Attack Protection → Bot Protection
 *    - メール認証ではボットによる大量登録・総当たり攻撃のリスクがあるため必須
 *
 * 2. CSP（Content Security Policy）に Cloudflare Turnstile を許可する
 *    - next.config.ts の script-src に https://challenges.cloudflare.com を追加
 *    - next.config.ts の frame-src に https://challenges.cloudflare.com を追加
 *
 * import { useState, type FormEvent } from "react";
 * import { useRouter } from "next/navigation";
 * import { Eye, EyeOff, Mail, Lock, ShieldCheck } from "lucide-react";
 * import { SignUpFormSchema, VerificationCodeSchema } from "@/lib/validation/auth.form";
 * import { getFirstZodErrorMessage } from "@/lib/validation/tag.form";
 *
 * // useSignUp() から setActive も取得する
 * // const { isLoaded, signUp, setActive } = useSignUp();
 * // const router = useRouter();
 *
 * // State:
 * // const [pendingVerification, setPendingVerification] = useState(false);
 * // const [emailAddress, setEmailAddress] = useState("");
 * // const [password, setPassword] = useState("");
 * // const [confirmPassword, setConfirmPassword] = useState("");
 * // const [showPassword, setShowPassword] = useState(false);
 * // const [verificationCode, setVerificationCode] = useState("");
 *
 * // ステップ1 Handler:
 * // const handleSignUp = async (e: FormEvent<HTMLFormElement>) => {
 * //   e.preventDefault();
 * //   if (!isLoaded || !signUp) return;
 * //   setErrorMessage(null);
 * //   const parsed = SignUpFormSchema.safeParse({ emailAddress, password, confirmPassword });
 * //   if (!parsed.success) { setErrorMessage(getFirstZodErrorMessage(parsed.error)); return; }
 * //   setIsSubmitting(true);
 * //   try {
 * //     await signUp.create({ emailAddress: parsed.data.emailAddress, password: parsed.data.password });
 * //     await signUp.prepareEmailAddressVerification({ strategy: "email_code" });
 * //     setPendingVerification(true);
 * //   } catch (err) { ... }
 * //   finally { setIsSubmitting(false); }
 * // };
 *
 * // ステップ2 Handler:
 * // const handleVerification = async (e: FormEvent<HTMLFormElement>) => {
 * //   e.preventDefault();
 * //   if (!isLoaded || !signUp) return;
 * //   setErrorMessage(null);
 * //   const parsed = VerificationCodeSchema.safeParse({ code: verificationCode });
 * //   if (!parsed.success) { setErrorMessage(getFirstZodErrorMessage(parsed.error)); return; }
 * //   setIsSubmitting(true);
 * //   try {
 * //     const result = await signUp.attemptEmailAddressVerification({ code: parsed.data.code });
 * //     if (result.status === "complete") {
 * //       await setActive({ session: result.createdSessionId });
 * //       router.push("/tags");
 * //     }
 * //   } catch (err) { ... }
 * //   finally { setIsSubmitting(false); }
 * // };
 *
 * // 再送信 Handler:
 * // const handleResendCode = async () => {
 * //   if (!isLoaded || !signUp) return;
 * //   setErrorMessage(null);
 * //   try { await signUp.prepareEmailAddressVerification({ strategy: "email_code" }); }
 * //   catch { setErrorMessage("コードの再送信に失敗しました。"); }
 * // };
 *
 * // ステップ2 JSX (pendingVerification === true の場合):
 * // <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
 * //   <div className="mb-6 text-center">
 * //     <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-blue-50">
 * //       <ShieldCheck className="h-6 w-6 text-blue-500" />
 * //     </div>
 * //     <h1 className="text-2xl font-bold text-gray-900">メール認証</h1>
 * //     <p className="mt-1 text-sm text-gray-500">{emailAddress} に送信された6桁のコードを入力してください</p>
 * //   </div>
 * //   <form onSubmit={handleVerification} className="space-y-4">
 * //     <input id="code" type="text" inputMode="numeric" autoComplete="one-time-code" maxLength={6}
 * //       value={verificationCode} onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, ""))}
 * //       placeholder="000000" className="w-full rounded-xl border ... text-center text-lg font-mono tracking-widest ..." />
 * //     <button type="submit" ...>認証する</button>
 * //   </form>
 * //   <p className="mt-4 text-center text-sm text-gray-500">
 * //     コードが届かない場合は <button onClick={handleResendCode}>再送信</button>
 * //   </p>
 * // </div>
 *
 * // ステップ1 JSX (登録フォーム):
 * // <form onSubmit={handleSignUp} className="space-y-4">
 * //   メールアドレス入力 + パスワード入力 + パスワード確認入力 + 送信ボタン
 * // </form>
 */
