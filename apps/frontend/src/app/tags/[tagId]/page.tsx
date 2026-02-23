import type { Metadata } from "next";
import { getTagDetail } from "@/lib/api/tags/detail";
import { TagDetailClient } from "./_components/TagDetailClient";

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
  return <TagDetailClient tagId={tagId} />;
}
