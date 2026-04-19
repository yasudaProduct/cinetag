"use client";

import { Modal } from "@/components/Modal";
import { Spinner } from "@/components/ui/spinner";

type TagDeleteConfirmModalProps = {
  open: boolean;
  title: string;
  followerCount: number;
  otherUsersMovieCount: number;
  isDeleting: boolean;
  errorMessage: string | null;
  onClose: () => void;
  onConfirm: () => void;
};

export function TagDeleteConfirmModal({
  open,
  title,
  followerCount,
  otherUsersMovieCount,
  isDeleting,
  errorMessage,
  onClose,
  onConfirm,
}: TagDeleteConfirmModalProps) {
  const showIntenseWarning = followerCount > 0 || otherUsersMovieCount > 0;

  const intenseLead = (() => {
    const hasF = followerCount > 0;
    const hasM = otherUsersMovieCount > 0;
    if (hasF && hasM) {
      return (
        <>
          このタグには{" "}
          <span className="font-bold text-gray-900">{followerCount}人のフォロワー</span> と{" "}
          <span className="font-bold text-gray-900">
            他のユーザーが追加した{otherUsersMovieCount}件の映画
          </span>
          があります。
        </>
      );
    }
    if (hasF) {
      return (
        <>
          このタグには{" "}
          <span className="font-bold text-gray-900">{followerCount}人のフォロワー</span>がいます。
        </>
      );
    }
    return (
      <>
        このタグには{" "}
        <span className="font-bold text-gray-900">
          他のユーザーが追加した{otherUsersMovieCount}件の映画
        </span>
        があります。
      </>
    );
  })();

  return (
    <Modal open={open} onClose={isDeleting ? undefined : onClose}>
      <div className="w-[min(100vw-2rem,28rem)] rounded-2xl border border-gray-200 bg-white p-6 shadow-xl">
        <h2 className="text-lg font-bold text-gray-900">タグを削除</h2>
        <p className="mt-1 text-sm text-gray-500 line-clamp-2">「{title}」</p>

        <div className="mt-4 space-y-3 text-sm text-gray-700 leading-relaxed">
          {showIntenseWarning ? (
            <>
              <p>
                {intenseLead}
                削除するとこれらのデータもすべて失われます。本当に削除しますか？
              </p>
              <p className="rounded-lg bg-amber-50 border border-amber-200 px-3 py-2 text-amber-900 text-xs font-medium">
                この操作は取り消せません。
              </p>
            </>
          ) : (
            <p>このタグを削除しますか？この操作は取り消せません。</p>
          )}
        </div>

        {errorMessage ? (
          <p className="mt-3 text-sm text-red-600" role="alert">
            {errorMessage}
          </p>
        ) : null}

        <div className="mt-6 flex flex-col-reverse sm:flex-row sm:justify-end gap-2 sm:gap-3">
          <button
            type="button"
            disabled={isDeleting}
            className="w-full sm:w-auto rounded-full px-4 py-2.5 text-sm font-bold border border-gray-300 text-gray-800 hover:bg-gray-50 disabled:opacity-50"
            onClick={onClose}
          >
            キャンセル
          </button>
          <button
            type="button"
            disabled={isDeleting}
            className="w-full sm:w-auto rounded-full px-4 py-2.5 text-sm font-bold bg-red-600 text-white hover:bg-red-700 disabled:opacity-50 flex items-center justify-center gap-2"
            onClick={onConfirm}
          >
            {isDeleting ? (
              <>
                <Spinner size="sm" className="text-white" />
                削除中…
              </>
            ) : (
              "削除する"
            )}
          </button>
        </div>
      </div>
    </Modal>
  );
}
