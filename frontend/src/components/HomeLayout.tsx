import type {MenuId} from '../config/menuItems';
import {AskPokedexFab} from './AskPokedexFab';
import {MenuPanel} from './MenuPanel';
import {StatusBar} from './StatusBar';
import {TrainerSidebar} from './TrainerSidebar';
import {WelcomeMain} from './WelcomeMain';

interface HomeLayoutProps {
    onSelectMenu: (id: MenuId) => void;
}

export function HomeLayout({onSelectMenu}: HomeLayoutProps) {
    return (
        <div className="home-layout">
            <StatusBar />
            <div className="home-layout__body">
                <TrainerSidebar />
                <WelcomeMain />
                <MenuPanel onSelect={onSelectMenu} />
            </div>
            <AskPokedexFab onClick={() => onSelectMenu('ask-pokedex')} />
        </div>
    );
}
