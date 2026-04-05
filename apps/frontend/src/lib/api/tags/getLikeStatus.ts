import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import {
  TagLikeStatusResponse,
  TagLikeStatusResponseSchema,
} from "@/lib/validation/tag.api";

/**
 * ユーザーがタグをいいねしているかチェックする
 * @param tagId タグID
 * @param token 認証トークン（必須）
 * @returns いいね状態
 */
export async function getTagLikeStatus(
  tagId: string,
  token: string
): Promise<TagLikeStatusResponse> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(tagId)}/like-status`,
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
        fallback: "いいね状態の取得に失敗しました",
      })
    );
  }

  const parsed = TagLikeStatusResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("いいね状態のレスポンス形式が不正です");
  }

  return parsed.data;
}
