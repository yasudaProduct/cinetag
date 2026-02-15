import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";
import { MovieDetailResponseSchema, type MovieDetailResponse } from "@/lib/validation/movie.api";

export async function getMovieDetail(tmdbMovieId: number): Promise<MovieDetailResponse> {
  const base = getPublicApiBaseOrThrow();
  const url = `${base}/api/v1/movies/${tmdbMovieId}`;

  const res = await fetch(url, { method: "GET" });

  if (!res.ok) {
    const body = await safeJson(res);
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "映画詳細の取得に失敗しました",
      }),
    );
  }

  const body = await safeJson(res);
  const parsed = MovieDetailResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid movie detail response:", parsed.error, body);
    throw new Error("映画詳細レスポンスの形式が不正です。");
  }

  return parsed.data;
}
