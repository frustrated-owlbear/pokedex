type SystemState = {
    label: string;
    value: string;
    detail?: string;
    healthy: boolean;
};

type SystemStatusCardProps = {
    states: SystemState[];
    compact?: boolean;
};

export function SystemStatusCard({states, compact = false}: SystemStatusCardProps) {
    const items =
        states.length > 0
            ? states
            : [
                  {
                      label: 'Ollama',
                      value: 'Checking...',
                      detail: 'Waiting for healthcheck',
                      healthy: false,
                  },
              ];

    return (
        <section
            className={`system-status-card${compact ? ' system-status-card--compact' : ''}`}
            aria-label="System status"
        >
            <p className="system-status-card__title">SYSTEM STATUS</p>
            <div className="system-status-card__divider" aria-hidden />
            <ul className="system-status-list">
                {items.map((state) => (
                    <li
                        key={state.label}
                        className={`system-status-item${
                            state.healthy ? '' : ' system-status-item--unhealthy'
                        }`}
                    >
                        <div className="system-status-row">
                            <div className="system-status-row__meta">
                                <span className="system-status-row__icon" aria-hidden>
                                    <StatusGlyph healthy={state.healthy} />
                                </span>
                                <span className="system-status-row__label">{state.label}</span>
                            </div>
                            <span className="system-status-row__value">{state.value}</span>
                        </div>
                        {state.detail && (
                            <p className="system-status-row__detail">{state.detail}</p>
                        )}
                    </li>
                ))}
            </ul>
        </section>
    );
}

function StatusGlyph({healthy}: {healthy: boolean}) {
    if (!healthy) {
        return (
            <svg viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="7" stroke="currentColor" strokeWidth="1.6" />
                <path
                    d="M9 9l6 6M15 9l-6 6"
                    stroke="currentColor"
                    strokeWidth="1.6"
                    strokeLinecap="round"
                />
            </svg>
        );
    }

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

export type {SystemState};
