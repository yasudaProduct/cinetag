import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";

/**
 * 指定タグのいいねを解除する
 * @param tagId いいね解除対象のタグID
 * @param token 認証トークン（必須）
 */
export async function unlikeTag(tagId: string, token: string): Promise<void> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(tagId)}/like`,
    {
      method: "DELETE",
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
        fallback: "いいね解除に失敗しました",
      })
    );
  }
}
