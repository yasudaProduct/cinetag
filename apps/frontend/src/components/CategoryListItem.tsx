import Image from "next/image";
import Link from "next/link";
import { Film, ThumbsUp } from "lucide-react";

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
  images,
  href,
}: CategoryListItemProps) => {
  const AuthorName = authorDisplayId ? (
    <Link
      href={`/${authorDisplayId}`}
      className="text-pink-500 font-medium hover:underline text-sm"
    >
      {author}
    </Link>
  ) : (
    <span className="text-pink-500 font-medium text-sm">{author}</span>
  );

  const MainContent = (
    <>
      {/* Image Grid - Smaller for list view */}
      <div className="flex-shrink-0 w-24 h-24">
        <div className="grid grid-cols-2 gap-1 w-full h-full rounded-md overflow-hidden bg-gray-100">
          {images.slice(0, 4).map((src, index) => (
            <div key={index} className="relative w-full h-full">
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
                  sizes="48px"
                />
              ) : (
                <div className="w-full h-full bg-gray-200 flex items-center justify-center text-gray-400 text-[10px]">
                  No Image
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="flex-grow min-w-0 flex flex-col justify-center h-full">
        <h3 className="font-bold text-lg text-gray-900 mb-1 line-clamp-1">
          {title}
        </h3>
        <p className="text-sm text-gray-500 line-clamp-2 mb-2">{description}</p>
        
        {/* Mobile only stats row */}
        <div className="flex md:hidden items-center gap-4 text-xs text-gray-600 font-medium mt-1">
           <div className="flex items-center gap-1.5">
            <Film className="w-3.5 h-3.5 text-blue-500" />
            <span>{movieCount}</span>
          </div>
          <div className="flex items-center gap-1.5">
            <ThumbsUp className="w-3.5 h-3.5 text-yellow-500" />
            <span>{likes}</span>
          </div>
        </div>
      </div>
    </>
  );

  return (
    <div className="bg-white rounded-xl border border-gray-200 shadow-sm hover:shadow-md transition-shadow duration-200 overflow-hidden">
      <div className="flex">
        {/* Left Side: Clickable */}
        <div className="flex-grow">
          {href ? (
            <Link href={href} className="flex gap-4 p-4 h-full items-center block">
              {MainContent}
            </Link>
          ) : (
            <div className="flex gap-4 p-4 h-full items-center">
              {MainContent}
            </div>
          )}
        </div>

        {/* Desktop Stats & Author - Right side */}
        <div className="hidden md:flex flex-col items-end justify-center gap-2 pr-4 pl-4 border-l border-gray-100 min-w-[120px] my-2">
          <div className="text-sm text-gray-600">by {AuthorName}</div>
          <div className="flex items-center gap-4 text-sm text-gray-600 font-medium">
            <div className="flex items-center gap-1.5">
              <Film className="w-4 h-4 text-blue-500" />
              <span>{movieCount}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <ThumbsUp className="w-4 h-4 text-yellow-500" />
              <span>{likes}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Mobile Author Footer */}
      <div className="md:hidden px-4 pb-3 pt-0 flex justify-end">
         <div className="text-xs text-gray-600">by {AuthorName}</div>
      </div>
    </div>
  );
};
