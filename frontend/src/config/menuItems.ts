export type MenuId =
    | 'current-situation'
    | 'my-team'
    | 'pokemon'
    | 'analysis'
    | 'knowledge-base'
    | 'ask-pokedex';

export interface MenuItem {
    id: MenuId;
    title: string;
    description: string;
    icon: MenuId;
}

export const menuItems: MenuItem[] = [
    {
        id: 'current-situation',
        title: 'Current Situation',
        description: 'Analyze and get advice',
        icon: 'current-situation',
    },
    {
        id: 'my-team',
        title: 'My Team',
        description: 'View your Pokémon',
        icon: 'my-team',
    },
    {
        id: 'pokemon',
        title: 'Pokémon',
        description: 'Browse Pokémon info',
        icon: 'pokemon',
    },
    {
        id: 'analysis',
        title: 'Analysis',
        description: 'Type matchups, strengths, and more',
        icon: 'analysis',
    },
    {
        id: 'knowledge-base',
        title: 'Knowledge Base',
        description: 'Articles and references',
        icon: 'knowledge-base',
    },
    {
        id: 'ask-pokedex',
        title: 'Ask Pokédex',
        description: 'Ask anything',
        icon: 'ask-pokedex',
    },
];
