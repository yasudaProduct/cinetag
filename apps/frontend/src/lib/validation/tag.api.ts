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
        author_display_id: z.string().optional(),
        movie_count: z.number().int().nonnegative(),
        follower_count: z.number().int().nonnegative(),
        // バックエンド実装差分で null が返ることがあるため、null/undefined は [] に正規化する
        images: z.preprocess((v) => (v == null ? [] : v), z.array(z.string()).default([])),
        created_at: z.string().optional(),
    })
    .passthrough()
    .transform((data) => ({
        id: data.id,
        title: data.title,
        description: data.description,
        author: data.author,
        authorDisplayId: data.author_display_id,
        movieCount: data.movie_count,
        followerCount: data.follower_count,
        images: data.images,
        createdAt: data.created_at,
    }));

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
                total_count: z.number().int().nonnegative().optional(),
                page: z.number().int().positive().optional(),
                page_size: z.number().int().positive().optional(),
            })
            .passthrough(),
        z.array(TagListItemSchema).transform((items) => ({ items })),
    ])
    .transform((data) => ({
        items: "items" in data ? data.items : [],
        totalCount: "total_count" in data ? data.total_count ?? 0 : 0,
        page: "page" in data ? data.page ?? 1 : 1,
        pageSize: "page_size" in data ? data.page_size ?? 20 : 20,
    }));

export type TagsListResponse = z.infer<typeof TagsListResponseSchema>;

export const AddMoviePolicySchema = z.enum(["everyone", "owner_only"]);
export type AddMoviePolicy = z.infer<typeof AddMoviePolicySchema>;

export const TagCreateResponseSchema = z
    .object({
        id: z.string().min(1),
        title: z.string(),
        description: z.string().nullable().optional(),
        cover_image_url: z.string().nullable().optional(),
        is_public: z.boolean(),
        add_movie_policy: AddMoviePolicySchema.optional().default("everyone"),
        movie_count: z.number().int().nonnegative(),
        follower_count: z.number().int().nonnegative(),
        created_at: z.string().optional(),
        updated_at: z.string().optional(),
    })
    .passthrough();

export type TagCreateResponse = z.infer<typeof TagCreateResponseSchema>;

/**
 * GET /api/v1/tags/:tagId のレスポンス（UI向けに正規化）
 * - API仕様上は owner を返す想定だが、実装差分・開発中の形の揺れを吸収する
 * - UIで使う形へ transform する
 */
const TagDetailAuthorNameSchema = z.union([
    z.string(),
    z
        .object({
            name: z.string(),
        })
        .passthrough()
        .transform((v) => v.name),
]);

const TagDetailOwnerNameSchema = z
    .object({
        id: z.string().optional(),
        display_id: z.string().optional(),
        display_name: z.string().optional(),
        username: z.string().optional(),
        avatar_url: z.string().nullable().optional(),
    })
    .passthrough()
    .transform((v) => ({
        id: v.id ?? "",
        displayId: v.display_id,
        name: v.display_name ?? v.username ?? "unknown",
        avatarUrl: v.avatar_url ?? undefined,
    }));

const TagDetailParticipantSchema = z.union([
    z.string().transform((name) => ({ name })),
    z
        .object({
            name: z.string(),
        })
        .passthrough(),
]);

export const TagDetailResponseSchema = z
    .object({
        id: z.string().optional(),
        title: z.string().optional(),
        description: z.string().optional(),
        // 旧: author(string or {name})
        author: TagDetailAuthorNameSchema.optional(),
        // 新: owner({display_name, username})
        owner: TagDetailOwnerNameSchema.optional(),
        can_edit: z.boolean().optional(),
        can_add_movie: z.boolean().optional(),
        add_movie_policy: AddMoviePolicySchema.optional(),
        participant_count: z.number().int().nonnegative().optional(),
        participants: z.array(TagDetailParticipantSchema).optional(),
    })
    .passthrough()
    .transform((data) => {
        const owner = data.owner;
        const authorName = data.author ?? owner?.name ?? "unknown";
        const participants = (data.participants ?? []).filter((p) => p.name.length > 0);
        return {
            id: data.id ?? "",
            title: data.title ?? "",
            description: data.description ?? "",
            author: { name: authorName },
            owner: owner ?? { id: "", name: authorName },
            canEdit: data.can_edit ?? false,
            canAddMovie: data.can_add_movie ?? false,
            addMoviePolicy: data.add_movie_policy ?? "everyone",
            participantCount: data.participant_count ?? 0,
            participants,
        };
    });

export type TagDetailResponse = z.infer<typeof TagDetailResponseSchema>;

/**
 * GET /api/v1/tags/:tagId/movies のレスポンス（items配列へ正規化）
 * - { items: [...] } と [...] の両方を許容
 * - 各 item の形の揺れ（movieネスト等）を許容し、最低限の情報を抽出できる形で返す
 */
const TagMovieItemSchema = z
    .object({
        id: z.union([z.string(), z.number()]).optional(),
        tmdb_movie_id: z.union([z.string(), z.number()]).optional(),
        added_by_user_id: z.string().optional(),
        title: z.string().optional(),
        year: z.number().optional(),
        release_year: z.number().optional(),
        poster_url: z.string().nullable().optional(),
        movie: z
            .object({
                title: z.string().optional(),
                poster_path: z.string().nullable().optional(),
                release_date: z.string().nullable().optional(),
            })
            .passthrough()
            .optional(),
    })
    .passthrough();

export const TagMoviesResponseSchema = z
    .union([
        z
            .object({
                items: z.array(TagMovieItemSchema),
            })
            .passthrough(),
        z.array(TagMovieItemSchema).transform((items) => ({ items })),
    ])
    .transform((data) => ({
        items: "items" in data ? data.items : [],
    }));

export type TagMoviesResponse = z.infer<typeof TagMoviesResponseSchema>;
