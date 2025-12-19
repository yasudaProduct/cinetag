import { ApiErrorSchema } from "@/lib/validation/tag.api";

/**
 * 公開APIのベースURLを取得する
 * @returns 環境変数 NEXT_PUBLIC_BACKEND_API_BASE が設定されている場合はその値を返し、設定されていない場合はエラーを投げる
 */
export function getPublicApiBaseOrThrow(): string {
  const base = process.env.NEXT_PUBLIC_BACKEND_API_BASE;
  if (!base) throw new Error("環境変数 NEXT_PUBLIC_BACKEND_API_BASE が設定されていません。");
  return base;
}

/**
 * 公開APIのベースURLを取得する
 * @returns 環境変数 NEXT_PUBLIC_BACKEND_API_BASE が設定されている場合はその値を返し、設定されていない場合は undefined を返す
 */
export function getPublicApiBase(): string | undefined {
  return process.env.NEXT_PUBLIC_BACKEND_API_BASE;
}

/**
 * ResponseオブジェクトのJSONを取得する
 * @param res Responseオブジェクト
 * @returns ResponseオブジェクトのJSONを返す。JSONがない場合は null を返す
 */
export async function safeJson(res: Response): Promise<unknown> {
  return await res.json().catch(() => null);
}

/**
 * APIエラーをメッセージに変換する
 * @param params エラー情報
 * @returns エラーメッセージ。APIエラーの場合は error フィールドの値を返し、それ以外の場合は fallback の値を返す
 */
export function toApiErrorMessage(params: {
  status: number;
  body: unknown;
  fallback: string;
}): string {
  const parsed = ApiErrorSchema.safeParse(params.body);
  if (parsed.success) return parsed.data.error;
  return `${params.fallback}（${params.status}）`;
}


