import {
  NotificationListResponseSchema,
  type NotificationListResponse,
} from "@/lib/validation/notification.api";
import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";

export async function listNotifications(
  token: string,
  page: number = 1,
  pageSize: number = 20,
  unreadOnly?: boolean,
): Promise<NotificationListResponse> {
  const base = getPublicApiBaseOrThrow();
  const url = new URL(`${base}/api/v1/notifications`);
  url.searchParams.set("page", page.toString());
  url.searchParams.set("page_size", pageSize.toString());
  if (unreadOnly) {
    url.searchParams.set("unread_only", "true");
  }

  const res = await fetch(url.toString(), {
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
        fallback: "通知一覧の取得に失敗しました",
      }),
    );
  }

  const parsed = NotificationListResponseSchema.safeParse(body);
  if (!parsed.success) {
    console.warn("Invalid notification list response:", parsed.error, body);
    throw new Error("通知一覧レスポンスの形式が不正です。");
  }

  return parsed.data;
}
