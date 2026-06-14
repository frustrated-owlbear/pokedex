import type {SessionView, SituationView} from '../types/agent';
import {MyTeamWidget, type TeamPokemon} from './MyTeamWidget';

type SituationPanelProps = {
    situation: SituationView;
    team: TeamPokemon[];
    teamLoading: boolean;
    teamError: string;
    analysis: string;
    activeSession: SessionView | null;
    onOpenTeam: () => void;
    onSelectPokemon: (id: number) => void;
};

function formatSessionStart(startedAt?: string): string {
    if (!startedAt) {
        return 'No active session';
    }
    const date = new Date(startedAt);
    if (Number.isNaN(date.getTime())) {
        return 'Recording events…';
    }
    return `Started ${date.toLocaleString(undefined, {
        month: 'short',
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit',
    })}`;
}

export function SituationPanel({
    situation,
    team,
    teamLoading,
    teamError,
    analysis,
    activeSession,
    onOpenTeam,
    onSelectPokemon,
}: SituationPanelProps) {
    return (
        <aside className="situation-panel" aria-label="Context and tools">
            <section className="situation-card">
                <p className="situation-card__title">CURRENT SITUATION</p>
                <div className="situation-card__divider" aria-hidden />
                <dl className="situation-list">
                    <div>
                        <dt>Location</dt>
                        <dd>{situation.location}</dd>
                    </div>
                    <div>
                        <dt>Time</dt>
                        <dd>
                            {situation.time} · {situation.period}
                        </dd>
                    </div>
                    <div>
                        <dt>Weather</dt>
                        <dd>{situation.weather}</dd>
                    </div>
                </dl>
            </section>

            <MyTeamWidget
                team={team}
                loading={teamLoading}
                error={teamError}
                onOpen={onOpenTeam}
                onSelectPokemon={onSelectPokemon}
            />

            <section className="situation-card">
                <p className="situation-card__title">ANALYSIS</p>
                <div className="situation-card__divider" aria-hidden />
                <p className="situation-card__body">
                    {analysis || 'Analysis will appear after the agent runs tools.'}
                </p>
            </section>

            <section className="situation-card">
                <p className="situation-card__title">MEMORY</p>
                <div className="situation-card__divider" aria-hidden />
                <p className="situation-card__body">{situation.memory}</p>
            </section>

            <section className="situation-card">
                <p className="situation-card__title">SESSION</p>
                <div className="situation-card__divider" aria-hidden />
                <p className="situation-card__body">
                    {activeSession?.active
                        ? formatSessionStart(activeSession.startedAt)
                        : 'No active session'}
                </p>
                {activeSession?.active ? (
                    <p className="situation-card__meta">
                        {activeSession.eventCount}{' '}
                        {activeSession.eventCount === 1 ? 'event' : 'events'} recorded
                    </p>
                ) : null}
                <p className="situation-card__meta">Session managed automatically by Pokédex</p>
            </section>

            <section className="situation-card">
                <p className="situation-card__title">TOOLS</p>
                <div className="situation-card__divider" aria-hidden />
                <ul className="situation-tools">
                    {situation.tools.map((tool) => (
                        <li key={tool}>{tool}</li>
                    ))}
                </ul>
            </section>
        </aside>
    );
}
