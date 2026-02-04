"use client";

type TabType = "created" | "registered" | "favorite";

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
    <div className="flex items-center justify-start md:justify-center gap-2 mb-8 border-b border-gray-200 overflow-x-auto pb-1 md:pb-0">
      <button
        onClick={() => onTabChange("created")}
        className={`px-4 py-2 md:px-6 md:py-3 font-medium transition-colors relative whitespace-nowrap text-sm md:text-base ${
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
        className={`px-4 py-2 md:px-6 md:py-3 font-medium transition-colors relative whitespace-nowrap text-sm md:text-base ${
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
        className={`px-4 py-2 md:px-6 md:py-3 font-medium transition-colors relative whitespace-nowrap text-sm md:text-base ${
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
    </div>
  );
}

export type { TabType };
