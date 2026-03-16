package gmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// withMessagesMock registers all message-related mock handlers on mux.
func withMessagesMock(mux *http.ServeMux) {
	// messages.list (GET) and messages.insert (POST)
	mux.HandleFunc("/gmail/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// messages.insert
			resp := map[string]string{
				"id":       "inserted1",
				"threadId": "thread-inserted1",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"messages": []map[string]string{
				{"id": "msg1", "threadId": "thread1"},
				{"id": "msg2", "threadId": "thread2"},
			},
			"resultSizeEstimate": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// messages.get msg1 (also handles DELETE for messages.delete)
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		msg := map[string]any{
			"id":       "msg1",
			"snippet":  "Hello world",
			"threadId": "thread1",
			"payload": map[string]any{
				"headers": []map[string]string{
					{"name": "From", "value": "alice@example.com"},
					{"name": "To", "value": "bob@example.com"},
					{"name": "Subject", "value": "Test Email"},
					{"name": "Date", "value": "Mon, 16 Mar 2026 10:00:00 -0500"},
					{"name": "Message-ID", "value": "<abc123@example.com>"},
				},
				"mimeType": "text/plain",
				"body": map[string]string{
					"data": "SGVsbG8gV29ybGQ=", // base64url of "Hello World"
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	// messages.get msg2
	mux.HandleFunc("/gmail/v1/users/me/messages/msg2", func(w http.ResponseWriter, r *http.Request) {
		msg := map[string]any{
			"id":       "msg2",
			"snippet":  "Second email",
			"threadId": "thread2",
			"payload": map[string]any{
				"headers": []map[string]string{
					{"name": "From", "value": "charlie@example.com"},
					{"name": "To", "value": "bob@example.com"},
					{"name": "Subject", "value": "Another Test"},
					{"name": "Date", "value": "Mon, 16 Mar 2026 11:00:00 -0500"},
				},
				"mimeType": "text/plain",
				"body": map[string]string{
					"data": "U2Vjb25kIGJvZHk=", // base64url of "Second body"
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msg)
	})

	// messages.send
	mux.HandleFunc("/gmail/v1/users/me/messages/send", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"id":       "sent1",
			"threadId": "thread-sent1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// messages.batchModify
	mux.HandleFunc("/gmail/v1/users/me/messages/batchModify", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// messages.batchDelete
	mux.HandleFunc("/gmail/v1/users/me/messages/batchDelete", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// messages.import
	mux.HandleFunc("/gmail/v1/users/me/messages/import", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"id":       "imported1",
			"threadId": "thread-imported1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// messages.trash, messages.untrash, messages.modify, messages.delete for msg1
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/trash", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "msg1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/untrash", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "msg1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/modify", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":       "msg1",
			"labelIds": []string{"INBOX", "STARRED"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withThreadsMock registers all thread-related mock handlers on mux.
func withThreadsMock(mux *http.ServeMux) {
	// threads.list
	mux.HandleFunc("/gmail/v1/users/me/threads", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"threads": []map[string]any{
				{"id": "thread1", "snippet": "First thread snippet", "historyId": "100"},
				{"id": "thread2", "snippet": "Second thread snippet", "historyId": "200"},
			},
			"resultSizeEstimate": 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// threads.get thread1 (also handles DELETE for threads.delete)
	mux.HandleFunc("/gmail/v1/users/me/threads/thread1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		thread := map[string]any{
			"id":        "thread1",
			"historyId": "100",
			"messages": []map[string]any{
				{
					"id":      "msg1",
					"snippet": "Hello world",
					"payload": map[string]any{
						"headers": []map[string]string{
							{"name": "From", "value": "alice@example.com"},
							{"name": "Subject", "value": "Test Email"},
							{"name": "Date", "value": "Mon, 16 Mar 2026 10:00:00 -0500"},
						},
					},
				},
				{
					"id":      "msg2",
					"snippet": "Second email",
					"payload": map[string]any{
						"headers": []map[string]string{
							{"name": "From", "value": "charlie@example.com"},
							{"name": "Subject", "value": "Re: Test Email"},
							{"name": "Date", "value": "Mon, 16 Mar 2026 11:00:00 -0500"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(thread)
	})

	// threads.trash
	mux.HandleFunc("/gmail/v1/users/me/threads/thread1/trash", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "thread1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// threads.untrash
	mux.HandleFunc("/gmail/v1/users/me/threads/thread1/untrash", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "thread1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// threads.modify
	mux.HandleFunc("/gmail/v1/users/me/threads/thread1/modify", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"id": "thread1"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withLabelsMock registers all label-related mock handlers on mux.
func withLabelsMock(mux *http.ServeMux) {
	// labels.list (GET) and labels.create (POST)
	mux.HandleFunc("/gmail/v1/users/me/labels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			resp := map[string]any{
				"id":             "Label_created1",
				"name":           "NewLabel",
				"type":           "user",
				"messagesTotal":  0,
				"messagesUnread": 0,
				"threadsTotal":   0,
				"threadsUnread":  0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"labels": []map[string]any{
				{
					"id":             "INBOX",
					"name":           "INBOX",
					"type":           "system",
					"messagesTotal":  42,
					"messagesUnread": 5,
					"threadsTotal":   38,
					"threadsUnread":  4,
				},
				{
					"id":             "SENT",
					"name":           "SENT",
					"type":           "system",
					"messagesTotal":  100,
					"messagesUnread": 0,
					"threadsTotal":   95,
					"threadsUnread":  0,
				},
				{
					"id":             "Label_1",
					"name":           "TestLabel",
					"type":           "user",
					"messagesTotal":  3,
					"messagesUnread": 1,
					"threadsTotal":   3,
					"threadsUnread":  1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// labels.get, labels.update (PUT), labels.patch (PATCH), labels.delete (DELETE) for Label_1
	mux.HandleFunc("/gmail/v1/users/me/labels/Label_1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodPut:
			resp := map[string]any{
				"id":             "Label_1",
				"name":           "UpdatedLabel",
				"type":           "user",
				"messagesTotal":  3,
				"messagesUnread": 1,
				"threadsTotal":   3,
				"threadsUnread":  1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		case http.MethodPatch:
			resp := map[string]any{
				"id":             "Label_1",
				"name":           "PatchedLabel",
				"type":           "user",
				"messagesTotal":  3,
				"messagesUnread": 1,
				"threadsTotal":   3,
				"threadsUnread":  1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		default:
			// GET
			resp := map[string]any{
				"id":                    "Label_1",
				"name":                  "TestLabel",
				"type":                  "user",
				"messagesTotal":         3,
				"messagesUnread":        1,
				"threadsTotal":          3,
				"threadsUnread":         1,
				"labelListVisibility":   "labelShow",
				"messageListVisibility": "show",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	})
}

// withDraftsMock registers all draft-related mock handlers on mux.
func withDraftsMock(mux *http.ServeMux) {
	// drafts.list (GET) and drafts.create (POST)
	mux.HandleFunc("/gmail/v1/users/me/drafts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			resp := map[string]any{
				"id": "draft1",
				"message": map[string]any{
					"id": "draftmsg1",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"drafts": []map[string]any{
				{
					"id": "draft1",
					"message": map[string]any{
						"id":      "draftmsg1",
						"snippet": "Draft preview...",
						"payload": map[string]any{
							"headers": []map[string]string{
								{"name": "From", "value": "user@example.com"},
								{"name": "To", "value": "recipient@example.com"},
								{"name": "Subject", "value": "Draft Subject"},
								{"name": "Date", "value": "Mon, 16 Mar 2026 10:00:00 -0500"},
							},
						},
					},
				},
				{
					"id": "draft2",
					"message": map[string]any{
						"id":      "draftmsg2",
						"snippet": "Another draft preview...",
						"payload": map[string]any{
							"headers": []map[string]string{
								{"name": "From", "value": "user@example.com"},
								{"name": "To", "value": "other@example.com"},
								{"name": "Subject", "value": "Another Draft"},
								{"name": "Date", "value": "Mon, 16 Mar 2026 11:00:00 -0500"},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// drafts.get, drafts.update (PUT), drafts.delete (DELETE) for draft1
	mux.HandleFunc("/gmail/v1/users/me/drafts/draft1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodPut:
			resp := map[string]any{
				"id": "draft1",
				"message": map[string]any{
					"id": "draftmsg1-updated",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		default:
			// GET
			resp := map[string]any{
				"id": "draft1",
				"message": map[string]any{
					"id":      "draftmsg1",
					"snippet": "Draft preview...",
					"payload": map[string]any{
						"headers": []map[string]string{
							{"name": "From", "value": "user@example.com"},
							{"name": "To", "value": "recipient@example.com"},
							{"name": "Subject", "value": "Draft Subject"},
							{"name": "Date", "value": "Mon, 16 Mar 2026 10:00:00 -0500"},
						},
						"mimeType": "text/plain",
						"body": map[string]string{
							"data": "RHJhZnQgYm9keSBjb250ZW50", // base64url of "Draft body content"
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	})

	// drafts.send
	mux.HandleFunc("/gmail/v1/users/me/drafts/send", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"id":       "sentmsg1",
			"threadId": "thread-sentmsg1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withAttachmentsMock registers attachment-related mock handlers on mux.
func withAttachmentsMock(mux *http.ServeMux) {
	// messages.attachments.get for att1 on msg1
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1/attachments/att1", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"attachmentId": "att1",
			"size":         1234,
			"data":         "SGVsbG8gV29ybGQ=", // base64 of "Hello World"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withHistoryMock registers history-related mock handlers on mux.
func withHistoryMock(mux *http.ServeMux) {
	mux.HandleFunc("/gmail/v1/users/me/history", func(w http.ResponseWriter, r *http.Request) {
		// The Gmail API serialises uint64 IDs as JSON strings (,string tag).
		resp := map[string]any{
			"history": []map[string]any{
				{
					"id": "12345",
					"messagesAdded": []map[string]any{
						{"message": map[string]any{"id": "msg1", "labelIds": []string{"INBOX"}}},
					},
				},
				{
					"id": "12346",
					"labelsAdded": []map[string]any{
						{
							"message":  map[string]any{"id": "msg2"},
							"labelIds": []string{"STARRED"},
						},
					},
				},
			},
			"historyId": "12347",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withSettingsMock registers all settings-related mock handlers on mux.
func withSettingsMock(mux *http.ServeMux) {
	// settings.getVacation (GET) and settings.updateVacation (PUT)
	// The Gmail API serialises startTime/endTime as JSON strings (int64 with ,string tag).
	mux.HandleFunc("/gmail/v1/users/me/settings/vacation", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"enableAutoReply":       true,
			"responseSubject":       "Out of office",
			"responseBodyPlainText": "I am out of office.",
			"restrictToContacts":    false,
			"restrictToDomain":      false,
			"startTime":             "1700000000000",
			"endTime":               "1700086400000",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// settings.getAutoForwarding (GET) and settings.updateAutoForwarding (PUT)
	mux.HandleFunc("/gmail/v1/users/me/settings/autoForwarding", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"enabled":      true,
			"emailAddress": "forward@example.com",
			"disposition":  "leaveInInbox",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// settings.getImap (GET) and settings.updateImap (PUT)
	mux.HandleFunc("/gmail/v1/users/me/settings/imap", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"enabled":         true,
			"autoExpunge":     true,
			"expungeBehavior": "archive",
			"maxFolderSize":   int64(0),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// settings.getPop (GET) and settings.updatePop (PUT)
	mux.HandleFunc("/gmail/v1/users/me/settings/pop", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"accessWindow": "allMail",
			"disposition":  "leaveInInbox",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// settings.getLanguage (GET) and settings.updateLanguage (PUT)
	mux.HandleFunc("/gmail/v1/users/me/settings/language", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"displayLanguage": "en",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withTokenMock registers a mock OAuth token endpoint on mux.
func withTokenMock(mux *http.ServeMux) {
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"access_token":  "new-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "new-refresh-token",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest.Server with all mock handlers registered.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withMessagesMock(mux)
	withThreadsMock(mux)
	withLabelsMock(mux)
	withDraftsMock(mux)
	withAttachmentsMock(mux)
	withHistoryMock(mux)
	withSettingsMock(mux)
	withTokenMock(mux)
	return httptest.NewServer(mux)
}

// buildTestThreadsCmd creates a `threads` subcommand tree for use in tests.
func buildTestThreadsCmd(factory ServiceFactory) *cobra.Command {
	threadsCmd := &cobra.Command{Use: "threads"}
	threadsCmd.AddCommand(newThreadsListCmd(factory))
	threadsCmd.AddCommand(newThreadsGetCmd(factory))
	threadsCmd.AddCommand(newThreadsTrashCmd(factory))
	threadsCmd.AddCommand(newThreadsUntrashCmd(factory))
	threadsCmd.AddCommand(newThreadsDeleteCmd(factory))
	threadsCmd.AddCommand(newThreadsModifyCmd(factory))
	return threadsCmd
}

// buildTestDraftsCmd creates a `drafts` subcommand tree for use in tests.
func buildTestDraftsCmd(factory ServiceFactory) *cobra.Command {
	draftsCmd := &cobra.Command{Use: "drafts"}
	draftsCmd.AddCommand(newDraftsListCmd(factory))
	draftsCmd.AddCommand(newDraftsGetCmd(factory))
	draftsCmd.AddCommand(newDraftsCreateCmd(factory))
	draftsCmd.AddCommand(newDraftsUpdateCmd(factory))
	draftsCmd.AddCommand(newDraftsSendCmd(factory))
	draftsCmd.AddCommand(newDraftsDeleteCmd(factory))
	return draftsCmd
}

// buildTestLabelsCmd creates a `labels` subcommand tree for use in tests.
func buildTestLabelsCmd(factory ServiceFactory) *cobra.Command {
	labelsCmd := &cobra.Command{Use: "labels"}
	labelsCmd.AddCommand(newLabelsListCmd(factory))
	labelsCmd.AddCommand(newLabelsGetCmd(factory))
	labelsCmd.AddCommand(newLabelsCreateCmd(factory))
	labelsCmd.AddCommand(newLabelsUpdateCmd(factory))
	labelsCmd.AddCommand(newLabelsPatchCmd(factory))
	labelsCmd.AddCommand(newLabelsDeleteCmd(factory))
	return labelsCmd
}

// buildTestAttachmentsCmd creates an `attachments` subcommand tree for use in tests.
func buildTestAttachmentsCmd(factory ServiceFactory) *cobra.Command {
	attachmentsCmd := &cobra.Command{Use: "attachments"}
	attachmentsCmd.AddCommand(newAttachmentsGetCmd(factory))
	return attachmentsCmd
}

// buildTestHistoryCmd creates a `history` subcommand tree for use in tests.
func buildTestHistoryCmd(factory ServiceFactory) *cobra.Command {
	historyCmd := &cobra.Command{Use: "history"}
	historyCmd.AddCommand(newHistoryListCmd(factory))
	return historyCmd
}

// buildTestSettingsCmd creates a `settings` subcommand tree for use in tests.
func buildTestSettingsCmd(factory ServiceFactory) *cobra.Command {
	settingsCmd := &cobra.Command{Use: "settings"}
	settingsCmd.AddCommand(newSettingsGetVacationCmd(factory))
	settingsCmd.AddCommand(newSettingsSetVacationCmd(factory))
	settingsCmd.AddCommand(newSettingsGetAutoForwardingCmd(factory))
	settingsCmd.AddCommand(newSettingsSetAutoForwardingCmd(factory))
	settingsCmd.AddCommand(newSettingsGetImapCmd(factory))
	settingsCmd.AddCommand(newSettingsSetImapCmd(factory))
	settingsCmd.AddCommand(newSettingsGetPopCmd(factory))
	settingsCmd.AddCommand(newSettingsSetPopCmd(factory))
	settingsCmd.AddCommand(newSettingsGetLanguageCmd(factory))
	settingsCmd.AddCommand(newSettingsSetLanguageCmd(factory))
	return settingsCmd
}

// newTestServiceFactory returns a ServiceFactory that creates a *gmail.Service
// backed by the given httptest server, bypassing OAuth entirely.
func newTestServiceFactory(server *httptest.Server) ServiceFactory {
	return func(ctx context.Context) (*api.Service, error) {
		return api.NewService(ctx,
			option.WithoutAuthentication(),
			option.WithEndpoint(server.URL+"/"),
			option.WithHTTPClient(server.Client()),
		)
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 65536)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}
