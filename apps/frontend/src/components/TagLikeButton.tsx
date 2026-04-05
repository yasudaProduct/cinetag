"use client";

import { useState } from "react";
import { ThumbsUp } from "lucide-react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuth, useUser } from "@clerk/nextjs";
import { likeTag } from "@/lib/api/tags/like";
import { unlikeTag } from "@/lib/api/tags/unlike";
import { Spinner } from "@/components/ui/spinner";

interface TagLikeButtonProps {
  tagId: string;
  initialLikeCount: number;
  initialIsLiked: boolean;
}

export function TagLikeButton({
  tagId,
  initialLikeCount,
  initialIsLiked,
}: TagLikeButtonProps) {
  const { getToken } = useAuth();
  const { isSignedIn } = useUser();
  const queryClient = useQueryClient();

  const [optimisticIsLiked, setOptimisticIsLiked] = useState(initialIsLiked);
  const [optimisticCount, setOptimisticCount] = useState(initialLikeCount);

  const likeMutation = useMutation({
    mutationFn: async (action: "like" | "unlike") => {
      const token = await getToken({ template: "cinetag-backend" });
      if (!token) throw new Error("認証が必要です");
      if (action === "unlike") {
        await unlikeTag(tagId, token);
      } else {
        await likeTag(tagId, token);
      }
    },
    onMutate: (action) => {
      setOptimisticIsLiked((prev) => !prev);
      setOptimisticCount((prev) => (action === "unlike" ? prev - 1 : prev + 1));
    },
    onError: () => {
      setOptimisticIsLiked(initialIsLiked);
      setOptimisticCount(initialLikeCount);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tagDetail", tagId] });
    },
  });

  if (!isSignedIn) {
    return (
      <div className="flex items-center gap-1.5 text-sm text-gray-500">
        <ThumbsUp className="w-4 h-4" />
        <span className="font-bold text-gray-900">{optimisticCount}</span>
      </div>
    );
  }

  return (
    <button
      type="button"
      disabled={likeMutation.isPending}
      onClick={() => likeMutation.mutate(optimisticIsLiked ? "unlike" : "like")}
      className={`w-full font-bold py-3 rounded-full flex items-center justify-center gap-2 shadow-sm hover:shadow transition-all ${
        optimisticIsLiked
          ? "bg-blue-100 text-blue-600 border border-blue-300 hover:bg-blue-200"
          : "bg-white text-gray-700 border border-gray-300 hover:bg-gray-50"
      }`}
    >
      <ThumbsUp
        className={`w-5 h-5 ${optimisticIsLiked ? "fill-current" : ""}`}
      />
      {likeMutation.isPending ? (
        <span className="flex items-center gap-2">
          <Spinner size="sm" />
          処理中
        </span>
      ) : (
        <>{optimisticIsLiked ? "いいね済み" : "いいねする"}</>
      )}
    </button>
  );
}
