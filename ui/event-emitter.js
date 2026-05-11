// event-emitter.js
function createEmitter() {
    // Map<string, Function[]>
    // e.g. { "data:loaded": [fn1, fn2], "error": [fn3] }
    const listeners = new Map();

    // event-emitter.js
    function on(event, callback) {
        if (!listeners.has(event)) {
            listeners.set(event, []);
        }
        listeners.get(event).push(callback);
    }

    // event-emitter.js
    function emit(event, data) {
        if (!listeners.has(event)) {
            return
        }
        listeners.get(event).forEach(fn => fn(data))
    }

    // event-emitter.js
    function off(event, callback) {
        if (!listeners.has(event)) {
            return
        }
        listeners.set(event, listeners.get(event).filter(cb => cb !== callback))
    }

    return { on, emit, off };
}

// One shared instance for the whole app
// emitter has access to all three functions - on, emit, off
export const emitter = createEmitter();
