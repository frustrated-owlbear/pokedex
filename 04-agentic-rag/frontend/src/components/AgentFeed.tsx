import {useMemo, useState} from 'react';
import type {TraceFilter, TraceStep} from '../types/agent';

type AgentFeedProps = {
    steps: TraceStep[];
    busy: boolean;
};

const FILTERS: {id: TraceFilter; label: string}[] = [
    {id: 'all', label: 'All'},
    {id: 'events', label: 'Events'},
    {id: 'actions', label: 'Actions'},
    {id: 'thoughts', label: 'Thoughts'},
    {id: 'tools', label: 'Tools'},
];

function formatTime(timestamp: string): string {
    const date = new Date(timestamp);
    if (Number.isNaN(date.getTime())) {
        return '--:--:--';
    }
    return date.toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
    });
}

function matchesFilter(step: TraceStep, filter: TraceFilter): boolean {
    switch (filter) {
        case 'events':
            return step.kind === 'Event';
        case 'actions':
            return step.kind === 'Action' || step.kind === 'Observation';
        case 'thoughts':
            return step.kind === 'Thought';
        case 'tools':
            return step.kind === 'Action' || step.kind === 'Observation';
        default:
            return true;
    }
}

function kindClass(kind: TraceStep['kind']): string {
    switch (kind) {
        case 'Event':
            return 'agent-feed__badge--event';
        case 'Thought':
            return 'agent-feed__badge--thought';
        case 'Action':
            return 'agent-feed__badge--action';
        case 'Observation':
            return 'agent-feed__badge--observation';
        case 'FinalAnswer':
            return 'agent-feed__badge--final';
        default:
            return '';
    }
}

export function AgentFeed({steps, busy}: AgentFeedProps) {
    const [filter, setFilter] = useState<TraceFilter>('all');

    const visibleSteps = useMemo(
        () => steps.filter((step) => matchesFilter(step, filter)),
        [filter, steps],
    );

    return (
        <section className="agent-feed" aria-label="Agent activity">
            <div className="agent-feed__header">
                <div className="agent-feed__heading">
                    <p className="agent-feed__eyebrow">AGENT FEED</p>
                    <h2 className="agent-feed__title">Execution trace</h2>
                </div>
                <div className="agent-feed__filters" role="tablist" aria-label="Trace filters">
                    {FILTERS.map((item) => (
                        <button
                            key={item.id}
                            type="button"
                            role="tab"
                            aria-selected={filter === item.id}
                            className={`agent-feed__filter${
                                filter === item.id ? ' agent-feed__filter--active' : ''
                            }`}
                            onClick={() => setFilter(item.id)}
                        >
                            {item.label}
                        </button>
                    ))}
                </div>
            </div>

            <div className="agent-feed__timeline">
                {visibleSteps.length === 0 && !busy && (
                    <p className="agent-feed__empty">
                        Submit an observation to watch the agent decide which tools to use.
                    </p>
                )}
                {visibleSteps.map((step) => (
                    <article key={step.id} className="agent-feed__item">
                        <div className="agent-feed__rail" aria-hidden />
                        <div className="agent-feed__card">
                            <div className="agent-feed__card-meta">
                                <span className="agent-feed__time">{formatTime(step.timestamp)}</span>
                                <span className={`agent-feed__badge ${kindClass(step.kind)}`}>
                                    {step.kind.toUpperCase()}
                                </span>
                            </div>
                            <h3 className="agent-feed__card-title">{step.title}</h3>
                            {step.detail && (
                                <p className="agent-feed__card-detail">{step.detail}</p>
                            )}
                            {step.tools && step.tools.length > 0 && (
                                <ul className="agent-feed__tools">
                                    {step.tools.map((tool) => (
                                        <li key={tool}>{tool}</li>
                                    ))}
                                </ul>
                            )}
                        </div>
                    </article>
                ))}
                {busy && (
                    <p className="agent-feed__running" aria-live="polite">
                        Agent is thinking…
                    </p>
                )}
            </div>
        </section>
    );
}
