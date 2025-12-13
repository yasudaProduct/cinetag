import { z } from "zod";

/**
 * API レスポンス/エラーの zod schema
 * - バックエンドが最終的な正としつつ、UI/UX と不正データ耐性のために実行時検証も行う。
 */

export const ApiErrorSchema = z
    .object({
        error: z.string(),
    })
    .passthrough();

export const TagListItemSchema = z
    .object({
        id: z.string().min(1),
        title: z.string(),
        description: z.string().nullable().optional(),
        author: z.string(),
        movie_count: z.number().int().nonnegative(),
        follower_count: z.number().int().nonnegative(),
        // バックエンド実装差分で null が返ることがあるため、null/undefined は [] に正規化する
        images: z.preprocess((v) => (v == null ? [] : v), z.array(z.string()).default([])),
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
