"use client";

type TabType =
  | "created"
  | "registered"
  | "favorite"
  | "following"
  | "followers"
  | "followingTags";

type UserPageTabsProps = {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
  isOwnPage: boolean;
};

export function UserPageTabs({
  activeTab,
  onTabChange,
  isOwnPage,
}: UserPageTabsProps) {
  return (
    <div className="flex items-center justify-center gap-2 mb-8 border-b border-gray-200 overflow-x-auto">
      <button
        onClick={() => onTabChange("created")}
        className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
          activeTab === "created"
            ? "text-pink-600"
            : "text-gray-600 hover:text-gray-900"
        }`}
      >
        作成したタグ
        {activeTab === "created" && (
          <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
        )}
      </button>
      <button
        onClick={() => onTabChange("registered")}
        className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
          activeTab === "registered"
            ? "text-pink-600"
            : "text-gray-600 hover:text-gray-900"
        }`}
      >
        フォローしたタグ
        {activeTab === "registered" && (
          <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
        )}
      </button>
      <button
        onClick={() => onTabChange("favorite")}
        className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
          activeTab === "favorite"
            ? "text-pink-600"
            : "text-gray-600 hover:text-gray-900"
        }`}
      >
        いいねしたタグ
        {activeTab === "favorite" && (
          <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
        )}
      </button>
      <button
        onClick={() => onTabChange("following")}
        className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
          activeTab === "following"
            ? "text-pink-600"
            : "text-gray-600 hover:text-gray-900"
        }`}
      >
        フォロー中のユーザー
        {activeTab === "following" && (
          <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
        )}
      </button>
      <button
        onClick={() => onTabChange("followers")}
        className={`px-6 py-3 font-medium transition-colors relative whitespace-nowrap ${
          activeTab === "followers"
            ? "text-pink-600"
            : "text-gray-600 hover:text-gray-900"
        }`}
      >
        フォロワーのユーザー
        {activeTab === "followers" && (
          <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-pink-600" />
        )}
      </button>
    </div>
  );
}

export type { TabType };
