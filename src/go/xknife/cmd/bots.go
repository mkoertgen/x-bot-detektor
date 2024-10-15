package cmd

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/resources"
	"github.com/michimani/gotwi/user/userlookup"
	"github.com/michimani/gotwi/user/userlookup/types"
	"github.com/spf13/cobra"
	"math"
	"time"
)

func init() {
	rootCmd.AddCommand(getUserCmd, recentFollowersCmd)
}

var getUserCmd = &cobra.Command{
	Use:   "get",
	Short: "Get user",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := cmd.Flag("username").Value.String()
		output, err := getUser(name)
		if err != nil {
			fmt.Println("Could not get user", name, err)
			return err
		}

		user := output.Data
		fmt.Println("ID:          ", gotwi.StringValue(user.ID))
		fmt.Println("Name:        ", gotwi.StringValue(user.Name))
		fmt.Println("Username:    ", gotwi.StringValue(user.Username))
		fmt.Println("CreatedAt:   ", user.CreatedAt)
		fmt.Printf("Score:       %.2f%%\n", score(user))
		if output.Includes.Tweets != nil {
			for _, t := range output.Includes.Tweets {
				fmt.Println("PinnedTweet: ", gotwi.StringValue(t.Text))
			}
		}
		return nil
	},
}

var recentFollowersCmd = &cobra.Command{
	Use:   "recent-followers",
	Short: "Check recent followers on Twitter",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func getUser(userName string) (*types.GetByUsernameOutput, error) {
	p := &types.GetByUsernameInput{
		Username: userName,
		Expansions: fields.ExpansionList{
			fields.ExpansionPinnedTweetID,
		},
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			//fields.UserFieldEntities,
			fields.UserFieldVerified,
			fields.UserFieldProtected,
			fields.UserFieldLocation,
			fields.UserFieldPublicMetrics,
			fields.UserFieldMostRecentTweetID,
			fields.UserFieldPinnedTweetID,
			fields.UserFieldProfileImageUrl,
			fields.UserFieldDescription,
		},
		TweetFields: fields.TweetFieldList{
			fields.TweetFieldCreatedAt,
		},
	}

	return userlookup.GetByUsername(context.Background(), xClient, p)
}

func score(u resources.User) float64 {
	// https://developer.x.com/en/docs/x-api/data-dictionary/object-model/user
	s := 100.0
	// 1. verified user
	if gotwi.BoolValue(u.Verified) {
		return s
	}
	// - private user (Protected)
	if gotwi.BoolValue(u.Protected) {
		return s
	}
	// 2. follower/following ratio
	// x: 10, 5   , 1  , 0.5 , 0.1, 0.01
	// y:  1, 0.95, 0.9, 0.85, 0.3,    0
	following := float64(gotwi.IntValue(u.PublicMetrics.FollowingCount))
	followers := float64(gotwi.IntValue(u.PublicMetrics.FollowersCount))
	x := followers / math.Max(following, 1.0)
	y := 2 * math.Atan(5*x) / 3 // <-- https://www.geogebra.org/graphing
	s *= y
	// - account duration compared to absolute following (bots are quick to follow many accounts, following more than 20/day avg is usually suspicious)
	// Cutoff date: March 2006, cf.: https://www.oldest.org/technology/oldest-twitter-accounts/
	cutOffDate := time.Date(2006, time.March, 21, 0, 0, 0, 0, time.UTC)
	days := gotwi.TimeValue(u.CreatedAt).Sub(cutOffDate).Hours() / 24.0
	x = math.Min(following/days, 40)
	// x: 1, 5   , 10  , 20 , 30 , 40
	// y: 1, 0.95, 0.85, 0.5, 0.1,  0
	y = (1 + math.Cos(x/(4*math.Pi))) / 2
	s *= y
	// 3. common followers
	// 4. avg number of hashtags used in tweets
	// ...
	return s
}
