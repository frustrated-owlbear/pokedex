import {
    ChangeEvent,
    DragEvent,
    FormEvent,
    KeyboardEvent,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react';
import DOMPurify from 'dompurify';
import {marked} from 'marked';
import './App.css';
import {AgentFeed} from './components/AgentFeed';
import {LeftNav} from './components/LeftNav';
import {SituationPanel} from './components/SituationPanel';
import {MyTeamScreen} from './views/MyTeamScreen';
import type {TeamPokemon} from './components/MyTeamWidget';
import {
    parseTraceStep,
    type SituationView,
    type TraceStep,
} from './types/agent';
import {AskPokedex, GetCurrentSituation, ListMyTeam} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

type SystemState = {
    label: string;
    value: string;
    detail?: string;
    healthy: boolean;
};

const defaultSituation: SituationView = {
    location: 'Viridian Forest',
    region: 'Kanto',
    time: '--:--',
    period: 'Morning',
    weather: 'Clear',
    memory: 'No previous sessions recorded yet.',
    tools: [],
};

function fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => {
            const result = reader.result;
            if (typeof result !== 'string') {
                reject(new Error('Could not read image'));
                return;
            }
            const comma = result.indexOf(',');
            resolve(comma >= 0 ? result.slice(comma + 1) : result);
        };
        reader.onerror = () => reject(new Error('Could not read image'));
        reader.readAsDataURL(file);
    });
}

type AppScreen = 'feed' | 'team';

function App() {
    const [screen, setScreen] = useState<AppScreen>('feed');
    const [selectedPokemonId, setSelectedPokemonId] = useState<number | null>(null);
    const [prompt, setPrompt] = useState('');
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [isDragging, setIsDragging] = useState(false);
    const [response, setResponse] = useState('');
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState('');
    const [traceSteps, setTraceSteps] = useState<TraceStep[]>([]);
    const [analysis, setAnalysis] = useState('');
    const fileInputRef = useRef<HTMLInputElement | null>(null);
    const [systemStates, setSystemStates] = useState<SystemState[]>([]);
    const [situation, setSituation] = useState<SituationView>(defaultSituation);
    const [team, setTeam] = useState<TeamPokemon[]>([]);
    const [teamLoading, setTeamLoading] = useState(true);
    const [teamError, setTeamError] = useState('');

    useEffect(() => {
        let cancelled = false;

        async function loadInitialData() {
            setTeamLoading(true);
            setTeamError('');
            try {
                const [teamResult, situationResult] = await Promise.all([
                    ListMyTeam(),
                    GetCurrentSituation(),
                ]);
                if (!cancelled) {
                    setTeam(teamResult ?? []);
                    if (situationResult) {
                        setSituation(parseSituation(situationResult));
                    }
                }
            } catch {
                if (!cancelled) {
                    setTeam([]);
                    setTeamError('Could not load team.');
                }
            } finally {
                if (!cancelled) {
                    setTeamLoading(false);
                }
            }
        }

        loadInitialData();
        return () => {
            cancelled = true;
        };
    }, []);

    useEffect(() => {
        const unsubscribeChunk = EventsOn('llm:chunk', (...args: unknown[]) => {
            const chunk =
                typeof args[0] === 'string' ? args[0] : String(args[0] ?? '');
            setResponse((prev) => prev + chunk);
        });

        const unsubscribeTrace = EventsOn('agent:trace', (...args: unknown[]) => {
            const step = parseTraceStep(args[0]);
            if (!step) {
                return;
            }
            setTraceSteps((prev) => [...prev, step]);
            if (step.kind === 'Observation' && step.detail) {
                setAnalysis(step.detail);
            }
        });

        const unsubscribeStatus = EventsOn('ollama:status', (...args: unknown[]) => {
            const status = parseSystemState(args[0]);
            setSystemStates((prev) => upsertSystemState(prev, status));
        });

        return () => {
            unsubscribeChunk();
            unsubscribeTrace();
            unsubscribeStatus();
        };
    }, []);

    const previewUrl = useMemo(() => {
        if (!selectedFile) {
            return '';
        }
        return URL.createObjectURL(selectedFile);
    }, [selectedFile]);

    useEffect(() => {
        return () => {
            if (previewUrl) {
                URL.revokeObjectURL(previewUrl);
            }
        };
    }, [previewUrl]);

    const canSubmit = Boolean(prompt.trim() || selectedFile);

    function openFilePicker() {
        fileInputRef.current?.click();
    }

    function updateSelectedFile(file?: File | null) {
        if (!file || !file.type.startsWith('image/')) {
            return;
        }
        setSelectedFile(file);
    }

    function handleFileInput(event: ChangeEvent<HTMLInputElement>) {
        updateSelectedFile(event.target.files?.[0] ?? null);
    }

    function handleDragOver(event: DragEvent<HTMLButtonElement>) {
        event.preventDefault();
        setIsDragging(true);
    }

    function handleDragLeave(event: DragEvent<HTMLButtonElement>) {
        event.preventDefault();
        setIsDragging(false);
    }

    function handleDrop(event: DragEvent<HTMLButtonElement>) {
        event.preventDefault();
        setIsDragging(false);
        updateSelectedFile(event.dataTransfer.files?.[0] ?? null);
    }

    function clearImage() {
        setSelectedFile(null);
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
    }

    async function refreshSituation() {
        try {
            const situationResult = await GetCurrentSituation();
            if (situationResult) {
                setSituation(parseSituation(situationResult));
            }
        } catch {
            // keep previous situation snapshot
        }
    }

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const question = prompt.trim();
        if (!canSubmit || busy) {
            return;
        }

        setBusy(true);
        setError('');
        setResponse('');
        setTraceSteps([]);
        setAnalysis('');
        try {
            const imageBase64 = selectedFile ? await fileToBase64(selectedFile) : '';
            const imageMIME = selectedFile?.type ?? '';
            await AskPokedex(question, imageBase64, imageMIME);
            await refreshSituation();
            setPrompt('');
            clearImage();
        } catch (err) {
            const message =
                typeof err === 'string'
                    ? err
                    : err instanceof Error
                      ? err.message
                      : '';
            setError(message.trim() || 'Could not reach the Pokédex. Try again.');
        } finally {
            setBusy(false);
        }
    }

    function handlePromptKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
        if (event.key === 'Enter' && (event.metaKey || event.ctrlKey)) {
            event.preventDefault();
            const form = event.currentTarget.form;
            if (form) {
                form.requestSubmit();
            }
        }
    }

    const renderedMarkdown = useMemo(() => {
        if (error || (!busy && !response)) {
            return '';
        }
        return DOMPurify.sanitize(marked.parse(response || 'Thinking...') as string);
    }, [busy, error, response]);

    function openTeamScreen(pokemonId: number | null = null) {
        setSelectedPokemonId(pokemonId);
        setScreen('team');
    }

    if (screen === 'team') {
        return (
            <MyTeamScreen
                onBack={() => setScreen('feed')}
                initialSelectedId={selectedPokemonId}
            />
        );
    }

    return (
        <main className="dashboard">
            <div className="dashboard__grid">
                <LeftNav
                    active="feed"
                    systemStates={systemStates}
                    onNavigate={(item) => {
                        if (item === 'team') {
                            openTeamScreen(null);
                        }
                    }}
                />

                <section className="dashboard__main" aria-label="Agent workspace">
                    <div className="dashboard__workspace">
                        <AgentFeed steps={traceSteps} busy={busy} />

                        {(error || busy || response) && (
                            <section
                                className={`agentic-answer${
                                    error ? ' agentic-answer--error' : ''
                                }`}
                                role="status"
                                aria-live="polite"
                            >
                                {error ? (
                                    <>
                                        <p className="agentic-answer__label">ERROR</p>
                                        <p className="agentic-answer__message">{error}</p>
                                    </>
                                ) : (
                                    <>
                                        <p className="agentic-answer__label">FINAL ANSWER</p>
                                        <div
                                            className="agentic-answer__markdown"
                                            dangerouslySetInnerHTML={{__html: renderedMarkdown}}
                                        />
                                        {busy && <span className="composer__caret" aria-hidden />}
                                    </>
                                )}
                            </section>
                        )}
                    </div>

                    <section className="composer-dock dashboard__composer" aria-label="Inference input">
                        <form className="composer" onSubmit={handleSubmit}>
                            <button
                                type="button"
                                className={`composer__dropzone${
                                    isDragging ? ' composer__dropzone--dragging' : ''
                                }`}
                                onClick={openFilePicker}
                                onDragOver={handleDragOver}
                                onDragLeave={handleDragLeave}
                                onDrop={handleDrop}
                                aria-label="Add image"
                            >
                                {previewUrl ? (
                                    <>
                                        <img
                                            src={previewUrl}
                                            alt={selectedFile?.name ?? 'Selected upload'}
                                            className="composer__preview"
                                        />
                                        <span className="composer__preview-name">
                                            {selectedFile?.name}
                                        </span>
                                    </>
                                ) : (
                                    <>
                                        <span className="composer__dropzone-icon" aria-hidden>
                                            <UploadGlyph />
                                        </span>
                                        <span className="composer__dropzone-text">Add image</span>
                                        <span className="composer__dropzone-subtext">
                                            click or drop
                                        </span>
                                    </>
                                )}
                            </button>

                            <div className="composer__panel">
                                <textarea
                                    value={prompt}
                                    onChange={(event) => setPrompt(event.target.value)}
                                    onKeyDown={handlePromptKeyDown}
                                    className="composer__textarea"
                                    placeholder="Describe what you see or ask Pokédex..."
                                    rows={3}
                                />

                                <button
                                    type="submit"
                                    className="composer__send"
                                    disabled={busy || !canSubmit}
                                    aria-label="Send input"
                                >
                                    <SendGlyph />
                                </button>
                            </div>

                            <input
                                ref={fileInputRef}
                                type="file"
                                accept="image/*"
                                className="composer__file-input"
                                onChange={handleFileInput}
                            />
                        </form>
                    </section>
                </section>

                <aside className="dashboard__aside" aria-label="Context and tools">
                    <SituationPanel
                        situation={situation}
                        team={team}
                        teamLoading={teamLoading}
                        teamError={teamError}
                        analysis={analysis}
                        onOpenTeam={() => openTeamScreen(null)}
                        onSelectPokemon={(id) => openTeamScreen(id)}
                    />
                </aside>
            </div>
        </main>
    );
}

function parseSystemState(payload: unknown): SystemState {
    if (!payload || typeof payload !== 'object') {
        return {
            label: 'Ollama',
            value: 'Unavailable',
            detail: 'Invalid healthcheck payload',
            healthy: false,
        };
    }

    const record = payload as Record<string, unknown>;
    return {
        label: typeof record.label === 'string' ? record.label : 'Ollama',
        value: typeof record.value === 'string' ? record.value : 'Unavailable',
        detail: typeof record.detail === 'string' ? record.detail : undefined,
        healthy: record.healthy === true,
    };
}

function upsertSystemState(states: SystemState[], next: SystemState): SystemState[] {
    const index = states.findIndex((state) => state.label === next.label);
    if (index === -1) {
        return [...states, next];
    }
    const copy = [...states];
    copy[index] = next;
    return copy;
}

function parseSituation(payload: unknown): SituationView {
    if (!payload || typeof payload !== 'object') {
        return defaultSituation;
    }
    const record = payload as Record<string, unknown>;
    return {
        location: typeof record.location === 'string' ? record.location : defaultSituation.location,
        region: typeof record.region === 'string' ? record.region : defaultSituation.region,
        time: typeof record.time === 'string' ? record.time : defaultSituation.time,
        period: typeof record.period === 'string' ? record.period : defaultSituation.period,
        weather: typeof record.weather === 'string' ? record.weather : defaultSituation.weather,
        memory: typeof record.memory === 'string' ? record.memory : defaultSituation.memory,
        tools: Array.isArray(record.tools)
            ? record.tools.filter((tool): tool is string => typeof tool === 'string')
            : [],
        analysis: typeof record.analysis === 'string' ? record.analysis : undefined,
    };
}

function UploadGlyph() {
    return (
        <svg viewBox="0 0 24 24" fill="none">
            <rect x="4" y="5" width="16" height="14" rx="2.5" stroke="currentColor" strokeWidth="1.5" />
            <path
                d="M8 15l2.8-3 2.4 2.4 1.8-2.1L18 15"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
            <circle cx="9" cy="9" r="1.1" fill="currentColor" />
        </svg>
    );
}

function SendGlyph() {
    return (
        <svg viewBox="0 0 24 24" fill="none">
            <path
                d="M5 12h11"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
            />
            <path
                d="M12 6l6 6-6 6"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
                strokeLinejoin="round"
            />
        </svg>
    );
}

export default App;
