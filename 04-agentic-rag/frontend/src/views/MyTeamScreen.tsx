import {useEffect, useMemo, useState} from 'react';
import {ListMyTeam} from '../../wailsjs/go/main/App';
import {PokemonDetailPanel} from '../components/PokemonDetailPanel';
import {PokemonTile} from '../components/PokemonTile';
import {type TeamPokemon} from '../components/MyTeamWidget';

interface MyTeamScreenProps {
    onBack: () => void;
    initialSelectedId: number | null;
}

export function MyTeamScreen({onBack, initialSelectedId}: MyTeamScreenProps) {
    const [team, setTeam] = useState<TeamPokemon[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [selectedId, setSelectedId] = useState<number | null>(initialSelectedId);

    useEffect(() => {
        let cancelled = false;

        async function loadTeam() {
            setLoading(true);
            setError('');
            try {
                const result = await ListMyTeam();
                if (!cancelled) {
                    setTeam(result ?? []);
                }
            } catch {
                if (!cancelled) {
                    setTeam([]);
                    setError('Could not load team.');
                }
            } finally {
                if (!cancelled) {
                    setLoading(false);
                }
            }
        }

        loadTeam();
        return () => {
            cancelled = true;
        };
    }, []);

    useEffect(() => {
        if (team.length === 0) {
            setSelectedId(null);
            return;
        }

        if (selectedId !== null && team.some((pokemon) => pokemon.id === selectedId)) {
            return;
        }

        const preferred =
            initialSelectedId !== null &&
            team.some((pokemon) => pokemon.id === initialSelectedId)
                ? initialSelectedId
                : team[0].id;
        setSelectedId(preferred);
    }, [team, selectedId, initialSelectedId]);

    const selectedPokemon = useMemo(
        () => team.find((pokemon) => pokemon.id === selectedId) ?? null,
        [team, selectedId],
    );

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
        <main className="team-screen">
            <div className="team-screen__shell">
                <header className="team-screen__header">
                    <button type="button" className="team-screen__back" onClick={onBack}>
                        ← Back
                    </button>
                    <h1 className="team-screen__title">My Team</h1>
                    <p className="team-screen__subtitle">Select a Pokémon to view details</p>
                </header>

                {loading ? (
                    <p className="team-screen__status">Loading team...</p>
                ) : error ? (
                    <p className="team-screen__status team-screen__status--error">{error}</p>
                ) : team.length === 0 ? (
                    <p className="team-screen__status">No Pokémon in your party yet.</p>
                ) : (
                    <div className="team-screen__layout">
                        <aside className="team-screen__rail" aria-label="Party Pokémon">
                            <p className="team-screen__rail-title">Party</p>
                            <ul className="team-screen__tiles">
                                {team.map((pokemon, index) => (
                                    <PokemonTile
                                        key={pokemon.id}
                                        pokemon={pokemon}
                                        slot={index + 1}
                                        selected={pokemon.id === selectedId}
                                        onSelect={setSelectedId}
                                    />
                                ))}
                            </ul>
                        </aside>

                        <div className="team-screen__detail">
                            {selectedPokemon ? (
                                <PokemonDetailPanel pokemon={selectedPokemon} />
                            ) : (
                                <p className="team-screen__status">Select a Pokémon.</p>
                            )}
                        </div>
                    </div>
                )}

                <footer className="team-screen__footer">
                    <span>Press B to go back</span>
                </footer>
            </div>
        </main>
    );
}
