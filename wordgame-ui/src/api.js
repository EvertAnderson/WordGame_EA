// src/api.js

export async function newGame() {
    const res = await fetch('/api/new', {
        method: 'POST',
    })

    if (!res.ok) {
        throw new Error(await res.text())
    }

    return res.json()
}

export async function makeGuess(id, guess) {
    const res = await fetch('/api/guess', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id, guess }),
    })

    if (!res.ok) {
        const text = await res.text()
        const err = new Error(text || `HTTP ${res.status}`)
        err.status = res.status
        throw err
    }

    return res.json()
}
