export type MovieDetail = {
  id: string;
  title: string;
  originalTitle: string;
  rating: number;
  releaseYear: number;
  runtime: number;
  country: string;
  director: string;
  genres: string[];
  overview: string;
  cast: string[];
  posterUrl: string;
};

export type RelatedTag = {
  id: string;
  title: string;
  followerCount: number;
  movieCount: number;
};

export function getMockMovieDetail(_movieId: string): MovieDetail {
  return {
    id: "godfather-1972",
    title: "ゴッドファーザー",
    originalTitle: "The Godfather",
    rating: 9.2,
    releaseYear: 1972,
    runtime: 175,
    country: "アメリカ",
    director: "フランシス・フォード・コッポラ",
    genres: ["ドラマ", "犯罪"],
    overview:
      "マフィアの世界を描いた映画史に残る傑作。コルレオーネ一族の栄光と悲劇を壮大なスケールで描いた不朽の名作。マーロン・ブランドとアル・パチーノの名演が光る。",
    cast: [
      "マーロン・ブランド",
      "アル・パチーノ",
      "ジェームズ・カーン",
      "ロバート・デュヴァル",
    ],
    posterUrl: "https://placehold.co/360x540/1a1a2e/ffffff/png?text=The+Godfather",
  };
}

export function getMockRelatedTags(_movieId: string): RelatedTag[] {
  return [
    { id: "tag-1", title: "クラシック名作集", followerCount: 156, movieCount: 24 },
    { id: "tag-2", title: "マフィア映画特集", followerCount: 89, movieCount: 12 },
    { id: "tag-3", title: "1970年代の傑作", followerCount: 67, movieCount: 18 },
    { id: "tag-4", title: "アカデミー賞作品", followerCount: 203, movieCount: 45 },
  ];
}
