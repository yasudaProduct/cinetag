import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { z } from "zod";

// フォロー統計のスキーマ
export const FollowStatsSchema = z.object({
  following_count: z.number(),
  followers_count: z.number(),
  is_following: z.boolean(),
});

export type FollowStats = z.infer<typeof FollowStatsSchema>;

/**
 * ユーザーのフォロー統計を取得する
 * @param displayId ユーザーのdisplay_id
 * @param token 認証トークン（オプション）
 */
export async function getFollowStats(
  displayId: string,
  token?: string
): Promise<FollowStats> {
  const base = getPublicApiBaseOrThrow();

  const headers: HeadersInit = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(
    `${base}/api/v1/users/${encodeURIComponent(displayId)}/follow-stats`,
    {
      method: "GET",
      headers,
    }
  );

  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "フォロー統計の取得に失敗しました",
      })
    );
  }

  const parsed = FollowStatsSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("フォロー統計の形式が不正です");
  }

  return parsed.data;
}
