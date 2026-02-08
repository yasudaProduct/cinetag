import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { UserProfileSchema, type UserProfile } from "./getUser";

export interface UpdateMeInput {
  display_name?: string;
}

/**
 * 認証済みユーザー自身の情報を更新する
 * @param token 認証トークン（必須）
 * @param input 更新するフィールド
 */
export async function updateMe(
  token: string,
  input: UpdateMeInput
): Promise<UserProfile> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(`${base}/api/v1/users/me`, {
    method: "PATCH",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "ユーザー情報の更新に失敗しました",
      })
    );
  }

  const parsed = UserProfileSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("ユーザー情報の形式が不正です");
  }

  return parsed.data;
}
