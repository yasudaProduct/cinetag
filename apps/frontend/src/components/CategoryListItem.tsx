"use client";

import Image from "next/image";
import Link from "next/link";
import { Film, Heart } from "lucide-react";

interface CategoryListItemProps {
  title: string;
  description: string;
  author: string;
  authorDisplayId?: string;
  movieCount: number;
  likes: string | number;
  images: string[];
  href?: string;
}

export const CategoryListItem = ({
  title,
  description,
  author,
  authorDisplayId,
  movieCount,
  likes,
  images = [],
  href,
}: CategoryListItemProps) => {
  const safeImages = Array.isArray(images) ? images : [];

  const AuthorName = authorDisplayId ? (
    <Link
      href={`/${authorDisplayId}`}
      className="text-pink-500 font-medium hover:underline"
    >
      {author}
    </Link>
  ) : (
    <span className="text-pink-500 font-medium">{author}</span>
  );

  const ImageGrid = (
    <div className="flex-shrink-0 w-28 h-28 rounded-md overflow-hidden grid grid-cols-2 grid-rows-2 gap-0.5 bg-gray-100 border border-gray-100">
      {Array.from({ length: 4 }).map((_, index) => {
        const src = safeImages[index];
        return (
          <div key={index} className="relative w-full h-full bg-gray-200">
            {src ? (
              <Image
                src={
                  src.startsWith("http")
                    ? src
                    : "https://image.tmdb.org/t/p/w200" + src
                }
                alt={`Movie poster ${index + 1}`}
                fill
                className="object-cover"
                sizes="56px"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-gray-400">
                <Film className="w-4 h-4 opacity-20" />
              </div>
            )}
          </div>
        );
      })}
    </div>
  );

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-sm hover:shadow-md transition-shadow duration-200 overflow-hidden">
      <div className="flex gap-4 p-4 h-full">
        {/* Left Side: Image Grid (Linked) */}
        {href ? (
          <Link href={href} className="block flex-shrink-0">
            {ImageGrid}
          </Link>
        ) : (
          <div className="flex-shrink-0">{ImageGrid}</div>
        )}

        {/* Right Side: Content */}
        <div className="flex flex-col flex-grow min-w-0 justify-between py-0.5">
          <div>
            {/* Title (Linked) */}
            {href ? (
              <Link href={href} className="block group">
                <h3 className="font-bold text-gray-900 line-clamp-1 mb-1 text-base group-hover:text-blue-600 transition-colors">
                  {title}
                </h3>
              </Link>
            ) : (
              <h3 className="font-bold text-gray-900 line-clamp-1 mb-1 text-base">
                {title}
              </h3>
            )}
            
            {/* Description (Linked) */}
             {href ? (
              <Link href={href} className="block">
                <p className="text-xs text-gray-500 line-clamp-2 leading-relaxed mb-2 hover:text-gray-700">
                  {description}
                </p>
              </Link>
            ) : (
              <p className="text-xs text-gray-500 line-clamp-2 leading-relaxed mb-2">
                {description}
              </p>
            )}
          </div>

          <div>
            <div className="text-xs text-pink-500 mb-1.5">by {AuthorName}</div>
            <div className="flex items-center gap-4 text-xs text-gray-500 font-medium">
              <div className="flex items-center gap-1.5">
                <Film className="w-3.5 h-3.5 text-blue-500" />
                <span>{movieCount}</span>
              </div>
              <div className="flex items-center gap-1.5">
                <Heart className="w-3.5 h-3.5 text-pink-500" />
                <span>{likes}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
