"use client";

import Image from "next/image";
import Link from "next/link";

export type AvatarCircleProps = {
  name: string;
  avatarUrl?: string;
  className?: string;
  /**
   * `next/image` の `sizes`。デフォルトは 40px（h-10 w-10 相当）を想定。
   * 例: "(max-width: 768px) 40px, 48px"
   */
  sizes?: string;
  /**
   * 指定するとユーザーページへのリンクになる
   */
  displayId?: string;
};

export const AvatarCircle = ({
  name,
  avatarUrl,
  className,
  sizes = "40px",
  displayId,
}: AvatarCircleProps) => {
  const initial = (name?.trim()?.[0] ?? "?").toUpperCase();

  const content = (
    <div
      className={[
        "relative overflow-hidden flex items-center justify-center rounded-full bg-white border border-gray-200 text-gray-700 font-bold",
        className ?? "",
      ].join(" ")}
      aria-label={name}
      title={name}
    >
      {avatarUrl ? (
        <Image
          src={avatarUrl}
          alt={name}
          fill
          sizes={sizes}
          className="rounded-full object-cover"
          referrerPolicy="no-referrer"
        />
      ) : (
        <span className="text-xs">{initial}</span>
      )}
    </div>
  );

  if (displayId) {
    return (
      <Link href={`/${displayId}`} className="hover:opacity-80 transition-opacity">
        {content}
      </Link>
    );
  }

  return content;
};
