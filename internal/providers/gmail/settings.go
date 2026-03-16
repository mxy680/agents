package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// VacationInfo is the JSON-serializable representation of vacation responder settings.
type VacationInfo struct {
	EnableAutoReply       bool   `json:"enableAutoReply"`
	ResponseSubject       string `json:"responseSubject"`
	ResponseBodyPlainText string `json:"responseBodyPlainText"`
	RestrictToContacts    bool   `json:"restrictToContacts"`
	RestrictToDomain      bool   `json:"restrictToDomain"`
	StartTime             int64  `json:"startTime"`
	EndTime               int64  `json:"endTime"`
}

// AutoForwardingInfo is the JSON-serializable representation of auto-forwarding settings.
type AutoForwardingInfo struct {
	Enabled      bool   `json:"enabled"`
	EmailAddress string `json:"emailAddress"`
	Disposition  string `json:"disposition"`
}

// ImapInfo is the JSON-serializable representation of IMAP settings.
type ImapInfo struct {
	Enabled         bool   `json:"enabled"`
	AutoExpunge     bool   `json:"autoExpunge"`
	ExpungeBehavior string `json:"expungeBehavior"`
	MaxFolderSize   int64  `json:"maxFolderSize"`
}

// PopInfo is the JSON-serializable representation of POP settings.
type PopInfo struct {
	AccessWindow string `json:"accessWindow"`
	Disposition  string `json:"disposition"`
}

// LanguageInfo is the JSON-serializable representation of language settings.
type LanguageInfo struct {
	DisplayLanguage string `json:"displayLanguage"`
}

// --- settings get-vacation ---

// newSettingsGetVacationCmd returns the `settings get-vacation` command.
func newSettingsGetVacationCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get-vacation",
		Short: "Get vacation responder settings",
		RunE:  makeRunSettingsGetVacation(factory),
	}
}

func makeRunSettingsGetVacation(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		v, err := svc.Users.Settings.GetVacation("me").Do()
		if err != nil {
			return fmt.Errorf("getting vacation settings: %w", err)
		}

		info := VacationInfo{
			EnableAutoReply:       v.EnableAutoReply,
			ResponseSubject:       v.ResponseSubject,
			ResponseBodyPlainText: v.ResponseBodyPlainText,
			RestrictToContacts:    v.RestrictToContacts,
			RestrictToDomain:      v.RestrictToDomain,
			StartTime:             v.StartTime,
			EndTime:               v.EndTime,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Enable Auto Reply:    %v", info.EnableAutoReply),
			fmt.Sprintf("Response Subject:     %s", info.ResponseSubject),
			fmt.Sprintf("Response Body:        %s", info.ResponseBodyPlainText),
			fmt.Sprintf("Restrict To Contacts: %v", info.RestrictToContacts),
			fmt.Sprintf("Restrict To Domain:   %v", info.RestrictToDomain),
			fmt.Sprintf("Start Time:           %d", info.StartTime),
			fmt.Sprintf("End Time:             %d", info.EndTime),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings set-vacation ---

// newSettingsSetVacationCmd returns the `settings set-vacation` command.
func newSettingsSetVacationCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-vacation",
		Short: "Update vacation responder settings",
		RunE:  makeRunSettingsSetVacation(factory),
	}
	cmd.Flags().Bool("enable-auto-reply", false, "Enable vacation auto-reply")
	cmd.Flags().String("subject", "", "Vacation response subject")
	cmd.Flags().String("body", "", "Vacation response body (plain text)")
	cmd.Flags().Int64("start-time", 0, "Start time as Unix epoch milliseconds")
	cmd.Flags().Int64("end-time", 0, "End time as Unix epoch milliseconds")
	cmd.Flags().Bool("restrict-to-contacts", false, "Only send auto-reply to contacts")
	cmd.Flags().Bool("restrict-to-domain", false, "Only send auto-reply to users in your domain")
	return cmd
}

func makeRunSettingsSetVacation(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		enableAutoReply, _ := cmd.Flags().GetBool("enable-auto-reply")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		startTime, _ := cmd.Flags().GetInt64("start-time")
		endTime, _ := cmd.Flags().GetInt64("end-time")
		restrictToContacts, _ := cmd.Flags().GetBool("restrict-to-contacts")
		restrictToDomain, _ := cmd.Flags().GetBool("restrict-to-domain")

		settings := &api.VacationSettings{
			EnableAutoReply:       enableAutoReply,
			ResponseSubject:       subject,
			ResponseBodyPlainText: body,
			StartTime:             startTime,
			EndTime:               endTime,
			RestrictToContacts:    restrictToContacts,
			RestrictToDomain:      restrictToDomain,
			ForceSendFields:       []string{"EnableAutoReply", "RestrictToContacts", "RestrictToDomain"},
		}

		if cli.IsDryRun(cmd) {
			info := vacationFromAPI(settings)
			return dryRunResult(cmd, "Would update vacation responder settings", info)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Settings.UpdateVacation("me", settings).Do()
		if err != nil {
			return fmt.Errorf("updating vacation settings: %w", err)
		}

		info := vacationFromAPI(updated)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Println("Vacation responder updated")
		return nil
	}
}

// vacationFromAPI converts a Gmail API VacationSettings to VacationInfo.
func vacationFromAPI(v *api.VacationSettings) VacationInfo {
	return VacationInfo{
		EnableAutoReply:       v.EnableAutoReply,
		ResponseSubject:       v.ResponseSubject,
		ResponseBodyPlainText: v.ResponseBodyPlainText,
		RestrictToContacts:    v.RestrictToContacts,
		RestrictToDomain:      v.RestrictToDomain,
		StartTime:             v.StartTime,
		EndTime:               v.EndTime,
	}
}

// --- settings get-auto-forwarding ---

// newSettingsGetAutoForwardingCmd returns the `settings get-auto-forwarding` command.
func newSettingsGetAutoForwardingCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get-auto-forwarding",
		Short: "Get auto-forwarding settings",
		RunE:  makeRunSettingsGetAutoForwarding(factory),
	}
}

func makeRunSettingsGetAutoForwarding(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		af, err := svc.Users.Settings.GetAutoForwarding("me").Do()
		if err != nil {
			return fmt.Errorf("getting auto-forwarding settings: %w", err)
		}

		info := AutoForwardingInfo{
			Enabled:      af.Enabled,
			EmailAddress: af.EmailAddress,
			Disposition:  af.Disposition,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Enabled:       %v", info.Enabled),
			fmt.Sprintf("Email Address: %s", info.EmailAddress),
			fmt.Sprintf("Disposition:   %s", info.Disposition),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings set-auto-forwarding ---

// newSettingsSetAutoForwardingCmd returns the `settings set-auto-forwarding` command.
func newSettingsSetAutoForwardingCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-auto-forwarding",
		Short: "Update auto-forwarding settings",
		RunE:  makeRunSettingsSetAutoForwarding(factory),
	}
	cmd.Flags().Bool("enabled", false, "Enable auto-forwarding")
	cmd.Flags().String("email", "", "Forwarding email address")
	cmd.Flags().String("disposition", "", "Action for forwarded messages (leaveInInbox, archive, trash, markRead)")
	return cmd
}

func makeRunSettingsSetAutoForwarding(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		enabled, _ := cmd.Flags().GetBool("enabled")
		email, _ := cmd.Flags().GetString("email")
		disposition, _ := cmd.Flags().GetString("disposition")

		settings := &api.AutoForwarding{
			Enabled:      enabled,
			EmailAddress: email,
			Disposition:  disposition,
			ForceSendFields: []string{"Enabled"},
		}

		if cli.IsDryRun(cmd) {
			info := AutoForwardingInfo{
				Enabled:      settings.Enabled,
				EmailAddress: settings.EmailAddress,
				Disposition:  settings.Disposition,
			}
			return dryRunResult(cmd, "Would update auto-forwarding settings", info)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Settings.UpdateAutoForwarding("me", settings).Do()
		if err != nil {
			return fmt.Errorf("updating auto-forwarding settings: %w", err)
		}

		info := AutoForwardingInfo{
			Enabled:      updated.Enabled,
			EmailAddress: updated.EmailAddress,
			Disposition:  updated.Disposition,
		}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Println("Auto-forwarding updated")
		return nil
	}
}

// --- settings get-imap ---

// newSettingsGetImapCmd returns the `settings get-imap` command.
func newSettingsGetImapCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get-imap",
		Short: "Get IMAP settings",
		RunE:  makeRunSettingsGetImap(factory),
	}
}

func makeRunSettingsGetImap(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		imap, err := svc.Users.Settings.GetImap("me").Do()
		if err != nil {
			return fmt.Errorf("getting IMAP settings: %w", err)
		}

		info := ImapInfo{
			Enabled:         imap.Enabled,
			AutoExpunge:     imap.AutoExpunge,
			ExpungeBehavior: imap.ExpungeBehavior,
			MaxFolderSize:   imap.MaxFolderSize,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Enabled:          %v", info.Enabled),
			fmt.Sprintf("Auto Expunge:     %v", info.AutoExpunge),
			fmt.Sprintf("Expunge Behavior: %s", info.ExpungeBehavior),
			fmt.Sprintf("Max Folder Size:  %d", info.MaxFolderSize),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings set-imap ---

// newSettingsSetImapCmd returns the `settings set-imap` command.
func newSettingsSetImapCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-imap",
		Short: "Update IMAP settings",
		RunE:  makeRunSettingsSetImap(factory),
	}
	cmd.Flags().Bool("enabled", false, "Enable IMAP access")
	cmd.Flags().Bool("auto-expunge", false, "Auto-expunge messages")
	cmd.Flags().Int64("max-folder-size", 0, "Maximum folder size (0 = unlimited)")
	cmd.Flags().String("expunge-behavior", "", "Expunge behavior (archive, deleteForever, trash)")
	return cmd
}

func makeRunSettingsSetImap(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		enabled, _ := cmd.Flags().GetBool("enabled")
		autoExpunge, _ := cmd.Flags().GetBool("auto-expunge")
		maxFolderSize, _ := cmd.Flags().GetInt64("max-folder-size")
		expungeBehavior, _ := cmd.Flags().GetString("expunge-behavior")

		settings := &api.ImapSettings{
			Enabled:         enabled,
			AutoExpunge:     autoExpunge,
			MaxFolderSize:   maxFolderSize,
			ExpungeBehavior: expungeBehavior,
			ForceSendFields: []string{"Enabled", "AutoExpunge"},
		}

		if cli.IsDryRun(cmd) {
			info := ImapInfo{
				Enabled:         settings.Enabled,
				AutoExpunge:     settings.AutoExpunge,
				ExpungeBehavior: settings.ExpungeBehavior,
				MaxFolderSize:   settings.MaxFolderSize,
			}
			return dryRunResult(cmd, "Would update IMAP settings", info)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Settings.UpdateImap("me", settings).Do()
		if err != nil {
			return fmt.Errorf("updating IMAP settings: %w", err)
		}

		info := ImapInfo{
			Enabled:         updated.Enabled,
			AutoExpunge:     updated.AutoExpunge,
			ExpungeBehavior: updated.ExpungeBehavior,
			MaxFolderSize:   updated.MaxFolderSize,
		}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Println("IMAP settings updated")
		return nil
	}
}

// --- settings get-pop ---

// newSettingsGetPopCmd returns the `settings get-pop` command.
func newSettingsGetPopCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get-pop",
		Short: "Get POP settings",
		RunE:  makeRunSettingsGetPop(factory),
	}
}

func makeRunSettingsGetPop(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		pop, err := svc.Users.Settings.GetPop("me").Do()
		if err != nil {
			return fmt.Errorf("getting POP settings: %w", err)
		}

		info := PopInfo{
			AccessWindow: pop.AccessWindow,
			Disposition:  pop.Disposition,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Access Window: %s", info.AccessWindow),
			fmt.Sprintf("Disposition:   %s", info.Disposition),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings set-pop ---

// newSettingsSetPopCmd returns the `settings set-pop` command.
func newSettingsSetPopCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-pop",
		Short: "Update POP settings",
		RunE:  makeRunSettingsSetPop(factory),
	}
	cmd.Flags().String("access-window", "", "POP access window (disabled, allMail, fromNowOn) (required)")
	cmd.Flags().String("disposition", "", "Action for accessed messages (leaveInInbox, archive, trash, markRead)")
	_ = cmd.MarkFlagRequired("access-window")
	return cmd
}

func makeRunSettingsSetPop(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		accessWindow, _ := cmd.Flags().GetString("access-window")
		disposition, _ := cmd.Flags().GetString("disposition")

		settings := &api.PopSettings{
			AccessWindow: accessWindow,
			Disposition:  disposition,
		}

		if cli.IsDryRun(cmd) {
			info := PopInfo{
				AccessWindow: settings.AccessWindow,
				Disposition:  settings.Disposition,
			}
			return dryRunResult(cmd, "Would update POP settings", info)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Settings.UpdatePop("me", settings).Do()
		if err != nil {
			return fmt.Errorf("updating POP settings: %w", err)
		}

		info := PopInfo{
			AccessWindow: updated.AccessWindow,
			Disposition:  updated.Disposition,
		}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Println("POP settings updated")
		return nil
	}
}

// --- settings get-language ---

// newSettingsGetLanguageCmd returns the `settings get-language` command.
func newSettingsGetLanguageCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get-language",
		Short: "Get language settings",
		RunE:  makeRunSettingsGetLanguage(factory),
	}
}

func makeRunSettingsGetLanguage(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		lang, err := svc.Users.Settings.GetLanguage("me").Do()
		if err != nil {
			return fmt.Errorf("getting language settings: %w", err)
		}

		info := LanguageInfo{
			DisplayLanguage: lang.DisplayLanguage,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		fmt.Printf("Display language: %s\n", info.DisplayLanguage)
		return nil
	}
}

// --- settings set-language ---

// newSettingsSetLanguageCmd returns the `settings set-language` command.
func newSettingsSetLanguageCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-language",
		Short: "Update language settings",
		RunE:  makeRunSettingsSetLanguage(factory),
	}
	cmd.Flags().String("display-language", "", "Display language code (e.g. en, fr, de) (required)")
	_ = cmd.MarkFlagRequired("display-language")
	return cmd
}

func makeRunSettingsSetLanguage(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		displayLanguage, _ := cmd.Flags().GetString("display-language")

		if cli.IsDryRun(cmd) {
			info := LanguageInfo{DisplayLanguage: displayLanguage}
			return dryRunResult(cmd, fmt.Sprintf("Would update language to %s", displayLanguage), info)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Settings.UpdateLanguage("me", &api.LanguageSettings{
			DisplayLanguage: displayLanguage,
		}).Do()
		if err != nil {
			return fmt.Errorf("updating language settings: %w", err)
		}

		info := LanguageInfo{DisplayLanguage: updated.DisplayLanguage}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Language updated to %s\n", updated.DisplayLanguage)
		return nil
	}
}
