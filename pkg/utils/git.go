package utils

import (
	stderrors "errors"
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	git2go "gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Clone a git repo into a local directory.
// Credentials need to be specified by embedding in gitRemote url.
// TODO: this pattern may not work on Bitbucket/GitLab
func GitClone(parentPath string, repositoryName string, gitRemote string) (string, error) {
	absPath, _ := filepath.Abs(path.Join(parentPath, repositoryName))

	if !utils.FileExists(absPath) {
		os.MkdirAll(absPath, os.ModePerm)
	} else {
		return "", errors.ScmFilesystemError(fmt.Sprintf("The local repository path already exists, this should never happen. %s", absPath))
	}

	_, err := git2go.Clone(gitRemote, absPath, new(git2go.CloneOptions))
	return absPath, err
}

//Add all modified files to index, and commit.
func GitCommit(repoPath string, message string, signature *git2go.Signature) error {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	//get repo index.
	idx, ierr := repo.Index()
	if ierr != nil {
		return ierr
	}
	aerr := idx.AddAll([]string{}, git2go.IndexAddDefault, nil)
	if aerr != nil {
		return aerr
	}
	treeId, wterr := idx.WriteTree()
	if wterr != nil {
		return wterr
	}
	werr := idx.Write()
	if werr != nil {
		return werr
	}

	tree, lerr := repo.LookupTree(treeId)
	if lerr != nil {
		return lerr
	}

	currentBranch, berr := repo.Head()
	if berr != nil {
		return berr
	}

	commitTarget, terr := repo.LookupCommit(currentBranch.Target())
	if terr != nil {
		return terr
	}

	_, cerr := repo.CreateCommit("HEAD", signature, signature, message, tree, commitTarget)
	//if(cerr != nil){return cerr}

	return cerr
}

func GitTag(repoPath string, version string, message string, signature *git2go.Signature) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}
	commitHead, herr := repo.Head()
	if herr != nil {
		return "", herr
	}

	commit, lerr := repo.LookupCommit(commitHead.Target())
	if lerr != nil {
		return "", lerr
	}

	//tagId, terr := repo.Tags.CreateLightweight(version, commit, false)
	tagId, terr := repo.Tags.Create(version, commit, signature, fmt.Sprintf("(%s) %s", version, message))
	if terr != nil {
		return "", terr
	}

	tagObj, terr := repo.LookupTag(tagId)
	if terr != nil {
		return "", terr
	}
	return tagObj.TargetId().String(), terr
}

// Get the nearest tag on branch.
// tag must be nearest, ie. sorted by their distance from the HEAD of the branch, not the date or tagname.
// basically `git describe --tags --abbrev=0`
func GitFindNearestTagName(repoPath string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	descOptions, derr := git2go.DefaultDescribeOptions()
	if derr != nil {
		return "", derr
	}
	descOptions.Strategy = git2go.DescribeTags

	formatOptions, ferr := git2go.DefaultDescribeFormatOptions()
	if ferr != nil {
		return "", ferr
	}
	formatOptions.AbbreviatedSize = 0

	descr, derr := repo.DescribeWorkdir(&descOptions)
	if derr != nil {
		return "", derr
	}

	nearestTag, ferr := descr.Format(&formatOptions)
	if ferr != nil {
		return "", ferr
	}

	return nearestTag, nil
}

func GitGenerateChangelog(repoPath string, baseSha string, headSha string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	markdown := utils.StripIndent(`Timestamp |  SHA | Message | Author
	------------- | ------------- | ------------- | -------------
	`)

	revWalk, werr := repo.Walk()
	if werr != nil {
		return "", werr
	}

	rerr := revWalk.PushRange(fmt.Sprintf("%s..%s", baseSha, headSha))
	if rerr != nil {
		return "", rerr
	}

	revWalk.Iterate(func(commit *git2go.Commit) bool {
		markdown += fmt.Sprintf("%s | %.8s | %s | %s\n", //TODO: this should have a link for the SHA.
			commit.Author().When.UTC().Format("2006-01-02T15:04Z"),
			commit.Id().String(),
			cleanCommitMessage(commit.Message()),
			commit.Author().Name,
		)
		return true
	})
	//for {
	//	err := revWalk.Next()
	//	if err != nil {
	//		break
	//	}
	//
	//	log.Info(gi.String())
	//}

	return markdown, nil
}

func GitGenerateGitIgnore(repoPath string, ignoreType string) error {
	//https://github.com/GlenDC/go-gitignore/blob/master/gitignore/provider/github.go

	gitIgnoreBytes, err := getGitIgnore(ignoreType)
	if err != nil {
		return err
	}

	gitIgnorePath := filepath.Join(repoPath, ".gitignore")
	return ioutil.WriteFile(gitIgnorePath, gitIgnoreBytes, 0644)
}

func GitGetTagDetails(repoPath string, tagName string) (*pipeline.GitTagDetails, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return nil, oerr
	}

	id, aerr := repo.References.Dwim(tagName)
	if aerr != nil {
		return nil, aerr
	}
	tag, lerr := repo.LookupTag(id.Target()) //assume its an annotated tag.

	var currentTag *pipeline.GitTagDetails
	if lerr != nil {
		//this is a lightweight tag, not an annotated tag.
		commitRef, rerr := repo.LookupCommit(id.Target())
		if rerr != nil {
			return nil, rerr
		}

		author := commitRef.Author()

		log.Printf("Light-weight tag (%s) Commit ID: %s, DATE: %s", tagName, commitRef.Id().String(), author.When.String())

		currentTag = &pipeline.GitTagDetails{
			TagShortName: tagName,
			CommitSha:    commitRef.Id().String(),
			CommitDate:   author.When,
		}

	} else {

		log.Printf("Annotated tag (%s) Tag ID: %s, Commit ID: %s, DATE: %s", tagName, tag.Id().String(), tag.TargetId().String(), tag.Tagger().When.String())

		currentTag = &pipeline.GitTagDetails{
			TagShortName: tagName,
			CommitSha:    tag.TargetId().String(),
			CommitDate:   tag.Tagger().When,
		}
	}
	return currentTag, nil

}

//private methods

func GitSignature(authorName string, authorEmail string) *git2go.Signature {
	return &git2go.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}
}

func cleanCommitMessage(commitMessage string) string {
	commitMessage = strings.TrimSpace(commitMessage)
	if commitMessage == "" {
		return "--"
	}

	commitMessage = strings.Replace(commitMessage, "|", "/", -1)
	commitMessage = strings.Replace(commitMessage, "\n", " ", -1)

	return commitMessage
}

func getGitIgnore(languageName string) ([]byte, error) {
	gitURL := fmt.Sprintf("https://raw.githubusercontent.com/github/gitignore/master/%s.gitignore", languageName)

	resp, err := http.Get(gitURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, stderrors.New(fmt.Sprintf("Could not find .gitignore for '%s'", languageName))
	}

	return ioutil.ReadAll(resp.Body)
}

//func gitRemoteCallbacks() *git.RemoteCallbacks {
//	return  &git.RemoteCallbacks{
//		CredentialsCallback: credentialsCallback,
//		CertificateCheckCallback: certificateCheckCallback,
//	}
//}
//
//func credentialsCallback(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
//	log.Printf("This is the CRED URL FOR PUSH: %s %s",url, username_from_url)
//	ret, cred := git.NewCredUserpassPlaintext("placeholder", "") //TODO: remote cred.
//
//	log.Printf("THIS IS THE CRED RESPONS: %s %s", ret, cred)
//	return git.ErrorCode(ret), &cred
//}
//
//func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
//	if hostname != "github.com" {
//		return git.ErrUser
//	}
//	return git.ErrOk
//}
