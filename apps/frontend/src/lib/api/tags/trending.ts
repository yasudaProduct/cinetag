import { TrendingTagsResponseSchema, type TagListItem } from "@/lib/validation/tag.api";
import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export type TrendingPeriod = "week";

export type ListTrendingTagsParams = {
  period?: TrendingPeriod;
  limit?: number;
};

export type TrendingTagsResult = {
  items: TagListItem[];
  period: string;
  limit: number;
};

/**
 * 直近の利用増加（映画追加・フォロー・いいね）が多い公開タグ上位を取得する。
 * - 増加分が無いタグは含まれず、件数が limit に満たない場合は存在分のみ返る。
 */
export async function listTrendingTags(
  params?: ListTrendingTagsParams,
): Promise<TrendingTagsResult> {
  const base = getPublicApiBaseOrThrow();
  const url = new URL(`${base}/api/v1/tags/trending`);

  url.searchParams.set("period", params?.period ?? "week");
  if (params?.limit) {
    url.searchParams.set("limit", params.limit.toString());
  }

  const res = await fetch(url.toString(), { method: "GET" });
  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "今週のタグの取得に失敗しました",
      }),
    );
  }

  const parsed = TrendingTagsResponseSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("今週のタグレスポンスの形式が不正です。");
  }

  return parsed.data;
}
