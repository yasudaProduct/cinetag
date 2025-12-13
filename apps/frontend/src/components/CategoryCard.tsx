import Image from "next/image";
import { Film, ThumbsUp } from "lucide-react";

interface CategoryCardProps {
  title: string;
  description: string;
  author: string;
  movieCount: number;
  likes: string | number;
  images: string[];
}

export const CategoryCard = ({
  title,
  description,
  author,
  movieCount,
  likes,
  images,
}: CategoryCardProps) => {
  return (
    <div className="bg-white rounded-xl border border-gray-200 p-4 shadow-sm hover:shadow-md transition-shadow duration-200 flex flex-col h-full">
      {/* Image Grid */}
      <div className="grid grid-cols-2 gap-2 mb-4 aspect-square rounded-lg overflow-hidden">
        {images.slice(0, 4).map((src, index) => (
          <div key={index} className="relative w-full h-full bg-gray-100">
            {/* Use a colored div fallback if src is empty, otherwise Image */}
            {src ? (
              <Image
                src={src.startsWith("http") ? src : "https://image.tmdb.org/t/p/w500" + src}
                alt={`Movie poster ${index + 1}`}
                fill
                className="object-cover"
              />
            ) : (
              <div className="w-full h-full bg-gray-200 flex items-center justify-center text-gray-400 text-xs">
                No Image
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Content */}
      <div className="flex flex-col flex-grow">
        <h3 className="font-bold text-lg text-gray-900 mb-1 line-clamp-1">
          {title}
        </h3>
        <p className="text-sm text-gray-500 mb-4 line-clamp-2 flex-grow">
          {description}
        </p>

        <div className="flex items-center justify-between mt-auto">
          <div className="text-sm text-gray-600">
            by <span className="text-pink-500 font-medium">{author}</span>
          </div>
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
