import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import {
  TagFollowStatusResponse,
  TagFollowStatusResponseSchema,
} from "@/lib/validation/tag.api";

/**
 * ユーザーがタグをフォローしているかチェックする
 * @param tagId タグID
 * @param token 認証トークン（必須）
 * @returns フォロー状態
 */
export async function getTagFollowStatus(
  tagId: string,
  token: string
): Promise<TagFollowStatusResponse> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(tagId)}/follow-status`,
    {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );

  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "フォロー状態の取得に失敗しました",
      })
    );
  }

  const parsed = TagFollowStatusResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("フォロー状態のレスポンス形式が不正です");
  }

  return parsed.data;
}
