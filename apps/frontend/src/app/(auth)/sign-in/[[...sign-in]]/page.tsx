"use client";

import { GoogleSignInForm } from "@/components/GoogleSignInForm";

export default function SignInPage() {
  return (
    <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-bold text-gray-900">ログイン</h1>
        <p className="mt-1 text-sm text-gray-500">
          アカウントにログインしてください
        </p>
      </div>

      <GoogleSignInForm />
    </div>
  );
}

/*
 * === メール/パスワード認証フォーム（将来用） ===
 *
 * import { useState, type FormEvent } from "react";
 * import { useRouter } from "next/navigation";
 * import { useSignIn } from "@clerk/nextjs";
 * import { isClerkAPIResponseError } from "@clerk/nextjs/errors";
 * import { Eye, EyeOff, Mail, Lock } from "lucide-react";
 * import { Spinner } from "@/components/ui/spinner";
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
 * // const [errorMessage, setErrorMessage] = useState<string | null>(null);
 * // const [isSubmitting, setIsSubmitting] = useState(false);
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
