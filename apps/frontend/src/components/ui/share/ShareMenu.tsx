"use client";

import { useState, useEffect, useRef } from "react";
import { Link, Check } from "lucide-react";
import { XIcon } from "./icons/XIcon";
import { LineIcon } from "./icons/LineIcon";
import { FacebookIcon } from "./icons/FacebookIcon";
import { buildShareUrl, copyToClipboard } from "@/lib/share";

type ShareMenuProps = {
  open: boolean;
  onClose: () => void;
  url: string;
  title: string;
};

export function ShareMenu({ open, onClose, url, title }: ShareMenuProps) {
  const [copied, setCopied] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // 外側クリックで閉じる
  useEffect(() => {
    if (!open) return;
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open, onClose]);

  // Escキーで閉じる
  useEffect(() => {
    if (!open) return;
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [open, onClose]);

  if (!open) return null;

  const handleCopy = async () => {
    const success = await copyToClipboard(url);
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleShareToSNS = (target: "x" | "line" | "facebook") => {
    const shareUrl = buildShareUrl(target, { url, title });
    window.open(shareUrl, "_blank", "noopener,noreferrer,width=600,height=500");
    onClose();
  };

  return (
    <div
      ref={menuRef}
      className="absolute z-50 w-56 bg-white rounded-2xl shadow-lg border border-gray-200 py-2 overflow-hidden"
    >
      <button
        type="button"
        onClick={handleCopy}
        className="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
      >
        {copied ? (
          <Check className="w-5 h-5 text-green-500" />
        ) : (
          <Link className="w-5 h-5 text-gray-600" />
        )}
        {copied ? "コピーしました！" : "URLをコピー"}
      </button>
      <button
        type="button"
        onClick={() => handleShareToSNS("x")}
        className="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
      >
        <XIcon className="w-5 h-5 text-black" />
        Xでシェア
      </button>
      <button
        type="button"
        onClick={() => handleShareToSNS("line")}
        className="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
      >
        <LineIcon className="w-5 h-5 text-[#06C755]" />
        LINEでシェア
      </button>
      <button
        type="button"
        onClick={() => handleShareToSNS("facebook")}
        className="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
      >
        <FacebookIcon className="w-5 h-5 text-[#1877F2]" />
        Facebookでシェア
      </button>
    </div>
  );
}
