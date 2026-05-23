import type {MenuId} from '../config/menuItems';

export function CameraIcon({className}: {className?: string}) {
    return (
        <svg className={className} viewBox="0 0 24 24" fill="none" aria-hidden>
            <path
                d="M4 8h3l1.5-2h7L17 8h3a2 2 0 012 2v8a2 2 0 01-2 2H4a2 2 0 01-2-2v-8a2 2 0 012-2z"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinejoin="round"
            />
            <circle cx="12" cy="13" r="3.5" stroke="currentColor" strokeWidth="1.5" />
        </svg>
    );
}

export function PokeballIcon({className}: {className?: string}) {
    return (
        <svg className={className} viewBox="0 0 24 24" fill="none" aria-hidden>
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="1.5" />
            <path d="M2 12h20" stroke="currentColor" strokeWidth="1.5" />
            <circle cx="12" cy="12" r="3.5" fill="currentColor" />
        </svg>
    );
}

export function MenuIcon({id}: {id: MenuId}) {
    switch (id) {
        case 'current-situation':
            return (
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <path
                        d="M5 5h14v14H5z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                    <path
                        d="M9 9h6v6H9z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                </svg>
            );
        case 'my-team':
            return (
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <circle cx="8" cy="10" r="3" stroke="currentColor" strokeWidth="1.5" />
                    <circle cx="16" cy="10" r="3" stroke="currentColor" strokeWidth="1.5" />
                    <circle cx="12" cy="16" r="3" stroke="currentColor" strokeWidth="1.5" />
                </svg>
            );
        case 'pokemon':
            return <PokeballIcon />;
        case 'analysis':
            return (
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <rect x="6" y="14" width="3" height="6" fill="currentColor" />
                    <rect x="10.5" y="8" width="3" height="12" fill="currentColor" />
                    <rect x="15" y="4" width="3" height="16" fill="currentColor" />
                </svg>
            );
        case 'knowledge-base':
            return (
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <path
                        d="M5 6h8a3 3 0 013 3v11H8a3 3 0 01-3-3V6z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                    <path
                        d="M11 6h8a3 3 0 013 3v11h-8a3 3 0 01-3-3V6z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                </svg>
            );
        case 'ask-pokedex':
            return (
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <path
                        d="M6 8a6 6 0 0112 0v5a6 6 0 01-12 0V8z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                    <circle cx="9" cy="11" r="1" fill="currentColor" />
                    <circle cx="12" cy="11" r="1" fill="currentColor" />
                    <circle cx="15" cy="11" r="1" fill="currentColor" />
                </svg>
            );
    }
}
