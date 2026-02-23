import type { Metadata } from "next";
import { getMovieDetail } from "@/lib/api/movies/detail";
import { MovieDetailClient } from "./_components/MovieDetailClient";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ movieId: string }>;
}): Promise<Metadata> {
  const { movieId } = await params;
  const tmdbMovieId = Number(movieId);

  if (Number.isNaN(tmdbMovieId) || tmdbMovieId <= 0) {
    return { title: "映画詳細 | cinetag" };
  }

  try {
    const movie = await getMovieDetail(tmdbMovieId);
    const title = `${movie.title} | cinetag`;
    const description =
      movie.overview || `「${movie.title}」の映画情報 - cinetag`;
    const posterUrl = movie.posterPath
      ? `https://image.tmdb.org/t/p/w500${movie.posterPath}`
      : undefined;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        type: "website",
        ...(posterUrl && { images: [posterUrl] }),
      },
      twitter: {
        card: posterUrl ? "summary_large_image" : "summary",
        title,
        description,
        ...(posterUrl && { images: [posterUrl] }),
      },
    };
  } catch {
    return {
      title: "映画詳細 | cinetag",
    };
  }
}

export default async function MovieDetailPage({
  params,
}: {
  params: Promise<{ movieId: string }>;
}) {
  const { movieId } = await params;
  return <MovieDetailClient movieId={movieId} />;
}
