import { safeJson, toApiErrorMessage, getPublicApiBaseOrThrow } from "@/lib/api/_shared/http";
import type { AddMoviePolicy } from "@/lib/validation/tag.api";

export type UpdateTagInput = {
  title?: string;
  description?: string | null;
  cover_image_url?: string | null;
  is_public?: boolean;
  add_movie_policy?: AddMoviePolicy;
};

export async function updateTag(params: { tagId: string; token: string; input: UpdateTagInput }) {
  const base = getPublicApiBaseOrThrow();
  const res = await fetch(`${base}/api/v1/tags/${encodeURIComponent(params.tagId)}`, {
    method: "PATCH",
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
        fallback: "タグの更新に失敗しました",
      })
    );
  }

  return body;
}


