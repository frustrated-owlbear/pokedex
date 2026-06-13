import type {SituationView} from '../types/agent';
import {MyTeamWidget, type TeamPokemon} from './MyTeamWidget';

type SituationPanelProps = {
    situation: SituationView;
    team: TeamPokemon[];
    teamLoading: boolean;
    teamError: string;
    analysis: string;
    onOpenTeam: () => void;
    onSelectPokemon: (id: number) => void;
};

export function SituationPanel({
    situation,
    team,
    teamLoading,
    teamError,
    analysis,
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
