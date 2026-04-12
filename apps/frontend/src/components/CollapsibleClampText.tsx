"use client";

import { useLayoutEffect, useRef, useState } from "react";

type CollapsibleClampTextProps = {
  text: string;
  /** 外側ラッパー（余白など） */
  className?: string;
  /** 本文 `<p>` のクラス（未指定時はタグ詳細と同系統のデフォルト） */
  paragraphClassName?: string;
};

/**
 * 長文を line-clamp で折りたたみ、はみ出す場合のみ「続きを表示 / 折りたたむ」を出す。
 */
export function CollapsibleClampText({
  text,
  className,
  paragraphClassName = "text-sm md:text-base text-gray-600",
}: CollapsibleClampTextProps) {
  const [expanded, setExpanded] = useState(false);
  const [needsToggle, setNeedsToggle] = useState(false);
  const paragraphRef = useRef<HTMLParagraphElement>(null);

  useLayoutEffect(() => {
    const el = paragraphRef.current;
    if (!el || !text) {
      const id = requestAnimationFrame(() => setNeedsToggle(false));
      return () => cancelAnimationFrame(id);
    }
    const id = requestAnimationFrame(() => {
      if (!paragraphRef.current) return;
      if (expanded) {
        setNeedsToggle(true);
        return;
      }
      setNeedsToggle(
        paragraphRef.current.scrollHeight > paragraphRef.current.clientHeight,
      );
    });
    return () => cancelAnimationFrame(id);
  }, [text, expanded]);

  if (!text) {
    return null;
  }

  return (
    <div className={className}>
      <p
        ref={paragraphRef}
        className={`${paragraphClassName} leading-relaxed whitespace-pre-wrap break-words ${
          !expanded ? "line-clamp-3" : ""
        }`}
      >
        {text}
      </p>
      {needsToggle ? (
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="mt-2 text-sm font-medium text-pink-600 hover:text-pink-700 hover:underline"
        >
          {expanded ? "折りたたむ" : "続きを表示"}
        </button>
      ) : null}
    </div>
  );
}
