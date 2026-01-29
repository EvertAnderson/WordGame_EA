import { useEffect, useMemo, useState } from 'react'
import { makeGuess, newGame } from './api'
import './styles.css'

const ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split('')

function normalizeGuess(input) {
    if (!input) return ''
    const s = input.trim().toUpperCase()
    if (s.length !== 1) return ''
    const c = s[0]
    if (c < 'A' || c > 'Z') return ''
    return c
}

export default function App() {
    const [gameId, setGameId] = useState('')
    const [current, setCurrent] = useState('')
    const [remaining, setRemaining] = useState(6)
    const [used, setUsed] = useState(() => new Set())
    const [status, setStatus] = useState('idle') // idle | playing | won | lost
    const [loading, setLoading] = useState(false)
    const [errMsg, setErrMsg] = useState('')
    const [guessInput, setGuessInput] = useState('')

    const masked = useMemo(() => {
        // display with spaces: _ P P _ _
        return current ? current.split('').join(' ') : ''
    }, [current])

    const start = async () => {
        setErrMsg('')
        setLoading(true)
        try {
            const g = await newGame()
            setGameId(g.id)
            setCurrent(g.current)
            setRemaining(g.guesses_remaining)
            setUsed(new Set())
            setStatus('playing')
            setGuessInput('')
        } catch (e) {
            setErrMsg(String(e.message || e))
            setStatus('idle')
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        // auto-start
        start()
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const completedFrom = (board, rem) => {
        if (!board) return 'idle'
        if (board.indexOf('_') === -1) return 'won'
        if (rem === 0) return 'lost'
        return 'playing'
    }

    const submitGuess = async (letter) => {
        const g = normalizeGuess(letter)
        if (!g) return
        if (status !== 'playing') return
        if (used.has(g)) return

        setErrMsg('')
        setLoading(true)

        try {
            const next = await makeGuess(gameId, g)
            setCurrent(next.current)
            setRemaining(next.guesses_remaining)

            setUsed(prev => {
                const copy = new Set(prev)
                copy.add(g)
                return copy
            })

            const nextStatus = completedFrom(next.current, next.guesses_remaining)
            setStatus(nextStatus)
        } catch (e) {
            // Si la API borra el juego al finalizar, podría devolver 404 en una jugada tardía.
            setErrMsg(String(e.message || e))
            if (e.status === 404) setStatus('idle')
        } finally {
            setLoading(false)
            setGuessInput('')
        }
    }

    const onFormSubmit = (e) => {
        e.preventDefault()
        submitGuess(guessInput)
    }

    const title = status === 'won'
        ? '¡Ganaste!'
        : status === 'lost'
            ? 'Perdiste'
            : 'Fleet Word Game'

    return (
        <div className="page">
            <div className="card">
                <header className="header">
                    <div>
                        <h1 className="h1">{title}</h1>
                        <p className="sub">
                            {status === 'playing'
                                ? `Intentos restantes: ${remaining}`
                                : status === 'idle'
                                    ? 'Inicia una partida para jugar.'
                                    : `Intentos restantes: ${remaining}`}
                        </p>
                    </div>

                    <button className="btn" onClick={start} disabled={loading}>
                        {status === 'playing' ? 'Reiniciar' : 'Nuevo juego'}
                    </button>
                </header>

                <section className="board">
                    <div className="word" aria-label="current board">
                        {current ? masked : '—'}
                    </div>
                    <div className="meta">
                        <span className="pill">ID: {gameId ? gameId.slice(0, 8) + '…' : '—'}</span>
                        <span className="pill">Usadas: {used.size}</span>
                    </div>
                </section>

                <section className="controls">
                    <form className="guessForm" onSubmit={onFormSubmit}>
                        <input
                            className="input"
                            value={guessInput}
                            onChange={(e) => setGuessInput(e.target.value)}
                            placeholder="Letra (A-Z)"
                            inputMode="text"
                            maxLength={2}
                            disabled={loading || status !== 'playing'}
                            aria-label="guess input"
                        />
                        <button className="btn primary" type="submit" disabled={loading || status !== 'playing'}>
                            Adivinar
                        </button>
                    </form>

                    {errMsg ? (
                        <div className="error" role="alert">{errMsg}</div>
                    ) : null}

                    <div className="keyboard" aria-label="on-screen keyboard">
                        {ALPHABET.map((l) => {
                            const isUsed = used.has(l)
                            const disabled = loading || status !== 'playing' || isUsed
                            return (
                                <button
                                    key={l}
                                    className={`key ${isUsed ? 'used' : ''}`}
                                    disabled={disabled}
                                    onClick={() => submitGuess(l)}
                                >
                                    {l}
                                </button>
                            )
                        })}
                    </div>

                    {(status === 'won' || status === 'lost') && (
                        <div className="end">
                            <div className="endMsg">
                                {status === 'won'
                                    ? 'Bien jugado. ¿Otra ronda?'
                                    : 'Se acabaron los intentos. ¿Quieres intentarlo de nuevo?'}
                            </div>
                            <button className="btn primary" onClick={start} disabled={loading}>
                                Jugar de nuevo
                            </button>
                        </div>
                    )}
                </section>
            </div>

            <footer className="footer">
                <span>Usa el teclado en pantalla o escribe una letra.</span>
            </footer>
        </div>
    )
}
