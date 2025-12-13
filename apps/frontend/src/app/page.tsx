"use client";

import { Header } from "@/components/Header";
import { CategoryCard } from "@/components/CategoryCard";
import { Search, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useState } from "react";
import { TagCreateModal, CreatedTagForList } from "@/components/TagCreateModal";
import { SignedIn, SignedOut, SignInButton } from "@clerk/nextjs";

// Mock Data
const MOCK_CATEGORIES = [
  {
    id: 1,
    title: "ジブリの名作",
    description: "スタジオジブリの不朽の名作アニメ。",
    author: "cinephile_jane",
    movieCount: 32,
    likes: "1.2k",
    images: [
      "https://placehold.co/400x600/orange/white?text=Spirited+Away",
      "https://placehold.co/400x600/green/white?text=Howl",
      "https://placehold.co/400x600/red/white?text=Mononoke",
      "https://placehold.co/400x600/blue/white?text=Totoro",
    ],
  },
  {
    id: 2,
    title: "頭が混乱するスリラー",
    description: "最後まであなたを惑わせ続ける映画。",
    author: "twist_lover",
    movieCount: 48,
    likes: "875",
    images: [
      "https://placehold.co/400x600/1a1a1a/white?text=Inception",
      "https://placehold.co/400x600/333/white?text=Shutter+Island",
      "https://placehold.co/400x600/000/white?text=Memento",
      "https://placehold.co/400x600/222/white?text=Fight+Club",
    ],
  },
  {
    id: 3,
    title: "90年代SFクラシック",
    description: "黄金時代の象徴的なSF作品。",
    author: "retro_future",
    movieCount: 55,
    likes: "2.1k",
    images: [
      "https://placehold.co/400x600/teal/white?text=Blade+Runner",
      "https://placehold.co/400x600/indigo/white?text=Matrix",
      "https://placehold.co/400x600/purple/white?text=Total+Recall",
      "https://placehold.co/400x600/blue/white?text=Fifth+Element",
    ],
  },
  {
    id: 4,
    title: "美しい映画",
    description: "美しい映像と色彩の映画。",
    author: "color_palette",
    movieCount: 71,
    likes: "3.4k",
    images: [
      "https://placehold.co/400x600/pink/white?text=La+La+Land",
      "https://placehold.co/400x600/rose/white?text=Grand+Budapest",
      "https://placehold.co/400x600/fef3c7/black?text=Amelie",
      "https://placehold.co/400x600/e0f2f1/black?text=Her",
    ],
  },
  {
    id: 5,
    title: "史上最高の映画",
    description: "評価が最も高く、称賛されている映画。",
    author: "critic_choice",
    movieCount: 100,
    likes: "5.8k",
    images: [
      "https://placehold.co/400x600/gray/white?text=Godfather",
      "https://placehold.co/400x600/black/white?text=Dark+Knight",
      "https://placehold.co/400x600/78350f/white?text=Shawshank",
      "https://placehold.co/400x600/1e3a8a/white?text=Pulp+Fiction",
    ],
  },
  {
    id: 6,
    title: "80年代スラッシャー映画",
    description: "悲鳴の世代を定義したクラシックホラー。",
    author: "horror_fan",
    movieCount: 25,
    likes: "666",
    images: [
      "https://placehold.co/400x600/991b1b/white?text=Nightmare",
      "https://placehold.co/400x600/b91c1c/white?text=Friday+13th",
      "https://placehold.co/400x600/7f1d1d/white?text=Halloween",
      "https://placehold.co/400x600/black/red?text=Scream",
    ],
  },
  {
    id: 7,
    title: "心温まる映画",
    description: "あなたの一日を明るくする感動的な物語。",
    author: "optimist_prime",
    movieCount: 42,
    likes: "1.8k",
    images: [
      "https://placehold.co/400x600/fcd34d/black?text=Paddington",
      "https://placehold.co/400x600/bef264/black?text=Little+Miss",
      "https://placehold.co/400x600/6ee7b7/black?text=Up",
      "https://placehold.co/400x600/93c5fd/black?text=Forrest",
    ],
  },
  {
    id: 8,
    title: "現代コミック映画ヒット",
    description: "壮大なスーパーヒーローバトルとアクション。",
    author: "mcu_fanatic",
    movieCount: 98,
    likes: "4.5k",
    images: [
      "https://placehold.co/400x600/b91c1c/white?text=Avengers",
      "https://placehold.co/400x600/1d4ed8/white?text=Spider-Man",
      "https://placehold.co/400x600/047857/white?text=Batman",
      "https://placehold.co/400x600/b45309/white?text=Wonder+Woman",
    ],
  },
];

interface Tag {
  id: string;
  title: string;
  description?: string | null;
  author: string;
  movie_count: number;
  follower_count: number;
  images: string[];
  created_at?: string;
}

export default function Home() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [tags, setTags] = useState<Tag[]>([]);

  useEffect(() => {
    const fetchTags = async () => {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_BACKEND_API_BASE}/api/v1/tags`
        );
        const data = await response.json();
        console.log("tags", data);
        setTags(data.items ?? []);
      } catch (error) {
        console.error("Error fetching tags:", error);
      }
    };
    fetchTags();
  }, []);

  return (
    <div className="min-h-screen bg-[#FFF5F5]">
      <Header />

      <main className="container mx-auto px-4 md:px-6 py-12">
        {/* Hero Section */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-10">
          <div>
            <h1 className="text-3xl md:text-4xl font-bold text-gray-900 mb-2">
              タグを探そう！
            </h1>
            <p className="text-gray-600">
              お気に入りの映画リストを見つけたり、自分だけのタグを作ってみよう。
            </p>
          </div>
          <SignedIn>
            <button
              type="button"
              className="bg-[#FFD75E] hover:bg-[#ffcf40] text-gray-900 font-bold py-3 px-6 rounded-full flex items-center gap-2 shadow-sm hover:shadow transition-all"
              onClick={() => setIsCreateModalOpen(true)}
            >
              <Plus className="w-5 h-5" />
              <span>新しいタグを作成</span>
            </button>
          </SignedIn>
          <SignedOut>
            <SignInButton mode="modal">
              <button
                type="button"
                className="bg-[#FFD75E] hover:bg-[#ffcf40] text-gray-900 font-bold py-3 px-6 rounded-full flex items-center gap-2 shadow-sm hover:shadow transition-all"
              >
                <Plus className="w-5 h-5" />
                <span>新しいタグを作成</span>
              </button>
            </SignInButton>
          </SignedOut>
        </div>

        {/* Search & Filter */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-10">
          {/* Search Bar */}
          <div className="relative w-full md:max-w-2xl">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Search className="h-5 w-5 text-gray-400" />
            </div>
            <input
              type="text"
              placeholder="「90年代SF」などで検索..."
              className="block w-full pl-12 pr-4 py-3.5 rounded-full border border-gray-900 bg-white text-gray-900 focus:ring-2 focus:ring-blue-500 focus:border-transparent shadow-sm"
            />
          </div>

          {/* Filter Tabs */}
          <div className="flex items-center bg-white rounded-full p-1 border border-gray-900">
            <button className="px-6 py-2 rounded-full bg-[#FFD75E] text-gray-900 font-bold text-sm">
              人気
            </button>
            <button className="px-6 py-2 rounded-full text-gray-600 hover:bg-gray-100 font-medium text-sm">
              新着
            </button>
            <button className="px-6 py-2 rounded-full text-gray-600 hover:bg-gray-100 font-medium text-sm">
              映画数
            </button>
          </div>
        </div>

        {/* Category Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
          {/* {MOCK_CATEGORIES.map((category) => ( */}
          {tags.length > 0 ? (
            tags.map((tag) => (
              <CategoryCard
                key={tag.id}
                title={tag.title}
                description={tag.description ?? ""}
                author={tag.author}
                movieCount={tag.movie_count}
                likes={tag.follower_count}
                images={tag.images || []}
              />
            ))
          ) : (
            <div>タグがありません</div>
          )}
          {/* ))} */}
        </div>

        {/* Pagination */}
        <div className="flex justify-center items-center gap-2">
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-300 bg-white hover:bg-gray-50 text-gray-600 disabled:opacity-50">
            <ChevronLeft className="w-5 h-5" />
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full bg-blue-500 text-white font-bold">
            1
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            2
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            3
          </button>
          <span className="text-gray-500 px-1">...</span>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            12
          </button>
          <button className="w-10 h-10 flex items-center justify-center rounded-full border border-gray-900 bg-white hover:bg-gray-50 text-gray-900 font-medium">
            <ChevronRight className="w-5 h-5" />
          </button>
        </div>
        {/* Create Tag Modal */}
        <TagCreateModal
          open={isCreateModalOpen}
          onClose={() => setIsCreateModalOpen(false)}
          onCreated={(created: CreatedTagForList) => {
            setTags((prev) => [created, ...prev]);
          }}
        />
      </main>
    </div>
  );
}
