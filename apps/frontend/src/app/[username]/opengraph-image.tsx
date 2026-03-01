import { ImageResponse } from "next/og";
import { getUserByDisplayId } from "@/lib/api/users/getUser";
import { getNotoSansJPBold } from "@/lib/og-font";

export const revalidate = 600;
export const runtime = "edge";
export const alt = "ユーザープロフィール";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function OGImage({
  params,
}: {
  params: Promise<{ username: string }>;
}) {
  const { username } = await params;

  let displayName = username;
  let bio = "";
  let avatarUrl = "";

  try {
    const user = await getUserByDisplayId(username);
    displayName = user.display_name;
    bio = user.bio || "";
    avatarUrl = user.avatar_url || "";
  } catch {
    // fallback to defaults
  }

  const notoSansJP = await getNotoSansJPBold();

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          padding: "60px",
          background: "linear-gradient(135deg, #FFF9F3 0%, #F3E8FF 50%, #DBEAFE 100%)",
          fontFamily: "'Noto Sans JP', sans-serif",
        }}
      >
        {avatarUrl ? (
           
          <img
            alt=""
            src={avatarUrl}
            width={120}
            height={120}
            style={{
              borderRadius: "60px",
              marginBottom: "24px",
              objectFit: "cover",
            }}
          />
        ) : (
          <div
            style={{
              width: "120px",
              height: "120px",
              borderRadius: "60px",
              background: "linear-gradient(135deg, #3B82F6, #8B5CF6)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              marginBottom: "24px",
              color: "white",
              fontSize: "48px",
              fontWeight: 700,
            }}
          >
            {displayName.charAt(0).toUpperCase()}
          </div>
        )}
        <div
          style={{
            fontSize: "56px",
            fontWeight: 700,
            color: "#111827",
            marginBottom: "8px",
          }}
        >
          {displayName}
        </div>
        <div
          style={{
            fontSize: "24px",
            color: "#9CA3AF",
            marginBottom: "16px",
          }}
        >
          @{username}
        </div>
        {bio && (
          <div
            style={{
              fontSize: "24px",
              color: "#6B7280",
              textAlign: "center",
              maxWidth: "800px",
              lineHeight: 1.4,
            }}
          >
            {bio.slice(0, 80)}
          </div>
        )}
        <div
          style={{
            position: "absolute",
            bottom: "40px",
            right: "60px",
            fontSize: "32px",
            fontWeight: 700,
            color: "#3B82F6",
          }}
        >
          cinetag
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
