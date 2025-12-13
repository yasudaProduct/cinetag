import Image from "next/image";

export type MoviePosterCardProps = {
  title: string;
  year?: number;
  posterUrl?: string;
};

export const MoviePosterCard = ({ title, year, posterUrl }: MoviePosterCardProps) => {
  const src = posterUrl || `https://placehold.co/360x540/png?text=${encodeURIComponent(title)}`;
  return (
    <div className="group">
      <div className="relative w-full aspect-[2/3] rounded-2xl overflow-hidden shadow-sm bg-white border border-gray-200 group-hover:shadow-md transition-shadow">
        <Image
          src={src}
          alt={`${title} poster`}
          fill
          className="object-cover"
          sizes="(max-width: 640px) 45vw, (max-width: 1024px) 22vw, 180px"
        />
      </div>
      <div className="mt-3">
        <div className="font-bold text-gray-900 text-sm line-clamp-1">{title}</div>
        {typeof year === "number" && year > 0 && (
          <div className="text-xs text-gray-500 mt-0.5">{year}</div>
        )}
      </div>
    </div>
  );
};


