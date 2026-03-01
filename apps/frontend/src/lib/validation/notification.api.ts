import { z } from "zod";

/**
 * 通知関連の API レスポンス zod schema
 */

const NotificationActorSchema = z.object({
  id: z.string(),
  display_id: z.string(),
  display_name: z.string(),
  avatar_url: z.string().nullable().optional(),
});

const NotificationTagSchema = z.object({
  id: z.string(),
  title: z.string(),
});

export const NotificationItemSchema = z
  .object({
    id: z.string(),
    notification_type: z.enum([
      "tag_movie_added",
      "tag_followed",
      "user_followed",
      "following_user_created_tag",
    ]),
    is_read: z.boolean(),
    created_at: z.string(),
    actor: NotificationActorSchema.nullable(),
    tag: NotificationTagSchema.nullable().optional(),
    movie_title: z.string().nullable().optional(),
  })
  .passthrough()
  .transform((data) => ({
    id: data.id,
    notificationType: data.notification_type,
    isRead: data.is_read,
    createdAt: data.created_at,
    actor: data.actor
      ? {
          id: data.actor.id,
          displayId: data.actor.display_id,
          displayName: data.actor.display_name,
          avatarUrl: data.actor.avatar_url ?? null,
        }
      : null,
    tag: data.tag
      ? {
          id: data.tag.id,
          title: data.tag.title,
        }
      : null,
    movieTitle: data.movie_title ?? null,
  }));

export type NotificationItem = z.infer<typeof NotificationItemSchema>;

export const NotificationListResponseSchema = z
  .object({
    notifications: z.array(NotificationItemSchema),
    total: z.number().int().nonnegative(),
    page: z.number().int().positive(),
    page_size: z.number().int().positive(),
  })
  .passthrough()
  .transform((data) => ({
    notifications: data.notifications,
    total: data.total,
    page: data.page,
    pageSize: data.page_size,
  }));

export type NotificationListResponse = z.infer<
  typeof NotificationListResponseSchema
>;

export const UnreadCountResponseSchema = z
  .object({
    unread_count: z.number().int().nonnegative(),
  })
  .passthrough()
  .transform((data) => ({
    unreadCount: data.unread_count,
  }));

export type UnreadCountResponse = z.infer<typeof UnreadCountResponseSchema>;
