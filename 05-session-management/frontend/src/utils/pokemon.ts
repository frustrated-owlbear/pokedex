export const typeColors: Record<string, string> = {
    GRASS: '#4a9c5d',
    NORMAL: '#8a8d94',
    FIRE: '#c45c3e',
    WATER: '#4a7fb5',
    ELECTRIC: '#c9a227',
    POISON: '#8b5a9e',
    FLYING: '#7a8db5',
    PSYCHIC: '#b55a7a',
};

export function typeColor(type: string): string {
    return typeColors[type.toUpperCase()] ?? '#8a8d94';
}

export function formatDisplayDate(isoDate: string): string {
    if (!isoDate) {
        return '—';
    }
    const parsed = new Date(`${isoDate}T00:00:00`);
    if (Number.isNaN(parsed.getTime())) {
        return isoDate;
    }
    return parsed.toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    });
}

export function hpPercent(hp: number, maxHp: number): number {
    return maxHp > 0 ? Math.min(100, (hp / maxHp) * 100) : 0;
}
