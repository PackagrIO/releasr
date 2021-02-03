package utils

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	git2go "gopkg.in/libgit2/git2go.v25"
	"os"
	"path"
	"path/filepath"
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

//private methods
func GitSignature(authorName string, authorEmail string) *git2go.Signature {
	return &git2go.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}
}
