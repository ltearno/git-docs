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




function afterFetch(responseContentType = 'application/json') {
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
}

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
    return `badge-color-${c % 4}`
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
}

function appStateSetCategory(category) {
    if (category == appState.category)
        return

    appState.category = category
    appState.document = null

    // TODO : redraw full UI
}

function appStateSetDocument(document, modeEditDocument) {
    if (document == appState.document && modeEditDocument == appState.modeEditDocument)
        return

    appState.document = document
    appState.modeEditDocument = modeEditDocument

    drawDocumentDetail()
}

function drawDocumentDetail() {
    if (appState.modeEditDocument)
        drawDocumentEdition(appState.document)
    else
        drawDocument(appState.document)
}






function clear() {
    el('board-documents-ul').innerHTML = ''
    el('board-opened-documents').innerHTML = ''
}

function drawDocumentEdition(name) {
    el('board-opened-documents').innerHTML = ''

    if (!name)
        return

    const documentElement = document.createElement('div')
    documentElement.classList.add('mui-panel')
    documentElement.innerHTML += `<input id='name-input' type='text' style='font-size:2em;'/></input>`
    const contentElement = document.createElement('div')
    contentElement.innerHTML += `<h2>Content</h2>`
    documentElement.appendChild(contentElement)
    documentElement.appendChild(elFromHtml(`<button onclick='appStateSetDocument("${name}", false)' class="mui-btn mui-btn--raised">Cancel</button>`))
    documentElement.appendChild(elFromHtml(`<button class="validate-edit mui-btn mui-btn--primary mui-btn--raised">Validate</button>`))

    el('board-opened-documents').appendChild(documentElement)

    documentElement.querySelector('#name-input').value = name

    getData(`/api/documents/issues/${name}/content`, 'application/mardown')
        .then(content => contentElement.innerHTML += `<textarea class='document-content-textarea' style='width:80em;height:30em;'>${content}</textarea>`)
        .catch(err => log(`get content for ${name} failed`))

    let validateButton = documentElement.getElementsByClassName('validate-edit').item(0)
    validateButton.addEventListener('click', () => {
        let waitCount = 1
        const maybeReload = name => {
            waitCount--
            if (!waitCount)
                reloadWithDocument(name)
        }

        const newName = documentElement.querySelector('#name-input').value
        if (newName != name) {
            waitCount++
            postData(`/api/documents/issues/${name}/rename`, { name: newName })
                .then(_ => {
                    log(`renamed document ${name}`)
                    maybeReload(newName)
                })
                .catch(err => log(`rename ${name} failed`))
        }

        const newContent = documentElement.getElementsByClassName('document-content-textarea').item(0).value
        if (newContent) {
            waitCount++
            putData(`/api/documents/issues/${name}/content`, newContent, 'application/markdown')
                .then(_ => {
                    log(`updated document ${name} content`)
                    maybeReload(newName)
                })
                .catch(err => log(`editDocument ${name} failed`))
        }
        else {
            log(`no change to content`)
        }

        maybeReload()
    })
}

function drawDocument(name) {
    el('board-opened-documents').innerHTML = ''

    if (!name)
        return

    const documentElement = document.createElement('div')
    documentElement.classList.add('mui-panel')
    documentElement.innerHTML += `<div class='mui--text-dark-secondary mui--text-caption' style='padding-top:1em;padding-bottom:1.7em;'>${name}</div>`
    const metadataElement = document.createElement('div')
    documentElement.appendChild(metadataElement)
    documentElement.appendChild(elFromHtml(`<form id='document-add-tag-form'>Tags: <label><input id='document-add-tag-text'/></label><button role='submit' class='mui-btn mui-btn--primary mui-btn--flat'>add tag</button></form>`))
    documentElement.appendChild(elFromHtml('<div class="mui-divider"></div>'))
    const contentElement = document.createElement('div')
    documentElement.appendChild(contentElement)
    documentElement.appendChild(elFromHtml('<div class="mui-divider"></div>'))
    documentElement.appendChild(elFromHtml(`<button onclick='appStateSetDocument("${name}", true)' class="mui-btn mui-btn--primary mui-btn--flat">Edit document</button>`))

    documentElement.querySelector('#document-add-tag-form').addEventListener('submit', event => {
        event.preventDefault()
        event.stopPropagation()

        let tag = documentElement.querySelector('#document-add-tag-text').value

        addTagToDocument(name, tag)
    })

    el('board-opened-documents').appendChild(documentElement)

    getData(`/api/documents/issues/${name}/metadata`)
        .then(metadata => {
            if (metadata && metadata.tags) {
                metadataElement.innerHTML += metadata.tags.map(tagToHtmlBadge).join('')
            }
            else {
                metadataElement.innerHTML += `<pre>${JSON.stringify(metadata, null, 2)}</pre>`
            }
        })

    getData(`/api/documents/issues/${name}/content?interpolated=true`, 'application/markdown')
        .then(content => {
            contentElement.innerHTML += marked(content)
        })
        .catch(err => log(`get content for ${name} failed`))
}

function fetchCategories() {
    return getData(`/api/categories`)
}

function addTagToDocument(name, tagToAdd) {
    getData(`/api/documents/issues/${name}/metadata`)
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
                putData(`/api/documents/issues/${name}/metadata`, metadata)
                    .then(_ => {
                        log(`update document metadata ${name}`)
                        reloadWithDocument(name)
                    })
                    .catch(err => log(`updateDocument metadata ${name} failed`))
            }
            else {
                log(`tag already present`)
            }
        })
        .catch(err => log(`get metadata for ${name} failed`))
}

function loadDocuments() {
    let search = el('search-document').value || null
    let columns = (el('columns-document').value || '').split(",").map(v => v.trim())

    if (search)
        localStorage.setItem('search-document', search)
    else
        localStorage.removeItem('search-document')
    if (columns)
        localStorage.setItem('columns-document', columns)
    else
        localStorage.removeItem('columns-document')

    if (!columns.length)
        columns.push(null)

    let columnsElement = el('board-documents-ul')
    columnsElement.innerHTML = ''

    let documentIndex = -1
    for (let column of columns) {
        documentIndex++

        let q = search ? (column ? `& ${search} ${column}` : search) : column

        let columnElement = elFromHtml(`<div><div style='text-align: center;font-weight: bold;padding-bottom: .5em;'>${q || 'All'}</div></div>`)
        if (documentIndex > 0)
            columnElement.style.marginLeft = '1em'
        columnsElement.appendChild(columnElement)

        getData(q ? `/api/documents/issues/?q=${encodeURIComponent(q)}` : "/api/documents/issues")
            .then(documents => {
                let prep = documents.map(name => `<div><span onclick='appStateSetDocument("${name}", false)'>${name}</span>&nbsp;<span x-id='tags'></span>&nbsp;&nbsp;&nbsp;<button onclick='deleteDocument("${name}")' class="delete mui-btn mui-btn--small mui-btn--flat mui-btn--danger">X</button></div>`).join('')

                let columnDocumentsElement = elFromHtml(`<div class='mui-panel'>${prep}</div>`)

                let documentToFetchTags = 0
                const maybeLoad = () => {
                    if (documentToFetchTags >= documents.length) {
                        columnElement.appendChild(columnDocumentsElement)
                        return
                    }

                    let loadedDocumentTags = documentToFetchTags++
                    let name = documents[loadedDocumentTags]

                    getData(`/api/documents/issues/${name}/metadata`)
                        .then(metadata => {
                            if (metadata && metadata.tags)
                                columnDocumentsElement.children.item(loadedDocumentTags).querySelector('[x-id=tags]').innerHTML = metadata.tags.map(tagToHtmlBadge).join('')
                            maybeLoad()
                        })
                        .catch(err => {
                            log(`get metadata for ${name} failed`)
                            maybeLoad()
                        })
                }

                maybeLoad()

            })
            .catch(err => log(`loadDocuments failed`))
    }

    loadTags()
}

function loadTags() {
    getData("/api/tags/issues")
        .then(tags => {
            el('tagsList').innerHTML = "All tags : " + tags.map(tagToHtmlBadge).join('')
        })
        .catch(err => log(`loadTags failed`))
}

function loadStatus() {
    getData("/api/status")
        .then(status => {
            el('board-status').innerHTML = `repository: ${status.gitRepository}<br/>`
            el('board-status').innerHTML += `<span style='color:${status.clean ? 'green' : 'red'};'>${status.clean ? 'ready for operations !' : 'working directory files not synced, commit your changes please'}</span>`
            if (!status.clean)
                el('board-status').innerHTML += `<pre>${status.text}</pre>`
        })
        .catch(err => log(`loadStatus failed`))
}

function reloadWithDocument(document) {
    appStateSetDocument(document, false)
    loadStatus()
    loadDocuments()
}

function deleteDocument(name) {
    deleteData(`/api/documents/issues/${name}`)
        .then(_ => {
            log(`delete document ${name}`)
            reloadWithDocument(null)
        })
        .catch(err => log(`deleteDocument ${name} failed`))
}

function addDocument(name) {
    postData(`/api/documents/issues/${name}`, {})
        .then(_ => {
            log(`add document ${name}`)
            reloadWithDocument(name)
        })
        .catch(err => log(`addDocument ${name} failed`))
}

function installUi() {
    el('search-document').value = localStorage.getItem('search-document') || null
    el('columns-document').value = localStorage.getItem('columns-document') || null

    let loadDocumentsTriggered = false
    const maybeLoadDocuments = () => {
        if (loadDocumentsTriggered)
            return

        loadDocumentsTriggered = true
        setTimeout(() => {
            loadDocumentsTriggered = false
            loadDocuments()
        }, 50)
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

        addDocument(name)
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
}

installUi()
loadDocuments()
loadStatus()