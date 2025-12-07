import { X, UploadCloud, Film } from "lucide-react";
import { useState, FormEvent } from "react";

interface TagCreateModalProps {
  open: boolean;
  onClose: () => void;
}

export const TagCreateModal = ({ open, onClose }: TagCreateModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  if (!open) return null;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    // TODO: API連携時に送信処理をここに実装
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      {/* Card */}
      <div className="w-full max-w-xl mx-4 rounded-3xl bg-[#FFF9F3] shadow-xl border border-[#F3E1D6]">
        {/* Header */}
        <div className="flex items-start justify-between px-8 pt-8">
          <div>
            <h2 className="text-2xl md:text-3xl font-extrabold text-[#1F1A2B] tracking-tight">
              Create a New Tag
            </h2>
            <p className="mt-2 text-sm md:text-base text-[#7C7288]">
              あなたの映画コレクションを世界とシェアしましょう。
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="閉じる"
            className="ml-4 inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#E4D3C7] bg-white text-[#7C7288] hover:bg-[#FDF1E7] hover:text-[#1F1A2B] transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="px-8 pb-8 pt-6 space-y-6">
          {/* Tag Name */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              Tag Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Mind-Bending Sci-Fi"
              className="w-full rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] px-4 py-3 text-sm text-[#1F1A2B] shadow-[0_1px_0_rgba(0,0,0,0.03)] focus:outline-none focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8]"
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="このタグについての簡単な説明を書いてください。"
              rows={4}
              className="w-full rounded-xl border border-[#E4D3C7] bg-[#FFFDF8] px-4 py-3 text-sm text-[#1F1A2B] shadow-[0_1px_0_rgba(0,0,0,0.03)] focus:outline-none focus:ring-2 focus:ring-[#FF8C75] focus:border-transparent placeholder:text-[#C2B5A8] resize-none"
            />
          </div>

          {/* Cover Image (dummy uploader) */}
          <div className="space-y-2">
            <label className="block text-xs font-semibold tracking-wide text-[#7C7288]">
              Cover Image
            </label>
            <div className="rounded-2xl border-2 border-dashed border-[#E4D3C7] bg-[#FFFDF8] px-6 py-8 flex flex-col items-center justify-center text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-[#FFF2E0] mb-3">
                <Film className="w-7 h-7 text-[#FF8C75]" />
              </div>
              <p className="text-sm">
                <span className="font-semibold text-[#FF5C5C]">
                  Upload a file
                </span>{" "}
                <span className="text-[#7C7288]">or drag and drop</span>
              </p>
              <p className="mt-1 text-xs text-[#B09EA0]">
                PNG, JPG, GIF up to 10MB
              </p>
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end pt-2">
            <button
              type="submit"
              className="inline-flex items-center justify-center rounded-full bg-[#FF5C5C] px-8 py-3 text-sm font-semibold text-white shadow-[0_8px_0_#D44242] hover:translate-y-0.5 hover:shadow-[0_6px_0_#D44242] active:translate-y-1 active:shadow-[0_3px_0_#D44242] transition-transform"
            >
              Create Tag
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
