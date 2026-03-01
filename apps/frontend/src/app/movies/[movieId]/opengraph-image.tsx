import { ImageResponse } from "next/og";
import { getMovieDetail } from "@/lib/api/movies/detail";

export const runtime = "edge";
export const alt = "映画詳細";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function OGImage({
  params,
}: {
  params: Promise<{ movieId: string }>;
}) {
  const { movieId } = await params;
  const tmdbMovieId = Number(movieId);

  let title = "映画詳細";
  let overview = "";
  let releaseYear = "";
  let genres: string[] = [];

  if (!Number.isNaN(tmdbMovieId) && tmdbMovieId > 0) {
    try {
      const movie = await getMovieDetail(tmdbMovieId);
      title = movie.title;
      overview = movie.overview || "";
      releaseYear = movie.releaseDate
        ? new Date(movie.releaseDate).getFullYear().toString()
        : "";
      genres = movie.genres.map((g) => g.name).slice(0, 3);
    } catch {
      // fallback to defaults
    }
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
          background: "linear-gradient(135deg, #FFF9F3 0%, #E0E7FF 50%, #DBEAFE 100%)",
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
            <span style={{ fontSize: "28px" }}>🎬</span>
            <span>映画</span>
            {releaseYear && <span>・{releaseYear}</span>}
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
          {overview && (
            <div
              style={{
                fontSize: "22px",
                color: "#6B7280",
                lineHeight: 1.4,
                overflow: "hidden",
                textOverflow: "ellipsis",
                maxHeight: "64px",
              }}
            >
              {overview.slice(0, 100)}
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
          <div style={{ display: "flex", gap: "12px" }}>
            {genres.map((genre) => (
              <span
                key={genre}
                style={{
                  padding: "6px 16px",
                  background: "rgba(59, 130, 246, 0.1)",
                  borderRadius: "20px",
                  color: "#3B82F6",
                  fontSize: "20px",
                }}
              >
                {genre}
              </span>
            ))}
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
