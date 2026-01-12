import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";

/**
 * 指定タグをフォローする
 * @param tagId フォロー対象のタグID
 * @param token 認証トークン（必須）
 */
export async function followTag(tagId: string, token: string): Promise<void> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(tagId)}/follow`,
    {
      method: "POST",
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
        fallback: "フォローに失敗しました",
      })
    );
  }
}
