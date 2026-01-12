import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import {
  TagFollowersListResponse,
  TagFollowersListResponseSchema,
} from "@/lib/validation/tag.api";

/**
 * タグのフォロワー一覧を取得する
 * @param tagId タグID
 * @param page ページ番号（デフォルト: 1）
 * @param pageSize 1ページあたり件数（デフォルト: 20）
 * @returns フォロワー一覧
 */
export async function listTagFollowers(
  tagId: string,
  page = 1,
  pageSize = 20
): Promise<TagFollowersListResponse> {
  const base = getPublicApiBaseOrThrow();

  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });

  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(tagId)}/followers?${params}`,
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
        fallback: "フォロワー一覧の取得に失敗しました",
      })
    );
  }

  const parsed = TagFollowersListResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("フォロワー一覧のレスポンス形式が不正です");
  }

  return parsed.data;
}
