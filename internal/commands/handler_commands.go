package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/GoFlux/internal/database"
	"github.com/twomotive/GoFlux/internal/rssfeeds"
	"github.com/twomotive/GoFlux/internal/state"
)

// MiddlewareLoggedIn wraps handlers requiring a logged-in user
// It takes a handler that expects a user object and returns a standard handler
func MiddlewareLoggedIn(handler func(s *state.State, cmd Command, user database.User) error) func(*state.State, Command) error {
	return func(s *state.State, cmd Command) error {
		// Check if there's a current user in config
		if s.Cfg.CurrentUsername == "" {
			return fmt.Errorf("no user logged in, please log in first")
		}

		// Get user from database
		currentUser, err := s.DB.GetUser(context.Background(), s.Cfg.CurrentUsername)
		if err != nil {
			return fmt.Errorf("user validation failed: %w", err)
		}

		// Call the original handler with the user
		return handler(s, cmd, currentUser)
	}
}

func HandlerLogin(s *state.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("empty argument")
	}
	name := cmd.Args[0]

	_, err := s.DB.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user with name '%s' doesnt exists", name)
	}

	err = s.Cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func HandlerRegister(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.DB.GetUser(context.Background(), name)
	if err == nil {
		// User exists (no error means the query found a user)
		return fmt.Errorf("user with name '%s' already exists", name)
	}

	newUser, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	})

	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	err = s.Cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set user :%v", err)
	}

	fmt.Printf("User created successfully: %s \n", newUser.Name)
	return nil

}

func HandlerReset(s *state.State, cmd Command) error {
	// The verification is unnecessary, as command name is already "reset"
	// Just check if user explicitly confirms the reset
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: reset")
	}

	err := s.DB.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("cannot reset users table: %w", err)
	}

	// Also clear the current user in configuration
	err = s.Cfg.SetUser("")
	if err != nil {
		return fmt.Errorf("cleared users table but couldn't reset current user: %w", err)
	}

	fmt.Println("All users have been deleted successfully!")
	return nil
}

func HandlerGetUsers(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: users")
	}

	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get users from database: %v", err)
	}

	currentUser := s.Cfg.CurrentUsername

	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}

	return nil

}

func HandlerAgg(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %v", cmd.Name)
	}

	link := "https://www.wagslane.dev/index.xml"

	newFeed, err := rssfeeds.FetchFeed(context.Background(), link)
	if err != nil {
		return fmt.Errorf("feed cannot fetched: %v", err)
	}

	fmt.Println("Channel Title:", newFeed.Channel.Title)
	fmt.Println("Channel Link:", newFeed.Channel.Link)
	fmt.Println("Channel Description:", newFeed.Channel.Description)
	fmt.Println("Number of Items:", len(newFeed.Channel.Item))
	fmt.Println("----------------------------------------")

	for i, item := range newFeed.Channel.Item {
		fmt.Printf("Item #%d\n", i+1)
		fmt.Println("Title:", item.Title)
		fmt.Println("Link:", item.Link)
		fmt.Println("Description:", item.Description)
		fmt.Println("Publication Date:", item.PubDate)
		fmt.Println("----------------------------------------")
	}

	return nil

}

func HandlerAddFeed(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %v <name> <url>", cmd.Name)
	}

	feedName := cmd.Args[0]
	url := cmd.Args[1]

	newFeed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      feedName,
		Url:       url,
		UserID:    user.ID,
	})

	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}

	// Automatically follow the feed after creation
	followResult, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("feed created but failed to follow: %v", err)
	}

	fmt.Printf("Feed '%s' added successfully and followed by %s!\n", newFeed.Name, followResult.UserName)
	return nil
}

func HandlerGetFeeds(s *state.State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: feeds")
	}

	feeds, err := s.DB.GetFeedsWithUserNames(context.Background())
	if err != nil {
		return fmt.Errorf("cannot get feeds from database: %v", err)
	}

	fmt.Println("╔════════════════════════════════════╗")
	fmt.Printf("║ Feed Information                  ║\n")
	fmt.Println("╟────────────────────────────────────╢")

	for _, feed := range feeds {

		fmt.Printf("║ Feed Name: %-22s ║\n", feed.FeedName)
		fmt.Printf("║ Feed URL:  %-22s ║\n", feed.FeedUrl)
		fmt.Printf("║ User Name: %-22s ║\n", feed.UserName)

	}
	fmt.Println("╚════════════════════════════════════╝")

	return nil
}

func HandlerFollow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.Name)
	}

	url := cmd.Args[0]

	// No need to query for the current user - it's passed in by middleware
	feedByUrl, err := s.DB.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("cannot get feed from database: %v", err)
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feedByUrl.ID,
	})
	if err != nil {
		return fmt.Errorf("cannot create follow row: %v", err)
	}

	fmt.Printf("Feed: %v\n", feedByUrl.Name)
	fmt.Printf("Following by : %v\n", user.Name)

	return nil
}

func HandlerFollowing(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %v", cmd.Name)
	}

	// No need to query for the user - it's passed in by middleware
	userFeeds, err := s.DB.GetFeedFollowsByUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("cannot get users feeds from database: %v", err)
	}

	fmt.Printf("Feeds followed by the %v\n", user.Name)
	for _, feed := range userFeeds {
		fmt.Printf(" * %v\n", feed.FeedName)
	}

	return nil
}

func HandlerUnfollow(s *state.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.Name)
	}

	url := cmd.Args[0]

	err := s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow: %v", err)
	}
	return nil
}
