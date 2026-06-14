import {type TeamPokemon} from './MyTeamWidget';
import {hpPercent, typeColor} from '../utils/pokemon';

interface PokemonTileProps {
    pokemon: TeamPokemon;
    slot: number;
    selected: boolean;
    onSelect: (id: number) => void;
}

export function PokemonTile({pokemon, slot, selected, onSelect}: PokemonTileProps) {
    const fill = hpPercent(pokemon.hp, pokemon.maxHp);

    return (
        <li>
            <button
                type="button"
                className={`pokemon-tile${selected ? ' pokemon-tile--selected' : ''}`}
                onClick={() => onSelect(pokemon.id)}
                aria-pressed={selected}
                aria-label={`${pokemon.name}, level ${pokemon.level}`}
            >
                <span className="pokemon-tile__slot">{slot}</span>
                <div className="pokemon-tile__thumb">
                    {pokemon.imageUrl ? (
                        <img
                            src={pokemon.imageUrl}
                            alt=""
                            className="pokemon-tile__image"
                        />
                    ) : (
                        <span className="pokemon-tile__fallback" aria-hidden>
                            ?
                        </span>
                    )}
                </div>
                <div className="pokemon-tile__meta">
                    <div className="pokemon-tile__name-row">
                        <span className="pokemon-tile__name">{pokemon.name}</span>
                        <span className="pokemon-tile__level">Lv. {pokemon.level}</span>
                    </div>
                    <span
                        className="pokemon-tile__type"
                        style={{backgroundColor: typeColor(pokemon.primaryType)}}
                    >
                        {pokemon.primaryType}
                    </span>
                    <div className="pokemon-tile__hp-row">
                        <div className="pokemon-tile__hp-bar" aria-hidden>
                            <span
                                className="pokemon-tile__hp-fill"
                                style={{width: `${fill}%`}}
                            />
                        </div>
                        <span className="pokemon-tile__hp-value">
                            {pokemon.hp}/{pokemon.maxHp}
                        </span>
                    </div>
                </div>
            </button>
        </li>
    );
}
