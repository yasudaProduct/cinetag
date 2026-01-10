import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { z } from "zod";
import { UserProfileSchema } from "./getUser";

// フォロワー一覧レスポンスのスキーマ
export const FollowersListResponseSchema = z.object({
  items: z.array(UserProfileSchema),
  page: z.number(),
  page_size: z.number(),
  total_count: z.number(),
});

export type FollowersListResponse = z.infer<typeof FollowersListResponseSchema>;

/**
 * ユーザーをフォローしているユーザー一覧を取得する
 * @param displayId ユーザーのdisplay_id
 * @param page ページ番号（デフォルト: 1）
 * @param pageSize ページサイズ（デフォルト: 20）
 */
export async function listFollowers(
  displayId: string,
  page: number = 1,
  pageSize: number = 20
): Promise<FollowersListResponse> {
  const base = getPublicApiBaseOrThrow();

  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  });

  const res = await fetch(
    `${base}/api/v1/users/${encodeURIComponent(displayId)}/followers?${params}`,
    {
      method: "GET",
    }
  );

  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "フォロワーの取得に失敗しました",
      })
    );
  }

  const parsed = FollowersListResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("フォロワーの形式が不正です");
  }

  return parsed.data;
}
