package imessage

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newFindMyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "findmy",
		Short:   "Find My devices and friends",
		Aliases: []string{"fm"},
	}

	cmd.AddCommand(newFindMyDevicesCmd(factory))
	cmd.AddCommand(newFindMyDevicesRefreshCmd(factory))
	cmd.AddCommand(newFindMyFriendsCmd(factory))
	cmd.AddCommand(newFindMyFriendsRefreshCmd(factory))

	return cmd
}

func newFindMyDevicesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List FindMy devices",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunFindMyDevices(factory)
	return cmd
}

func makeRunFindMyDevices(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "icloud/findmy/devices", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var devices []FindMyDevice
		if err := json.Unmarshal(data, &devices); err != nil {
			// data may be a single object or non-standard; fall back to raw
			return printResult(cmd, data, []string{string(data)})
		}

		lines := make([]string, 0, len(devices))
		for _, d := range devices {
			battery := ""
			if d.Battery > 0 {
				battery = fmt.Sprintf(" (battery: %.0f%%)", d.Battery*100)
			}
			name := d.Name
			if name == "" {
				name = d.ID
			}
			lines = append(lines, fmt.Sprintf("%-30s  %s%s", truncate(name, 28), d.ID, battery))
		}
		if len(lines) == 0 {
			lines = []string{"No devices found."}
		}

		return printResult(cmd, devices, lines)
	}
}

func newFindMyDevicesRefreshCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices-refresh",
		Short: "Refresh FindMy devices",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunFindMyDevicesRefresh(factory)
	return cmd
}

func makeRunFindMyDevicesRefresh(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "icloud/findmy/devices/refresh", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{"FindMy devices refreshed."})
	}
}

func newFindMyFriendsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "friends",
		Short: "List FindMy friends",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunFindMyFriends(factory)
	return cmd
}

func makeRunFindMyFriends(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "icloud/findmy/friends", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var friends []FindMyFriend
		if err := json.Unmarshal(data, &friends); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		lines := make([]string, 0, len(friends))
		for _, f := range friends {
			name := f.Name
			if name == "" {
				name = f.Handle
			}
			if name == "" {
				name = f.ID
			}
			lines = append(lines, fmt.Sprintf("%-30s  %s", truncate(name, 28), f.ID))
		}
		if len(lines) == 0 {
			lines = []string{"No friends found."}
		}

		return printResult(cmd, friends, lines)
	}
}

func newFindMyFriendsRefreshCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "friends-refresh",
		Short: "Refresh FindMy friends",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunFindMyFriendsRefresh(factory)
	return cmd
}

func makeRunFindMyFriendsRefresh(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "icloud/findmy/friends/refresh", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{"FindMy friends refreshed."})
	}
}
