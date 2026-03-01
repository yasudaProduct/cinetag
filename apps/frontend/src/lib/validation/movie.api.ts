import { z } from "zod";

// GET /api/v1/movies/:tmdbMovieId のレスポンス

const GenreItemSchema = z.object({
  id: z.number(),
  name: z.string(),
});

const ProductionCountrySchema = z.object({
  iso_3166_1: z.string(),
  name: z.string(),
});

const CastMemberSchema = z.object({
  name: z.string(),
  character: z.string(),
});

export const MovieDetailResponseSchema = z
  .object({
    tmdb_movie_id: z.number(),
    title: z.string(),
    original_title: z.string().nullable().optional(),
    poster_path: z.string().nullable().optional(),
    release_date: z.string().nullable().optional(),
    vote_average: z.number().nullable().optional(),
    vote_count: z.number().int().nullable().optional(),
    overview: z.string().nullable().optional(),
    genres: z.array(GenreItemSchema).default([]),
    runtime: z.number().nullable().optional(),
    production_countries: z.array(ProductionCountrySchema).default([]),
    directors: z.array(z.string()).default([]),
    cast: z.array(CastMemberSchema).default([]),
  })
  .passthrough()
  .transform((data) => ({
    tmdbMovieId: data.tmdb_movie_id,
    title: data.title,
    originalTitle: data.original_title ?? undefined,
    posterPath: data.poster_path ?? undefined,
    releaseDate: data.release_date ?? undefined,
    voteAverage: data.vote_average ?? undefined,
    voteCount: data.vote_count ?? undefined,
    overview: data.overview ?? undefined,
    genres: data.genres,
    runtime: data.runtime ?? undefined,
    productionCountries: data.production_countries,
    directors: data.directors,
    cast: data.cast,
  }));

export type MovieDetailResponse = z.infer<typeof MovieDetailResponseSchema>;

// GET /api/v1/movies/:tmdbMovieId/tags のレスポンス

const MovieRelatedTagItemSchema = z.object({
  tag_id: z.string(),
  title: z.string(),
  follower_count: z.number().int().nonnegative(),
  movie_count: z.number().int().nonnegative(),
});

export const MovieRelatedTagsResponseSchema = z
  .object({
    items: z.array(MovieRelatedTagItemSchema),
  })
  .passthrough()
  .transform((data) => ({
    items: data.items.map((t) => ({
      tagId: t.tag_id,
      title: t.title,
      followerCount: t.follower_count,
      movieCount: t.movie_count,
    })),
  }));

export type MovieRelatedTagsResponse = z.infer<typeof MovieRelatedTagsResponseSchema>;
export type MovieRelatedTagItem = MovieRelatedTagsResponse["items"][number];
