"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { useUser } from "@clerk/nextjs";
import { useRouter } from "next/navigation";
import { AddToTagModal } from "./AddToTagModal";

type AddToTagButtonProps = {
  tmdbMovieId: number;
  movieTitle: string;
  relatedTagIds: string[];
};

export function AddToTagButton({
  tmdbMovieId,
  movieTitle,
  relatedTagIds,
}: AddToTagButtonProps) {
  const { isSignedIn } = useUser();
  const router = useRouter();
  const [open, setOpen] = useState(false);

  const handleClick = () => {
    if (!isSignedIn) {
      router.push("/sign-in");
      return;
    }
    setOpen(true);
  };

  return (
    <>
      <button
        type="button"
        onClick={handleClick}
        className="inline-flex items-center gap-2 rounded-full border-2 border-[#FF5C5C] px-5 py-2 text-sm font-semibold text-[#FF5C5C] hover:bg-[#FF5C5C] hover:text-white transition-colors"
      >
        <Plus className="w-4 h-4" />
        タグに追加
      </button>

      <AddToTagModal
        open={open}
        onClose={() => setOpen(false)}
        tmdbMovieId={tmdbMovieId}
        movieTitle={movieTitle}
        relatedTagIds={relatedTagIds}
      />
    </>
  );
}
