import { getMockTagDetail, type TagDetail } from "@/lib/mock/tagDetail";
import { TagDetailResponseSchema } from "@/lib/validation/tag.api";
import { getPublicApiBase, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export async function getTagDetail(
  tagId: string,
  options?: { token?: string }
): Promise<TagDetail> {
  const base = getPublicApiBase();
  if (!base) return getMockTagDetail(tagId);

  const url = `${base}/api/v1/tags/${encodeURIComponent(tagId)}`;
  const res = await fetch(url, {
    method: "GET",
    headers: options?.token ? { Authorization: `Bearer ${options.token}` } : undefined,
  });

  if (!res.ok) {
    // 未実装/開発中の想定（404）はモックにフォールバック
    if (res.status === 404) return getMockTagDetail(tagId);

    const body = await safeJson(res);
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "タグ詳細の取得に失敗しました",
      })
    );
  }

  const body = await safeJson(res);
  const parsed = TagDetailResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid tag detail response:", parsed.error, body);
    throw new Error("タグ詳細レスポンスの形式が不正です。");
  }

  console.log("tag detail response:", parsed.data);

  return { ...parsed.data, id: parsed.data.id || tagId };
}


