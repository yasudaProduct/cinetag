export type ClerkGetToken = (options?: { template?: string }) => Promise<string | null>;

/**
 * Clerk から backend 用の JWT を取得する（取得できない場合は例外）
 *
 * - `useAuth().getToken` は React hook 経由でしか取れないため、
 *   API層では「getToken関数」を注入してもらう前提にする。
 */
export async function getBackendTokenOrThrow(
    getToken: ClerkGetToken,
    options?: { template?: string; errorMessage?: string }
): Promise<string> {
    const template = options?.template ?? "cinetag-backend";
    const token = await getToken({ template });
    if (!token) {
        throw new Error(
            options?.errorMessage ??
            "認証情報の取得に失敗しました。再ログインしてください。"
        );
    }
    return token;
}


