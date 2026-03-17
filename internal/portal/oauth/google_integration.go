package oauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/sheets/v4"
)

// NewGoogleIntegrationConfig creates an OAuth2 config for full Google integration
// covering Gmail, Sheets, Calendar, and Drive.
func NewGoogleIntegrationConfig(clientID, clientSecret, baseURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  baseURL + "/integrations/google/callback",
		Scopes: []string{
			gmail.MailGoogleComScope,
			gmail.GmailSettingsBasicScope,
			gmail.GmailSettingsSharingScope,
			sheets.SpreadsheetsScope,
			drive.DriveFileScope,
			calendar.CalendarScope,
			drive.DriveScope,
		},
	}
}
