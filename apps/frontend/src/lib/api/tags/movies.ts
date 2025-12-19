import { getMockTagMovies, type TagMovie } from "@/lib/mock/tagDetail";
import { TagMoviesResponseSchema } from "@/lib/validation/tag.api";
import { getPublicApiBase, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

function parseYearFromDate(date: string | null | undefined): number {
  if (!date) return 0;
  const y = Number(date.slice(0, 4));
  return Number.isFinite(y) ? y : 0;
}

function buildTmdbPosterUrl(posterPath: string | null | undefined): string | undefined {
  if (!posterPath) return undefined;
  // 仕様: TMDB の poster_path が来た場合は w400 を採用（必要になったら設定化する）
  return `https://image.tmdb.org/t/p/w400${posterPath}`;
}

export async function listTagMovies(tagId: string): Promise<TagMovie[]> {
  const base = getPublicApiBase();
  if (!base) return getMockTagMovies(tagId);

  const url = `${base}/api/v1/tags/${encodeURIComponent(tagId)}/movies`;
  const res = await fetch(url, { method: "GET" });

  if (!res.ok) {
    if (res.status === 404) return getMockTagMovies(tagId);

    const body = await safeJson(res);
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "タグの映画一覧の取得に失敗しました",
      })
    );
  }

  const body = await safeJson(res);
  const parsed = TagMoviesResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid tag movies response:", parsed.error, body);
    throw new Error("タグ映画一覧レスポンスの形式が不正です。");
  }

  return (parsed.data.items ?? []).map((item, idx) => {
    const idRaw = item.id ?? item.tmdb_movie_id;
    const id = idRaw != null && `${idRaw}`.length > 0 ? `${idRaw}` : `${tagId}-${idx + 1}`;

    const title = item.title ?? item.movie?.title ?? "";

    const year =
      (typeof item.year === "number" ? item.year : undefined) ??
      (typeof item.release_year === "number" ? item.release_year : undefined) ??
      parseYearFromDate(item.movie?.release_date) ??
      0;

    const posterUrl = item.poster_url ?? buildTmdbPosterUrl(item.movie?.poster_path);

    return {
      id,
      title,
      year,
      posterUrl: posterUrl ?? undefined,
    };
  });
}


