import {useEffect, useState} from 'react';
import {GetPokemonDetail} from '../../wailsjs/go/main/App';
import {domain} from '../../wailsjs/go/models';
import {PokemonImage} from './PokemonImage';

interface PokemonDetailPanelProps {
    name: string;
    onClose: () => void;
}

export function PokemonDetailPanel({name, onClose}: PokemonDetailPanelProps) {
    const [detail, setDetail] = useState<domain.PokemonDetail | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(false);

    useEffect(() => {
        setLoading(true);
        setError(false);
        GetPokemonDetail(name)
            .then(setDetail)
            .catch(() => setError(true))
            .finally(() => setLoading(false));
    }, [name]);

    return (
        <div className="pokemon-detail-overlay" onClick={onClose} role="presentation">
            <div
                className="pokemon-detail"
                onClick={(e) => e.stopPropagation()}
                role="dialog"
                aria-labelledby="pokemon-detail-title"
            >
                <button
                    type="button"
                    className="pokemon-detail__close"
                    onClick={onClose}
                    aria-label="Close"
                >
                    ×
                </button>

                {loading ? (
                    <p className="pokemon-detail__loading">Loading…</p>
                ) : error || !detail ? (
                    <p className="pokemon-detail__loading">
                        Could not load Pokémon details.
                    </p>
                ) : (
                    <>
                        <div className="pokemon-detail__header">
                            <PokemonImage
                                src={detail.imageUrl}
                                alt={detail.name}
                                className="pokemon-detail__image"
                            />
                            <div>
                                <h2 id="pokemon-detail-title">{detail.name}</h2>
                                {detail.id > 0 && (
                                    <p className="pokemon-detail__id">#{detail.id}</p>
                                )}
                            </div>
                        </div>

                        <dl className="pokemon-detail__stats">
                            <div>
                                <dt>Types</dt>
                                <dd>{detail.types}</dd>
                            </div>
                            <div>
                                <dt>Region</dt>
                                <dd>{detail.region}</dd>
                            </div>
                            <div>
                                <dt>Level</dt>
                                <dd>{detail.level}</dd>
                            </div>
                            <div>
                                <dt>HP</dt>
                                <dd>
                                    {detail.hp} / {detail.maxHp}
                                </dd>
                            </div>
                            <div>
                                <dt>Ability</dt>
                                <dd>{detail.ability}</dd>
                            </div>
                        </dl>

                        <section className="pokemon-detail__matchups">
                            <h3>Type matchups</h3>
                            <ul>
                                <li>
                                    <strong>Strong vs</strong>{' '}
                                    {joinList(detail.strengths)}
                                </li>
                                <li>
                                    <strong>Weak vs</strong>{' '}
                                    {joinList(detail.weaknesses)}
                                </li>
                                <li>
                                    <strong>Resists</strong>{' '}
                                    {joinList(detail.resistances)}
                                </li>
                            </ul>
                        </section>
                    </>
                )}
            </div>
        </div>
    );
}

function joinList(items: string[]): string {
    return items.length ? items.join(', ') : '—';
}
