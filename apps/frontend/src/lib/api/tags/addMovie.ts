import { safeJson, toApiErrorMessage, getPublicApiBaseOrThrow } from "@/lib/api/_shared/http";

export type MovieInput = {
  tmdb_movie_id: number;
  note?: string;
  position?: number;
};

export type MovieResultItem = {
  tmdb_movie_id: number;
  status: "created" | "already_exists" | "error";
  error?: string;
};

export type AddMoviesResponse = {
  results: MovieResultItem[];
  summary: {
    created: number;
    already_exists: number;
    failed: number;
  };
};

export async function addMoviesToTag(params: {
  tagId: string;
  token: string;
  movies: MovieInput[];
}): Promise<AddMoviesResponse> {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(
    `${base}/api/v1/tags/${encodeURIComponent(params.tagId)}/movies`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${params.token}`,
      },
      body: JSON.stringify({
        movies: params.movies.map((m) => ({
          tmdb_movie_id: m.tmdb_movie_id,
          note: m.note,
          position: m.position ?? 0,
        })),
      }),
    },
  );

  const body = await safeJson(res);

  // 404/403/400/500 等（タグ全体の問題）
  if (!res.ok && res.status !== 207) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "映画の追加に失敗しました",
      }),
    );
  }

  // 201 or 207: 部分成功を含む結果
  return body as AddMoviesResponse;
}
