import {useState} from 'react';
import './App.css';
import './HomeScreen.css';
import {HomeLayout} from './components/HomeLayout';
import type {MenuId} from './config/menuItems';
import {FeatureView} from './views/FeatureView';

function App() {
    const [screen, setScreen] = useState<MenuId | null>(null);

    if (screen) {
        return <FeatureView screen={screen} onBack={() => setScreen(null)} />;
    }

    return <HomeLayout onSelectMenu={setScreen} />;
}

export default App;
