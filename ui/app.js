// app.js — the complete wiring
import { emitter }     from './event-emitter.js'
import { DataService } from './data-service.js'
  
// ── 1. State — single source of truth ──
const state = { 
    classes: [],
    metadata: null,
    loading: false,
    error: null,
};

// ── 2. render() — only function that writes the DOM ──
function render() {
    const app = document.querySelector('#app')

    if (state.loading) {
        app.innerHTML = `<div class="status-loading">
        <span class="spinner"></span> Loading...</div>`
        return
    }

    if (state.error) {
        app.innerHTML = `<div class="status-error">⚠ ${state.error}</div>`
        return
    }

    if (state.classes.length === 0) {
        app.innerHTML = `<div class="status-empty">No classes yet.</div>`
        return
    }
    let classesHTML = `
        <div class="table">
            <div class="table-headers">
                <span>Name</span>
                <span>Capacity Limit</span>
                <span>Membership Tier</span>
                <span>Status</span>
            </div>
            <div class="table-content">`

    classesHTML += state.classes.map(c => `
                <div class="table-row">
                    <strong>${c.name}</strong>
                    <span>${c.capacity_limit}</span>
                    <span>${c.membership_tier}</span>
                    <span>${c.terminated ? 'Finished' : 'Ongoing'}</span>
                </div>
            `)
    .join('')

    let current_page = state.metadata.current_page
    let last_page = state.metadata.last_page

    classesHTML += `
            </div>
        </div>
        <div class="metadata">
            <div id="icon-first" class="${current_page == 1 ? 'icon-disabled' : 'icon-enabled'} icon"><<</i></div>
            <div id="icon-previous" class="${current_page == 1 ? 'icon-disabled' : 'icon-enabled'} icon"><</i></div>
            <div>Page ${current_page} of ${last_page}</div>
            <div id="icon-next" class="${current_page == last_page ? 'icon-disabled' : 'icon-enabled'} icon">></i></div>
            <div id="icon-last" class="${current_page == last_page ? 'icon-disabled' : 'icon-enabled'} icon">>></i></div>
        </div>
    `

    app.innerHTML = classesHTML


    document.getElementById('icon-first').addEventListener('click', () => {
        if(current_page == 1) return
        emitter.emit('classes:read', 1)
    })

    document.getElementById('icon-previous').addEventListener('click', () => {
        if(current_page == 1) return
        emitter.emit('classes:read', current_page - 1)
    })

    document.getElementById('icon-next').addEventListener('click', () => {
        if(current_page == last_page) return
        emitter.emit('classes:read', current_page + 1)
    })

    document.getElementById('icon-last').addEventListener('click', () => {
        if(current_page == last_page) return
        emitter.emit('classes:read', last_page)
    })

    // lucide.createIcons()
}

// ── 3. Observers — update state, then call render() ───
emitter.on('classes:loading', () => {
    state.loading = true
    state.error   = null
    render()
})
 
emitter.on('classes:loaded', (classes) => {
    state.classes   = classes.classes
    state.metadata   = classes['@metadata']
    state.loading = false
    render()
})

emitter.on('classes:error', (msg) => {
    state.error   = msg
    state.loading = false
    render()
})

emitter.on('classes:read', (page) => {
    render()
    DataService.fetchClasses(page)
})

    // ── 4. Boot ──
render()                    // initial paint (empty state)
DataService.fetchClasses(1)   // kick off the first fetch
