export type AddMoviePolicy = "everyone" | "owner_only";

export type TagDetail = {
  id: string;
  title: string;
  description: string;
  author: {
    name: string;
    avatarUrl?: string;
  };
  owner: {
    id: string;
    name: string;
    avatarUrl?: string;
  };
  canEdit: boolean;
  canAddMovie: boolean;
  addMoviePolicy: AddMoviePolicy;
  participantCount: number;
  participants: { name: string }[];
};

export type TagMovie = {
  id: string;
  title: string;
  year: number;
  posterUrl?: string;
};

const DEFAULT_MOVIES: Omit<TagMovie, "id">[] = [
  { title: "Blade Runner 2049", year: 2017, posterUrl: "https://placehold.co/360x540/png?text=Blade+Runner+2049" },
  { title: "2001: A Space Odyssey", year: 1968, posterUrl: "https://placehold.co/360x540/png?text=2001%3A+A+Space+Odyssey" },
  { title: "Dune", year: 2021, posterUrl: "https://placehold.co/360x540/png?text=Dune" },
  { title: "Arrival", year: 2016, posterUrl: "https://placehold.co/360x540/png?text=Arrival" },
  { title: "Annihilation", year: 2018, posterUrl: "https://placehold.co/360x540/png?text=Annihilation" },
  { title: "Ex Machina", year: 2014, posterUrl: "https://placehold.co/360x540/png?text=Ex+Machina" },
  { title: "Interstellar", year: 2014, posterUrl: "https://placehold.co/360x540/png?text=Interstellar" },
  { title: "Children of Men", year: 2006, posterUrl: "https://placehold.co/360x540/png?text=Children+of+Men" },
  { title: "Her", year: 2013, posterUrl: "https://placehold.co/360x540/png?text=Her" },
  { title: "The Matrix", year: 1999, posterUrl: "https://placehold.co/360x540/png?text=The+Matrix" },
  { title: "Minority Report", year: 2002, posterUrl: "https://placehold.co/360x540/png?text=Minority+Report" },
  { title: "Solaris", year: 1972, posterUrl: "https://placehold.co/360x540/png?text=Solaris" },
  { title: "Moon", year: 2009, posterUrl: "https://placehold.co/360x540/png?text=Moon" },
  { title: "Gattaca", year: 1997, posterUrl: "https://placehold.co/360x540/png?text=Gattaca" },
  { title: "Edge of Tomorrow", year: 2014, posterUrl: "https://placehold.co/360x540/png?text=Edge+of+Tomorrow" },
];

function stableHash(input: string): number {
  let h = 0;
  for (let i = 0; i < input.length; i++) h = (h * 31 + input.charCodeAt(i)) >>> 0;
  return h;
}

export function getMockTagDetail(tagId: string): TagDetail {
  const h = stableHash(tagId);
  const title = h % 3 === 0 ? "SF映画の映像美" : h % 3 === 1 ? "心を揺さぶる傑作SF" : "夜に観たい静かなSF";
  const description =
    "革新的なビジュアルと息をのむようなカメラワークで知られるSF映画のコレクション。未来を、フレームごとに探検しよう。";
  const participants = ["Aki", "Yuta", "Ema", "Ren", "Hana", "Sora", "Kenta", "Mio"].map((name) => ({ name }));
  return {
    id: tagId,
    title,
    description,
    author: { name: "Eleanor Vance" },
    owner: { id: "", name: "Eleanor Vance", avatarUrl: undefined },
    canEdit: false,
    canAddMovie: true,
    addMoviePolicy: "everyone" as AddMoviePolicy,
    participantCount: 8,
    participants,
  };
}

export function getMockTagMovies(tagId: string): TagMovie[] {
  // tagId によって並びを少しだけ変える（画面を行き来しても毎回同じ）
  const h = stableHash(tagId);
  const rotated = [...DEFAULT_MOVIES.slice(h % DEFAULT_MOVIES.length), ...DEFAULT_MOVIES.slice(0, h % DEFAULT_MOVIES.length)];
  return rotated.map((m, idx) => ({ id: `${tagId}-${idx + 1}`, ...m }));
}


