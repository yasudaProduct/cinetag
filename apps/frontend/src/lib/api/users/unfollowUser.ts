import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";

/**
 * 指定ユーザーをアンフォローする
 * @param displayId アンフォロー対象のユーザーのdisplay_id
 * @param token 認証トークン（必須）
 */
export async function unfollowUser(
  displayId: string,
  token: string
): Promise<void> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(
    `${base}/api/v1/users/${encodeURIComponent(displayId)}/follow`,
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
        fallback: "アンフォローに失敗しました",
      })
    );
  }
}
