import { safeJson, toApiErrorMessage, getPublicApiBaseOrThrow } from "@/lib/api/_shared/http";

export async function deleteMovieFromTag(params: { tagId: string; tagMovieId: string; token: string }) {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(params.tagId)}/movies/${encodeURIComponent(params.tagMovieId)}`,
    {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${params.token}`,
      },
    }
  );

  if (!res.ok) {
    const body = await safeJson(res);
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "映画の削除に失敗しました",
      })
    );
  }

  // 204 No Content の場合はボディなし
  return;
}
