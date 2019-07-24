const el = document.getElementById.bind(document)

const elFromHtml = html => {
    const e = document.createElement('div')
    e.innerHTML = html
    return e.children.item(0)
}

const memoize = f => {
    const cache = new Map()
    return arg => {
        if (cache.has(arg))
            return cache.get(arg)
        let value = f(arg)
        cache.set(arg, value)
        return value
    }
}




const afterFetch = memoize((responseContentType = 'application/json') => {
    return (response, error) => {
        if (error)
            return null

        if (!response.ok) {
            log(`bad response : ${JSON.stringify(response)}`)
            return null
        }

        if (responseContentType == 'application/json')
            return response.json()
        else
            return response.text()
    }
})

function getData(url, responseContentType = 'application/json') {
    return fetch(
        url, {
            method: 'GET',
            mode: 'cors',
            cache: 'no-cache',
            credentials: 'same-origin',
            redirect: 'follow',
            referrer: 'no-referrer'
        })
        .then(afterFetch(responseContentType))
}

function postData(url = '', data = {}, contentType = 'application/json', responseContentType = 'application/json') {
    return fetch(
        url, {
            method: 'POST',
            mode: 'cors',
            cache: 'no-cache',
            credentials: 'same-origin',
            redirect: 'follow',
            referrer: 'no-referrer',
            headers: { "Content-Type": contentType },
            body: contentType == 'application/json' ? JSON.stringify(data) : data
        })
        .then(afterFetch(responseContentType))
}

function putData(url = '', data = {}, contentType = 'application/json', responseContentType = 'application/json') {
    return fetch(
        url, {
            method: 'PUT',
            mode: 'cors',
            cache: 'no-cache',
            credentials: 'same-origin',
            redirect: 'follow',
            referrer: 'no-referrer',
            headers: { "Content-Type": contentType },
            body: contentType == 'application/json' ? JSON.stringify(data) : data
        })
        .then(afterFetch(responseContentType))
}

function deleteData(url = '', data = {}, contentType = 'application/json', responseContentType = 'application/json') {
    return fetch(
        url, {
            method: 'DELETE',
            mode: 'cors',
            cache: 'no-cache',
            credentials: 'same-origin',
            redirect: 'follow',
            referrer: 'no-referrer',
            headers: { "Content-Type": contentType },
            body: contentType == 'application/json' ? JSON.stringify(data) : data
        })
        .then(afterFetch(responseContentType))
}




const badgeColorClass = tag => {
    let c = 0
    for (let i = 0; i < tag.length; i++) {
        c += tag.charCodeAt(i) * 5
    }
    return `badge-color-${c % 9}`
}

const tagToHtmlBadge = memoize(tag => `<div class="badge ${badgeColorClass(tag)}">${tag}</div>`)



let logMessages = []
const log = msg => {
    logMessages.push(msg)
    if (logMessages.length > 10)
        logMessages = logMessages.slice(-10)
    el('log').innerHTML = logMessages.map(msg => `<div>${msg}</div>`).join('')
}




let appState = {
    category: null,
    document: null,
    modeEditDocument: false,

    search: "",
    split: ""
}

function appStateSetCategory(category, dbChanged = false) {
    if (!dbChanged && category == appState.category)
        return

    localStorage.setItem('selected-category', category)

    appState.category = category
    appState.document = null
    appState.search = localStorage.getItem('search-document-' + encodeURIComponent(appState.category)) || ''
    appState.split = localStorage.getItem('columns-document-' + encodeURIComponent(appState.category)) || ''

    appStateAfterChange()
}

function appStateSetBoardSearch(search, split) {
    search = search || ""
    split = split || ""

    if (search == appState.search && split == appState.split)
        return

    appState.search = search
    appState.split = split

    localStorage.setItem('search-document-' + encodeURIComponent(appState.category), appState.search)
    localStorage.setItem('columns-document-' + encodeURIComponent(appState.category), appState.split)

    loadDocuments(appState.category, appState.search, appState.split)
}

function appStateSetDocument(document, modeEditDocument, dbChanged = false) {
    if (!dbChanged && document == appState.document && modeEditDocument == appState.modeEditDocument)
        return

    appState.document = document
    appState.modeEditDocument = modeEditDocument

    if (dbChanged)
        appStateAfterChange()
    else
        drawDocumentDetail()
}

function appStateAfterChange() {
    loadStatus()
    loadCategories(appState.category)
    loadTags(appState.category)
    loadDocuments(appState.category, appState.search, appState.split)
    drawDocumentDetail()
}

function drawDocumentDetail() {
    if (appState.modeEditDocument)
        drawDocumentEdition(appState.category, appState.document)
    else
        drawDocument(appState.category, appState.document)
}








function fetchCategories() {
    return getData(`/git-docs/api/categories`)
}

function deleteDocument(category, name) {
    deleteData(`/git-docs/api/documents/${category}/${name}`)
        .then(_ => {
            log(`deleted document ${name}`)
            appStateSetDocument(null, false, true)
        })
}

function addDocument(category, name) {
    postData(`/git-docs/api/documents/${category}/${name}`, {})
        .then(_ => {
            log(`add document ${name}`)
            appStateSetDocument(name, false, true)
        })
}

function addTagToDocument(category, name, tagToAdd) {
    return getData(`/git-docs/api/documents/${category}/${name}/metadata`)
        .then(metadata => {
            let update = false

            if (!metadata) {
                metadata = {}
                update = true
            }

            if (!metadata.tags) {
                metadata.tags = []
                update = true
            }

            if (!metadata.tags.includes(tagToAdd)) {
                metadata.tags.push(tagToAdd)
                update = true
            }

            if (update) {
                return chooseWorkflowAction(category, false, tagToAdd)
                    .then(actionName => {
                        putData(`/git-docs/api/documents/${category}/${name}/metadata?action_name=${encodeURIComponent(actionName || '')}`, metadata)
                            .then(_ => {
                                log(`update document metadata ${name}`)
                                appStateSetDocument(name, false, true)
                            })
                    })
                    .catch(() => {
                        log(`cancelled tag add`)
                    })
            }
            else {
                log(`tag already present`)
            }
        })
}

function deleteTagToDocument(category, name, tagToRemove) {
    getData(`/git-docs/api/documents/${category}/${name}/metadata`)
        .then(metadata => {
            if (!metadata || !metadata.tags || !metadata.tags.includes(tagToRemove))
                return

            return chooseWorkflowAction(category, true, tagToRemove)
                .then(actionName => {
                    if (actionName)
                        log(`choosen action : ${actionName}`)

                    metadata.tags = metadata.tags.filter(tag => tag != tagToRemove)

                    putData(`/git-docs/api/documents/${category}/${name}/metadata?action_name=${encodeURIComponent(actionName || '')}`, metadata)
                        .then(_ => {
                            log(`update document metadata ${name}`)
                            appStateSetDocument(name, false, true)
                        })
                })
                .catch(() => {
                    log(`cancelled tag remove`)
                })
        })
}

// si action (add/rem tag) a un workflow, et que ce workflow a plusieurs alternatives, on fait choisir une de ces alternatives

function getWorkflowPossibleActions(category, removal, tag) {
    return getData(`/git-docs/api/workflows/${category}`)
        .then(workflow => {
            if (!workflow)
                return []

            let key = `when-${removal ? 'removed' : 'added'}-${tag}`
            let elements = workflow[key]
            if (!elements || !elements.length)
                return []

            let actionNames = elements.map(element => ({ name: element.name || null, description: element.description || null }))
            return actionNames
        })
}

function chooseWorkflowAction(category, removal, tag) {
    return getWorkflowPossibleActions(category, removal, tag)
        .then(possibleActions => {
            if (!possibleActions || !possibleActions.length)
                return null
            if (possibleActions.length == 1)
                return possibleActions[0].name

            let resolver = null
            let rejecter = null
            let promise = new Promise((resolve, reject) => {
                resolver = resolve
                rejecter = reject
            })

            let choiceUi = elFromHtml(`
                <div class='mui-panel'>
                    <h2>Please choose an action for ${removal ? 'removing' : 'adding'} the tag ${tag}</h2>
                    <div class="mui--text-caption mui--text-dark-secondary">When tags are added or removed, a workflow can occur. This workflow may lead to different actions depending on your action's intent.</div>
                    <div class="mui-divider"></div>
                    <ul style='cursor:pointer;'>${possibleActions.map((action, i) => `<li x-id='action-${i}'>${action.name || 'DEFAULT'} ${action.description ? `<i>(${action.description})</i>` : ''}</li>`).join('')}</ul>
                    <button x-id='cancel' class="mui-btn mui-btn--flat">Cancel</button>
                </div>
            `)
            choiceUi.style.width = '400px';
            choiceUi.style.margin = '100px auto';
            choiceUi.style.backgroundColor = '#fff';

            for (let i = 0; i < possibleActions.length; i++) {
                choiceUi.querySelector(`[x-id=action-${i}]`).addEventListener('click', () => {
                    mui.overlay('off')
                    resolver(possibleActions[i].name)
                })
            }
            choiceUi.querySelector(`[x-id=cancel]`).addEventListener('click', () => {
                mui.overlay('off')
                rejecter(null)
            })

            mui.overlay('on', choiceUi)

            return promise
        })
}





function drawDocumentEdition(category, name) {
    el('board-opened-documents').innerHTML = ''

    if (!name)
        return

    const documentElement = document.createElement('div')
    documentElement.classList.add('mui-panel')
    documentElement.innerHTML += `<input id='name-input' type='text' style='font-size:2em;'/></input>`
    const contentElement = document.createElement('div')
    contentElement.innerHTML += `<h2>Content</h2>`
    documentElement.appendChild(contentElement)
    documentElement.appendChild(elFromHtml(`<button onclick='appStateSetDocument("${name}", false, false)' class="mui-btn mui-btn--flat">Cancel</button>`))
    documentElement.appendChild(elFromHtml(`<button onclick='deleteDocument("${category}","${name}")' class="delete mui-btn mui-btn--flat mui-btn--danger">Delete</button>`))
    documentElement.appendChild(elFromHtml(`<button class="validate-edit mui-btn mui-btn--primary mui-btn--raised">Validate</button>`))

    el('board-opened-documents').appendChild(documentElement)

    documentElement.querySelector('#name-input').value = name

    getData(`/git-docs/api/documents/${category}/${name}/content`, 'application/mardown')
        .then(content => contentElement.innerHTML += `<textarea class='document-content-textarea' style='width:80em;height:30em;'>${content}</textarea>`)

    let validateButton = documentElement.getElementsByClassName('validate-edit').item(0)
    validateButton.addEventListener('click', () => {
        let waitCount = 1
        const maybeReload = name => {
            waitCount--
            if (!waitCount)
                appStateSetDocument(name, false, true)
        }

        const newName = documentElement.querySelector('#name-input').value
        if (newName != name) {
            waitCount++
            postData(`/git-docs/api/documents/${category}/${name}/rename`, { name: newName })
                .then(_ => {
                    log(`renamed document ${name}`)
                    maybeReload(newName)
                })
        }

        const newContent = documentElement.getElementsByClassName('document-content-textarea').item(0).value
        if (newContent) {
            waitCount++
            putData(`/git-docs/api/documents/${category}/${name}/content`, newContent, 'application/markdown')
                .then(_ => {
                    log(`updated document ${name} content`)
                    maybeReload(newName)
                })
        }
        else {
            log(`no change to content`)
        }

        maybeReload()
    })
}

function drawDocument(category, name) {
    if (!name) {
        el('board-opened-documents').innerHTML = ``
        return
    }

    const documentElement = document.createElement('div')
    documentElement.classList.add('mui-panel')
    documentElement.innerHTML += `<div class='mui--text-dark-secondary mui--text-caption' style='padding-top:1em;padding-bottom:1.7em;'>${name}</div>`
    const metadataElement = document.createElement('div')
    documentElement.appendChild(metadataElement)
    documentElement.appendChild(elFromHtml(`<form id='document-add-tag-form'>Tags: <label><input id='document-add-tag-text'/></label> <button role='submit' class='mui-btn mui-btn--primary mui-btn--flat'>add tag</button></form>`))
    documentElement.appendChild(elFromHtml('<div class="mui-divider"></div>'))
    const contentElement = document.createElement('div')
    documentElement.appendChild(contentElement)
    documentElement.appendChild(elFromHtml('<div class="mui-divider"></div>'))
    documentElement.appendChild(elFromHtml(`<button onclick='deleteDocument("${category}", "${name}")' class="delete mui-btn mui-btn--small mui-btn--flat mui-btn--danger">Delete</button>`))
    documentElement.appendChild(elFromHtml(`<button onclick='appStateSetDocument("${name}", true, false)' class="mui-btn mui-btn--primary mui-btn--flat">Edit</button>`))

    documentElement.querySelector('#document-add-tag-form').addEventListener('submit', event => {
        event.preventDefault()
        event.stopPropagation()

        let tag = documentElement.querySelector('#document-add-tag-text').value

        addTagToDocument(category, name, tag)
    })

    const asyncCount = runAtLast => {
        let waited = 0
        return {
            add: function () {
                waited++
            },
            remove: function () {
                waited--
                if (waited == 0)
                    runAtLast()
            }
        }
    }

    const count = asyncCount(() => {
        el('board-opened-documents').innerHTML = ''
        el('board-opened-documents').appendChild(documentElement)
    })

    count.add()
    count.add()
    getData(`/git-docs/api/documents/${category}/${name}/metadata`)
        .then(metadata => {
            if (metadata && metadata.tags) {
                metadataElement.innerHTML += metadata.tags.map(tag => `<div class="badge ${badgeColorClass(tag)}" >${tag}&nbsp;<span onclick='deleteTagToDocument("${category}","${name}","${tag}")'>[X]</span></div>`).join('')
            }
            else {
                metadataElement.innerHTML += `<pre>${JSON.stringify(metadata, null, 2)}</pre>`
            }

            count.remove()
        })

    count.add()
    getData(`/git-docs/api/documents/${category}/${name}/content?interpolated=true`, 'application/markdown')
        .then(content => {
            contentElement.innerHTML += marked(content)

            count.remove()
        })
    count.remove()
}

function loadDocuments(category, search, split) {
    if (!category) {
        el('board-documents-ul').innerHTML = `No document can appear here until a category selected.`
        return
    }

    el('search-document').value = search || ''
    el('columns-document').value = split || ''

    let columns = split.split(",").map(v => v.trim())
    if (!columns.length)
        columns.push(null)

    let columnsElement = document.createElement('div')

    let nbFinishedColumns = 0
    const finishedOneColumn = () => {
        nbFinishedColumns++
        if (nbFinishedColumns != columns.length)
            return

        el('board-documents-ul').innerHTML = columnsElement.innerHTML
    }

    let documentIndex = -1
    for (let column of columns) {
        documentIndex++

        let q = search ? (column ? `& ${search} ${column}` : search) : column

        let columnElement = elFromHtml(`<div style='${documentIndex > 0 ? 'margin-left:1em;' : ''}'><div style='text-align: center;font-weight: bold;padding-bottom: .5em'>${q || 'All'}</div></div>`)
        columnsElement.appendChild(columnElement)

        getData(q ? `/git-docs/api/documents/${category}/?q=${encodeURIComponent(q)}` : `/git-docs/api/documents/${category}`)
            .then(documents => {
                let prep = documents.map(name => `<div><span style='cursor: pointer;' onclick='appStateSetDocument("${name}", false, false)'>${name}</span>&nbsp;<span x-id='tags'></span></div>`).join('')

                let columnDocumentsElement = elFromHtml(`<div class='mui-panel'>${prep}</div>`)

                let documentToFetchTags = 0
                const maybeLoad = () => {
                    if (documentToFetchTags >= documents.length) {
                        columnElement.appendChild(columnDocumentsElement)
                        finishedOneColumn()
                        return
                    }

                    let loadedDocumentTags = documentToFetchTags++
                    let name = documents[loadedDocumentTags]

                    getData(`/git-docs/api/documents/${category}/${name}/metadata`)
                        .then(metadata => {
                            if (metadata && metadata.tags)
                                columnDocumentsElement.children.item(loadedDocumentTags).querySelector('[x-id=tags]').innerHTML = metadata.tags.map(tagToHtmlBadge).join('')
                            maybeLoad()
                        })
                }

                maybeLoad()

            })
    }
}

function loadTags(category) {
    getData(`/git-docs/api/tags/${category}`)
        .then(tags => {
            el('tagsList').innerHTML = "All tags : " + tags.map(tagToHtmlBadge).join('')
        })
}

function loadStatus() {
    getData("/git-docs/api/status")
        .then(status => {
            if (!status) {
                log(`loadStatus failed`)
                return
            }

            let html = ''
            html += `working directory: ${status.workingDirectory}<br/>`
            html += `git repository: ${status.gitRepository ? status.gitRepository : '-'}<br/>`
            html += `<span style='color:${status.clean ? 'green' : 'red'};'>${status.clean ? 'ready for operations !' : 'working directory files not synced, commit your changes please'}</span>`
            if (!status.clean)
                html += `<pre>${status.text}</pre>`
            el('board-status').innerHTML = html
        })
}

function loadCategories(currentCategory) {
    fetchCategories().then(categories => {
        if (!categories) {
            el('board-categories').innerHTML = `fetch categories failed`
            return
        }

        el('board-categories').innerHTML = `Categories : ` + categories.map(category => `<div onclick='appStateSetCategory("${category}",false)' class='badge ${category == currentCategory ? 'badge-color-0' : ''}'>${category}</div>`).join('')
    })
}

function installUi() {
    let lastTimeTriggered = 0
    const runLoadDocuments = () => {
        lastTimeTriggered = Date.now()
        if (timer) {
            clearTimeout(timer)
            timer = 0
        }

        let search = el('search-document').value || ''
        let split = el('columns-document').value || ''

        appStateSetBoardSearch(search, split)
    }
    let timer = 0
    const DELAY = 50
    const maybeLoadDocuments = () => {
        const now = Date.now()

        if (lastTimeTriggered + DELAY > now) {
            if (!timer)
                timer = setTimeout(runLoadDocuments, lastTimeTriggered + DELAY - now)
            return
        }

        runLoadDocuments()
    }
    el('search-document').addEventListener('input', event => {
        maybeLoadDocuments()
    })

    el('columns-document').addEventListener('input', event => {
        maybeLoadDocuments()
    })

    el('new-document-form').addEventListener('submit', event => {
        event.preventDefault()
        event.stopPropagation()

        let name = el('new-document-name').value
        el('new-document-name').value = ''

        addDocument(appState.category, name)
    })

    const $bodyEl = document.body
    const $sidedrawerEl = el('sidedrawer')

    function showSidedrawer() {
        var options = {
            onclose: function () {
                $sidedrawerEl
                    .removeClass('active')
                    .appendTo(document.body);
            }
        };

        //var $overlayEl = $(mui.overlay('on', options));

        overlayEl.appendChild($sidedrawerEl)
        setTimeout(function () {
            $sidedrawerEl.classList.add('active')
        }, 20)
    }


    function hideSidedrawer() {
        $bodyEl.classList.toggle('hide-sidedrawer')
    }

    el('js-show-sidedrawer').addEventListener('click', showSidedrawer)
    el('js-hide-sidedrawer').addEventListener('click', hideSidedrawer)

    titleEls = document.getElementsByClassName('sidedrawer-title')

    for (let i = 0; i < titleEls.length; i++) {
        const titleEl = titleEls.item(i)
        let toToggle = titleEl.nextElementSibling
        toToggle.style.display = "none"
        titleEl.addEventListener('click', () => {
            toToggle.style.display = toToggle.style.display == "none" ? "" : "none"
        })
    }

    el('new-category-form').addEventListener('submit', event => {
        event.preventDefault()
        event.stopPropagation()

        let name = el('new-category-name').value
        el('new-category-name').value = null
        postData(`/git-docs/api/categories/${name}`).then(() => {
            appStateSetCategory(name, true)
        })
    })
}

installUi()

let selectedCategory = localStorage.getItem('selected-category')
if (!selectedCategory || selectedCategory == "") {
    fetchCategories().then(categories => {
        if (!categories || !categories.length)
            log(`no categories yet !`)
        else
            appStateSetCategory(categories[0], true)
    })
}
else {
    appStateSetCategory(selectedCategory, true)
}