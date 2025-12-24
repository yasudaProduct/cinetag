import { TagsListResponseSchema, type TagListItem } from "@/lib/validation/tag.api";
import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export type UserTagsList = TagListItem[];

export type ListUserTagsResult = {
  items: UserTagsList;
  totalCount: number;
  page: number;
  pageSize: number;
};

export type ListUserTagsParams = {
  displayId: string;
  page?: number;
  pageSize?: number;
  /** 認証トークン（任意。本人の非公開タグを取得する場合に必要） */
  token?: string;
};

/**
 * ユーザーが作成したタグ一覧を取得する
 * ログイン状態であれば認証トークンを送信し、本人の非公開タグも取得可能
 */
export async function listUserTags(params: ListUserTagsParams): Promise<ListUserTagsResult> {
  const base = getPublicApiBaseOrThrow();
  const url = new URL(`${base}/api/v1/users/${encodeURIComponent(params.displayId)}/tags`);

  if (params.page) {
    url.searchParams.set("page", params.page.toString());
  }
  if (params.pageSize) {
    url.searchParams.set("page_size", params.pageSize.toString());
  }

  const headers: HeadersInit = {};
  if (params.token) {
    headers["Authorization"] = `Bearer ${params.token}`;
  }

  const res = await fetch(url.toString(), {
    method: "GET",
    headers,
  });
  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "ユーザーのタグ一覧の取得に失敗しました",
      })
    );
  }

  const parsed = TagsListResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid user tags list response:", parsed.error, body);
    throw new Error("ユーザーのタグ一覧レスポンスの形式が不正です。");
  }

  return {
    items: parsed.data.items ?? [],
    totalCount: parsed.data.totalCount ?? 0,
    page: parsed.data.page ?? 1,
    pageSize: parsed.data.pageSize ?? 20,
  };
}
