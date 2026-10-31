import { serializeJsonLd } from "@/lib/jsonld";

/**
 * 構造化データ（JSON-LD）を出力する
 *
 * ユーザー入力を含むデータも安全に埋め込めるよう、
 * dangerouslySetInnerHTML を直接使わずこのコンポーネントを利用すること。
 */
export function JsonLd({ data }: { data: Record<string, unknown> }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: serializeJsonLd(data) }}
    />
  );
}
