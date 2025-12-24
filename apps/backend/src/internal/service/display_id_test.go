package service

import (
	"testing"
)

func TestGenerateDisplayID(t *testing.T) {
	tests := []struct {
		name         string
		clerkUserID  string
		username     string
		email        string
		wantPrefix   string // 期待される接頭辞（完全一致でなくても良い場合）
		wantExact    string // 期待される完全一致の値
		checkValid   bool   // isValidDisplayID でチェックするか
	}{
		{
			name:        "Clerk username がある場合（GitHub ログイン）",
			clerkUserID: "user_2abc123",
			username:    "cinephile-alex",
			email:       "alex@example.com",
			wantExact:   "cinephile-alex",
			checkValid:  true,
		},
		{
			name:        "Clerk username が空で email がある場合（Google ログイン）",
			clerkUserID: "user_2abc123",
			username:    "",
			email:       "yamada.taro@gmail.com",
			wantExact:   "yamada-taro",
			checkValid:  true,
		},
		{
			name:        "Clerk username も email も使えない場合（フォールバック）",
			clerkUserID: "user_2abc123def",
			username:    "",
			email:       "",
			wantExact:   "user-2abc123d",
			checkValid:  true,
		},
		{
			name:        "username に大文字が含まれる場合",
			clerkUserID: "user_2abc123",
			username:    "MovieFan123",
			email:       "",
			wantExact:   "moviefan123",
			checkValid:  true,
		},
		{
			name:        "username に記号が含まれる場合",
			clerkUserID: "user_2abc123",
			username:    "movie_fan@123",
			email:       "",
			wantExact:   "movie-fan-123",
			checkValid:  true,
		},
		{
			name:        "email のローカル部分が長い場合",
			clerkUserID: "user_2abc123",
			username:    "",
			email:       "this-is-a-very-long-email-address@example.com",
			wantExact:   "this-is-a-very-long",
			checkValid:  true,
		},
		{
			name:        "email のローカル部分に記号が含まれる場合",
			clerkUserID: "user_2abc123",
			username:    "",
			email:       "user+tag@example.com",
			wantExact:   "user-tag",
			checkValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateDisplayID(tt.clerkUserID, tt.username, tt.email)

			if tt.wantExact != "" && got != tt.wantExact {
				t.Errorf("GenerateDisplayID() = %v, want %v", got, tt.wantExact)
			}

			if tt.wantPrefix != "" && got[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("GenerateDisplayID() = %v, want prefix %v", got, tt.wantPrefix)
			}

			if tt.checkValid && !isValidDisplayID(got) {
				t.Errorf("GenerateDisplayID() = %v, but it's not a valid display_id", got)
			}
		})
	}
}

func TestSanitizeDisplayID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "小文字化",
			input: "MovieFan",
			want:  "moviefan",
		},
		{
			name:  "記号をハイフンに変換",
			input: "movie_fan@123",
			want:  "movie-fan-123",
		},
		{
			name:  "連続したハイフンを1つに",
			input: "movie---fan",
			want:  "movie-fan",
		},
		{
			name:  "先頭・末尾のハイフンを削除",
			input: "-movie-fan-",
			want:  "movie-fan",
		},
		{
			name:  "20文字を超える場合は切り詰め",
			input: "this-is-a-very-long-username-that-exceeds-limit",
			want:  "this-is-a-very-long",
		},
		{
			name:  "日本語はハイフンに変換される",
			input: "movie太郎fan",
			want:  "movie-fan",
		},
		{
			name:  "アンダースコアは削除",
			input: "movie_fan_123",
			want:  "movie-fan-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeDisplayID(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeDisplayID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidDisplayID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "有効なID（英小文字）",
			input: "moviefan",
			want:  true,
		},
		{
			name:  "有効なID（数字含む）",
			input: "moviefan123",
			want:  true,
		},
		{
			name:  "有効なID（ハイフン含む）",
			input: "movie-fan-123",
			want:  true,
		},
		{
			name:  "有効なID（最小長3文字）",
			input: "abc",
			want:  true,
		},
		{
			name:  "有効なID（最大長20文字）",
			input: "a2345678901234567890",
			want:  true,
		},
		{
			name:  "無効：2文字（短すぎる）",
			input: "ab",
			want:  false,
		},
		{
			name:  "無効：21文字（長すぎる）",
			input: "a23456789012345678901",
			want:  false,
		},
		{
			name:  "無効：先頭がハイフン",
			input: "-moviefan",
			want:  false,
		},
		{
			name:  "無効：末尾がハイフン",
			input: "moviefan-",
			want:  false,
		},
		{
			name:  "無効：大文字が含まれる",
			input: "MovieFan",
			want:  false,
		},
		{
			name:  "無効：アンダースコアが含まれる",
			input: "movie_fan",
			want:  false,
		},
		{
			name:  "無効：記号が含まれる",
			input: "movie@fan",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidDisplayID(tt.input)
			if got != tt.want {
				t.Errorf("isValidDisplayID(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
