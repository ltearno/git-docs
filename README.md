# Magic Git

Make git tools working offline :
- issue management,
- ...

Magic Git is a little tool (web server rest + ui, and a cli) to manage your project's task event offline.

All data are stored in the git repository itself.

Persistent data are stored as commited files (by default).

Index files are stored in a git ignored directory. Index is incremental and follows the git commits to make operations blazing fast.

## How to use ?

Go inside a git repository and launch this to start serving the UI :

    mgit serve

Then in a browser go to `http://127.0.0.1:8080/webui/index.html`

## TODO

Plugins:
- Notifications : when things happen concerning the git user, send email !
- Trigger CI build...
- actions de flow habituel : wip last, new feat, ...

obtenir un lien vers l'issue dans le presse-papier pour pouvoir le coller dans les commits. ce lien sera compatible avec le "moteur de recherche" pour indexation

syntaxe dans les markdowns pour avoir des données avec sémantique (champs boolean, etc) => pour indexation...

étendre à plus d'utilisations => adr, documents de fonctionnements d'une boite, etc

TODO : ui par défaut sur branche courante mais sélecteur pour changer de branche

TODO : historique d'une issue, grâce à git log...