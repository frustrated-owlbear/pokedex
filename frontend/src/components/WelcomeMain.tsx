import {useEffect, useRef, useState} from 'react';
import {CameraIcon, PokeballIcon} from './icons';

export function WelcomeMain() {
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [importedImage, setImportedImage] = useState<string | null>(null);

    useEffect(() => {
        return () => {
            if (importedImage) {
                URL.revokeObjectURL(importedImage);
            }
        };
    }, [importedImage]);

    function openImageImport() {
        fileInputRef.current?.click();
    }

    function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0];
        if (!file || !file.type.startsWith('image/')) {
            return;
        }
        setImportedImage((prev) => {
            if (prev) {
                URL.revokeObjectURL(prev);
            }
            return URL.createObjectURL(file);
        });
        e.target.value = '';
    }

    return (
        <main className="welcome-main">
            <button
                type="button"
                className="welcome-main__camera-btn"
                onClick={openImageImport}
                title="Take photo"
                aria-label="Take photo"
            >
                <CameraIcon />
            </button>
            <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="welcome-main__file-input"
                onChange={handleFileChange}
                tabIndex={-1}
                aria-hidden
            />

            {importedImage ? (
                <img
                    src={importedImage}
                    alt="Imported"
                    className="welcome-main__preview"
                />
            ) : (
                <PokeballIcon className="welcome-main__bg-icon" />
            )}

            <h1 className="welcome-main__title">Pokédex is ready!</h1>
            <p className="welcome-main__subtitle">
                {importedImage
                    ? 'Photo imported. Select a function from the menu to get started.'
                    : 'Select a function from the menu to get started.'}
            </p>
        </main>
    );
}
