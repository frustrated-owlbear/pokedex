import {useEffect, useState} from 'react';
import {GetTrainerProfile} from '../../wailsjs/go/main/App';
import {domain} from '../../wailsjs/go/models';
import {PokeballIcon} from './icons';

export function TrainerSidebar() {
    const [profile, setProfile] = useState<domain.TrainerProfile | null>(null);

    useEffect(() => {
        GetTrainerProfile()
            .then(setProfile)
            .catch(() => {
                setProfile(
                    domain.TrainerProfile.createFrom({
                        trainerId: '000123',
                        avatarUrl: '',
                        connectionStatus: 'ONLINE',
                    }),
                );
            });
    }, []);

    return (
        <aside className="trainer-sidebar">
            <div className="trainer-sidebar__brand">
                <PokeballIcon className="trainer-sidebar__logo" />
                <span>POKÉDEX</span>
            </div>

            <div className="trainer-card">
                <div className="trainer-card__avatar" aria-hidden>
                    <svg viewBox="0 0 64 64" fill="none">
                        <circle cx="32" cy="24" r="12" fill="#c8c8cc" />
                        <path
                            d="M12 58c4-14 14-20 20-20s16 6 20 20"
                            fill="#c8c8cc"
                        />
                    </svg>
                </div>
                <p className="trainer-card__label">TRAINER ID</p>
                <p className="trainer-card__id">{profile?.trainerId ?? '———'}</p>
            </div>
        </aside>
    );
}
