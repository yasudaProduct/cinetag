import Image from "next/image";
import Link from "next/link";
import { Film, ThumbsUp } from "lucide-react";

interface CategoryCardProps {
  title: string;
  description: string;
  author: string;
  authorDisplayId?: string;
  movieCount: number;
  likes: string | number;
  images: string[];
  href?: string;
}

export const CategoryCard = ({
  title,
  description,
  author,
  authorDisplayId,
  movieCount,
  likes,
  images,
  href,
}: CategoryCardProps) => {
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

  // 上半分（画像〜description）をリンクにするコンテンツ
  const LinkableContent = (
    <>
      {/* Image Grid */}
      <div className="grid grid-cols-2 gap-2 mb-4 aspect-square rounded-lg overflow-hidden">
        {images.slice(0, 4).map((src, index) => (
          <div key={index} className="relative w-full h-full bg-gray-100">
            {src ? (
              <Image
                src={
                  src.startsWith("http")
                    ? src
                    : "https://image.tmdb.org/t/p/w500" + src
                }
                alt={`Movie poster ${index + 1}`}
                fill
                className="object-cover"
                sizes="(max-width: 640px) 45vw, (max-width: 768px) 22vw, 150px"
              />
            ) : (
              <div className="w-full h-full bg-gray-200 flex items-center justify-center text-gray-400 text-xs">
                No Image
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Title & Description */}
      <h3 className="font-bold text-lg text-gray-900 mb-1 line-clamp-1">
        {title}
      </h3>
      <p className="text-sm text-gray-500 mb-4 line-clamp-2">{description}</p>
    </>
  );

  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4 shadow-sm hover:shadow-md transition-shadow duration-200 flex flex-col h-full">
      {/* 上半分: タグ詳細へのリンク */}
      {href ? (
        <Link href={href} className="block flex-grow">
          {LinkableContent}
        </Link>
      ) : (
        <div className="flex-grow">{LinkableContent}</div>
      )}

      {/* 下半分: 作成者とステータス（リンク外） */}
      <div className="mt-auto">
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-600">by {AuthorName}</div>
        </div>

        <div className="flex items-center gap-4 mt-3 pt-3 border-t border-gray-100 text-sm text-gray-600 font-medium">
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
  );
};
