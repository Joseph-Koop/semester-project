// data-service.js
import { emitter } from './event-emitter.js'
const API_BASE = 'http://localhost:4000/';

// Common Pattern: do not export each function; instead put them in
// an object and export only the object
export const DataService = {
    async fetchClasses() {
        // 1. announce that loading has started
        emitter.emit('classes:loading');
        try {
            const res  = await fetch(`${API_BASE}/classes`);
            if(!res.ok) {
                throw new Error(`Server error: ${res.status}`);
            }
            const classes = await res.json();
            // 2. announce success — pass the data to all listeners
            // console.log(classes);
            emitter.emit('classes:loaded', classes.classes);
        } catch(err) {
        // 3. announce failure — pass the error message
            emitter.emit('classes:error', err.message);
        }
    }
    // async createUser(payload) {
    //     emitter.emit('classes:loading');
    //     try {
    //         const res = await fetch(`${API_BASE}/classes`, {
    //             method:  'POST',
    //             headers: { 'Content-Type': 'application/json' },
    //             body:    JSON.stringify(payload),
    //         });
    //         if (!res.ok) { 
    //             throw new Error(`Server error: ${res.status}`);
    //         }
    //         const newUser = await res.json();
    //         emitter.emit('classes:created', newUser);
    //     } catch (err) {
    //         emitter.emit('classes:error', err.message);
    //     }
    // }
}

// // state.js — the starting shape (of the data)
// // The complete state for a user list UI
// const state = {
//     classes:   [],         // Array — the data itself
//     loading: false,     // Boolean — is a fetch in flight?
//     error:   null,      // String | null — error message, or null if none
// };

//     // class-toggle pattern
//     function render() {
//         // CSS defines the states — JS just toggles the class
//         loadingBanner.classList.toggle('visible', state.loading);
//         errorBanner.classList.toggle('visible', !!state.error);
//         if (state.error) {
//             errorBanner.textContent = state.error;
//         }
//         // innerHTML only for the dynamic list content
//         if (!state.loading && !state.error) {
//             list.innerHTML = state.classes.map(u => `...`).join('');
//         }
//     }

