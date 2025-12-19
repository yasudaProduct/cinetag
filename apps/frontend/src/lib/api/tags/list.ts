import { TagsListResponseSchema, type TagListItem } from "@/lib/validation/tag.api";
import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export type TagsList = TagListItem[];

export async function listTags(): Promise<TagsList> {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/tags`, { method: "GET" });
  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "タグ一覧の取得に失敗しました",
      })
    );
  }

  const parsed = TagsListResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid tags list response:", parsed.error, body);
    throw new Error("タグ一覧レスポンスの形式が不正です。");
  }

  return parsed.data.items ?? [];
}


