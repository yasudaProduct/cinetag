import { z } from "zod";

/**
 * フォーム入力の zod schema
 */

export const AddMoviePolicyFormSchema = z.enum(["everyone", "owner_only"]);
export type AddMoviePolicyForm = z.infer<typeof AddMoviePolicyFormSchema>;

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
    add_movie_policy: AddMoviePolicyFormSchema.optional().default("everyone"),
});

export type TagCreateInput = z.infer<typeof TagCreateInputSchema>;

export function getFirstZodErrorMessage(error: z.ZodError): string {
    return error.issues[0]?.message ?? "入力が正しくありません。";
}


