package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// Tags represents response of the API that gets Docker tags
type Tags struct {
	Count   int `json:"count"`
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

// Status represents status in pulling image
type Status struct {
	Status         string `json:"status"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
}

var faintText = promptui.Styler(promptui.FGFaint)
var boldText = promptui.Styler(promptui.FGBold)
var greenText = promptui.Styler(promptui.FGGreen)
var redText = promptui.Styler(promptui.FGRed)

// NewCmdRoot create root cmd
func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dimg",
		Short: "docker pull image suppoter",
		Run: func(cmd *cobra.Command, args []string) {

			imageName, err := imagePrompt()
			if err != nil {
				cmd.Println(err.Error())
				return
			}

			cmd.Printf("Searching %s tags...\n", boldText(imageName))

			url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/library/%s/tags/?page_size=10000", imageName)

			resp, err := http.Get(url)
			if err != nil {
				cmd.Println("Failed to Get Request: ", err.Error())
				return
			}
			defer resp.Body.Close()

			bytes, _ := ioutil.ReadAll(resp.Body)
			jsonBytes := ([]byte)(bytes)

			tags := new(Tags)
			if err := json.Unmarshal(jsonBytes, tags); err != nil {
				cmd.Println("JSON Unmarshal error: ", err.Error())
				return
			}

			// tag list
			var tagNames []string
			for _, res := range tags.Results {
				tagNames = append(tagNames, res.Name)
			}

			// select tag
			if len(tagNames) > 0 {

				cmd.Printf("%s has %s tags.\n", boldText(imageName), greenText(tags.Count))

				tagName, err := tagSelect(tagNames)
				if err != nil {
					cmd.Println(err.Error())
					return
				}

				image := fmt.Sprintf("%s:%s", imageName, tagName)

				confirmLabel := fmt.Sprintf("Start pulling %s ", image)
				isConfirm, err := confirm(confirmLabel)
				if !isConfirm {
					if err != nil {
						cmd.Println(err.Error())
					}
					return
				}

				if commandExists("docker") {

					pull := exec.Command("docker", "pull", image)
					pull.Stdout = os.Stdout
					pull.Stderr = os.Stderr
					if err := pull.Run(); err != nil {
						cmd.Println(err.Error())
						return
					}

				} else {
					ctx := context.Background()
					cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
					if err != nil {
						cmd.Println(err.Error())
						return
					}

					out, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
					defer out.Close()
					if err != nil {
						cmd.Println(err.Error())
						return
					}

					scanner := bufio.NewScanner(out)
					for scanner.Scan() {
						line := scanner.Bytes()
						status := new(Status)
						if err := json.Unmarshal(line, status); err != nil {
							cmd.Println("JSON Unmarshal error: ", err.Error())
							return
						}
						cmd.Printf("%s: %s %s\n", status.ID, status.Status, status.Progress)
					}
					if scanner.Err() != nil {
						cmd.Println(err.Error())
						return
					}
				}
			} else {
				cmd.Printf("%q not found or no tags.\n", imageName)
			}
		},
	}
	return cmd
}

func imagePrompt() (string, error) {

	validate := func(input string) error {
		if len(input) < 1 {
			return errors.New("Please input Image Name")
		}
		if strings.Contains(input, " ") {
			return errors.New("Imaage Name must not has spaces")
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
		return "", err
	}

	return imageName, nil
}

func tagSelect(tagNames []string) (string, error) {

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
		Size:     15,
	}

	_, tagName, err := selectTag.Run()
	if err != nil {
		return "", err
	}

	return tagName, err
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

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

/*
Execute executes the CLI root command
*/
func Execute() {
	cmd := NewCmdRoot()
	cmd.SetOutput(os.Stdout)
	if err := cmd.Execute(); err != nil {
		cmd.SetOutput(os.Stderr)
		cmd.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
}
