"use client";

import { useState, useRef, useCallback } from "react";
import { X, Upload, Image as ImageIcon } from "lucide-react";
import Image from "next/image";
import { Modal } from "@/components/Modal";
import { Spinner } from "@/components/ui/spinner";

type ImageUploadModalProps = {
  open: boolean;
  onClose: () => void;
  onUpload: (file: File) => Promise<void>;
  currentImageUrl?: string;
  maxSizeMB?: number;
};

export const ImageUploadModal = ({
  open,
  onClose,
  onUpload,
  currentImageUrl,
  maxSizeMB = 5,
}: ImageUploadModalProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const resetState = useCallback(() => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setErrorMessage(null);
    setIsDragging(false);
  }, []);

  const handleClose = useCallback(() => {
    resetState();
    onClose();
  }, [resetState, onClose]);

  const validateFile = useCallback(
    (file: File): string | null => {
      // ファイルサイズチェック
      if (file.size > maxSizeMB * 1024 * 1024) {
        return `画像サイズは${maxSizeMB}MB以下にしてください`;
      }

      // ファイルタイプチェック
      if (!file.type.startsWith("image/")) {
        return "画像ファイルを選択してください";
      }

      return null;
    },
    [maxSizeMB]
  );

  const processFile = useCallback(
    (file: File) => {
      const error = validateFile(file);
      if (error) {
        setErrorMessage(error);
        return;
      }

      setErrorMessage(null);
      setSelectedFile(file);

      // プレビュー用URLを生成
      const url = URL.createObjectURL(file);
      console.log(url);
      setPreviewUrl(url);
    },
    [validateFile]
  );

  const handleFileChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (file) {
        processFile(file);
      }
      // inputをリセット（同じファイルを再選択できるように）
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    },
    [processFile]
  );

  const handleDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      setIsDragging(false);

      const file = event.dataTransfer.files[0];
      if (file) {
        processFile(file);
      }
    },
    [processFile]
  );

  const handleUpload = useCallback(async () => {
    if (!selectedFile) return;

    setIsUploading(true);
    setErrorMessage(null);

    try {
      await onUpload(selectedFile);
      handleClose();
    } catch (err) {
      setErrorMessage(
        err instanceof Error ? err.message : "アップロードに失敗しました"
      );
    } finally {
      setIsUploading(false);
    }
  }, [selectedFile, onUpload, handleClose]);

  return (
    <Modal open={open} onClose={handleClose}>
      <div className="w-full max-w-md mx-auto rounded-2xl bg-white shadow-xl border border-gray-200 overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h2 className="text-lg font-bold text-gray-900">
            プロフィール画像を変更
          </h2>
          <button
            type="button"
            onClick={handleClose}
            aria-label="閉じる"
            className="inline-flex h-8 w-8 items-center justify-center rounded-full text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-6 space-y-6">
          {/* 現在の画像とプレビュー */}
          <div className="flex items-center justify-center gap-8">
            {/* 現在の画像 */}
            <div className="text-center">
              <p className="text-xs text-gray-500 mb-2">現在</p>
              <div className="w-20 h-20 rounded-full overflow-hidden bg-gray-100 border-2 border-gray-200">
                {currentImageUrl ? (
                  <Image
                    src={currentImageUrl}
                    alt="現在のプロフィール画像"
                    width={80}
                    height={80}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400">
                    <ImageIcon className="w-8 h-8" />
                  </div>
                )}
              </div>
            </div>

            {/* 矢印 */}
            {previewUrl && (
              <>
                <div className="text-gray-300 text-2xl">→</div>

                {/* プレビュー */}
                <div className="text-center">
                  <p className="text-xs text-gray-500 mb-2">新しい画像</p>
                  <div className="w-20 h-20 rounded-full overflow-hidden bg-gray-100 border-2 border-[#FFD75E]">
                    <Image
                      src={previewUrl}
                      alt="新しいプロフィール画像"
                      width={80}
                      height={80}
                      className="w-full h-full object-cover"
                    />
                  </div>
                </div>
              </>
            )}
          </div>

          {/* ドロップゾーン */}
          <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`
              relative rounded-xl border-2 border-dashed p-8 text-center cursor-pointer transition-colors
              ${
                isDragging
                  ? "border-[#FFD75E] bg-[#FFF9F0]"
                  : "border-gray-200 bg-gray-50 hover:border-gray-300 hover:bg-gray-100"
              }
            `}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              onChange={handleFileChange}
              className="hidden"
            />
            <Upload
              className={`w-10 h-10 mx-auto mb-3 ${
                isDragging ? "text-[#FFD75E]" : "text-gray-400"
              }`}
            />
            <p className="text-sm text-gray-600">
              <span className="font-semibold text-[#FFD75E]">
                クリックして選択
              </span>
              <span className="text-gray-500"> またはドラッグ＆ドロップ</span>
            </p>
            <p className="mt-2 text-xs text-gray-400">
              PNG, JPG, GIF ({maxSizeMB}MB以下)
            </p>
          </div>

          {/* エラーメッセージ */}
          {errorMessage && (
            <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-sm text-red-600">
              {errorMessage}
            </div>
          )}

          {/* ボタン */}
          <div className="flex gap-3">
            <button
              type="button"
              onClick={handleClose}
              className="flex-1 px-4 py-3 bg-gray-100 text-gray-700 font-semibold rounded-xl hover:bg-gray-200 transition-colors"
            >
              キャンセル
            </button>
            <button
              type="button"
              onClick={handleUpload}
              disabled={!selectedFile || isUploading}
              className="flex-1 px-4 py-3 bg-[#FFD75E] text-gray-900 font-semibold rounded-xl hover:bg-[#ffcf40] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {isUploading ? (
                <>
                  <Spinner size="sm" />
                  アップロード中...
                </>
              ) : (
                "変更する"
              )}
            </button>
          </div>
        </div>
      </div>
    </Modal>
  );
};
