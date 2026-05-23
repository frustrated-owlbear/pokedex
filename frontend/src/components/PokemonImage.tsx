import {useState} from 'react';
import {PokeballIcon} from './icons';

interface PokemonImageProps {
    src: string;
    alt: string;
    className?: string;
}

export function PokemonImage({src, alt, className}: PokemonImageProps) {
    const [failed, setFailed] = useState(false);

    if (!src || failed) {
        return (
            <div className={`pokemon-image pokemon-image--fallback ${className ?? ''}`}>
                <PokeballIcon />
            </div>
        );
    }

    return (
        <img
            src={src}
            alt={alt}
            className={`pokemon-image ${className ?? ''}`}
            onError={() => setFailed(true)}
        />
    );
}
