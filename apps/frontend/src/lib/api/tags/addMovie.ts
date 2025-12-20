import { safeJson, toApiErrorMessage, getPublicApiBaseOrThrow } from "@/lib/api/_shared/http";

export type AddMovieToTagInput = {
  tmdb_movie_id: number;
  note?: string;
  position?: number;
};

export async function addMovieToTag(params: { tagId: string; token: string; input: AddMovieToTagInput }) {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/tags/${encodeURIComponent(params.tagId)}/movies`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${params.token}`,
    },
    body: JSON.stringify({
      tmdb_movie_id: params.input.tmdb_movie_id,
      note: params.input.note,
      position: params.input.position ?? 0,
    }),
  });

  const body = await safeJson(res);
  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "映画の追加に失敗しました",
      })
    );
  }

  return body;
}


