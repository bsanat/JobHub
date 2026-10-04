package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetMeAndInvalidTokenAreSafe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.URL.Path, "/getMe") {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"ok":true,"result":{"id":42,"is_bot":true,"username":"jhubkz_bot"}}`)
		}))
		defer server.Close()
		client, err := newTestClient("test-token", server.URL, server.Client())
		if err != nil {
			t.Fatal(err)
		}
		bot, err := client.GetMe(context.Background())
		if err != nil || bot.Username != "jhubkz_bot" {
			t.Fatalf("unexpected getMe result: %#v %v", bot, err)
		}
	})

	t.Run("redacted auth failure", func(t *testing.T) {
		const secret = "123456:VERY_SECRET_VALUE"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"ok":false,"error_code":401,"description":"Unauthorized"}`)
		}))
		defer server.Close()
		client, err := newTestClient(secret, server.URL, server.Client())
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.GetMe(context.Background())
		if err == nil || strings.Contains(err.Error(), secret) || err.Error() != "TELEGRAM_AUTH_FAILED" {
			t.Fatalf("unsafe or unexpected error: %v", err)
		}
	})
}

func TestGetChannelAndRateLimit(t *testing.T) {
	t.Run("explicit channel lookup", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"chat_id":"@jobhub_test"`) {
				t.Fatalf("unexpected lookup payload: %s", body)
			}
			_, _ = io.WriteString(w, `{"ok":true,"result":{"id":-10042,"type":"channel","title":"JobHub Test","username":"jobhub_test"}}`)
		}))
		defer server.Close()
		client, _ := newTestClient("test-token", server.URL, server.Client())
		chat, err := client.GetChannel(context.Background(), "@jobhub_test")
		if err != nil || chat.ID != -10042 {
			t.Fatalf("lookup failed: %#v %v", chat, err)
		}
	})

	t.Run("retry after", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"ok":false,"error_code":429,"parameters":{"retry_after":7}}`)
		}))
		defer server.Close()
		client, _ := newTestClient("test-token", server.URL, server.Client())
		_, err := client.GetUpdates(context.Background(), 0)
		apiErr, ok := err.(*APIError)
		if !ok || apiErr.Category != "TELEGRAM_RATE_LIMITED" || apiErr.RetryAfter != 7*time.Second {
			t.Fatalf("unexpected rate limit: %#v", err)
		}
	})
}

func TestVerifyChannelRequiresMatchingIDAndPostPermission(t *testing.T) {
	for _, test := range []struct {
		name      string
		channelID int64
		canPost   bool
		wantError string
	}{
		{name: "allowed", channelID: -10042, canPost: true},
		{name: "wrong id", channelID: -10099, canPost: true, wantError: "TELEGRAM_CHANNEL_ID_MISMATCH"},
		{name: "missing post permission", channelID: -10042, canPost: false, wantError: "TELEGRAM_CHANNEL_POST_FORBIDDEN"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/getChat"):
					_, _ = io.WriteString(w, `{"ok":true,"result":{"id":-10042,"type":"channel","username":"jobhub_test"}}`)
				case strings.HasSuffix(r.URL.Path, "/getMe"):
					_, _ = io.WriteString(w, `{"ok":true,"result":{"id":42,"username":"jhubkz_bot"}}`)
				case strings.HasSuffix(r.URL.Path, "/getChatMember"):
					body, _ := io.ReadAll(r.Body)
					if !strings.Contains(string(body), `"chat_id":-10042`) || !strings.Contains(string(body), `"user_id":42`) {
						t.Fatalf("wrong membership lookup: %s", body)
					}
					if test.canPost {
						_, _ = io.WriteString(w, `{"ok":true,"result":{"status":"administrator","can_post_messages":true}}`)
					} else {
						_, _ = io.WriteString(w, `{"ok":true,"result":{"status":"administrator","can_post_messages":false}}`)
					}
				default:
					t.Fatalf("unexpected Telegram method: %s", r.URL.Path)
				}
			}))
			defer server.Close()
			client, _ := newTestClient("test-token", server.URL, server.Client())
			err := client.VerifyChannel(context.Background(), test.channelID, "jobhub_test")
			if test.wantError == "" && err != nil || test.wantError != "" && (err == nil || err.Error() != test.wantError) {
				t.Fatalf("unexpected channel verification: %v", err)
			}
		})
	}
}

func TestLatestPendingUsesNonblockingNegativeOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		value := string(body)
		if !strings.Contains(value, `"offset":-1`) || !strings.Contains(value, `"timeout":0`) {
			t.Fatalf("unexpected baseline payload: %s", value)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":[]}`)
	}))
	defer server.Close()
	client, _ := newTestClient("test-token", server.URL, server.Client())
	updates, err := client.LatestPending(context.Background())
	if err != nil || len(updates) != 0 {
		t.Fatalf("baseline failed: %#v %v", updates, err)
	}
}

func TestSendMessageReturnsMessageID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		value := string(body)
		if !strings.Contains(value, `"chat_id":-10042`) || !strings.Contains(value, `"text":"Backend Engineer"`) || !strings.Contains(value, `"disable_web_page_preview":true`) {
			t.Fatalf("unexpected payload: %s", value)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":77}}`)
	}))
	defer server.Close()
	client, _ := newTestClient("test-token", server.URL, server.Client())
	messageID, err := client.SendMessageAndReturnID(context.Background(), -10042, "Backend Engineer")
	if err != nil || messageID != 77 {
		t.Fatalf("send failed: %d %v", messageID, err)
	}
}

func TestSendHTMLMessageSetsParseMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Text      string `json:"text"`
			ParseMode string `json:"parse_mode"`
		}
		if err := json.Unmarshal(body, &payload); err != nil || payload.ParseMode != "HTML" || payload.Text != "<b>Backend</b>" {
			t.Fatalf("unexpected formatted payload: %s", body)
		}
		_, _ = io.WriteString(w, `{"ok":true,"result":{"message_id":78}}`)
	}))
	defer server.Close()
	client, _ := newTestClient("test-token", server.URL, server.Client())
	messageID, err := client.SendHTMLMessageAndReturnID(context.Background(), -10042, "<b>Backend</b>")
	if err != nil || messageID != 78 {
		t.Fatalf("formatted send failed: %d %v", messageID, err)
	}
}

func TestMapUpdateAllowlistTextCaptionAndLinks(t *testing.T) {
	allowed := map[int64]struct{}{-10042: {}}
	base := message{MessageID: 7, Date: 1_700_000_000, Text: "Вакансия: Go developer"}
	base.Chat.ID, base.Chat.Type, base.Chat.Username, base.Chat.Title = -10042, "channel", "jobhub_test", "JobHub Test"

	post, reason, handled := MapUpdate(Update{ID: 1, ChannelPost: &base}, allowed, time.Now())
	if !handled || reason != "" || post.Text != base.Text || post.OriginalURL != "https://t.me/jobhub_test/7" {
		t.Fatalf("text post not mapped: %#v %q %v", post, reason, handled)
	}

	caption := base
	caption.Text, caption.Caption = "", "Лауазым: Инженер"
	post, reason, handled = MapUpdate(Update{ID: 2, ChannelPost: &caption}, allowed, time.Now())
	if !handled || reason != "" || post.Text != caption.Caption {
		t.Fatalf("caption post not mapped: %#v %q", post, reason)
	}

	unknown := base
	unknown.Chat.ID = -10099
	_, reason, handled = MapUpdate(Update{ID: 3, ChannelPost: &unknown}, allowed, time.Now())
	if !handled || reason != SkipSourceNotAllowed {
		t.Fatalf("unknown channel was not rejected: %q", reason)
	}

	private := base
	private.Chat.Username = ""
	post, reason, handled = MapUpdate(Update{ID: 4, ChannelPost: &private}, allowed, time.Now())
	if !handled || reason != "" || post.OriginalURL != "" {
		t.Fatalf("private channel URL was invented: %#v %q", post, reason)
	}
}

func TestNormalizeDeterministicConservativeAndStable(t *testing.T) {
	post := Post{ChatID: -10042, MessageID: 7, ChannelTitle: "JobHub Test", OriginalURL: "https://t.me/jobhub_test/7",
		Text: "Вакансия: Go-разработчик\nКомпания: Example\nГород: Алматы\nЗарплата: 500 000 ₸\nОписание без перевода.", TelegramDate: time.Unix(1_700_000_000, 0)}
	job, reason := Normalize(post)
	if reason != "" || job.Source != "telegram:-10042" || job.ExternalID != "telegram:-10042:7" || job.Title != "Go-разработчик" || job.CompanyNameRaw != "Example" || job.LocationRaw != "Алматы" || job.SalaryRaw != "500 000 ₸" || job.Description != post.Text {
		t.Fatalf("unexpected normalized job: %#v reason=%q", job, reason)
	}
	if job.CanonicalCityID != "" || job.EmploymentTypeRaw != "" {
		t.Fatalf("unknown values were invented: %#v", job)
	}

	for name, test := range map[string]struct {
		post Post
		want SkipReason
	}{
		"empty media":   {Post{OriginalURL: post.OriginalURL}, SkipEmptyText},
		"private":       {Post{ChatID: -1, MessageID: 1, Text: "Вакансия: Dev"}, SkipMissingSourceURL},
		"missing title": {Post{ChatID: -1, MessageID: 1, Text: "Просто текст", OriginalURL: "https://t.me/tester/1"}, SkipMissingTitle},
	} {
		t.Run(name, func(t *testing.T) {
			_, got := Normalize(test.post)
			if got != test.want {
				t.Fatalf("got %q want %q", got, test.want)
			}
		})
	}
}

func TestOffsetRoundTrip(t *testing.T) {
	path := t.TempDir() + "/state/offset"
	if got, err := LoadOffset(path); err != nil || got != 0 {
		t.Fatalf("initial offset: %d %v", got, err)
	}
	if err := SaveOffset(path, 44); err != nil {
		t.Fatal(err)
	}
	if got, err := LoadOffset(path); err != nil || got != 44 {
		t.Fatalf("saved offset: %d %v", got, err)
	}
}
