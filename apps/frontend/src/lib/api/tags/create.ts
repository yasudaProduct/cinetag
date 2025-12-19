import { TagCreateResponseSchema } from "@/lib/validation/tag.api";
import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export type CreateTagInput = {
  title: string;
  description?: string;
  is_public: boolean;
};

export async function createTag(params: { token: string; input: CreateTagInput }) {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/tags`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${params.token}`,
    },
    body: JSON.stringify(params.input),
  });

  const body = await safeJson(res);
  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "作成に失敗しました",
      })
    );
  }

  const parsedCreated = TagCreateResponseSchema.safeParse(body);
  if (!parsedCreated.success) {
    console.warn("Invalid create tag response:", parsedCreated.error, body);
    throw new Error("作成レスポンスの形式が不正です。");
  }

  return parsedCreated.data;
}


