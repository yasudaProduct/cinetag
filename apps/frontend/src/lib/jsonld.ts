/**
 * JSON-LD を <script> タグへ埋め込むための文字列にシリアライズする
 *
 * JSON.stringify は `</script>` や `<!--` をエスケープしないため、
 * ユーザー入力（タグタイトル・表示名など）をそのまま埋め込むと
 * script 要素を抜け出して任意のスクリプトを実行される（XSS）。
 * HTML として意味を持つ文字を Unicode エスケープに置き換えて防ぐ。
 * エスケープ後も JSON としての値は変わらない。
 */
export function serializeJsonLd(data: Record<string, unknown>): string {
  return JSON.stringify(data)
    .replace(/</g, "\\u003c")
    .replace(/>/g, "\\u003e")
    .replace(/&/g, "\\u0026")
    .replace(/\u2028/g, "\\u2028")
    .replace(/\u2029/g, "\\u2029");
}
