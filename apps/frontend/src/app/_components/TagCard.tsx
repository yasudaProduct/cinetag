import Image from "next/image";
import { Film } from "lucide-react";

export type MarqueeTag = {
  title: string;
  movieCount: number;
  color: string;
  images: string[];
  fallbackPosters?: { title: string; bg: string }[];
};

export function TagCard({ tag }: { tag: MarqueeTag }) {
  const hasRealImages = tag.images.length > 0;

  return (
    <div className="flex items-center gap-3 bg-white border border-gray-200 rounded-2xl pl-5 pr-3 py-3 shadow-sm hover:shadow-md transition-shadow shrink-0 cursor-default">
      <div className={`w-3 h-3 rounded-full ${tag.color} shrink-0`} />
      <span className="text-sm font-bold text-gray-900 whitespace-nowrap">
        {tag.title}
      </span>
      <span className="flex items-center gap-1 text-xs text-gray-400 whitespace-nowrap">
        <Film className="w-3.5 h-3.5" />
        {tag.movieCount}本
      </span>
      {/* Poster thumbnails */}
      <div className="flex -space-x-2 ml-1">
        {hasRealImages
          ? tag.images.map((url, j) => (
              <div
                key={j}
                className="relative w-8 h-12 rounded-md overflow-hidden shadow-sm border border-white/50 shrink-0"
              >
                <Image
                  src={url}
                  alt=""
                  fill
                  className="object-cover"
                  sizes="32px"
                />
              </div>
            ))
          : tag.fallbackPosters?.map((movie, j) => (
              <div
                key={j}
                className={`w-8 h-12 rounded-md bg-gradient-to-br ${movie.bg} shadow-sm border border-white/30 flex items-end overflow-hidden shrink-0`}
                title={movie.title}
              >
                <span className="text-[5px] leading-tight text-white/80 font-medium px-0.5 pb-0.5 line-clamp-2">
                  {movie.title}
                </span>
              </div>
            ))}
      </div>
    </div>
  );
}
