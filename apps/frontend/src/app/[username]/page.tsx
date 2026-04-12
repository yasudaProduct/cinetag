import type { Metadata } from "next";
import { auth } from "@clerk/nextjs/server";
import { notFound } from "next/navigation";
import UserPageClient from "./UserPageClient";
import { getUserByDisplayId } from "@/lib/api/users/getUser";
import { getMe } from "@/lib/api/users/getMe";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ username: string }>;
}): Promise<Metadata> {
  const { username } = await params;

  try {
    const user = await getUserByDisplayId(username);
    const title = `${user.display_name} | cinetag`;
    const description =
      user.bio || `${user.display_name}のプロフィール - cinetag`;

    return {
      title,
      description,
      alternates: {
        canonical: `/${username}`,
      },
      openGraph: {
        title,
        description,
        type: "profile",
        ...(user.avatar_url && { images: [user.avatar_url] }),
      },
      twitter: {
        card: "summary",
        title,
        description,
      },
    };
  } catch {
    return {
      title: "ユーザー | cinetag",
    };
  }
}

async function resolveIsOwnPage(username: string): Promise<boolean> {
  try {
    const { getToken } = await auth();
    const token = await getToken({ template: "cinetag-backend" });
    if (!token) return false;
    const me = await getMe(token);
    return me.display_id === username;
  } catch {
    return false;
  }
}

export default async function UserPage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;
  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

  const [user, isOwnPage] = await Promise.all([
    getUserByDisplayId(username).catch((error: unknown) => {
      if (error instanceof Error && error.message.includes("404")) {
        notFound();
      }
      throw error;
    }),
    resolveIsOwnPage(username),
  ]);

  const profileJsonLd = {
    "@context": "https://schema.org",
    "@type": "ProfilePage",
    mainEntity: {
      "@type": "Person",
      name: user.display_name,
      url: `${siteUrl}/${user.display_id}`,
      ...(user.bio && { description: user.bio }),
      ...(user.avatar_url && { image: user.avatar_url }),
    },
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(profileJsonLd) }}
      />
      <UserPageClient
        username={username}
        initialProfileUser={user}
        isOwnPage={isOwnPage}
      />
    </>
  );
}
