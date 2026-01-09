import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { z } from "zod";
import { UserProfileSchema } from "./getUser";

// フォロー一覧レスポンスのスキーマ
export const FollowListResponseSchema = z.object({
  items: z.array(UserProfileSchema),
  page: z.number(),
  page_size: z.number(),
  total_count: z.number(),
});

export type FollowListResponse = z.infer<typeof FollowListResponseSchema>;

/**
 * ユーザーがフォローしているユーザー一覧を取得する
 * @param displayId ユーザーのdisplay_id
 * @param page ページ番号（デフォルト: 1）
 * @param pageSize ページサイズ（デフォルト: 20）
 */
export async function listFollowing(
  displayId: string,
  page: number = 1,
  pageSize: number = 20
): Promise<FollowListResponse> {
  const base = getPublicApiBaseOrThrow();

  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  });

  const res = await fetch(
    `${base}/api/v1/users/${encodeURIComponent(displayId)}/following?${params}`,
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
        fallback: "フォロー中ユーザーの取得に失敗しました",
      })
    );
  }

  const parsed = FollowListResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("フォロー中ユーザーの形式が不正です");
  }

  return parsed.data;
}
