import {menuItems, type MenuId} from '../config/menuItems';
import {MenuIcon} from './icons';

interface MenuPanelProps {
    onSelect: (id: MenuId) => void;
}

export function MenuPanel({onSelect}: MenuPanelProps) {
    return (
        <nav className="menu-panel" aria-label="Main menu">
            <ul className="menu-panel__list">
                {menuItems
                    .filter((item) => item.id !== 'ask-pokedex')
                    .map((item) => (
                    <li key={item.id}>
                        <button
                            type="button"
                            className="menu-item"
                            onClick={() => onSelect(item.id)}
                        >
                            <span className="menu-item__icon">
                                <MenuIcon id={item.icon} />
                            </span>
                            <span className="menu-item__text">
                                <span className="menu-item__title">{item.title}</span>
                                <span className="menu-item__desc">{item.description}</span>
                            </span>
                            <span className="menu-item__chevron" aria-hidden>
                                ›
                            </span>
                        </button>
                    </li>
                ))}
            </ul>
        </nav>
    );
}
