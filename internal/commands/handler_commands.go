package commands

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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
	fmt.Printf("successfuly unfollowed '%v'", url)
	return nil
}

func HandlerBrowse(s *state.State, cmd Command, user database.User) error {
	var limit int32 = 2 // Default limit
	if len(cmd.Args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
		limit = int32(parsedLimit)
	}

	posts, err := s.DB.GetPostsByUser(context.Background(), database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("error getting posts: %v", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found. Follow some feeds first!")
		return nil
	}

	fmt.Printf("Found %d posts:\n\n", len(posts))
	for i, post := range posts {
		fmt.Printf("=== Post %d ===\n", i+1)
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("URL: %s\n", post.Url)

		if post.Description.Valid {
			// Only show a preview of the description to avoid too much text
			desc := post.Description.String
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			fmt.Printf("Description: %s\n", desc)
		}

		if post.PublishedAt.Valid {
			fmt.Printf("Published: %s\n", post.PublishedAt.Time.Format(time.RFC1123))
		}
		fmt.Printf("Feed: %s\n\n", post.FeedName)
	}

	return nil
}

func HandlerAgg(s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
	}

	timeStr := cmd.Args[0]
	timeBetweenRequests, err := time.ParseDuration(timeStr)
	if err != nil {
		return fmt.Errorf("invalid duration format: %v", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)

	// Run immediately and then on ticker
	ticker := time.NewTicker(timeBetweenRequests)
	// Immediate first run, then wait for ticker
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
		}
	}
}
func scrapeFeeds(s *state.State) error {
	ctx := context.Background()

	// Get the next feed to fetch
	feed, err := s.DB.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("error getting next feed to fetch: %v", err)
	}

	// Mark it as fetched
	err = s.DB.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %v", err)
	}

	fmt.Printf("Fetching feed: %s (%s)\n", feed.Name, feed.Url)

	// Fetch feed using URL
	rssFeed, err := rssfeeds.FetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}

	fmt.Printf("Found %d items in feed %s\n", len(rssFeed.Channel.Item), feed.Name)

	// Save posts to database
	for _, item := range rssFeed.Channel.Item {
		// Parse the published date
		var publishedAt sql.NullTime
		if item.PubDate != "" {
			// Try different time formats
			parsedTime, err := parseTime(item.PubDate)
			if err == nil {
				publishedAt.Time = parsedTime
				publishedAt.Valid = true
			} else {
				fmt.Printf("Warning: Could not parse date '%s': %v\n", item.PubDate, err)
			}
		}

		// Create post in database
		_, err := s.DB.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})

		if err != nil {
			// If it's a duplicate URL, just ignore the error
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			// Otherwise log the error
			fmt.Printf("Error saving post '%s': %v\n", item.Title, err)
		} else {
			fmt.Printf("Saved post: %s\n", item.Title)
		}
	}

	return nil
}

// Helper function to parse different time formats
func parseTime(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		time.RFC822,
		time.RFC822Z,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"02 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", timeStr)
}
