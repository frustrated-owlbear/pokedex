import {type TeamPokemon} from './MyTeamWidget';
import {formatDisplayDate, hpPercent, typeColor} from '../utils/pokemon';

interface PokemonDetailPanelProps {
    pokemon: TeamPokemon;
}

export function PokemonDetailPanel({pokemon}: PokemonDetailPanelProps) {
    const fill = hpPercent(pokemon.hp, pokemon.maxHp);

    return (
        <section className="pokemon-detail-panel" aria-label={`${pokemon.name} details`}>
            <div className="pokemon-detail-panel__hero">
                {pokemon.imageUrl ? (
                    <img
                        src={pokemon.imageUrl}
                        alt={pokemon.name}
                        className="pokemon-detail-panel__image"
                    />
                ) : (
                    <span className="pokemon-detail-panel__fallback" aria-hidden>
                        ?
                    </span>
                )}
            </div>

            <div className="pokemon-detail-panel__main">
                <div className="pokemon-detail-panel__heading">
                    <h2 className="pokemon-detail-panel__name">{pokemon.name}</h2>
                    <span className="pokemon-detail-panel__level">Lv. {pokemon.level}</span>
                    <span
                        className="pokemon-detail-panel__type"
                        style={{backgroundColor: typeColor(pokemon.primaryType)}}
                    >
                        {pokemon.primaryType}
                    </span>
                </div>

                <div className="pokemon-detail-panel__hp-row">
                    <div className="pokemon-detail-panel__hp-bar" aria-hidden>
                        <span
                            className="pokemon-detail-panel__hp-fill"
                            style={{width: `${fill}%`}}
                        />
                    </div>
                    <span className="pokemon-detail-panel__hp-value">
                        {pokemon.hp}/{pokemon.maxHp} HP
                    </span>
                </div>
            </div>

            <dl className="pokemon-detail-panel__facts">
                <div className="pokemon-detail-panel__fact">
                    <dt>Caught</dt>
                    <dd>{formatDisplayDate(pokemon.caughtDate)}</dd>
                </div>
                <div className="pokemon-detail-panel__fact">
                    <dt>Birthday</dt>
                    <dd>
                        {pokemon.birthday
                            ? formatDisplayDate(pokemon.birthday)
                            : 'Not recorded'}
                    </dd>
                </div>
                <div className="pokemon-detail-panel__fact">
                    <dt>National dex</dt>
                    <dd>#{pokemon.dexId}</dd>
                </div>
            </dl>
        </section>
    );
}
