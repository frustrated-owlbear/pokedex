import {FormEvent, KeyboardEvent, useEffect, useMemo, useState} from 'react';
import DOMPurify from 'dompurify';
import {marked} from 'marked';
import './App.css';
import {AskPokedex} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

type SystemState = {
    label: string;
    value: string;
    detail?: string;
    healthy: boolean;
};

function App() {
    const [prompt, setPrompt] = useState('');
    const [response, setResponse] = useState('');
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState('');
    const [lastSubmittedAt, setLastSubmittedAt] = useState<string | null>(null);
    const [systemState, setSystemState] = useState<SystemState>({
        label: 'Ollama',
        value: 'Checking...',
        detail: 'Waiting for healthcheck',
        healthy: false,
    });

    useEffect(() => {
        const unsubscribeChunk = EventsOn('llm:chunk', (...args: unknown[]) => {
            const chunk =
                typeof args[0] === 'string' ? args[0] : String(args[0] ?? '');
            setResponse((prev) => prev + chunk);
        });

        const unsubscribeStatus = EventsOn('ollama:status', (...args: unknown[]) => {
            setSystemState(parseSystemState(args[0]));
        });

        return () => {
            unsubscribeChunk();
            unsubscribeStatus();
        };
    }, []);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const question = prompt.trim();
        if (!question || busy) {
            return;
        }

        setBusy(true);
        setError('');
        setResponse('');
        setLastSubmittedAt(
            new Date().toLocaleTimeString([], {hour: '2-digit', minute: '2-digit', hour12: false}),
        );
        try {
            await AskPokedex(question);
        } catch {
            setError('Could not reach the Pokédex. Try again.');
        } finally {
            setBusy(false);
        }
    }

    function handlePromptKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
        if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
            event.preventDefault();
            const form = event.currentTarget.form;
            if (form) {
                form.requestSubmit();
            }
        }
    }

    const renderedMarkdown = useMemo(() => {
        if (error || (!busy && !response)) {
            return '';
        }
        return DOMPurify.sanitize(marked.parse(response || 'Thinking...') as string);
    }, [busy, error, response]);

    return (
        <main className="inference-screen">
            <div className="inference-screen__header">
                <section className="system-status-card" aria-label="System status">
                    <p className="system-status-card__title">SYSTEM STATUS</p>
                    <div className="system-status-card__divider" aria-hidden />
                    <ul className="system-status-list">
                        <li
                            className={`system-status-item${
                                systemState.healthy ? '' : ' system-status-item--unhealthy'
                            }`}
                        >
                            <div className="system-status-row">
                                <div className="system-status-row__meta">
                                    <span className="system-status-row__icon" aria-hidden>
                                        <StatusGlyph />
                                    </span>
                                    <span className="system-status-row__label">{systemState.label}</span>
                                </div>
                                <span className="system-status-row__value">{systemState.value}</span>
                            </div>
                            {systemState.detail && (
                                <p className="system-status-row__detail">{systemState.detail}</p>
                            )}
                        </li>
                    </ul>
                </section>

                <section className="composer-dock" aria-label="Inference input">
                    <form className="composer" onSubmit={handleSubmit}>
                        <div className="composer__panel">
                            <textarea
                                value={prompt}
                                onChange={(event) => setPrompt(event.target.value)}
                                onKeyDown={handlePromptKeyDown}
                                className="composer__textarea"
                                placeholder="Describe or ask Pokédex..."
                                rows={4}
                            />

                            <button
                                type="submit"
                                className="composer__send"
                                disabled={busy || !prompt.trim()}
                                aria-label="Send input"
                            >
                                <SendGlyph />
                            </button>
                        </div>
                    </form>

                    <div className="composer__footer">
                        <span>Press Cmd/Ctrl + Enter to send</span>
                        {lastSubmittedAt && (
                            <span className="composer__footer-time">Last submitted at {lastSubmittedAt}</span>
                        )}
                    </div>
                </section>
            </div>

            <div className="inference-screen__body">
                <div className="response-shell">
                    <section
                        className={`response-panel${error ? ' response-panel--error' : ''}`}
                        role="status"
                        aria-live="polite"
                    >
                        {error ? (
                            <p className="response-panel__message">{error}</p>
                        ) : busy || response ? (
                            <>
                                <div
                                    className="response-panel__markdown"
                                    dangerouslySetInnerHTML={{__html: renderedMarkdown}}
                                />
                                {busy && <span className="composer__caret" aria-hidden />}
                            </>
                        ) : null}
                    </section>
                </div>
            </div>
        </main>
    );
}

function parseSystemState(payload: unknown): SystemState {
    if (!payload || typeof payload !== 'object') {
        return {
            label: 'Ollama',
            value: 'Unavailable',
            detail: 'Invalid healthcheck payload',
            healthy: false,
        };
    }

    const record = payload as Record<string, unknown>;
    return {
        label: typeof record.label === 'string' ? record.label : 'Ollama',
        value: typeof record.value === 'string' ? record.value : 'Unavailable',
        detail: typeof record.detail === 'string' ? record.detail : undefined,
        healthy: record.healthy === true,
    };
}

function StatusGlyph() {
    return (
        <svg viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="7" stroke="currentColor" strokeWidth="1.6" />
            <path
                d="M9.25 12.25l1.8 1.8 3.9-4.1"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}

function SendGlyph() {
    return (
        <svg viewBox="0 0 24 24" fill="none">
            <path
                d="M5 12h11"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
            />
            <path
                d="M12 6l6 6-6 6"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}

export default App;
