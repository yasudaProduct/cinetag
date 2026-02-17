"use client";

import { useState } from "react";
import { useSignIn } from "@clerk/nextjs";
import { isClerkAPIResponseError } from "@clerk/nextjs/errors";
import Link from "next/link";
import { Spinner } from "@/components/ui/spinner";
import { GoogleIcon } from "@/components/ui/google-icon";

type GoogleSignInFormProps = {
  onNavigate?: () => void;
};

export function GoogleSignInForm({ onNavigate }: GoogleSignInFormProps) {
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
    <>
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
          onClick={onNavigate}
          className="text-blue-600 hover:text-blue-700 font-semibold"
        >
          新規登録
        </Link>
      </p>
    </>
  );
}
