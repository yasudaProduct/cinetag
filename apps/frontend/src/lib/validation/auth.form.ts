import { z } from "zod";

/**
 * 認証フォームの zod schema
 */

// --- サインイン ---
export const SignInFormSchema = z.object({
  identifier: z
    .string()
    .trim()
    .min(1, "メールアドレスを入力してください。")
    .email("有効なメールアドレスを入力してください。"),
  password: z.string().min(1, "パスワードを入力してください。"),
});
export type SignInFormInput = z.infer<typeof SignInFormSchema>;

// --- サインアップ ---
export const SignUpFormSchema = z
  .object({
    emailAddress: z
      .string()
      .trim()
      .min(1, "メールアドレスを入力してください。")
      .email("有効なメールアドレスを入力してください。"),
    password: z
      .string()
      .min(8, "パスワードは8文字以上で入力してください。"),
    confirmPassword: z
      .string()
      .min(1, "パスワード（確認）を入力してください。"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "パスワードが一致しません。",
    path: ["confirmPassword"],
  });
export type SignUpFormInput = z.infer<typeof SignUpFormSchema>;

// --- メール認証コード ---
export const VerificationCodeSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "認証コードを入力してください。")
    .regex(/^\d{6}$/, "6桁の数字で入力してください。"),
});
export type VerificationCodeInput = z.infer<typeof VerificationCodeSchema>;
