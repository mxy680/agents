package places

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
)

func newPhotosListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List photos for a place",
		Long:  "Retrieve photo references for a place (use photo names with 'photos get' to download).",
		RunE:  makeRunPhotosList(factory),
	}
	cmd.Flags().String("place-id", "", "Place ID (required)")
	_ = cmd.MarkFlagRequired("place-id")
	return cmd
}

func makeRunPhotosList(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		placeID, _ := cmd.Flags().GetString("place-id")
		name := "places/" + placeID

		place, err := svc.Places.Get(name).Fields(googleapi.Field("photos")).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("get place photos: %w", err)
		}

		photos := make([]PhotoReference, 0, len(place.Photos))
		for _, ph := range place.Photos {
			photos = append(photos, PhotoReference{
				Name:     ph.Name,
				WidthPx:  ph.WidthPx,
				HeightPx: ph.HeightPx,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(photos)
		}

		if len(photos) == 0 {
			fmt.Println("No photos found.")
			return nil
		}

		lines := make([]string, 0, len(photos)+1)
		lines = append(lines, fmt.Sprintf("%-70s  %s", "PHOTO NAME", "DIMENSIONS"))
		for _, p := range photos {
			lines = append(lines, fmt.Sprintf("%-70s  %dx%d", p.Name, p.WidthPx, p.HeightPx))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPhotosGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a place photo",
		Long:  "Download a place photo or get its URL. Use photo names from 'photos list'.",
		RunE:  makeRunPhotosGet(factory),
	}
	cmd.Flags().String("photo-name", "", "Full photo resource name (required)")
	cmd.Flags().Int("max-width", 800, "Max width in pixels (1-4800)")
	cmd.Flags().Int("max-height", 0, "Max height in pixels (1-4800)")
	cmd.Flags().String("output", "", "File path to save the photo")
	cmd.Flags().Bool("url-only", false, "Only return the photo URL, don't download")
	_ = cmd.MarkFlagRequired("photo-name")
	return cmd
}

func makeRunPhotosGet(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		photoName, _ := cmd.Flags().GetString("photo-name")
		maxWidth, _ := cmd.Flags().GetInt("max-width")
		maxHeight, _ := cmd.Flags().GetInt("max-height")
		output, _ := cmd.Flags().GetString("output")
		urlOnly, _ := cmd.Flags().GetBool("url-only")

		mediaName := photoName + "/media"
		call := svc.Places.Photos.GetMedia(mediaName).
			SkipHttpRedirect(true).
			MaxWidthPx(int64(maxWidth))

		if maxHeight > 0 {
			call = call.MaxHeightPx(int64(maxHeight))
		}

		media, err := call.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("get photo: %w", err)
		}

		result := PhotoMedia{
			Name:     media.Name,
			PhotoURI: media.PhotoUri,
		}

		if urlOnly || output == "" {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}
			fmt.Println(result.PhotoURI)
			return nil
		}

		// Download the photo
		resp, err := http.Get(result.PhotoURI)
		if err != nil {
			return fmt.Errorf("download photo: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download photo: HTTP %d", resp.StatusCode)
		}

		dir := filepath.Dir(output)
		if dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}
		}

		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()

		n, err := io.Copy(f, resp.Body)
		if err != nil {
			return fmt.Errorf("write photo: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"photoUri": result.PhotoURI,
				"output":   output,
				"bytes":    n,
			})
		}
		fmt.Printf("Saved %d bytes to %s\n", n, output)
		return nil
	}
}
