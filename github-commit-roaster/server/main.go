package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: No .env file found")
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/roast", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		ctx := context.Background()
		token := os.Getenv("GITHUB_TOKEN")
		var client *github.Client

		if token != "" {
			ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			tc := oauth2.NewClient(ctx, ts)
			client = github.NewClient(tc)
		} else {
			client = github.NewClient(nil)
			fmt.Println("Warning: Using unauthenticated API - rate limits will apply")
		}

		// Verify user exists
		_, _, err := client.Users.Get(ctx, username)
		if err != nil {
			if _, ok := err.(*github.RateLimitError); ok {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "GitHub API rate limit exceeded",
					"solution": "Please provide a GitHub token in server/.env file",
				})
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub user not found"})
			return
		}

		// Get repositories (limit to 10 most recent)
		repos, _, err := client.Repositories.List(ctx, username, &github.RepositoryListOptions{
			Type:      "owner",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{PerPage: 10},
		})
		if err != nil {
			handleGitHubError(c, err)
			return
		}

		// Get commits from last 30 days
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		var allCommits []*github.RepositoryCommit

		for _, repo := range repos {
			commits, _, err := client.Repositories.ListCommits(ctx, username, *repo.Name, &github.CommitsListOptions{
				Since: thirtyDaysAgo,
			})
			if err != nil {
				continue // Skip repo if we can't get commits
			}
			allCommits = append(allCommits, commits...)
		}

		roast := generateRoast(allCommits)
		
		c.JSON(http.StatusOK, gin.H{
			"username": username,
			"roast":    roast,
			"stats": gin.H{
				"total_commits": len(allCommits),
				"repos_analyzed": len(repos),
			},
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🚀 Server running on port %s\n", port)
	r.Run(":" + port)
}

func handleGitHubError(c *gin.Context, err error) {
	if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		resetTime := rateLimitErr.Rate.Reset.Format(time.RFC1123)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "GitHub API rate limit exceeded",
			"reset_time": resetTime,
			"solution": "Create a .env file with GITHUB_TOKEN in your server directory",
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch GitHub data",
			"details": err.Error(),
		})
	}
}

func generateRoast(commits []*github.RepositoryCommit) string {
	if len(commits) == 0 {
		return "Wow, you haven't committed anything recently. Are you even a developer?"
	}

	// Analysis counters
	lateNightCommits := 0
	swearWords := 0
	mergeCommits := 0
	fixCommits := 0
	genericMessages := 0
	
	for _, commit := range commits {
		msg := strings.ToLower(*commit.Commit.Message)
		commitTime := commit.Commit.Committer.Date
		
		// Check for late night commits (10pm-4am)
		if commitTime.Hour() >= 22 || commitTime.Hour() <= 4 {
			lateNightCommits++
		}
		
		// Check message content
		if containsAny(msg, "fix", "bug", "error") {
			fixCommits++
		}
		if containsAny(msg, "merge", "pull") {
			mergeCommits++
		}
		if containsAny(msg, "fuck", "shit", "damn", "wtf") {
			swearWords++
		}
		if strings.HasPrefix(msg, "update") || strings.HasPrefix(msg, "changes") {
			genericMessages++
		}
	}
	
	// Generate roast lines
	var roastLines []string
	
	if lateNightCommits > len(commits)/2 {
		roastLines = append(roastLines, "Over 50% of your commits are late at night. Do you even sleep?")
	}
	
	if swearWords > 0 {
		roastLines = append(roastLines, fmt.Sprintf("Found %d swear words in commits. Someone needs a stress ball!", swearWords))
	}
	
	if mergeCommits > len(commits)/3 {
		roastLines = append(roastLines, "You merge more than you code. Git plumber much?")
	}
	
	if fixCommits > len(commits)/2 {
		roastLines = append(roastLines, "Most of your commits are fixes. Maybe test before committing?")
	}
	
	if genericMessages > len(commits)/3 {
		roastLines = append(roastLines, "Your commit messages are as generic as a motivational poster.")
	}
	
	if len(roastLines) == 0 {
		roastLines = append(roastLines, "Your commits are suspiciously clean. Are you even trying?")
	}
	
	return strings.Join(roastLines, "\n\n")
}

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}