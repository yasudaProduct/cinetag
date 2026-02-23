"use client";

import { X } from "lucide-react";
import { Modal } from "@/components/Modal";
import { GoogleSignInForm } from "@/components/GoogleSignInForm";

type LoginModalProps = {
  open: boolean;
  onClose: () => void;
};

export function LoginModal({ open, onClose }: LoginModalProps) {
  return (
    <Modal open={open} onClose={onClose}>
      <div className="w-full max-w-sm mx-auto rounded-2xl bg-white shadow-xl border border-gray-100 p-8">
        <div className="flex items-start justify-between mb-6">
          <div>
            <h2 className="text-xl font-bold text-gray-900">ログイン</h2>
            <p className="mt-1 text-sm text-gray-500">
              この機能を使うにはログインが必要です
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="閉じる"
            className="ml-4 inline-flex h-8 w-8 items-center justify-center rounded-full text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <GoogleSignInForm onNavigate={onClose} />
      </div>
    </Modal>
  );
}
