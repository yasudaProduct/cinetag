import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import {
  TagsListResponseSchema,
  type TagsListResponse,
} from "@/lib/validation/tag.api";

/**
 * ログインユーザーがフォローしているタグ一覧を取得する
 * @param token 認証トークン（必須）
 * @param page ページ番号（デフォルト: 1）
 * @param pageSize 1ページあたり件数（デフォルト: 20）
 * @returns フォロー中のタグ一覧
 */
export async function listFollowingTags(
  token: string,
  page = 1,
  pageSize = 20
): Promise<TagsListResponse> {
  const base = getPublicApiBaseOrThrow();

  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });

  const res = await fetch(`${base}/api/v1/me/following-tags?${params}`, {
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
        fallback: "フォロー中のタグ一覧の取得に失敗しました",
      })
    );
  }

  const parsed = TagsListResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("フォロー中のタグ一覧のレスポンス形式が不正です");
  }

  return parsed.data;
}
