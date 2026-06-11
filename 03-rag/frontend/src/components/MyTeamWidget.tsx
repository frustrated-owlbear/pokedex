import {domain} from '../../wailsjs/go/models';
import {PokemonTile} from './PokemonTile';

export type TeamPokemon = domain.TeamPokemon;

interface MyTeamWidgetProps {
    team: TeamPokemon[];
    loading: boolean;
    error?: string;
    onOpen: () => void;
    onSelectPokemon: (id: number) => void;
}

export function MyTeamWidget({team, loading, error, onOpen, onSelectPokemon}: MyTeamWidgetProps) {
    return (
        <section className="my-team-widget" aria-label="My team">
            <header className="my-team-widget__header">
                <div className="my-team-widget__title-group">
                    <span className="my-team-widget__icon" aria-hidden>
                        <UsersGlyph />
                    </span>
                    <h2 className="my-team-widget__title">MY TEAM</h2>
                </div>
                <button type="button" className="my-team-widget__view-all" onClick={onOpen}>
                    View all
                </button>
            </header>

            <div className="my-team-widget__divider" aria-hidden />

            {loading ? (
                <p className="my-team-widget__status">Loading team...</p>
            ) : error ? (
                <p className="my-team-widget__status my-team-widget__status--error">{error}</p>
            ) : team.length === 0 ? (
                <p className="my-team-widget__status">No Pokémon in your party yet.</p>
            ) : (
                <ul className="my-team-widget__list">
                    {team.map((pokemon, index) => (
                        <PokemonTile
                            key={pokemon.id}
                            pokemon={pokemon}
                            slot={index + 1}
                            selected={false}
                            onSelect={onSelectPokemon}
                        />
                    ))}
                </ul>
            )}
        </section>
    );
}

function UsersGlyph() {
    return (
        <svg viewBox="0 0 24 24" fill="none">
            <circle cx="9" cy="8" r="3" stroke="currentColor" strokeWidth="1.6" />
            <path
                d="M3.5 19c0-2.8 2.2-5 5.5-5s5.5 2.2 5.5 5"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
            />
            <circle cx="16.5" cy="9" r="2.5" stroke="currentColor" strokeWidth="1.6" />
            <path
                d="M13 19c0-2 1.4-3.7 3.5-3.7"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
            />
        </svg>
    );
}
