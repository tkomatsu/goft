package cmd

import (
	"fmt"
	"goft/pkg/ftapi"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func NewRepoPathCmd(api *ftapi.APIInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show project repogitory path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var user string
			var err error
			if user, err = cmd.PersistentFlags().GetString("user"); err != nil {
				return err
			}
			for i := 1; ; i++ {
				projects, err := (*api).GetUserProjects(user, nil, nil, i)
				if err != nil {
					color.Set(color.FgRed)
					cmd.PrintErr("GetUserProjects:", err)
					color.Set(color.Reset)
					return err
				}
				if len(projects) == 0 {
					break
				}
				for _, project := range projects {
					color.Set(color.Reset)
					if args[0] != project.Project.Slug {
						continue
					}
					if len(project.Teams) == 0 {
						continue
					}
					team, err := currentTeam(project)
					if err != nil {
						return err
					}
					if team.RepoURL == "" {
						fmt.Fprintf(os.Stderr, "repository not found: %s\n", args[0])
						return nil
					}
					fmt.Println(team.RepoURL)
					return nil
				}
			}
			cmd.Printf("Team of %s is not locked.", args[0])
			return fmt.Errorf("%s is not in your projects", args[0])
		},
	}
}

var repoPathCmd = NewRepoPathCmd(&API)

func init() {
	repoPathCmd.PersistentFlags().StringP("user", "u", os.Getenv("USER"), "Set specific user")
	projectsCmd.AddCommand(repoPathCmd)
}
