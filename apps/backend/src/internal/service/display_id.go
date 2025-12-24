package service

import (
	"regexp"
	"strings"
)

// GenerateDisplayID は Clerk のユーザー情報から display_id を生成します。
//
// 優先順位:
// 1. Clerk の username（GitHubログインなど）
// 2. email のローカル部分（Google/Emailログイン）
// 3. clerk_user_id から生成（フォールバック）
func GenerateDisplayID(clerkUserID, username, email string) string {
	// 1. Clerkのusernameがあればそれを使用
	if username != "" {
		sanitized := sanitizeDisplayID(username)
		if isValidDisplayID(sanitized) {
			return sanitized
		}
	}

	// 2. emailのローカル部分を使用
	if email != "" {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			localPart := parts[0]
			sanitized := sanitizeDisplayID(localPart)
			if isValidDisplayID(sanitized) {
				return sanitized
			}
		}
	}

	// 3. clerk_user_idから生成（フォールバック）
	// "user_2abc123def..." → "user-2abc123d"
	shortID := strings.TrimPrefix(clerkUserID, "user_")
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return "user-" + shortID
}

// sanitizeDisplayID は文字列を display_id のフォーマットに変換します。
//
// ルール:
// - 小文字化
// - 英数字以外をハイフンに変換
// - 連続したハイフンを1つに
// - 先頭・末尾のハイフンを削除
// - 3-20文字に制限
func sanitizeDisplayID(s string) string {
	// 小文字化
	s = strings.ToLower(s)

	// 英数字以外をハイフンに変換
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// 先頭・末尾のハイフンを削除
	s = strings.Trim(s, "-")

	// 長さ制限（3-20文字）
	if len(s) > 20 {
		s = s[:20]
		// 末尾がハイフンになった場合は削除
		s = strings.TrimRight(s, "-")
	}

	return s
}

// isValidDisplayID は display_id が有効なフォーマットかチェックします。
//
// 条件:
// - 3-20文字
// - 英小文字、数字、ハイフンのみ
// - 先頭と末尾はハイフン不可
func isValidDisplayID(s string) bool {
	if len(s) < 3 || len(s) > 20 {
		return false
	}

	// 正規表現でフォーマットチェック
	// ^[a-z0-9] : 先頭は英小文字または数字
	// [a-z0-9-]{1,18} : 中間は英小文字、数字、ハイフンで1-18文字
	// [a-z0-9]$ : 末尾は英小文字または数字
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]{1,18}[a-z0-9]$`, s)
	return matched
}
