import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { z } from "zod";

// ユーザープロフィールのスキーマ
export const UserProfileSchema = z.object({
  id: z.string(),
  display_id: z.string(),
  display_name: z.string(),
  avatar_url: z.string().nullable().optional(),
  bio: z.string().nullable().optional(),
});

export type UserProfile = z.infer<typeof UserProfileSchema>;

/**
 * display_id からユーザー情報を取得する
 */
export async function getUserByDisplayId(
  displayId: string
): Promise<UserProfile> {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/users/${encodeURIComponent(displayId)}`, {
    method: "GET",
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
