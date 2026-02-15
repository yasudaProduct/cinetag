import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";
import { MovieRelatedTagsResponseSchema, type MovieRelatedTagItem } from "@/lib/validation/movie.api";

export async function getMovieRelatedTags(tmdbMovieId: number): Promise<MovieRelatedTagItem[]> {
  const base = getPublicApiBaseOrThrow();
  const url = `${base}/api/v1/movies/${tmdbMovieId}/tags`;

  const res = await fetch(url, { method: "GET" });

  if (!res.ok) {
    const body = await safeJson(res);
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "関連タグの取得に失敗しました",
      }),
    );
  }

  const body = await safeJson(res);
  const parsed = MovieRelatedTagsResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid movie related tags response:", parsed.error, body);
    throw new Error("関連タグレスポンスの形式が不正です。");
  }

  return parsed.data.items;
}
