let fontCache: ArrayBuffer | null = null;

export async function getNotoSansJPBold(): Promise<ArrayBuffer> {
  if (fontCache) return fontCache;

  const css = await fetch(
    "https://fonts.googleapis.com/css2?family=Noto+Sans+JP:wght@700&display=swap",
    { next: { revalidate: 86400 } },
  ).then((res) => res.text());

  const fontUrl = css.match(/src: url\((.+?)\) format\('woff2'\)/)?.[1];
  if (!fontUrl) throw new Error("Font URL not found");

  const data = await fetch(fontUrl, { next: { revalidate: 86400 } }).then(
    (res) => res.arrayBuffer(),
  );
  fontCache = data;

  return data;
}
