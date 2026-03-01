import { UnreadCountResponseSchema } from "@/lib/validation/notification.api";
import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";

export async function getUnreadCount(token: string): Promise<number> {
  const base = getPublicApiBaseOrThrow();

  const res = await fetch(`${base}/api/v1/notifications/unread-count`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const body = await safeJson(res);

  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "未読通知数の取得に失敗しました",
      }),
    );
  }

  const parsed = UnreadCountResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid unread count response:", parsed.error, body);
    throw new Error("未読通知数レスポンスの形式が不正です。");
  }

  return parsed.data.unreadCount;
}
