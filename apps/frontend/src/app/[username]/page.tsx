import { notFound } from "next/navigation";
import UserPageClient from "./UserPageClient";
import {
  getPublicApiBaseOrThrow,
  safeJson,
  toApiErrorMessage,
} from "@/lib/api/_shared/http";
import { UserProfileSchema } from "@/lib/api/users/getUser";

export default async function UserPage({
  params,
}: {
  params: { username: string };
}) {
  const { username } = params;

  const base = getPublicApiBaseOrThrow();
  const res = await fetch(
    `${base}/api/v1/users/${encodeURIComponent(username)}`,
    {
      method: "GET",
      cache: "no-store",
    }
  );

  if (res.status === 404) {
    notFound();
  }

  const body = await safeJson(res);
  if (!res.ok) {
    throw new Error(
      toApiErrorMessage({
        status: res.status,
        body,
        fallback: "ユーザー情報の取得に失敗しました",
      })
    );
  }

  const parsed = UserProfileSchema.safeParse(body);
  if (!parsed.success) {
    throw new Error("ユーザー情報の形式が不正です");
  }

  return (
    <UserPageClient username={username} initialProfileUser={parsed.data} />
  );
}
