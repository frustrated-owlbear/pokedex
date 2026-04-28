import {useEffect, useState} from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import './App.css';
import {AnswerQuestion} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

function App() {
    const [answer, setAnswer] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);

    useEffect(() => {
        const unsubscribe = EventsOn('llm:chunk', (...args: unknown[]) => {
            const chunk =
                typeof args[0] === 'string' ? args[0] : String(args[0] ?? '');
            setAnswer((prev) => (prev ?? '') + chunk);
        });
        return unsubscribe;
    }, []);

    async function ask() {
        setBusy(true);
        setAnswer(null);
        try {
            await AnswerQuestion('');
        } catch {
            setAnswer((prev) => prev ?? 'Something went wrong. Try again!');
        } finally {
            setBusy(false);
        }
    }

    const showPanel = busy || answer != null;
    const waitingForFirstChunk = busy && (answer === null || answer === '');

    return (
        <div className="poke-app">
            <button
                type="button"
                className="ask-pokedex-btn"
                onClick={ask}
                disabled={busy}
            >
                Ask pokedex
            </button>
            {showPanel && (
                <div className="poke-answer" role="status">
                    {waitingForFirstChunk ? (
                        'Thinking…'
                    ) : (
                        <div className="poke-answer-md">
                            <ReactMarkdown remarkPlugins={[remarkGfm]}>
                                {answer ?? ''}
                            </ReactMarkdown>
                            {busy && (
                                <span
                                    className="poke-stream-caret"
                                    aria-hidden
                                />
                            )}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

export default App;
