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

type Tags struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Result    `json:"results"`
}

type Result struct {
	Name                string      `json:"name"`
	FullSize            int64       `json:"full_size"`
	Images              []Image     `json:"images"`
	ID                  int         `json:"id"`
	Repository          int         `json:"repository"`
	Creator             int         `json:"creator"`
	LastUpdater         int         `json:"last_updater"`
	LastUpdaterUsername string      `json:"last_updater_username"`
	ImageID             interface{} `json:"image_id"`
	V2                  bool        `json:"v2"`
	LastUpdated         time.Time   `json:"last_updated"`
}

type Image struct {
	Size         int64       `json:"size"`
	Digest       string      `json:"digest"`
	Architecture string      `json:"architecture"`
	Os           string      `json:"os"`
	OsVersion    string      `json:"os_version"`
	OsFeatures   string      `json:"os_features"`
	Variant      interface{} `json:"variant"`
	Features     string      `json:"features"`
}

var rootCmd = &cobra.Command{
	Use:   "dimg",
	Short: "docker imags command suppoter",

	Run: func(cmd *cobra.Command, args []string) {
		validate := func(input string) error {
			if len(input) < 1 {
				return errors.New("Please input Image Name")
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

		fmt.Printf("Search %q tags\n", imageName)

		url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/library/%s/tags/", imageName)

		// tag list
		var tagNames []string

		for url != "" {
			fmt.Printf(".")

			resp, _ := http.Get(url)
			defer resp.Body.Close()

			bytes, _ := ioutil.ReadAll(resp.Body)
			jsonBytes := ([]byte)(bytes)

			tags := new(Tags)
			if err := json.Unmarshal(jsonBytes, tags); err != nil {
				fmt.Println("JSON Unmarshal error:", err)
				return
			}

			for _, res := range tags.Results {
				tagNames = append(tagNames, res.Name)
			}

			// next page
			url = tags.Next
		}

		// select tag
		if len(tagNames) > 0 {
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
			}

			_, tagName, err := selectTag.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			fmt.Printf("You select %q\n", tagName)

			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				panic(err)
			}

			image := fmt.Sprintf("%s:%s", imageName, tagName)
			fmt.Printf("\x1b[32mDocker pull %s\x1b[0m\n", image)

			out, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			if err != nil {
				panic(err)
			}
			io.Copy(os.Stdout, out)
		} else {
			fmt.Printf("%q has no tags.\n", imageName)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
