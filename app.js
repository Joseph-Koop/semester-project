// app.js — the complete wiring
import { emitter }     from './event-emitter.js'
import { DataService } from './data-service.js'
  
// ── 1. State — single source of truth ──
const state = { 
    classes: [],
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
    app.innerHTML = state.classes.map(c => `<div class="card">
            <strong>${c.name}</strong>
            <span class="badge">${c.studio_id}</span></div>
            <span class="badge">${c.trainer_id}</span></div>
            <span class="badge">${c.capacity_limit}</span></div>
            <span class="badge">${c.membership_tier}</span></div>
            <span class="badge">${c.terminated}</span></div>`)
        .join('')
}

// ── 3. Observers — update state, then call render() ───
emitter.on('classes:loading', () => {
    state.loading = true
    state.error   = null
    render()
})
 
emitter.on('classes:loaded', (classes) => {
    state.classes   = classes
    state.loading = false
    render()
})

emitter.on('classes:error', (msg) => {
    state.error   = msg
    state.loading = false
    render()
})

    // ── 4. Boot ──
render()                    // initial paint (empty state)
DataService.fetchClasses()   // kick off the first fetch
