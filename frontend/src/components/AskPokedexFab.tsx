import {MenuIcon} from './icons';

interface AskPokedexFabProps {
    onClick: () => void;
}

export function AskPokedexFab({onClick}: AskPokedexFabProps) {
    return (
        <button
            type="button"
            className="ask-pokedex-fab"
            onClick={onClick}
            aria-label="Ask Pokédex"
        >
            <span className="ask-pokedex-fab__icon">
                <MenuIcon id="ask-pokedex" />
            </span>
            <span className="ask-pokedex-fab__label">Ask Pokédex</span>
        </button>
    );
}
