import type { MetadataRoute } from "next";
import { listTags } from "@/lib/api/tags/list";

export const revalidate = 3600;

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const siteUrl =
    process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

  const staticPages: MetadataRoute.Sitemap = [
    {
      url: siteUrl,
      changeFrequency: "daily",
      priority: 1.0,
    },
    {
      url: `${siteUrl}/tags`,
      changeFrequency: "daily",
      priority: 0.9,
    },
    {
      url: `${siteUrl}/terms`,
      changeFrequency: "yearly",
      priority: 0.3,
    },
    {
      url: `${siteUrl}/privacy`,
      changeFrequency: "yearly",
      priority: 0.3,
    },
  ];

  const tagPages: MetadataRoute.Sitemap = [];
  const userDisplayIds = new Set<string>();
  const pageSize = 100;
  let page = 1;

  try {
    while (true) {
      const result = await listTags({ page, pageSize });

      for (const tag of result.items) {
        tagPages.push({
          url: `${siteUrl}/tags/${tag.id}`,
          lastModified: tag.createdAt ? new Date(tag.createdAt) : undefined,
          changeFrequency: "weekly",
          priority: 0.8,
        });

        if (tag.authorDisplayId) {
          userDisplayIds.add(tag.authorDisplayId);
        }
      }

      if (result.items.length < pageSize) break;
      page++;
    }
  } catch (error) {
    console.error("sitemap: タグ一覧の取得に失敗しました:", error);
  }

  const userPages: MetadataRoute.Sitemap = Array.from(userDisplayIds).map(
    (displayId) => ({
      url: `${siteUrl}/${displayId}`,
      changeFrequency: "weekly" as const,
      priority: 0.6,
    }),
  );

  return [...staticPages, ...tagPages, ...userPages];
}
