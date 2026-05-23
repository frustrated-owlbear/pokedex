import {useEffect, useState} from 'react';
import {GetTrainerProfile} from '../../wailsjs/go/main/App';
import {domain} from '../../wailsjs/go/models';

function formatTime(date: Date): string {
    return date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit', hour12: false});
}

export function StatusBar() {
    const now = new Date();
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

    const online = profile?.connectionStatus === 'ONLINE';
    const statusLabel = profile?.connectionStatus ?? 'OFFLINE';

    return (
        <header className="status-bar">
            <span className="status-bar__online">
                <span
                    className={`status-bar__dot${online ? ' status-bar__dot--online' : ''}`}
                    aria-hidden
                />
                {statusLabel}
            </span>
            <time className="status-bar__clock">{formatTime(now)}</time>
            <div className="status-bar__icons" aria-label="System status">
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <path
                        d="M4 16l4-4 4 4 6-8"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                    />
                    <path d="M2 20h20" stroke="currentColor" strokeWidth="1.5" />
                </svg>
                <svg viewBox="0 0 24 24" fill="none" aria-hidden>
                    <rect
                        x="3"
                        y="7"
                        width="16"
                        height="10"
                        rx="2"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                    <path d="M21 10v4" stroke="currentColor" strokeWidth="1.5" />
                    <rect x="5" y="9" width="10" height="6" rx="1" fill="currentColor" />
                </svg>
            </div>
        </header>
    );
}
