import { getPublicApiBaseOrThrow, safeJson, toApiErrorMessage } from "@/lib/api/_shared/http";

export async function deleteTag(params: { tagId: string; token: string }) {
    const base = getPublicApiBaseOrThrow();
    const res = await fetch(
        `${base}/api/v1/tags/${encodeURIComponent(params.tagId)}`,
        {
            method: "DELETE",
            headers: { Authorization: `Bearer ${params.token}` },
        }
    );
    if (!res.ok) {
        const body = await safeJson(res);
        throw new Error(
            toApiErrorMessage({
                status: res.status,
                body,
                fallback: "タグの削除に失敗しました",
            })
        );
    }
}