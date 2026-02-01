package api

// Build metadata injected at build time via -ldflags.
var (
	BuildGitTag    = ""
	BuildDockerTag = ""
	BuildCommitSHA = ""
)

type BuildInfo struct {
	GitTag    string `json:"gitTag"`
	DockerTag string `json:"dockerTag"`
	CommitSHA string `json:"commitSha"`
}
