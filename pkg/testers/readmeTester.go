package testers

import (
	"context"
	"net/http"

	"github.com/catena-x/gh-org-checks/pkg/data"
	"github.com/google/go-github/v45/github"
	log "github.com/sirupsen/logrus"
)

type ReadmeTester struct {
	testType     string
	ctx          context.Context
	owner        string
	githubClient *github.Client
}

func NewReadmeTester(ctx context.Context, owner string, githubClient *github.Client) GithubTester {
	log.Printf("creating new readme tester")
	return ReadmeTester{
		testType:     "Readme",
		ctx:          ctx,
		owner:        owner,
		githubClient: githubClient,
	}
}

func (checker ReadmeTester) PerformTest(repoName string) data.RepositoryReport {
	log.Infof("perform readme test on repo %s", repoName)
	_, resp, err := checker.githubClient.Repositories.GetReadme(checker.ctx, checker.owner, repoName, &github.RepositoryContentGetOptions{})

	if resp.StatusCode != http.StatusOK {
		log.Infof("readme test failed on repo %s", repoName)
		log.Infof("status code: %d", resp.StatusCode)
		return data.RepositoryReport{
			TestName:    checker.testType,
			GithubRepo:  repoName,
			TestSucceed: false,
			Log:         []string{"Readme file is missing!"},
		}
	}

	if err != nil {
		log.Infof("readme test failed on repo %s", repoName)
		log.Debugf("error message: %s", err.Error())
		return data.RepositoryReport{
			TestName:    checker.testType,
			GithubRepo:  repoName,
			TestSucceed: false,
			Log:         []string{err.Error()},
		}
	}

	log.Infof("readme test is successful on repo %s", repoName)
	return data.RepositoryReport{
		TestName:    checker.testType,
		GithubRepo:  repoName,
		TestSucceed: true,
	}

}
