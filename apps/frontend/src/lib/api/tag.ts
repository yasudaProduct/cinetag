import { getMockTagDetail, getMockTagMovies, type TagDetail, type TagMovie } from "@/lib/mock/tagDetail";

type FetchResult<T> = { ok: true; data: T } | { ok: false; error: string; status?: number };

async function safeJson(res: Response): Promise<unknown> {
  return await res.json().catch(() => null);
}

export async function fetchTagDetail(tagId: string): Promise<FetchResult<TagDetail>> {
  const base = process.env.NEXT_PUBLIC_BACKEND_API_BASE;
  if (!base) {
    return { ok: true, data: getMockTagDetail(tagId) };
  }

  const url = `${base}/api/v1/tags/${encodeURIComponent(tagId)}`;
  try {
    const res = await fetch(url, { method: "GET" });
    if (!res.ok) {
      // 未実装の想定（404）ではモックにフォールバック
      if (res.status === 404) return { ok: true, data: getMockTagDetail(tagId) };
      const body = await safeJson(res);
      return { ok: false, error: `タグ詳細の取得に失敗しました（${res.status}）`, status: res.status };
    }

    const body = await safeJson(res);
    // バックエンド実装が揃ったら zod で厳密にパースする（現時点はモック優先）
    const anyBody = body as any;
    const detail: TagDetail = {
      id: String(anyBody?.id ?? tagId),
      title: String(anyBody?.title ?? ""),
      description: String(anyBody?.description ?? ""),
      author: { name: String(anyBody?.author?.name ?? anyBody?.author ?? "unknown") },
      participantCount: Number(anyBody?.participant_count ?? 0),
      participants: Array.isArray(anyBody?.participants) ? anyBody.participants.map((p: any) => ({ name: String(p?.name ?? "") })) : [],
    };
    return { ok: true, data: detail };
  } catch {
    // ネットワーク等のエラーは開発時に辛いのでモックにフォールバック
    return { ok: true, data: getMockTagDetail(tagId) };
  }
}

export async function fetchTagMovies(tagId: string): Promise<FetchResult<TagMovie[]>> {
  const base = process.env.NEXT_PUBLIC_BACKEND_API_BASE;
  if (!base) {
    return { ok: true, data: getMockTagMovies(tagId) };
  }

  const url = `${base}/api/v1/tags/${encodeURIComponent(tagId)}/movies`;
  try {
    const res = await fetch(url, { method: "GET" });
    if (!res.ok) {
      if (res.status === 404) return { ok: true, data: getMockTagMovies(tagId) };
      return { ok: false, error: `タグの映画一覧の取得に失敗しました（${res.status}）`, status: res.status };
    }

    const body = await safeJson(res);
    const items = Array.isArray((body as any)?.items) ? (body as any).items : Array.isArray(body) ? (body as any) : [];
    const movies: TagMovie[] = items.map((m: any, idx: number) => ({
      id: String(m?.id ?? m?.tmdb_movie_id ?? `${tagId}-${idx + 1}`),
      title: String(m?.title ?? ""),
      year: Number(m?.year ?? m?.release_year ?? 0),
      posterUrl: typeof m?.poster_url === "string" ? m.poster_url : undefined,
    }));
    return { ok: true, data: movies };
  } catch {
    return { ok: true, data: getMockTagMovies(tagId) };
  }
}


