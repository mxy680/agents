package framer

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/emdash-projects/agents/internal/cli"
)

func newNodesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Manage Framer canvas nodes",
	}
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	cmd.AddCommand(
		newNodesGetCmd(factory),
		newNodesChildrenCmd(factory),
		newNodesParentCmd(factory),
		newNodesListByTypeCmd(factory),
		newNodesCreateFrameCmd(factory),
		newNodesCreateTextCmd(factory),
		newNodesCreateComponentCmd(factory),
		newNodesCreateWebPageCmd(factory),
		newNodesCreateDesignPageCmd(factory),
		newNodesCloneCmd(factory),
		newNodesRemoveCmd(factory),
		newNodesSetAttributesCmd(factory),
		newNodesSetParentCmd(factory),
		newNodesRectCmd(factory),
	)

	return cmd
}

func newNodesGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a node by ID",
		RunE:  makeRunNodesGet(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID (required)")
	_ = cmd.MarkFlagRequired("node-id")
	return cmd
}

func makeRunNodesGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		nodeID, _ := cmd.Flags().GetString("node-id")
		result, err := client.Call("getNode", map[string]any{"nodeId": nodeID})
		if err != nil {
			return fmt.Errorf("get node: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("ID:   %s", node.ID),
			fmt.Sprintf("Name: %s", node.Name),
			fmt.Sprintf("Type: %s", node.Type),
		})
	}
}

func newNodesChildrenCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "children",
		Short: "Get children of a node",
		RunE:  makeRunNodesChildren(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID (required)")
	_ = cmd.MarkFlagRequired("node-id")
	return cmd
}

func makeRunNodesChildren(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		nodeID, _ := cmd.Flags().GetString("node-id")
		result, err := client.Call("getChildren", map[string]any{"nodeId": nodeID})
		if err != nil {
			return fmt.Errorf("get children: %w", err)
		}

		var nodes []NodeSummary
		if err := json.Unmarshal(result, &nodes); err != nil {
			return fmt.Errorf("parse children: %w", err)
		}

		lines := make([]string, 0, len(nodes))
		for _, n := range nodes {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%s", n.ID, n.Type, n.Name))
		}

		return cli.PrintResult(cmd, nodes, lines)
	}
}

func newNodesParentCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parent",
		Short: "Get the parent of a node",
		RunE:  makeRunNodesParent(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID (required)")
	_ = cmd.MarkFlagRequired("node-id")
	return cmd
}

func makeRunNodesParent(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		nodeID, _ := cmd.Flags().GetString("node-id")
		result, err := client.Call("getParent", map[string]any{"nodeId": nodeID})
		if err != nil {
			return fmt.Errorf("get parent: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse parent: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("ID:   %s", node.ID),
			fmt.Sprintf("Name: %s", node.Name),
			fmt.Sprintf("Type: %s", node.Type),
		})
	}
}

func newNodesListByTypeCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-by-type",
		Short: "List nodes of a given type",
		RunE:  makeRunNodesListByType(factory),
	}
	cmd.Flags().String("type", "", "Node type (FrameNode|TextNode|SVGNode|ComponentInstanceNode|WebPageNode|DesignPageNode|ComponentNode) (required)")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func makeRunNodesListByType(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		nodeType, _ := cmd.Flags().GetString("type")
		result, err := client.Call("getNodesWithType", map[string]any{"type": nodeType})
		if err != nil {
			return fmt.Errorf("list nodes by type: %w", err)
		}

		var nodes []NodeSummary
		if err := json.Unmarshal(result, &nodes); err != nil {
			return fmt.Errorf("parse nodes: %w", err)
		}

		lines := make([]string, 0, len(nodes))
		for _, n := range nodes {
			lines = append(lines, fmt.Sprintf("%s\t%s", n.ID, n.Name))
		}

		return cli.PrintResult(cmd, nodes, lines)
	}
}

func newNodesCreateFrameCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-frame",
		Short: "Create a new frame node",
		RunE:  makeRunNodesCreateFrame(factory),
	}
	cmd.Flags().String("attributes", "", "Node attributes as JSON")
	cmd.Flags().String("parent-id", "", "Parent node ID")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunNodesCreateFrame(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		attrsStr, _ := cmd.Flags().GetString("attributes")
		parentID, _ := cmd.Flags().GetString("parent-id")

		params := map[string]any{}
		if attrsStr != "" {
			attrs, err := parseJSONFlag(attrsStr)
			if err != nil {
				return fmt.Errorf("parse attributes: %w", err)
			}
			params["attributes"] = attrs
		}
		if parentID != "" {
			params["parentId"] = parentID
		}

		if isDry, err := dryRunResult(cmd, "create frame node", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createFrameNode", params)
		if err != nil {
			return fmt.Errorf("create frame node: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Created frame node: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesCreateTextCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-text",
		Short: "Create a new text node",
		RunE:  makeRunNodesCreateText(factory),
	}
	cmd.Flags().String("attributes", "", "Node attributes as JSON")
	cmd.Flags().String("parent-id", "", "Parent node ID")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunNodesCreateText(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		attrsStr, _ := cmd.Flags().GetString("attributes")
		parentID, _ := cmd.Flags().GetString("parent-id")

		params := map[string]any{}
		if attrsStr != "" {
			attrs, err := parseJSONFlag(attrsStr)
			if err != nil {
				return fmt.Errorf("parse attributes: %w", err)
			}
			params["attributes"] = attrs
		}
		if parentID != "" {
			params["parentId"] = parentID
		}

		if isDry, err := dryRunResult(cmd, "create text node", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createTextNode", params)
		if err != nil {
			return fmt.Errorf("create text node: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Created text node: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesCreateComponentCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-component",
		Short: "Create a new component node",
		RunE:  makeRunNodesCreateComponent(factory),
	}
	cmd.Flags().String("name", "", "Component name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunNodesCreateComponent(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		params := map[string]any{"name": name}

		if isDry, err := dryRunResult(cmd, "create component node", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createComponentNode", params)
		if err != nil {
			return fmt.Errorf("create component node: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Created component node: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesCreateWebPageCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-web-page",
		Short: "Create a new web page node",
		RunE:  makeRunNodesCreateWebPage(factory),
	}
	cmd.Flags().String("path", "", "Page path (required)")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunNodesCreateWebPage(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("path")
		params := map[string]any{"path": path}

		if isDry, err := dryRunResult(cmd, "create web page", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createWebPage", params)
		if err != nil {
			return fmt.Errorf("create web page: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Created web page: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesCreateDesignPageCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-design-page",
		Short: "Create a new design page node",
		RunE:  makeRunNodesCreateDesignPage(factory),
	}
	cmd.Flags().String("name", "", "Design page name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunNodesCreateDesignPage(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		params := map[string]any{"name": name}

		if isDry, err := dryRunResult(cmd, "create design page", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createDesignPage", params)
		if err != nil {
			return fmt.Errorf("create design page: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Created design page: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesCloneCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a node",
		RunE:  makeRunNodesClone(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID to clone (required)")
	_ = cmd.MarkFlagRequired("node-id")
	cmd.Flags().Bool("dry-run", false, "Preview without cloning")
	return cmd
}

func makeRunNodesClone(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		nodeID, _ := cmd.Flags().GetString("node-id")
		params := map[string]any{"nodeId": nodeID}

		if isDry, err := dryRunResult(cmd, fmt.Sprintf("clone node %s", nodeID), params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("cloneNode", params)
		if err != nil {
			return fmt.Errorf("clone node: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Cloned node: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesRemoveCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove nodes by ID",
		RunE:  makeRunNodesRemove(factory),
	}
	cmd.Flags().String("node-ids", "", "Comma-separated node IDs to remove (required)")
	_ = cmd.MarkFlagRequired("node-ids")
	cmd.Flags().Bool("confirm", false, "Confirm destructive operation")
	cmd.Flags().Bool("dry-run", false, "Preview without removing")
	return cmd
}

func makeRunNodesRemove(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		nodeIDsStr, _ := cmd.Flags().GetString("node-ids")
		nodeIDs := parseStringList(nodeIDsStr)
		params := map[string]any{"nodeIds": nodeIDs}

		if isDry, err := dryRunResult(cmd, fmt.Sprintf("remove nodes %v", nodeIDs), params); isDry {
			return err
		}

		if !confirmDestructive(cmd, fmt.Sprintf("remove nodes %v", nodeIDs)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("removeNodes", params)
		if err != nil {
			return fmt.Errorf("remove nodes: %w", err)
		}

		return cli.PrintResult(cmd, result, []string{
			fmt.Sprintf("Removed %d node(s)", len(nodeIDs)),
		})
	}
}

func newNodesSetAttributesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-attributes",
		Short: "Set attributes on a node",
		RunE:  makeRunNodesSetAttributes(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID (required)")
	_ = cmd.MarkFlagRequired("node-id")
	cmd.Flags().String("attributes", "", "Attributes as JSON (required)")
	_ = cmd.MarkFlagRequired("attributes")
	cmd.Flags().Bool("dry-run", false, "Preview without updating")
	return cmd
}

func makeRunNodesSetAttributes(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		nodeID, _ := cmd.Flags().GetString("node-id")
		attrsStr, _ := cmd.Flags().GetString("attributes")

		attrs, err := parseJSONFlag(attrsStr)
		if err != nil {
			return fmt.Errorf("parse attributes: %w", err)
		}

		params := map[string]any{
			"nodeId":     nodeID,
			"attributes": attrs,
		}

		if isDry, err := dryRunResult(cmd, fmt.Sprintf("set attributes on node %s", nodeID), params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setAttributes", params)
		if err != nil {
			return fmt.Errorf("set attributes: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Updated node: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesSetParentCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-parent",
		Short: "Set the parent of a node",
		RunE:  makeRunNodesSetParent(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID (required)")
	_ = cmd.MarkFlagRequired("node-id")
	cmd.Flags().String("parent-id", "", "Parent node ID (required)")
	_ = cmd.MarkFlagRequired("parent-id")
	cmd.Flags().Int("index", -1, "Position index within parent")
	cmd.Flags().Bool("dry-run", false, "Preview without updating")
	return cmd
}

func makeRunNodesSetParent(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		nodeID, _ := cmd.Flags().GetString("node-id")
		parentID, _ := cmd.Flags().GetString("parent-id")
		index, _ := cmd.Flags().GetInt("index")

		params := map[string]any{
			"nodeId":   nodeID,
			"parentId": parentID,
		}
		if index >= 0 {
			params["index"] = index
		}

		if isDry, err := dryRunResult(cmd, fmt.Sprintf("set parent of node %s to %s", nodeID, parentID), params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setParent", params)
		if err != nil {
			return fmt.Errorf("set parent: %w", err)
		}

		var node NodeSummary
		if err := json.Unmarshal(result, &node); err != nil {
			return fmt.Errorf("parse node: %w", err)
		}

		return cli.PrintResult(cmd, node, []string{
			fmt.Sprintf("Updated parent of node: %s (%s)", node.Name, node.ID),
		})
	}
}

func newNodesRectCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rect",
		Short: "Get the bounding rectangle of a node",
		RunE:  makeRunNodesRect(factory),
	}
	cmd.Flags().String("node-id", "", "Node ID (required)")
	_ = cmd.MarkFlagRequired("node-id")
	return cmd
}

func makeRunNodesRect(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		nodeID, _ := cmd.Flags().GetString("node-id")
		result, err := client.Call("getRect", map[string]any{"nodeId": nodeID})
		if err != nil {
			return fmt.Errorf("get rect: %w", err)
		}

		var rect Rect
		if err := json.Unmarshal(result, &rect); err != nil {
			return fmt.Errorf("parse rect: %w", err)
		}

		return cli.PrintResult(cmd, rect, []string{
			fmt.Sprintf("X:      %s", strconv.FormatFloat(rect.X, 'f', -1, 64)),
			fmt.Sprintf("Y:      %s", strconv.FormatFloat(rect.Y, 'f', -1, 64)),
			fmt.Sprintf("Width:  %s", strconv.FormatFloat(rect.Width, 'f', -1, 64)),
			fmt.Sprintf("Height: %s", strconv.FormatFloat(rect.Height, 'f', -1, 64)),
		})
	}
}
