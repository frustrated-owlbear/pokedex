import {useCallback, useEffect, useState} from 'react';
import {ListSessions} from '../../wailsjs/go/main/App';
import type {SessionView} from '../types/agent';

interface SessionsScreenProps {
    onBack: () => void;
    refreshKey?: number;
}

function formatDateRange(startedAt: string, endedAt?: string): string {
    const start = new Date(startedAt);
    if (Number.isNaN(start.getTime())) {
        return startedAt;
    }
    const startLabel = start.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
    });
    if (!endedAt) {
        return `${startLabel} · active`;
    }
    const end = new Date(endedAt);
    if (Number.isNaN(end.getTime())) {
        return startLabel;
    }
    const endLabel = end.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
    });
    return `${startLabel} – ${endLabel}`;
}

export function SessionsScreen({onBack, refreshKey = 0}: SessionsScreenProps) {
    const [sessions, setSessions] = useState<SessionView[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    const loadSessions = useCallback(async () => {
        setLoading(true);
        setError('');
        try {
            const result = await ListSessions();
            setSessions(result ?? []);
        } catch {
            setSessions([]);
            setError('Could not load sessions.');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        loadSessions();
    }, [loadSessions, refreshKey]);

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent) {
            if (event.key !== 'b' && event.key !== 'B') {
                return;
            }
            const target = event.target;
            if (
                target instanceof HTMLInputElement ||
                target instanceof HTMLTextAreaElement ||
                target instanceof HTMLSelectElement ||
                (target instanceof HTMLElement && target.isContentEditable)
            ) {
                return;
            }
            event.preventDefault();
            onBack();
        }

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [onBack]);

    return (
        <main className="sessions-screen">
            <div className="sessions-screen__shell">
                <header className="sessions-screen__header">
                    <button type="button" className="sessions-screen__back" onClick={onBack}>
                        ← Back
                    </button>
                    <h1 className="sessions-screen__title">Sessions</h1>
                    <p className="sessions-screen__subtitle">
                        Past gameplay sessions and summaries
                    </p>
                </header>

                {loading ? (
                    <p className="sessions-screen__status">Loading sessions...</p>
                ) : error ? (
                    <p className="sessions-screen__status sessions-screen__status--error">
                        {error}
                    </p>
                ) : sessions.length === 0 ? (
                    <p className="sessions-screen__status">No sessions recorded yet.</p>
                ) : (
                    <ul className="sessions-screen__list">
                        {sessions.map((session) => (
                            <li key={session.id} className="sessions-screen__card">
                                <div className="sessions-screen__card-header">
                                    <p className="sessions-screen__card-dates">
                                        {formatDateRange(session.startedAt, session.endedAt)}
                                    </p>
                                    {session.active ? (
                                        <span className="sessions-screen__badge sessions-screen__badge--active">
                                            Active
                                        </span>
                                    ) : null}
                                </div>
                                <p className="sessions-screen__card-meta">
                                    {session.eventCount}{' '}
                                    {session.eventCount === 1 ? 'observation' : 'observations'}
                                </p>
                                <p className="sessions-screen__card-summary">
                                    {session.summary ||
                                        (session.active
                                            ? 'In progress…'
                                            : 'No summary recorded.')}
                                </p>
                            </li>
                        ))}
                    </ul>
                )}

                <footer className="sessions-screen__footer">
                    <span>Press B to go back</span>
                </footer>
            </div>
        </main>
    );
}
