# Git Docs

Git Docs is a little tool (web server rest + ui, and a cli) to manage your project's task event offline.
In the spirit of Open Web, this tool serves a handy UI to manage usual doc bases.

It can work completely offline (for instance you can still manage your issues even if you don't have an internet connection)

You get the same user experience whatever is your git service provider (GitLab, GitHub, BitBucket, bare git)

The typical use cases for this tools are :

- ADR (Architecture Design Records) management,
- Issues management,
- Documentation,
- ...

All data are stored in the git repository itself. But the tool can also work _without_ a git repository.

Persistent data are stored as commited files (by default).

Index files will be stored in a git ignored directory. Index is incremental and follows the git commits to keep operations blazing fast.

Stored data files are text based and always editable by hand. Typically those are JSON or markdown files.

## Build

Just run :

    make

This will generate embedded assets and compile the project.

An executable `git-docs` will be generated and installed in the `$GOPATH/bin` directory.

## How to use ?

Go inside a git repository and launch this to start serving the UI :

    git-docs serve

Then in a browser go to `http://127.0.0.1:8080/git-docs/webui/index.html`

## Features

[to be redacted]

### Categories

[to be redacted]

### Documents

[to be redacted]

#### Tags

[to be redacted]

#### Workflows

[to be redacted]

### Git management

[to be redacted]

### Files layout

[to be redacted]

## To do list

- Notifications : when things happen concerning the git user, send email !
- Trigger CI build...
- actions de flow habituel : wip last, new feat, ...
- obtenir un lien vers le document dans le presse-papier pour pouvoir le coller dans les commits. ce lien sera compatible avec le "moteur de recherche" pour indexation
- syntaxe dans les markdowns pour avoir des données avec sémantique (champs boolean, etc) => pour indexation...
- ui par défaut sur branche courante mais sélecteur pour changer de branche
- historique d'une document, grâce à git log...
- multi repositories
- demo how to use user's context data (like secrets, apikeys...) to interact with third party services (Rocket Chat, Emails, CIs...)
- plugins