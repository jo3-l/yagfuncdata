package yagfuncdata

import (
	"context"
	"fmt"

	"github.com/google/go-github/v55/github"
)

// DefaultFileContentProvider is a FileContentProvider that accesses files on
// the master branch of the YAGPDB repository using an unauthenticated GitHub
// client.
var DefaultFileContentProvider = NewGitHubFileProvider(github.NewClient(nil), "botlabs-gg", "yagpdb", "master")

// A FileContentProvider provides access to file content.
type FileContentProvider interface {
	Get(ctx context.Context, path string) (content string, err error)
}

var _ FileContentProvider = (*GitHubFileProvider)(nil)

type GitHubFileProvider struct {
	client              *github.Client
	repoOwner, repoName string
	ref                 string
}

// NewGitHubFileProvider creates a new FileContentProvider that accesses files
// from a GitHub repository at a specific Git reference using the client
// provided.
func NewGitHubFileProvider(client *github.Client, repoOwner, repoName, ref string) *GitHubFileProvider {
	return &GitHubFileProvider{client, repoOwner, repoName, ref}
}

func (g *GitHubFileProvider) Get(ctx context.Context, path string) (string, error) {
	f, _, _, err := g.client.Repositories.GetContents(ctx, g.repoOwner, g.repoName, path,
		&github.RepositoryContentGetOptions{Ref: g.ref})
	if err != nil {
		return "", fmt.Errorf("fetching %s from GitHub: %w", path, err)
	}
	if f == nil {
		return "", fmt.Errorf("%s is a directory, not a file", path)
	}

	content, err := f.GetContent()
	if err != nil {
		return "", fmt.Errorf("could not decode content of %s: %w", path, err)
	}
	return content, nil
}

var _ FileContentProvider = (*StaticFileProvider)(nil)

type StaticFileProvider struct {
	files map[string]string
}

// NewStaticFileProvider creates a new FileContentProvider based on the contents
// of the provided map. For every path such that files[path] == content,
// Get(ctx, path) will return content.
func NewStaticFileProvider(files map[string]string) *StaticFileProvider {
	return &StaticFileProvider{files}
}

func (s *StaticFileProvider) Get(_ context.Context, path string) (string, error) {
	content, ok := s.files[path]
	if !ok {
		return "", fmt.Errorf("could not access %s", path)
	}
	return content, nil
}
