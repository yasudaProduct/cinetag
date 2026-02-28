import type { Metadata } from "next";
import { getTagDetail } from "@/lib/api/tags/detail";
import { TagDetailClient } from "./_components/TagDetailClient";

// ISR: 5分ごとに再生成
export const revalidate = 300;

const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ tagId: string }>;
}): Promise<Metadata> {
  const { tagId } = await params;

  try {
    const tag = await getTagDetail(tagId);
    const title = `${tag.title} | cinetag`;
    const description =
      tag.description || `「${tag.title}」タグの映画一覧 - cinetag`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        type: "website",
      },
      twitter: {
        card: "summary",
        title,
        description,
      },
    };
  } catch {
    return {
      title: "タグ詳細 | cinetag",
    };
  }
}

export default async function TagDetailPage({
  params,
}: {
  params: Promise<{ tagId: string }>;
}) {
  const { tagId } = await params;

  let tagTitle = "";
  try {
    const tag = await getTagDetail(tagId);
    tagTitle = tag.title;
  } catch {
    // ignore - client will handle loading
  }

  const breadcrumbJsonLd = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: [
      {
        "@type": "ListItem",
        position: 1,
        name: "ホーム",
        item: siteUrl,
      },
      {
        "@type": "ListItem",
        position: 2,
        name: "タグ一覧",
        item: `${siteUrl}/tags`,
      },
      ...(tagTitle
        ? [
            {
              "@type": "ListItem",
              position: 3,
              name: tagTitle,
              item: `${siteUrl}/tags/${tagId}`,
            },
          ]
        : []),
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
      />
      <TagDetailClient tagId={tagId} />
    </>
  );
}
