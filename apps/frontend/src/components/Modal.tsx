"use client";

import type { ReactNode } from "react";

type ModalProps = {
  open: boolean;
  children: ReactNode;
  onClose?: () => void;
};

/**
 * 汎用モーダルのオーバーレイ + センター配置ラッパー。
 * - 背景クリックで onClose を呼び出して閉じる
 * - コンテンツ内クリックでは閉じない
 * 中身のカードやレイアウトは子コンポーネント側で定義します。
 */
export function Modal({ open, children, onClose }: ModalProps) {
  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
      onClick={() => {
        if (onClose) onClose();
      }}
    >
      <div
        className="w-full"
        onClick={(e) => {
          // モーダル内容クリックでは閉じないようにバブリングを止める
          e.stopPropagation();
        }}
      >
        {children}
      </div>
    </div>
  );
}
