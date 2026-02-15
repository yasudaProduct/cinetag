import Image from "next/image";
import Link from "next/link";
import { Trash2 } from "lucide-react";

export type MoviePosterCardProps = {
  title: string;
  year?: number;
  posterUrl?: string;
  href?: string;
  onDelete?: () => void;
  isDeleting?: boolean;
};

export const MoviePosterCard = ({ title, year, posterUrl, href, onDelete, isDeleting }: MoviePosterCardProps) => {
  const src = posterUrl || `https://placehold.co/360x540/png?text=${encodeURIComponent(title)}`;

  const Wrapper = href
    ? ({ children, className }: { children: React.ReactNode; className?: string }) => (
        <Link href={href} className={className}>
          {children}
        </Link>
      )
    : ({ children, className }: { children: React.ReactNode; className?: string }) => (
        <div className={className}>{children}</div>
      );

  return (
    <Wrapper className="group relative">
      <div className="relative w-full aspect-[2/3] rounded-2xl overflow-hidden shadow-sm bg-white border border-gray-200 group-hover:shadow-md transition-shadow">
        <Image
          src={src}
          alt={`${title} poster`}
          fill
          className="object-cover"
          sizes="(max-width: 640px) 45vw, (max-width: 1024px) 22vw, 180px"
        />
        {onDelete && (
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onDelete();
            }}
            disabled={isDeleting}
            className="absolute top-2 right-2 p-2 rounded-full bg-red-500 text-white opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-600 disabled:bg-gray-400 disabled:cursor-not-allowed shadow-md"
            aria-label="映画を削除"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        )}
      </div>
      <div className="mt-3">
        <div className="font-bold text-gray-900 text-sm line-clamp-1">{title}</div>
        {typeof year === "number" && year > 0 && (
          <div className="text-xs text-gray-500 mt-0.5">{year}</div>
        )}
      </div>
    </Wrapper>
  );
};
