import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";

export async function markAllAsRead(token: string): Promise<void> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(`${base}/api/v1/notifications/read-all`, {
    method: "PATCH",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) {
    const body = await safeJson(res);
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "全通知の既読化に失敗しました",
      }),
    );
  }
}
