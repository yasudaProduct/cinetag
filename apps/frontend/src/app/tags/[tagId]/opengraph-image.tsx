import { ImageResponse } from "next/og";
import { getTagDetail } from "@/lib/api/tags/detail";

export const runtime = "edge";
export const alt = "タグ詳細";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function OGImage({
  params,
}: {
  params: Promise<{ tagId: string }>;
}) {
  const { tagId } = await params;

  let title = "タグ詳細";
  let description = "";
  let movieCount = 0;
  let ownerName = "";

  try {
    const tag = await getTagDetail(tagId);
    title = tag.title;
    description = tag.description || "";
    movieCount = tag.movieCount;
    ownerName = tag.owner?.name || "";
  } catch {
    // fallback to defaults
  }

  const notoSansJP = await fetch(
    "https://fonts.googleapis.com/css2?family=Noto+Sans+JP:wght@700&display=swap",
  ).then(async (css) => {
    const text = await css.text();
    const fontUrl = text.match(
      /src: url\((.+?)\) format\('woff2'\)/,
    )?.[1];
    if (!fontUrl) throw new Error("Font URL not found");
    return fetch(fontUrl).then((res) => res.arrayBuffer());
  });

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          padding: "60px",
          background: "linear-gradient(135deg, #FFF9F3 0%, #FEE2E2 50%, #DBEAFE 100%)",
          fontFamily: "'Noto Sans JP', sans-serif",
        }}
      >
        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "12px",
              color: "#6B7280",
              fontSize: "24px",
            }}
          >
            <span style={{ fontSize: "28px" }}>🏷️</span>
            <span>タグ</span>
          </div>
          <div
            style={{
              fontSize: title.length > 20 ? "48px" : "64px",
              fontWeight: 700,
              color: "#111827",
              lineHeight: 1.2,
              overflow: "hidden",
              textOverflow: "ellipsis",
              display: "-webkit-box",
              WebkitLineClamp: 2,
              WebkitBoxOrient: "vertical",
            }}
          >
            {title}
          </div>
          {description && (
            <div
              style={{
                fontSize: "24px",
                color: "#6B7280",
                lineHeight: 1.4,
                overflow: "hidden",
                textOverflow: "ellipsis",
                maxHeight: "68px",
              }}
            >
              {description.slice(0, 80)}
            </div>
          )}
        </div>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-end",
          }}
        >
          <div style={{ display: "flex", gap: "32px", color: "#6B7280", fontSize: "22px" }}>
            {movieCount > 0 && <span>🎬 {movieCount}本の映画</span>}
            {ownerName && <span>👤 {ownerName}</span>}
          </div>
          <div
            style={{
              fontSize: "32px",
              fontWeight: 700,
              color: "#3B82F6",
            }}
          >
            cinetag
          </div>
        </div>
      </div>
    ),
    {
      ...size,
      fonts: [
        {
          name: "Noto Sans JP",
          data: notoSansJP,
          style: "normal",
          weight: 700,
        },
      ],
    },
  );
}
