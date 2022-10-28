package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	// List repository tags path paramters
	// ref: https://docs.docer.com/docker-hub/api/latest/#tag/repositories
	max_page_size = 100
)

// Tags represents response of the API that gets Docker tags
type Tags struct {
	Count   int    `json:"count"`
	Next    string `json:"next"`
	Results []struct {
		Name string `json:"name"`
	}
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
		Short: "docker pull image supporter",
		Run: func(cmd *cobra.Command, args []string) {

			imageName, err := imagePrompt()
			if err != nil {
				cmd.Println(err.Error())
				return
			}

			tagNames, err := getTags(cmd, imageName)
			if err != nil {
				cmd.Println(err.Error())
				return
			}

			// select tag
			if len(tagNames) > 0 {

				cmd.Printf("%s has %s tags.\n", boldText(imageName), greenText(len(tagNames)))

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

				pull := exec.Command("docker", "pull", image)
				pull.Stdout = os.Stdout
				pull.Stderr = os.Stderr
				if err := pull.Run(); err != nil {
					cmd.Println(err.Error())
					return
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
			return errors.New("Image Name must not has spaces")
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

	if !strings.Contains(imageName, "/") {
		imageName = "library/" + imageName
	}

	return imageName, nil
}

func getTags(cmd *cobra.Command, imageName string) ([]string, error) {
	cmd.Printf("Searching %s tags...\n", boldText(imageName))

	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags?page=1&page_size=%v", imageName, max_page_size)
	var tagNames []string
	for url != "" {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bytes, _ := ioutil.ReadAll(resp.Body)
		jsonBytes := ([]byte)(bytes)

		tags := new(Tags)
		if err := json.Unmarshal(jsonBytes, tags); err != nil {
			return nil, err
		}

		for _, res := range tags.Results {
			tagNames = append(tagNames, res.Name)
		}

		url = tags.Next
		if len(tagNames)%500 < 100 {
			cmd.Printf("loaded %v out of %v tags\n", len(tagNames), tags.Count)
		}
	}

	return tagNames, nil
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

	if !commandExists("docker") {
		cmd.Println("'docker' command not found")
		os.Exit(1)
	}

	cmd.SetOutput(os.Stdout)
	if err := cmd.Execute(); err != nil {
		cmd.SetOutput(os.Stderr)
		cmd.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {}
