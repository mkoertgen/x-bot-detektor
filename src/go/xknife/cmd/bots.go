package cmd

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/user/userlookup"
	"github.com/michimani/gotwi/user/userlookup/types"
	"github.com/spf13/cobra"
)

var recentFollowersCmd = &cobra.Command{
	Use:   "recent-followers",
	Short: "Check recent followers on Twitter",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := gotwi.NewClient(&gotwi.NewClientInput{
			AuthenticationMethod: gotwi.AuthenMethodOAuth2BearerToken,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		p := &types.GetByUsernameInput{
			Username: "michimani210",
			Expansions: fields.ExpansionList{
				fields.ExpansionPinnedTweetID,
			},
			UserFields: fields.UserFieldList{
				fields.UserFieldCreatedAt,
			},
			TweetFields: fields.TweetFieldList{
				fields.TweetFieldCreatedAt,
			},
		}

		u, err := userlookup.GetByUsername(context.Background(), c, p)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("ID:          ", gotwi.StringValue(u.Data.ID))
		fmt.Println("Name:        ", gotwi.StringValue(u.Data.Name))
		fmt.Println("Username:    ", gotwi.StringValue(u.Data.Username))
		fmt.Println("CreatedAt:   ", u.Data.CreatedAt)
		if u.Includes.Tweets != nil {
			for _, t := range u.Includes.Tweets {
				fmt.Println("PinnedTweet: ", gotwi.StringValue(t.Text))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(recentFollowersCmd)
}
