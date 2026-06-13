import {SystemStatusCard, type SystemState} from './SystemStatusCard';

type NavItem = 'feed' | 'team';

type NavEntry = {
    id: NavItem | 'observations' | 'sessions' | 'knowledge' | 'settings';
    label: string;
    disabled?: boolean;
};

type LeftNavProps = {
    active: NavItem;
    systemStates: SystemState[];
    onNavigate: (item: NavItem) => void;
};

const ITEMS: NavEntry[] = [
    {id: 'feed', label: 'Agent Feed'},
    {id: 'observations', label: 'Observations', disabled: true},
    {id: 'team', label: 'Pokémon'},
    {id: 'sessions', label: 'Sessions', disabled: true},
    {id: 'knowledge', label: 'Knowledge Base', disabled: true},
    {id: 'settings', label: 'Settings', disabled: true},
];

export function LeftNav({active, systemStates, onNavigate}: LeftNavProps) {
    return (
        <nav className="dashboard-sidebar" aria-label="Primary">
            <div className="dashboard-sidebar__profile">
                <p className="dashboard-sidebar__eyebrow">TRAINER ID</p>
                <p className="dashboard-sidebar__trainer-id">000123</p>
                <p className="dashboard-sidebar__rank">Novice Trainer</p>
            </div>

            <ul className="dashboard-sidebar__nav" aria-label="Sections">
                {ITEMS.map((item) => {
                    const isActive = item.id === active;
                    const isInteractive = !item.disabled && (item.id === 'feed' || item.id === 'team');

                    return (
                        <li key={item.id}>
                            <button
                                type="button"
                                disabled={!isInteractive}
                                aria-current={isActive ? 'page' : undefined}
                                className={`dashboard-sidebar__link${
                                    isActive ? ' dashboard-sidebar__link--active' : ''
                                }${item.disabled ? ' dashboard-sidebar__link--disabled' : ''}`}
                                onClick={() => {
                                    if (item.id === 'feed' || item.id === 'team') {
                                        onNavigate(item.id);
                                    }
                                }}
                            >
                                {item.label}
                            </button>
                        </li>
                    );
                })}
            </ul>

            <div className="dashboard-sidebar__footer">
                <SystemStatusCard states={systemStates} compact />
            </div>
        </nav>
    );
}
