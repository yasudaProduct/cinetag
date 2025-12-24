import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { UserProfileSchema, type UserProfile } from "./getUser";

/**
 * 認証済みユーザー自身の情報を取得する
 * @param token 認証トークン（必須）
 */
export async function getMe(token: string): Promise<UserProfile> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(`${base}/api/v1/users/me`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "ユーザー情報の取得に失敗しました",
      })
    );
  }

  const parsed = UserProfileSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("ユーザー情報の形式が不正です");
  }

  return parsed.data;
}
