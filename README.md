# Releasr

<p align="center">
  <a href="https://github.com/PackagrIO/docs">
  <img width="300" alt="portfolio_view" src="https://github.com/PackagrIO/releasr/raw/master/images/releasr.png">
  </a>
</p>

Language agnostic tool to package a git repo. Commit any local changes and create a git tag. It should also generate artifacts. Nothing should be pushed to remote repository

# Documentation
Full documentation is available at [PackagrIO/docs](https://github.com/PackagrIO/docs)

# Usage

```
cd /path/to/git/repo
git log
# commit 1b98710322765393ac839b9315619d8b94b945d5 (HEAD -> master, origin/master)

packagr-releasr start --scm github --package_type golang

git log
# commit 2477cbc9df830d79bf08922162ad8594b4cf173b (tag: v1.0.86)
```


# Inputs
- `package_type`
- `scm`
- `version_metadata_path`
- `generic_version_template`
- `engine_git_author_name`
- `engine_git_author_email`
- `engine_repo_config_path`
- `mgr_keep_lock_file`

# Outputs
N/A


# Logo

- [Tag By Travis Avery, US](https://thenounproject.com/search/?q=tag&i=2453778)
