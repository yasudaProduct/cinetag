import { notFound } from "next/navigation";
import UserPageClient from "./UserPageClient";
import { getUserByDisplayId } from "@/lib/api/users/getUser";

export default async function UserPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;

  let user;
  try {
    user = await getUserByDisplayId(username, { cache: "no-store" });
  } catch (error) {
    if (error instanceof Error && error.message.includes("404")) {
      notFound();
    }
    throw error;
  }

  return <UserPageClient username={username} initialProfileUser={user} />;
}
