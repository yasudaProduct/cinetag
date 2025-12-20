import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export type MovieSearchItem = {
  tmdb_movie_id: number;
  title: string;
  original_title?: string | null;
  poster_path?: string | null;
  release_date?: string | null;
  vote_average?: number | null;
};

export type MovieSearchResponse = {
  items: MovieSearchItem[];
  page: number;
  total_count: number;
};

export async function searchMovies(params: { q: string; page?: number }): Promise<MovieSearchResponse> {
  const base = getPublicApiBaseOrThrow();
  const q = params.q.trim();
  const page = params.page ?? 1;

  const url = new URL(`${base}/api/v1/movies/search`);
  url.searchParams.set("q", q);
  url.searchParams.set("page", String(page));

  const res = await fetch(url.toString(), { method: "GET" });
  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "映画検索に失敗しました",
      })
    );
  }

  // 形式は固定で返しているが、念のためガードする
  const data = body as Partial<MovieSearchResponse> | null;
  return {
    items: Array.isArray(data?.items) ? (data.items as MovieSearchItem[]) : [],
    page: typeof data?.page === "number" ? data.page : page,
    total_count: typeof data?.total_count === "number" ? data.total_count : 0,
  };
}


