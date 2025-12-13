import { z } from "zod";

/**
 * フロントエンド側バリデーション（入力/レスポンス）用の zod schema を集約する。
 * - バックエンドが最終的な正としつつ、UI/UX と不正データ耐性のために実行時検証も行う。
 */

export const ApiErrorSchema = z
  .object({
    error: z.string(),
  })
  .passthrough();

export const TagCreateInputSchema = z.object({
  title: z
    .string()
    .trim()
    .min(1, "名前を入力してください。")
    .max(100, "名前は100文字以内で入力してください。"),
  description: z
    .string()
    .trim()
    .max(500, "説明は500文字以内で入力してください。")
    .optional(),
});

export type TagCreateInput = z.infer<typeof TagCreateInputSchema>;

export const TagListItemSchema = z
  .object({
    id: z.string().min(1),
    title: z.string(),
    description: z.string().nullable().optional(),
    author: z.string(),
    movie_count: z.number().int().nonnegative(),
    follower_count: z.number().int().nonnegative(),
    images: z.array(z.string()).optional().default([]),
    created_at: z.string().optional(),
  })
  .passthrough();

export type TagListItem = z.infer<typeof TagListItemSchema>;

/**
 * GET /api/v1/tags のレスポンス
 * - ドキュメント上は { items: Tag[] ... } を想定
 * - 実装差分に備え、配列のみ返るケースも許容し { items } に正規化する
 */
export const TagsListResponseSchema = z
  .union([
    z
      .object({
        items: z.array(TagListItemSchema),
      })
      .passthrough(),
    z.array(TagListItemSchema).transform((items) => ({ items })),
  ])
  .transform((data) => ({
    items: "items" in data ? data.items : [],
  }));

export type TagsListResponse = z.infer<typeof TagsListResponseSchema>;

export const TagCreateResponseSchema = z
  .object({
    id: z.string().min(1),
    title: z.string(),
    description: z.string().nullable().optional(),
    cover_image_url: z.string().nullable().optional(),
    is_public: z.boolean(),
    movie_count: z.number().int().nonnegative(),
    follower_count: z.number().int().nonnegative(),
    created_at: z.string().optional(),
    updated_at: z.string().optional(),
  })
  .passthrough();

export type TagCreateResponse = z.infer<typeof TagCreateResponseSchema>;

export function getFirstZodErrorMessage(error: z.ZodError): string {
  return error.issues[0]?.message ?? "入力が正しくありません。";
}


