import {domain} from '../../wailsjs/go/models';
import {PokemonImage} from './PokemonImage';

interface TeamPokemonCardsProps {
    team: domain.Pokemon[];
    onSelect: (name: string) => void;
}

export function TeamPokemonCards({team, onSelect}: TeamPokemonCardsProps) {
    if (team.length === 0) {
        return <p className="team-cards__empty">No Pokémon in your party yet.</p>;
    }

    return (
        <ul className="team-cards">
            {team.map((pokemon, index) => (
                <li key={`${pokemon.name}-${index}`}>
                    <button
                        type="button"
                        className="team-card"
                        onClick={() => onSelect(pokemon.name)}
                    >
                        <span className="team-card__slot">{index + 1}</span>
                        <PokemonImage
                            src={pokemon.imageUrl}
                            alt={pokemon.name}
                            className="team-card__image"
                        />
                        <span className="team-card__name">{pokemon.name}</span>
                    </button>
                </li>
            ))}
        </ul>
    );
}
