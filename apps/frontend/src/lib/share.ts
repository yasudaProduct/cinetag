export type ShareTarget = "x" | "line" | "facebook";

export function buildShareUrl(
  target: ShareTarget,
  params: { url: string; title: string },
): string {
  const encodedUrl = encodeURIComponent(params.url);
  const encodedTitle = encodeURIComponent(params.title);

  switch (target) {
    case "x":
      return `https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}`;
    case "line":
      return `https://social-plugins.line.me/lineit/share?url=${encodedUrl}`;
    case "facebook":
      return `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`;
  }
}

export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

export function canUseNativeShare(): boolean {
  return (
    typeof navigator !== "undefined" &&
    typeof navigator.share === "function"
  );
}

export async function nativeShare(params: {
  url: string;
  title: string;
  text?: string;
}): Promise<boolean> {
  try {
    await navigator.share(params);
    return true;
  } catch {
    return false;
  }
}
