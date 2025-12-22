import { TagsListResponseSchema, type TagListItem } from "@/lib/validation/tag.api";
import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export type TagsList = TagListItem[];

export type ListTagsResult = {
  items: TagsList;
  totalCount: number;
  page: number;
  pageSize: number;
};

export type ListTagsParams = {
  q?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
};

export async function listTags(params?: ListTagsParams): Promise<ListTagsResult> {
  const base = getPublicApiBaseOrThrow();
  const url = new URL(`${base}/api/v1/tags`);

  if (params?.q) {
    url.searchParams.set("q", params.q);
  }
  if (params?.sort) {
    url.searchParams.set("sort", params.sort);
  }
  if (params?.page) {
    url.searchParams.set("page", params.page.toString());
  }
  if (params?.pageSize) {
    url.searchParams.set("page_size", params.pageSize.toString());
  }

  const res = await fetch(url.toString(), { method: "GET" });
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

  return {
    items: parsed.data.items ?? [],
    totalCount: parsed.data.totalCount ?? 0,
    page: parsed.data.page ?? 1,
    pageSize: parsed.data.pageSize ?? 20,
  };
}


