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

## Install

Precompiled releases for OS X, Windows and Linux are available as a `tar.gz` file. Here are the instructions to install from the precompiled binaries.

### On Linux

Run that :

    # fetch the release archive
    wget https://github.com/ltearno/git-docs/releases/download/v0.1/git-docs-releases.tag.gz
    # extract the archive
    tar xvzf git-docs-releases.tag.gz
    # add the binary to the PATH
    export PATH=$PATH:$(pwd)/git-docs-releases/linux-amd64

You can then run it :

    git-docs

### On other platform

TO BE DONE (MacOS and Windows releases exists but are not yet documented)

## How to use ?

Go inside a git repository and launch this to start serving the UI :

    git-docs serve

Then in a browser go to `http://127.0.0.1:8080/git-docs/webui/index.html`

## Concepts

Documents that have metadata and content are contained in categories.

### Categories

A _category_ is a referential for documents together with configuration data, regarding :

- _document templates_,
- _tags_,
- _boards_,
- _workflows_.

Those elements are described below.

### Documents

This is the main concept. A _document_ is made of :

- _metadata_ : any data, JSON stored. For instance the list of _tags_ associated with a _document_ is contained in its _metadata_
- _content_ : a Markdown formatted textual content. The _content_ can use the _Golang_ templating syntax to inject dynamic values in the _content_.

A _document_ is identified by its _name_.

Documents are stored in their own directories. Two files are used for its _metadata_ (`metadata.json`) and its _content_ (`content.md`).

#### Metadata structure

Although the metadata has no fixed predefined structure, the tool stores and reads _tags_ in the `tags` field. Other fields can be used by plugins and RESP API clients...

#### Document context for interpolation

Content can be interpolated according to the Golang templating utility.

Here are the data available to the content :

[to be redacted]
- document name,
- document metadata,
- category,
-...

#### Tags

Tags are simply texts (not containing spaces) that are associated with documents.

A lot of features use the tagging system.

#### Search language

The search language is used at different places : document _search_, _boards_, _worflows_, ...
So it is described in its own section.

The query language generally matches against document _tags_. It is used to extract documents according to which tag it contains.

Here is the syntax of the language, depending on the given _expression_ :

- `term` : matches if the _document_ has at least one _tag_ **containing** the _expression_ (case-insensitive)
- `TERM` : matches if the _document_ has at least one _tag_ **equal** to the _expression_ (case-insensitive)
- `!` _sub-expression_ : matches if the _document_ **does not match** the _sub-expression_
- `&` _sub-expression-1_ sub-expression-2_ : matches if the document **both** matches _sub-expression-1_ **and** _sub-expression-2_
- `|` _sub-expression-1_ sub-expression-2_ : matches if the document **either** matches _sub-expression-1_ **or** _sub-expression-2_

Examples :

`todo` : documents with a tag containing "todo" (case-insensitive)
`DEV` : documents with a tag equal to "dev" (case-insensitive)
`!todo` : documents with no tag containing "todo" (case-insensitive)
`!DEV` : documents with no tag equal to "dev" (case-insensitive)
`& toto !DOC` : documents with a tag containing "todo" and no tag equal to "doc" (case-insensitive)
`` : the empty string matched all documents

#### Boards

Boards allow to visualize and spread documents according to criterias.

Boards are specified as follow :
- first documents are matched according to a _search_ query. This defines which documents will be displayed in the board.
- then a _comma separated_ list of search queries define the different columns of the board.

For instance, the board with `feature` as the search query and `todo,doing,done,validated,` as the columns definition will display all the documents with the "feature" tag and split them according those which contain "todo", "doing", "done", "validated" and the last empty column matches all documents. The empty search can be positionned on any column.

#### Workflows

Workflows are actions that can be automatically applied when some event happens.

Events can be the addition or removal of a _tag_. The corresponding _action_ is triggered when the _condition_ is satisfied. Multiple intent paths are possible, the path is choosen according to the action _intent_. The _intent_ is communicated to the tool when changing a document metadata.

For instance here is the description of a workflow :

```json
{
    // self-describing
    "when-removed-todo": [
        // first intent path, this is the default one (if no intent is specified)
        {
            // human description
            "description": "normal operations",
            // this workflow will only trigger when this condition applies for the document 
            // on which the "todo" tag is removed (before removal)
            // here, the actions are done if the document contains a tag "todo" or "done"
            "condition": "| todo done",
            // those tags will be added to the document
            "addTags": [
                "done",
                "to-inspect"
            ],
            // those tags will be removed from the document
            "removeTags": [
                "todo",
                "a faire",
                "to-be-redacted"
            ]
        },
        // second intent path
        {
            // the name is matched against the intent path
            "name": "garbage",
            "description": "you remove the 'todo' tag because the issue is now obsolete",
            "addTags": [
                "to-delete",
                "to-terminate"
            ]
        }
    ],
    "when-added-doc": [
        {
            "addTags": [
                "to-be-redacted",
                "hello",
                "goodbye"
            ]
        }
    ]
}
```

### Git management

[to be redacted]

### Files layout

[to be redacted]

### REST API

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
- tag colors