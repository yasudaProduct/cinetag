"use client";

import { useState } from "react";
import { useSignIn } from "@clerk/nextjs";
import { isClerkAPIResponseError } from "@clerk/nextjs/errors";
import Link from "next/link";
import { Spinner } from "@/components/ui/spinner";
import { GoogleIcon } from "@/components/ui/google-icon";

export default function SignInPage() {
  const { isLoaded, signIn } = useSignIn();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleGoogleSignIn = async () => {
    if (!isLoaded || !signIn) return;

    setErrorMessage(null);
    setIsSubmitting(true);
    try {
      await signIn.authenticateWithRedirect({
        strategy: "oauth_google",
        redirectUrl: "/sso-callback",
        redirectUrlComplete: "/tags",
      });
    } catch (err) {
      if (isClerkAPIResponseError(err)) {
        const clerkMessage =
          err.errors[0]?.longMessage || err.errors[0]?.message;
        setErrorMessage(clerkMessage || "ログインに失敗しました。");
      } else {
        setErrorMessage("ログインに失敗しました。もう一度お試しください。");
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
        <h1 className="text-2xl font-bold text-gray-900">ログイン</h1>
        <p className="mt-1 text-sm text-gray-500">
          アカウントにログインしてください
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
        onClick={handleGoogleSignIn}
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
            Googleでログイン
          </>
        )}
      </button>

      <p className="mt-6 text-center text-sm text-gray-600">
        アカウントをお持ちでない方は{" "}
        <Link
          href="/sign-up"
          className="text-blue-600 hover:text-blue-700 font-semibold"
        >
          新規登録
        </Link>
      </p>
    </div>
  );
}

/*
 * === メール/パスワード認証フォーム（将来用） ===
 *
 * import { useState, type FormEvent } from "react";
 * import { useRouter } from "next/navigation";
 * import { Eye, EyeOff, Mail, Lock } from "lucide-react";
 * import { SignInFormSchema } from "@/lib/validation/auth.form";
 * import { getFirstZodErrorMessage } from "@/lib/validation/tag.form";
 *
 * // useSignIn() から setActive も取得する
 * // const { isLoaded, signIn, setActive } = useSignIn();
 * // const router = useRouter();
 *
 * // State:
 * // const [identifier, setIdentifier] = useState("");
 * // const [password, setPassword] = useState("");
 * // const [showPassword, setShowPassword] = useState(false);
 *
 * // Handler:
 * // const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
 * //   e.preventDefault();
 * //   if (!isLoaded || !signIn) return;
 * //   setErrorMessage(null);
 * //   const parsed = SignInFormSchema.safeParse({ identifier, password });
 * //   if (!parsed.success) {
 * //     setErrorMessage(getFirstZodErrorMessage(parsed.error));
 * //     return;
 * //   }
 * //   setIsSubmitting(true);
 * //   try {
 * //     const result = await signIn.create({
 * //       strategy: "password",
 * //       identifier: parsed.data.identifier,
 * //       password: parsed.data.password,
 * //     });
 * //     if (result.status === "complete") {
 * //       await setActive({ session: result.createdSessionId });
 * //       router.push("/tags");
 * //     }
 * //   } catch (err) { ... }
 * //   finally { setIsSubmitting(false); }
 * // };
 *
 * // JSX:
 * // <form onSubmit={handleSubmit} className="space-y-4">
 * //   <div>
 * //     <label htmlFor="identifier" className="block text-sm font-medium text-gray-700 mb-1">メールアドレス</label>
 * //     <div className="relative">
 * //       <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
 * //       <input id="identifier" type="email" autoComplete="email" value={identifier}
 * //         onChange={(e) => setIdentifier(e.target.value)} placeholder="mail@example.com"
 * //         className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 pl-11 text-sm ..." />
 * //     </div>
 * //   </div>
 * //   <div>
 * //     <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-1">パスワード</label>
 * //     <div className="relative">
 * //       <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
 * //       <input id="password" type={showPassword ? "text" : "password"} autoComplete="current-password"
 * //         value={password} onChange={(e) => setPassword(e.target.value)} placeholder="8文字以上"
 * //         className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 pl-11 pr-11 text-sm ..." />
 * //       <button type="button" onClick={() => setShowPassword(!showPassword)} tabIndex={-1}
 * //         className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
 * //         {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
 * //       </button>
 * //     </div>
 * //   </div>
 * //   <button type="submit" disabled={!isLoaded || isSubmitting}
 * //     className="w-full rounded-xl bg-blue-500 px-6 py-3 text-sm font-bold text-white hover:bg-blue-600 ...">
 * //     {isSubmitting ? (<><Spinner size="sm" />ログイン中...</>) : "ログイン"}
 * //   </button>
 * // </form>
 */
