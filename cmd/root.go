package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// Tags represents the Response of the API that gets Docker tags
type Tags struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []struct {
		Name     string `json:"name"`
		FullSize int    `json:"full_size"`
		Images   []struct {
			Size         int         `json:"size"`
			Digest       string      `json:"digest"`
			Architecture string      `json:"architecture"`
			Os           string      `json:"os"`
			OsVersion    interface{} `json:"os_version"`
			OsFeatures   string      `json:"os_features"`
			Variant      interface{} `json:"variant"`
			Features     string      `json:"features"`
		} `json:"images"`
		ID                  int         `json:"id"`
		Repository          int         `json:"repository"`
		Creator             int         `json:"creator"`
		LastUpdater         int         `json:"last_updater"`
		LastUpdaterUsername string      `json:"last_updater_username"`
		ImageID             interface{} `json:"image_id"`
		V2                  bool        `json:"v2"`
		LastUpdated         time.Time   `json:"last_updated"`
	} `json:"results"`
}

var faintText = promptui.Styler(promptui.FGFaint)
var boldText = promptui.Styler(promptui.FGBold)
var greenText = promptui.Styler(promptui.FGGreen)
var redText = promptui.Styler(promptui.FGRed)

var rootCmd = &cobra.Command{
	Use:   "dimg",
	Short: "docker pull image suppoter",

	Run: func(cmd *cobra.Command, args []string) {

		validate := func(input string) error {
			if len(input) < 1 {
				return errors.New("Please input Image Name")
			}
			if strings.Contains(input, " ") {
				return errors.New("Input must not has spaces")
			}
			return nil
		}

		// input image name
		typeImage := promptui.Prompt{
			Label:    "Image Name",
			Validate: validate,
		}

		imageName, err := typeImage.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		fmt.Printf("Searching %s tags...\n", boldText(imageName))

		url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/library/%s/tags/?page_size=10000", imageName)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Failed to Get Request:", err)
			return
		}
		defer resp.Body.Close()

		bytes, _ := ioutil.ReadAll(resp.Body)
		jsonBytes := ([]byte)(bytes)

		tags := new(Tags)
		if err := json.Unmarshal(jsonBytes, tags); err != nil {
			fmt.Println("JSON Unmarshal error:", err)
			return
		}

		// tag list
		var tagNames []string
		for _, res := range tags.Results {
			tagNames = append(tagNames, res.Name)
		}

		// select tag
		if len(tagNames) > 0 {

			fmt.Printf("%s has %s tags.\n", boldText(imageName), greenText(tags.Count))

			searcher := func(input string, index int) bool {
				tagNames := tagNames[index]
				name := strings.Replace(strings.ToLower(tagNames), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)

				return strings.Contains(name, input)
			}

			selectTag := promptui.Select{
				Label:    "Select Tag",
				Items:    tagNames,
				Searcher: searcher,
				Size:     10,
			}

			_, tagName, err := selectTag.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			image := fmt.Sprintf("%s:%s", imageName, tagName)

			confirmLabel := fmt.Sprintf("Start pulling %s", image)
			isConfirm, err := confirm(confirmLabel)
			if !isConfirm {
				if err != nil {
					fmt.Printf("%v\n", err)
				}
				return
			}

			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				fmt.Printf("CreateClient failed %v\n", err)
				return
			}

			out, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			if err != nil {
				fmt.Printf("ImagePull failed %v\n", err)
				return
			}
			io.Copy(os.Stdout, out)
		} else {
			fmt.Printf("%q has no tags.\n", imageName)
		}
	},
}

func confirm(label string) (bool, error) {

	prompt := promptui.Prompt{
		Label:     label,
		Default:   "y",
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		return false, err
	}

	return true, nil
}

/*
Execute executes the CLI root command
*/
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
