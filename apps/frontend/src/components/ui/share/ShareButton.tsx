"use client";

import { useState } from "react";
import { Share2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { canUseNativeShare, nativeShare } from "@/lib/share";
import { ShareMenu } from "./ShareMenu";

type ShareButtonProps = {
  title: string;
  description?: string;
  variant?: "default" | "icon-only";
  className?: string;
};

export function ShareButton({
  title,
  description,
  variant = "default",
  className,
}: ShareButtonProps) {
  const [menuOpen, setMenuOpen] = useState(false);

  const handleClick = async () => {
    const url = window.location.href;

    if (canUseNativeShare()) {
      await nativeShare({ url, title, text: description });
      return;
    }

    setMenuOpen((prev) => !prev);
  };

  const isIconOnly = variant === "icon-only";

  return (
    <div className="relative">
      <button
        type="button"
        onClick={handleClick}
        aria-label="共有"
        aria-haspopup="menu"
        aria-expanded={menuOpen}
        className={cn(
          isIconOnly
            ? "inline-flex items-center justify-center w-10 h-10 rounded-full bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors"
            : "w-full font-bold py-3 rounded-full flex items-center justify-center gap-2 bg-white text-gray-700 border border-gray-300 hover:bg-gray-50 shadow-sm hover:shadow transition-all",
          className,
        )}
      >
        <Share2 className={isIconOnly ? "w-5 h-5" : "w-5 h-5"} />
        {!isIconOnly && "シェアする"}
      </button>

      <div className={isIconOnly ? "absolute right-0 mt-2" : "absolute left-0 bottom-full mb-2"}>
        <ShareMenu
          open={menuOpen}
          onClose={() => setMenuOpen(false)}
          url={typeof window !== "undefined" ? window.location.href : ""}
          title={title}
        />
      </div>
    </div>
  );
}
