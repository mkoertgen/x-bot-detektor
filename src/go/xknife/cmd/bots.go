package cmd

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/resources"
	"github.com/michimani/gotwi/user/follow"
	ftypes "github.com/michimani/gotwi/user/follow/types"
	"github.com/michimani/gotwi/user/userlookup"
	utypes "github.com/michimani/gotwi/user/userlookup/types"
	"github.com/spf13/cobra"
	"math"
	"time"
)

func init() {
	rootCmd.AddCommand(getUserCmd, getFollowersCmd)
}

var getUserCmd = &cobra.Command{
	Use:   "get",
	Short: "Get user",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := getUser(userName)
		if err != nil {
			fmt.Println("Could not get user", userName, err)
			return err
		}
		printUser(output.Data)
		return nil
	},
}

var getFollowersCmd = &cobra.Command{
	Use:   "followers",
	Short: "Check recent followers on Twitter",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureUserId(); err != nil {
			return err
		}
		fmt.Printf("Getting followers for %s (%s)...\n", userName, userId)
		followers, err := getFollowers(userId, pageSize)
		if err != nil {
			return err
		}
		for _, user := range followers {
			printUser(user)
		}
		return nil
	},
}

func ensureUserId() error {
	if len(userId) == 0 {
		u, err := getUser(userName)
		if err != nil {
			return err
		}
		userId = gotwi.StringValue(u.Data.ID)
	}
	return nil
}

func getUser(name string) (*utypes.GetByUsernameOutput, error) {
	p := &utypes.GetByUsernameInput{
		Username: name,
		//Expansions: fields.ExpansionList{
		//	fields.ExpansionPinnedTweetID,
		//},
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			//fields.UserFieldEntities,
			fields.UserFieldVerified,
			fields.UserFieldProtected,
			fields.UserFieldLocation,
			fields.UserFieldPublicMetrics,
			//fields.UserFieldMostRecentTweetID,
			//fields.UserFieldPinnedTweetID,
			//fields.UserFieldProfileImageUrl,
			//fields.UserFieldDescription,
		},
		//TweetFields: fields.TweetFieldList{
		//	fields.TweetFieldCreatedAt,
		//	fields.TweetFieldPublicMetrics,
		//},
	}

	return userlookup.GetByUsername(context.Background(), xClient, p)
}

func getFollowers(id string, size int) ([]resources.User, error) {
	p := &ftypes.ListFollowersInput{
		ID:              id,
		MaxResults:      ftypes.ListMaxResults(size),
		PaginationToken: "",
		Expansions:      nil,
		TweetFields:     nil,
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
			fields.UserFieldVerified,
			fields.UserFieldProtected,
			fields.UserFieldPublicMetrics,
		},
	}
	output, err := follow.ListFollowers(context.Background(), xClient, p)
	if err != nil {
		return nil, err
	}
	return output.Data, nil
}

func printUser(u resources.User) {
	fmt.Println("ID:          ", gotwi.StringValue(u.ID))
	fmt.Println("Name:        ", gotwi.StringValue(u.Name))
	fmt.Println("Username:    ", gotwi.StringValue(u.Username))
	fmt.Println("CreatedAt:   ", gotwi.TimeValue(u.CreatedAt))
	fmt.Println("Verified:    ", gotwi.BoolValue(u.Verified))
	fmt.Println("Protected:   ", gotwi.BoolValue(u.Protected))
	fmt.Println("Location:    ", gotwi.StringValue(u.Location))
	if u.PublicMetrics != nil {
		m := *u.PublicMetrics
		fmt.Printf("Following: %d, Followers: %d, Tweets: %d, Lists: %d\n",
			gotwi.IntValue(m.FollowingCount), gotwi.IntValue(m.FollowersCount),
			gotwi.IntValue(m.TweetCount), gotwi.IntValue(m.ListedCount),
		)
	}
	fmt.Printf("Score:       %.2f%%\n", score(u))
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
	return math.Min(100, math.Max(0, s))
}
